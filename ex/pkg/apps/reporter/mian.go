package main

import (
	"context"
	"log"
	"os"

	"assistant/pkg/apps"
	"assistant/pkg/llm"

	_ "assistant/pkg/llm/providers/deepseek"
)

func main() {
	ctx := context.Background()

	// 1️⃣ 初始化 LLM Client（按你的 llm 包实现）
	client, err := llm.NewClient("deepseek", llm.Config{
		APIKey: os.Getenv("DEEPSEEK_API_KEY"),
		Model:  "deepseek-chat",
	})
	if err != nil {
		log.Fatalf("init llm client failed: %v", err)
	}

	// 2️⃣ 创建 Reporter
	reporter := apps.NewReporter(client)

	// 3️⃣ 构造复杂、真实的工作内容（唯一事实来源）
	workContent := `
周一：
 早上主要在看之前订单系统的一些老代码，历史包袱比较重，很多逻辑都耦合在一起，
 下午开始拆订单和支付的部分，把下单流程先单独抽出来，
 中间改了一版接口定义，前端那边同步了一下。

周二：
 继续做订单模块重构，发现库存扣减的逻辑之前是写死在订单里的，
 临时调整了一下结构，把库存相关逻辑独立成一个模块，
 顺便补了几个单元测试，但覆盖率还不高。
 晚上回头排查了一个线上偶发的下单失败问题，看日志发现和超时配置有关。

周三：
 上午主要在和前端联调支付流程，来回改了好几次参数和状态码，
 有几个异常场景之前没考虑到，比如支付成功但回调超时，
 下午参加了架构评审会，主要讨论新订单系统后续的扩展方案，
 会后根据建议调整了一下接口的幂等处理逻辑。

周四：
 处理线上问题比较多，有一个用户反馈重复扣款，
 紧急查了下发现是并发场景下支付回调没做好幂等，
 临时加了防重逻辑并发了热修复。
 顺便把之前遗留的几个小 bug 一起修了，但没来得及写太详细的说明。

周五：
 主要做了一些收尾工作，把这周改的代码整理合并，
 简单跑了一下压测，对比了一下改造前后的性能，
 接口平均响应时间比之前快了一些。
 下午把相关改动和问题简单记录了一下，准备下周再补文档和监控。

其他：
 这周整体感觉事情比较零碎，很多时间花在沟通和排查问题上，
 老系统文档缺失的问题还是比较明显，
 下周计划把权限模块的设计方案先整理出来。
`

	// 4️⃣ 调用 Reporter 生成报告
	result, err := reporter.Generate(ctx, apps.ReportInput{
		ReportType:  apps.WeeklyReport,
		Author:      "张三",
		Role:        "后端工程师",
		Period:      "2026-01-01~2026-01-07",
		Language:    "简体中文",
		WorkContent: workContent,
	})
	if err != nil {
		log.Fatalf("generate report failed: %v", err)
	}

	// 5️⃣ 使用模型返回的 file_name 写入本地文件
	if err := os.WriteFile(result.FileName, []byte(result.Markdown), 0644); err != nil {
		log.Fatalf("write report file failed: %v", err)
	}

	// 6️⃣ 打印关键信息
	log.Println("report generated successfully")
	log.Println("file name :", result.FileName)
	log.Printf("confidence : %.2f\n", result.Confidence)
}
