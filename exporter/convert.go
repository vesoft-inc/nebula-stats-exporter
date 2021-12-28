package exporter

import (
	"strings"
)

type MetricValue struct {
	Value  string
	Labels map[string]string
}

func convertToMap(metrics []string) map[string]MetricValue {
	matches := make(map[string]MetricValue, len(metrics))
	/*
		match metric format
		slow_query_latency_us.p95.5=0
		slow_query_latency_us{space=nba}.p95.5=0
	*/
	for _, metric := range metrics {
		metric, label := splitMetric(metric)

		s := strings.Split(metric, "=")
		if len(s) != 2 {
			continue
		}
		matches[s[0]] = MetricValue{Value: s[1], Labels: label}
	}

	return matches
}

// split slow_query_latency_us{space=nba}.p95.5=0 => slow_query_latency_us.p95.5=0, map[space:nba]
func splitMetric(metric string) (string, map[string]string) {
	start := strings.Index(metric, "{")
	end := strings.LastIndex(metric, "}")

	if start == -1 || end == -1 {
		return metric, nil
	}

	label := make(map[string]string)
	labelsStr := strings.Split(metric[start+1:end], ",")
	for _, labelStr := range labelsStr {
		s := strings.Split(labelStr, "=")
		if len(s) != 2 {
			continue
		}
		label[s[0]] = s[1]
	}

	return metric[:start] + metric[end+1:], label
}
