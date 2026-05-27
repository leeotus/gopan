/*
GoPan HTTP 压测脚本

使用方式：
  go run test/bench/http_bench.go -c 10 -n 100

参数：
  -c  并发数（默认 10）
  -n  总请求数（默认 100）
  -u  API 地址（默认 http://localhost:8888/api/video/list?cursor=0&limit=10）

输出：
  - QPS
  - P50 / P95 / P99 延迟
  - 错误率
  - 同时打印 Prometheus 指标查询链接
*/
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	concurrency = flag.Int("c", 10, "并发数")
	totalReqs   = flag.Int("n", 100, "总请求数")
	url         = flag.String("u", "http://localhost:8888/api/video/list?cursor=0&limit=10", "API 地址")
	token       = flag.String("t", "", "JWT token（需要登录的接口必填）")
)

func main() {
	flag.Parse()

	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║        GoPan HTTP 压测工具              ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Printf("\n目标: %s\n并发: %d | 总请求: %d\n\n", *url, *concurrency, *totalReqs)

	var wg sync.WaitGroup
	var success, fail int64
	var latencies []time.Duration
	var mu sync.Mutex
	sem := make(chan struct{}, *concurrency)

	start := time.Now()

	for i := 0; i < *totalReqs; i++ {
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			reqStart := time.Now()
			req, _ := http.NewRequest("GET", *url, nil)
			if *token != "" {
				req.Header.Set("Authorization", "Bearer "+*token)
			}

			resp, err := http.DefaultClient.Do(req)
			reqEnd := time.Since(reqStart)

			mu.Lock()
			latencies = append(latencies, reqEnd)
			mu.Unlock()

			if err != nil {
				atomic.AddInt64(&fail, 1)
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				atomic.AddInt64(&success, 1)
			} else {
				atomic.AddInt64(&fail, 1)
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	p50 := latencies[len(latencies)*50/100]
	p95 := latencies[len(latencies)*95/100]
	p99 := latencies[len(latencies)*99/100]
	avgMs := float64(elapsed.Milliseconds()) / float64(*totalReqs)

	// ── 打印报告 ──
	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("总耗时:     %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("成功:       %d\n", success)
	fmt.Printf("失败:       %d\n", fail)
	fmt.Printf("QPS:        %.1f req/s\n", float64(*totalReqs)/elapsed.Seconds())
	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("Avg 延迟:   %.2f ms\n", avgMs)
	fmt.Printf("P50 延迟:   %v\n", p50.Round(time.Millisecond))
	fmt.Printf("P95 延迟:   %v\n", p95.Round(time.Millisecond))
	fmt.Printf("P99 延迟:   %v\n", p99.Round(time.Millisecond))
	fmt.Println(strings.Repeat("─", 50))

	// Prometheus 查询链接
	if strings.Contains(*url, ":8888") {
		host := strings.Replace(*url, "http://", "", 1)
		host = strings.Split(host, ":")[0] + ":9090"
		fmt.Printf("\nPrometheus 指标:\n")
		fmt.Printf("  平均延迟: http://%s/graph?g0.expr=rate(http_server_requests_duration_ms_sum[1m])/rate(http_server_requests_duration_ms_count[1m])\n", host)
		fmt.Printf("  错误率:   http://%s/graph?g0.expr=rate(http_server_requests_duration_ms_count{status=\"500\"}[1m])\n", host)
	}
}
