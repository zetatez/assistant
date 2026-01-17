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
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set the DEEPSEEK_API_KEY environment variable")
	}

	cfg := llm.Config{
		APIKey:  apiKey,
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-chat",
	}

	client, err := llm.NewClient("deepseek", cfg)
	if err != nil {
		log.Fatalf("failed to create llm client: %v", err)
	}

	engine := apps.NewDiagnoser(client)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	examples := map[string]string{
		"应用问题 - 内存泄漏": `Java应用运行一段时间后变得很慢，应用日志：
[ERROR] 2024-01-15 10:23:45 [OutOfMemoryError] Java heap space
[WARN]  2024-01-15 10:23:46 [GC] GC overhead limit exceeded
[INFO]  2024-01-15 10:23:47 [Thread] "Thread-1234" daemon prio=10 tid=0x00007f1234567890 nid=0x1234 runnable [0x00007f1234567000]
[ERROR] 2024-01-15 10:23:48 [JVM] Exception in thread "main" java.lang.OutOfMemoryError: Java heap space
[ERROR] 2024-01-15 10:23:49 [JVM]  at com.example.Processor.process(Processor.java:456)
[ERROR] 2024-01-15 10:23:49 [JVM]  at com.example.Worker.run(Worker.java:789)`,
	}

	for name, input := range examples {
		result, err := engine.Diagnose(ctx, input)
		if err != nil {
			log.Printf("Failed to diagnose %s: %v", name, err)
			continue
		}

		printJSONAndMarkdown(result)
	}
}

func printJSONAndMarkdown(v interface{}) {
	// JSON 输出
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(v)

	// Markdown 输出
	os.Stdout.WriteString("\n## 诊断结果\n\n")

	if r, ok := v.(*apps.DiagnoseResult); ok {
		os.Stdout.WriteString(fmt.Sprintf("- **问题域**: %s\n", r.ProblemDomain))
		os.Stdout.WriteString(fmt.Sprintf("- **问题类型**: %s\n", r.ProblemType))
		os.Stdout.WriteString(fmt.Sprintf("- **严重级别**: %s\n", r.Severity))
		os.Stdout.WriteString(fmt.Sprintf("- **影响范围**: %s\n", r.ImpactScope))
		os.Stdout.WriteString(fmt.Sprintf("- **简要描述**: %s\n", r.Summary))
		os.Stdout.WriteString(fmt.Sprintf("- **置信度**: %.2f\n\n", r.Confidence))

		if len(r.Issues) > 0 {
			os.Stdout.WriteString("### 识别的问题\n\n")
			for i, issue := range r.Issues {
				os.Stdout.WriteString(fmt.Sprintf("%d. **%s** [%s]\n", i+1, issue.Message, issue.Severity))
				if issue.Type != "" {
					os.Stdout.WriteString(fmt.Sprintf("   - 类型: %s\n", issue.Type))
				}
				if issue.Location != "" {
					os.Stdout.WriteString(fmt.Sprintf("   - 位置: %s\n", issue.Location))
				}
				if issue.ErrorCode != "" {
					os.Stdout.WriteString(fmt.Sprintf("   - 错误码: %s\n", issue.ErrorCode))
				}
				if issue.Timestamp != "" {
					os.Stdout.WriteString(fmt.Sprintf("   - 时间: %s\n", issue.Timestamp))
				}
			}
			os.Stdout.WriteString("\n")
		}

		os.Stdout.WriteString("### 根因分析\n\n")
		os.Stdout.WriteString(fmt.Sprintf("- **主要原因**: %s\n", r.RootCause.Primary))
		os.Stdout.WriteString(fmt.Sprintf("- **根因分类**: %s\n", r.RootCause.Category))
		os.Stdout.WriteString(fmt.Sprintf("- **置信度**: %s\n", r.RootCause.Confidence))
		if len(r.RootCause.ContributingFactors) > 0 {
			os.Stdout.WriteString("- **影响因素**:\n")
			for _, factor := range r.RootCause.ContributingFactors {
				os.Stdout.WriteString(fmt.Sprintf("  - %s\n", factor))
			}
		}
		os.Stdout.WriteString("\n")

		if len(r.DiagnosisSteps) > 0 {
			os.Stdout.WriteString("### 诊断步骤\n\n")
			for i, step := range r.DiagnosisSteps {
				os.Stdout.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
			}
			os.Stdout.WriteString("\n")
		}

		if len(r.Solutions) > 0 {
			os.Stdout.WriteString("### 解决方案\n\n")
			for i, sol := range r.Solutions {
				os.Stdout.WriteString(fmt.Sprintf("%d. **%s** [%s]\n", i+1, sol.Description, sol.Priority))
				if sol.Category != "" {
					os.Stdout.WriteString(fmt.Sprintf("   - 方案类型: %s\n", sol.Category))
				}
				if sol.EstimatedEffort != "" {
					os.Stdout.WriteString(fmt.Sprintf("   - 工作量: %s\n", sol.EstimatedEffort))
				}
				if sol.Actionable {
					os.Stdout.WriteString("   - 可执行: 是\n")
				}
				if len(sol.SideEffects) > 0 {
					os.Stdout.WriteString("   - 副作用:\n")
					for _, effect := range sol.SideEffects {
						os.Stdout.WriteString(fmt.Sprintf("     - %s\n", effect))
					}
				}
			}
			os.Stdout.WriteString("\n")
		}

		if len(r.AffectedComponents) > 0 {
			os.Stdout.WriteString("### 受影响的组件\n\n")
			for _, comp := range r.AffectedComponents {
				os.Stdout.WriteString(fmt.Sprintf("- %s\n", comp))
			}
			os.Stdout.WriteString("\n")
		}

		if len(r.Dependencies) > 0 {
			os.Stdout.WriteString("### 相关依赖\n\n")
			for _, dep := range r.Dependencies {
				os.Stdout.WriteString(fmt.Sprintf("- %s\n", dep))
			}
			os.Stdout.WriteString("\n")
		}

		if len(r.PreventionMeasures) > 0 {
			os.Stdout.WriteString("### 预防措施\n\n")
			for _, measure := range r.PreventionMeasures {
				os.Stdout.WriteString(fmt.Sprintf("- %s\n", measure))
			}
		}
	}

	os.Stdout.WriteString("\n---\n\n")
}
