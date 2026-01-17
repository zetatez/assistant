package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"assistant/pkg/apps"
	"assistant/pkg/llm"

	_ "assistant/pkg/llm/providers/deepseek"
)

func main() {
	// 1. 从环境变量读取 API Key
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set the DEEPSEEK_API_KEY environment variable")
	}

	// 2. 创建 LLM 配置
	cfg := llm.Config{
		APIKey:  apiKey,
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-chat",
	}

	// 3. 初始化 LLM Client
	client, err := llm.NewClient("deepseek", cfg)
	if err != nil {
		log.Fatalf("failed to create llm client: %v", err)
	}

	advisor := apps.NewSQLIndexAdvisor(client)

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	input := apps.SQLIndexInput{
		SQL: `
SELECT
    u.id,
    u.name,
    SUM(oi.price * oi.quantity) AS total_amount
FROM orders o
JOIN order_items oi ON oi.order_id = o.id
JOIN users u ON u.id = o.user_id
JOIN products p ON p.id = oi.product_id
WHERE o.created_at >= '2025-01-01'
  AND o.status = 'paid'
  AND p.category_id IN (3, 5, 7)
  AND u.is_vip = 1
GROUP BY u.id, u.name
ORDER BY total_amount DESC
LIMIT 50;
`,
		TableDDLs: []string{
			`CREATE TABLE orders (
		id BIGINT PRIMARY KEY,
		user_id BIGINT NOT NULL,
		status VARCHAR(20),
		created_at DATETIME,
		total_amount DECIMAL(12,2),
		KEY idx_created_at (created_at)
	);`,

			`CREATE TABLE order_items (
		id BIGINT PRIMARY KEY,
		order_id BIGINT NOT NULL,
		product_id BIGINT NOT NULL,
		price DECIMAL(10,2),
		quantity INT
	);`,

			`CREATE TABLE users (
		id BIGINT PRIMARY KEY,
		name VARCHAR(64),
		is_vip TINYINT
	);`,

			`CREATE TABLE products (
		id BIGINT PRIMARY KEY,
		category_id INT,
		price DECIMAL(10,2)
	);`,
		},
		StatsText: `
orders:
  row_count ≈ 15 million
  NDV(user_id) ≈ 1.2 million
  NDV(status) = 4
  created_at 最近 30 天约占 8%

order_items:
  row_count ≈ 120 million
  NDV(order_id) ≈ 15 million

users:
  row_count ≈ 1.3 million
  vip 用户约占 6%

products:
  row_count ≈ 500k
  category_id in (3,5,7) 覆盖约 15%
`,
	}

	result, err := advisor.OptimizeIndexes(ctx, input)
	if err != nil {
		log.Fatal(err)
	}

	printJSONAndMarkdown(result)
}

func printJSONAndMarkdown(v interface{}) {
	// JSON 输出
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(v)

	// Markdown 输出
	os.Stdout.WriteString("\n## 索引优化建议\n\n")

	if r, ok := v.(*apps.SQLIndexResult); ok {
		if len(r.Plans) == 0 {
			os.Stdout.WriteString("无需添加新索引，现有索引已足够。\n\n")
			return
		}

		for i, plan := range r.Plans {
			os.Stdout.WriteString(fmt.Sprintf("### 表 %d: %s\n\n", i+1, plan.TableName))

			if len(plan.Actions) == 0 {
				os.Stdout.WriteString("该表无需添加新索引。\n\n")
				continue
			}

			for j, act := range plan.Actions {
				os.Stdout.WriteString(fmt.Sprintf("%d. **索引方案**\n\n", j+1))
				os.Stdout.WriteString("```sql\n")
				os.Stdout.WriteString(fmt.Sprintf("%s\n", act.DDL))
				os.Stdout.WriteString("```\n\n")

				os.Stdout.WriteString(fmt.Sprintf("- **原因**: %s\n", act.Reason))
				os.Stdout.WriteString(fmt.Sprintf("- **风险**: %s\n\n", act.Risk))
			}

			os.Stdout.WriteString("---\n\n")
		}
	}
}
