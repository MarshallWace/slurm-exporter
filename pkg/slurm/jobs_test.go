package slurm

import (
	"os"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestJobsMetrics(t *testing.T) {
	collector := NewJobsCollector(true)
	f, err := os.Open(showJobsTestDataProm)
	assert.NoError(t, err)
	defer f.Close()
	err = testutil.CollectAndCompare(collector, f)
	assert.NoError(t, err)
}

// TestGenerateJobsMetrics is only used to getnerate the .prom file for
// func TestGenerateJobsMetrics(t *testing.T) {
// 	reg := prometheus.NewRegistry()
// 	collector := NewJobsCollector(true)
// 	err := reg.Register(collector)
// 	assert.NoError(t, err)
// 	gatherers := prometheus.Gatherers{reg}
// 	err = prometheus.WriteToTextfile(showJobsTestDataProm, gatherers)
// 	assert.NoError(t, err)
// }
