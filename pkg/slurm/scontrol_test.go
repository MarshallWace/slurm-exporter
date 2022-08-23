package slurm

import (
	"os"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestScontrolNodeMetrics(t *testing.T) {
	collector := NewScontrolCollector(true)
	f, err := os.Open(showNodesDetailsTestDataProm)
	assert.NoError(t, err)
	defer f.Close()
	err = testutil.CollectAndCompare(collector, f)
	assert.NoError(t, err)
}

// TestGenerateScontrolNodeMetrics is only used to getnerate the .prom file for
func TestGenerateScontrolNodeMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	collector := NewScontrolCollector(true)
	err := reg.Register(collector)
	assert.NoError(t, err)
	err = reg.Register(ExporterErrors)
	assert.NoError(t, err)
	_, err = reg.Gather()
	assert.NoError(t, err)
	gatherers := prometheus.Gatherers{reg}
	err = prometheus.WriteToTextfile(showNodesDetailsTestDataProm, gatherers)
	assert.NoError(t, err)
}
