package exporter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	pcg "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func getNebulaMetrics(ipAddress string, port int32) (map[string]*pcg.MetricFamily, error) {
	httpClient := http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := httpClient.Get(fmt.Sprintf("http://%s:%d/metrics?format=prometheus", ipAddress, port))
	if err != nil {
		return nil, err
	}

	body, readErr := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		return nil, fmt.Errorf("read or close metrics response: %w", err)
	}

	parser := expfmt.TextParser{}
	return parser.TextToMetricFamilies(bytes.NewReader(body))
}
