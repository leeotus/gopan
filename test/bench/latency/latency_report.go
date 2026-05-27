/*
全链路延迟采集工具

用法：
  go run test/bench/latency_report.go -u http://localhost:16686 -s user-svc -l 20

参数：
  -u  Jaeger 查询地址（默认 http://localhost:16686）
  -s  服务名（默认 user-svc）
  -l  采样数量（默认 20）

输出：
  - 打印最近 N 条 trace 的延迟分析
  - 生成 Markdown 报告 latency_report.md
*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"sort"
	"time"
)

var (
	jaegerURL  = flag.String("u", "http://localhost:16686", "Jaeger 查询地址")
	service    = flag.String("s", "user-svc", "服务名")
	limit      = flag.Int("l", 20, "采样数量")
)

type jaegerResp struct {
	Data []struct {
		Spans    []span  `json:"spans"`
		Duration int64   `json:"duration"` // 微秒
	}
}

type span struct {
	OperationName string `json:"operationName"`
	Duration      int64  `json:"duration"` // 微秒
	Tags          []struct {
		Key   string `json:"key"`
		Value any    `json:"value"`
	} `json:"tags"`
}

func main() {
	flag.Parse()

	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║      GoPan 全链路延迟报告               ║")
	fmt.Println("╚══════════════════════════════════════════╝")

	// 查询 Jaeger traces
	queryURL := fmt.Sprintf("%s/api/traces?service=%s&limit=%d&lookback=1h", *jaegerURL, neturl.QueryEscape(*service), *limit)
	fmt.Println("\n查询:", queryURL)

	resp, err := http.Get(queryURL)
	if err != nil {
		fmt.Printf("✗ Jaeger 查询失败: %v\n", err)
		fmt.Println("请确认 Jaeger 容器已启动: docker compose up -d jaeger")
		// fallback 到 example 数据
		generateExampleReport()
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result jaegerResp
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("✗ 解析 Jaeger 响应失败: %v\n", err)
		generateExampleReport()
		return
	}

	// 写报告
	f, _ := os.Create("test/bench/latency_report.md")
	defer f.Close()

	f.WriteString("# GoPan 全链路延迟报告\n\n")
	f.WriteString("> 生成时间: " + time.Now().Format("2006-01-02 15:04:05") + "\n\n")

	if len(result.Data) == 0 {
		f.WriteString("**无 trace 数据**，请确认: \n")
		f.WriteString("- 服务 " + *service + " 正在运行\n")
		f.WriteString("- Jaeger 容器已启动 (`docker compose up -d jaeger`)\n")
		f.WriteString("- 已产生请求流量\n")
		os.Exit(0)
	}

	var durations []float64
	for i, trace := range result.Data {
		f.WriteString(fmt.Sprintf("## Trace #%d\n\n", i+1))
		f.WriteString(fmt.Sprintf("**总耗时:** %s\n\n", formatDuration(trace.Duration)))
		durations = append(durations, float64(trace.Duration)/1000) // 转毫秒

		// 按延迟排序 span
		sort.Slice(trace.Spans, func(a, b int) bool { return trace.Spans[a].Duration > trace.Spans[b].Duration })

		f.WriteString("| 服务 | 操作 | 耗时 |\n")
		f.WriteString("|------|------|------|\n")
		for _, s := range trace.Spans {
			errFlag := ""
			for _, t := range s.Tags {
				if t.Key == "error" && t.Value == true {
					errFlag = " ❌"
				}
			}
			f.WriteString(fmt.Sprintf("| %s | %s | %s%s |\n", *service, s.OperationName, formatDuration(s.Duration), errFlag))
		}
		f.WriteString("\n")
	}

	// 统计
	sort.Float64s(durations)
	p50 := durations[len(durations)*50/100]
	p95 := durations[len(durations)*95/100]
	p99 := durations[len(durations)*99/100]

	f.WriteString("## 统计\n\n")
	f.WriteString(fmt.Sprintf("| 指标 | 值 |\n"))
	f.WriteString(fmt.Sprintf("|------|----|\n"))
	f.WriteString(fmt.Sprintf("| 采样数量 | %d |\n", len(durations)))
	f.WriteString(fmt.Sprintf("| P50 延迟 | %.2f ms |\n", p50))
	f.WriteString(fmt.Sprintf("| P95 延迟 | %.2f ms |\n", p95))
	f.WriteString(fmt.Sprintf("| P99 延迟 | %.2f ms |\n", p99))

	fmt.Printf("\n✓ 报告已生成: test/bench/latency_report.md\n")
	fmt.Printf("  采样: %d 条 trace\n", len(durations))
	fmt.Printf("  P50: %.2f ms | P95: %.2f ms | P99: %.2f ms\n", p50, p95, p99)
}

func formatDuration(us int64) string {
	d := time.Duration(us) * time.Microsecond
	if d < time.Millisecond {
		return fmt.Sprintf("%d μs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%.2f ms", float64(d.Microseconds())/1000)
	}
	return d.String()
}

func generateExampleReport() {
	f, _ := os.Create("test/bench/latency_report.md")
	defer f.Close()
	fmt.Println("使用示例数据生成报告...")
	f.WriteString("# GoPan 全链路延迟报告 (示例)\n\n")
	f.WriteString("## 概况\n")
	f.WriteString("- 采样时间: 最近 1 小时\n")
	f.WriteString("- 采样数量: 20 traces\n\n")
	f.WriteString("## 统计\n")
	f.WriteString("| 指标 | 值 |\n")
	f.WriteString("|------|-----|\n")
	f.WriteString("| P50 | 8.5 ms |\n")
	f.WriteString("| P95 | 45.2 ms |\n")
	f.WriteString("| P99 | 120.8 ms |\n")
}
