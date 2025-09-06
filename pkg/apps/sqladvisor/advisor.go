package sqladvisor

import (
	"context"
	"encoding/json"
	"fmt"

	"assistant/pkg/llm"
)

type SQLAdvisor struct {
	client llm.Client
}

func NewSQLAdvisor(client llm.Client) *SQLAdvisor {
	return &SQLAdvisor{client: client}
}

func (o *SQLAdvisor) Optimize(ctx context.Context, sql string) (*Result, error) {
	prompt := fmt.Sprintf(PromptTpl, sql)
	req := llm.ChatRequest{
		Model: o.client.Model(),
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: prompt},
		},
		Temperature: 0.1, // SQL 优化必须极低随机性
		MaxTokens:   2000,
	}

	resp, err := o.client.Chat(ctx, req)
	if err != nil {
		return nil, err
	}

	content, err := llm.ExtractAndValidateJSONObject(resp.Content)
	if err != nil {
		return nil, err
	}

	var result Result
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf(
			"invalid JSON response: %w\nraw: %s",
			err,
			resp.Content,
		)
	}

	return &result, nil
}

const PromptTpl = `
你是一个严格的 SQL 优化引擎。

你不是聊天助手，不是教学工具，不是解释器。

你的唯一职责是：在不改变语义的前提下，对 SQL 进行性能与结构优化。

任务要求：
1. 自动识别 SQL 所属数据库类型（oceanbase_mysql / mysql / postgres / sqlite / unknown）
2. 分析 SQL 的性能问题，包括但不限于：
   - 不必要的 SELECT *
   - 无效或冗余的子查询
   - 可以提前过滤的条件
   - JOIN 顺序或方式问题
   - 可简化的表达式
3. 输出一个 **语义等价但性能更优** 的 SQL
4. 如果 SQL 已经是最优，optimized_sql 可以与 original_sql 相同

输出规范（必须严格遵守）：
- 仅输出一个 JSON 对象
- 不允许输出除 JSON 以外的任何字符
- 不允许使用 Markdown
- JSON 必须是合法且可直接解析的

JSON 字段定义：
{
  "database_type": "oceanbase_mysql | mysql | postgres | sqlite | unknown",
  "original_sql": "原始 SQL",
  "optimized_sql": "优化后的 SQL",
  "optimizations": [
    "具体优化点 1",
    "具体优化点 2"
  ],
  "risk_level": "low | medium | high",
  "confidence": 0.0 到 1.0 之间的小数
}

风险评估说明：
- low：仅结构或性能优化，不影响结果集
- medium：重写查询结构，但逻辑等价
- high：可能依赖隐式行为（如 NULL、去重、排序）

重要约束：
- 不要改变 SQL 的业务语义
- 不要添加不存在的字段或表
- 不要假设索引一定存在（可建议但不强制）
- 不要输出解释性文本
- 不要省略任何字段

待优化 SQL：
%s
`
