package smartapi

import (
	"context"
	"fmt"

	"assistant/pkg/llm"
)

type EmailComposer struct {
	engine *Engine
}

func NewEmailComposer(client llm.Client) *EmailComposer {
	return &EmailComposer{engine: NewEngine(client)}
}

type EmailType string

const (
	EmailTypeInquiry      EmailType = "inquiry"
	EmailTypeResponse     EmailType = "response"
	EmailTypeNotification EmailType = "notification"
	EmailTypeApology      EmailType = "apology"
	EmailTypeRequest      EmailType = "request"
	EmailTypeThankYou     EmailType = "thank_you"
	EmailTypeReminder     EmailType = "reminder"
	EmailTypeCustom       EmailType = "custom"
)

type EmailTone string

const (
	ToneFormal     EmailTone = "formal"
	ToneSemiFormal EmailTone = "semi_formal"
	ToneCasual     EmailTone = "casual"
)

type EmailInput struct {
	EmailType   EmailType `json:"email_type"`
	Tone        EmailTone `json:"tone,omitempty"`
	Language    string    `json:"language,omitempty"`
	Recipient   string    `json:"recipient"`
	Sender      string    `json:"sender"`
	Subject     string    `json:"subject,omitempty"`
	Content     string    `json:"content"`
	CC          []string  `json:"cc,omitempty"`
	BCC         []string  `json:"bcc,omitempty"`
	Attachments []string  `json:"attachments,omitempty"`
	Context     string    `json:"context,omitempty"`
}

type EmailResult struct {
	Subject    string   `json:"subject"`
	Body       string   `json:"body"`
	CC         []string `json:"cc,omitempty"`
	BCC        []string `json:"bcc,omitempty"`
	Language   string   `json:"language"`
	Confidence float64  `json:"confidence"`
}

const emailComposePrompt = `
	你是一个专业的商务邮件撰写引擎，不是聊天助手。

	你的唯一职责是：
	根据用户提供的邮件类型和内容，撰写一封专业、规范的商务邮件。

	【邮件类型说明】
	- inquiry：询盘/咨询邮件
	- response：回复邮件
	- notification：通知邮件
	- apology：道歉邮件
	- request：请求邮件
	- thank_you：感谢邮件
	- reminder：提醒邮件
	- custom：自定义邮件

	【邮件语气说明】
	- formal：正式语气，适用于正式商务场景
	- semi_formal：半正式语气，适用于一般商务场景
	- casual：轻松语气，适用于熟悉的工作伙伴

	【邮件结构规范】
	1. 称呼：Dear [姓名] / Hi [姓名]
	2. 开篇：简要说明写信目的
	3. 正文：详细说明事项，条理清晰
	4. 结尾：期待回复或说明后续行动
	5. 署名：Best regards / Best / Thanks

	【输出规范】
	- 仅输出 JSON 对象
	- 不允许输出 JSON 以外的任何字符
	- JSON 必须合法且可直接解析

	【JSON 字段】
	{
	  "subject": "邮件主题",
	  "body": "完整的邮件正文（Markdown 格式）",
	  "cc": ["抄送人1", "抄送人2"],
	  "bcc": ["密送人"],
	  "language": "zh-CN（默认中文，报告内容可根据情况掺杂英文等其他语言）",
	  "confidence": 0.0 到 1.0 之间的小数
	}

	【写作要求】
	- 根据邮件类型选择合适的结构和语气
	- 语言专业、礼貌、得体
	- 内容简洁有条理，避免冗长
	- 如有附件，在正文末尾注明
	- 不添加无关内容
	- 不省略任何字段（cc/bcc 为空时使用空数组）
`

func (e *EmailComposer) Compose(ctx context.Context, input EmailInput) (*EmailResult, error) {
	tone := input.Tone
	if tone == "" {
		tone = ToneSemiFormal
	}

	lang := input.Language
	if lang == "" {
		lang = "zh-CN"
	}

	prompt := buildEmailComposePrompt(input)
	systemPrompt := fmt.Sprintf(emailComposePrompt+`

	【本次任务】
	- 邮件类型：%s
	- 语气：%s
	`, input.EmailType, tone)

	return CompleteJSON[EmailResult](
		ctx,
		e.engine,
		prompt,
		systemPrompt,
		0.4,
		2048,
	)
}

func buildEmailComposePrompt(input EmailInput) string {
	prompt := "邮件基本信息：\n" +
		"- 收件人：" + input.Recipient + "\n" +
		"- 发件人：" + input.Sender + "\n" +
		"- 语言：" + langOrDefault(input.Language, "zh-CN") + "\n"

	if input.Subject != "" {
		prompt += "- 邮件主题：" + input.Subject + "\n"
	}

	if len(input.CC) > 0 {
		prompt += "- 抄送："
		for _, cc := range input.CC {
			prompt += cc + ", "
		}
		prompt += "\n"
	}

	if len(input.Context) > 0 {
		prompt += "\n背景/上下文：\n" + input.Context + "\n"
	}

	prompt += "\n邮件内容/要点：\n" + input.Content + "\n"

	if len(input.Attachments) > 0 {
		prompt += "\n附件：\n"
		for _, att := range input.Attachments {
			prompt += "- " + att + "\n"
		}
	}

	return prompt
}

func langOrDefault(lang, defaultLang string) string {
	if lang == "" {
		return defaultLang
	}
	return lang
}
