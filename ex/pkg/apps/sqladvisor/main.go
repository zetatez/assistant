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
		BaseURL: "https://api.deepseek.com", // 可选
		Model:   "deepseek-chat",            // 可选
	}

	// 3. 初始化 LLM Client
	client, err := llm.NewClient("deepseek", cfg)
	if err != nil {
		log.Fatalf("failed to create llm client: %v", err)
	}

	// 4. 创建 SQL advisor
	advisor := apps.NewSQLAdvisor(client)

	// 5. 设置上下文（生产建议带 timeout）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 6. 待优化 SQL
	sql := `
SELECT *
FROM orders o
LEFT JOIN users u ON o.user_id = u.id
WHERE YEAR(o.created_at) = 2023
ORDER BY o.created_at DESC
`

	// 7. 执行优化
	result, err := advisor.Optimize(ctx, sql)
	if err != nil {
		log.Fatalf("sql optimize failed: %v", err)
	}

	// 8. 输出结果
	printJSONAndMarkdown(result)
}

func printJSONAndMarkdown(v interface{}) {
	// JSON 输出
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(v)

	// Markdown 输出
	os.Stdout.WriteString("\n## SQL 优化结果\n\n")

	if r, ok := v.(*apps.SQLOptimizeResult); ok {
		os.Stdout.WriteString(fmt.Sprintf("- **数据库类型**: %s\n", r.DatabaseType))
		os.Stdout.WriteString(fmt.Sprintf("- **风险级别**: %s\n", r.RiskLevel))
		os.Stdout.WriteString(fmt.Sprintf("- **置信度**: %.2f\n\n", r.Confidence))

		os.Stdout.WriteString("### 原始 SQL\n\n")
		os.Stdout.WriteString(fmt.Sprintf("```sql\n%s\n```\n\n", r.OriginalSQL))

		os.Stdout.WriteString("### 优化后的 SQL\n\n")
		os.Stdout.WriteString(fmt.Sprintf("```sql\n%s\n```\n\n", r.OptimizedSQL))

		if len(r.Optimizations) > 0 {
			os.Stdout.WriteString("### 优化点\n\n")
			for i, opt := range r.Optimizations {
				os.Stdout.WriteString(fmt.Sprintf("%d. %s\n", i+1, opt))
			}
		}
	}

	os.Stdout.WriteString("\n---\n\n")
}
