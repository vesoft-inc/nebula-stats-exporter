package exporter

import (
	"strings"
)

type BaseMetric struct {
	Name   string
	Value  string
	Labels map[string]string
}

func convertToMetrics(originMetrics []string) []BaseMetric {
	metrics := make([]BaseMetric, len(originMetrics))
	/*
		match metric format
		slow_query_latency_us.p95.5=0
		slow_query_latency_us{space=nba}.p95.5=0
	*/
	for _, origin := range originMetrics {
		metric, label := splitMetric(origin)

		s := strings.Split(metric, "=")
		if len(s) != 2 {
			continue
		}
		metrics = append(metrics, BaseMetric{
			Name:   s[0],
			Value:  s[1],
			Labels: label,
		})
	}

	return metrics
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
