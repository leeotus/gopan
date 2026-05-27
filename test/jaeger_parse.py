#!/usr/bin/env python3
import json, sys

output_md = sys.argv[1] if len(sys.argv) > 1 else "/dev/stdout"
data = json.load(sys.stdin)
traces = data.get('data', [])

with open(output_md, 'w') as f:
    f.write("## 请求延迟 Top 列表\n\n")
    f.write("| # | Trace ID | 总耗时 | 最慢操作 | 操作耗时 | 错误 |\n")
    f.write("|---|----------|--------|---------|---------|------|\n")

    durations = []

    for i, trace in enumerate(traces):
        total_ms = trace.get('duration', 0) / 1000
        spans = trace.get('spans', [])
        slowest_op = '-'
        slowest_ms = 0
        has_error = '否'

        for s in spans:
            dur_ms = s.get('duration', 0) / 1000
            if dur_ms > slowest_ms:
                slowest_ms = dur_ms
                slowest_op = s.get('operationName', '-')
            for t in s.get('tags', []):
                if t.get('key') == 'error' and t.get('value') == True:
                    has_error = '是'

        f.write(f'| {i+1} | {trace["traceID"][:12]} | {total_ms:.1f}ms | {slowest_op} | {slowest_ms:.1f}ms | {has_error} |\n')

    # 统计
    for t in traces:
        dur = t.get('duration', 0) / 1000
        if dur > 0:
            durations.append(dur)

    durations.sort()
    n = len(durations)
    if n > 0:
        p50 = durations[n * 50 // 100]
        p95 = durations[n * 95 // 100]
        p99 = durations[n * 99 // 100]
        avg = sum(durations) / n
        f.write("\n## 统计\n\n")
        f.write("| 指标 | 值 |\n")
        f.write("|------|----|\n")
        f.write(f"| 采样数量 | {n} |\n")
        f.write(f"| 平均延迟 | {avg:.1f}ms |\n")
        f.write(f"| P50 | {p50:.1f}ms |\n")
        f.write(f"| P95 | {p95:.1f}ms |\n")
        f.write(f"| P99 | {p99:.1f}ms |\n")
