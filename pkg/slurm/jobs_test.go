/* Copyright 2022 Marshall Wace Asset Management

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>. */

package slurm

// import (
// 	"os"
// 	"testing"

// 	"github.com/prometheus/client_golang/prometheus/testutil"
// 	"github.com/stretchr/testify/assert"
// )

// func TestJobsMetrics(t *testing.T) {
// 	collector := NewJobsCollector(true, nil)
// 	f, err := os.Open(showJobsTestDataProm)
// 	assert.NoError(t, err)
// 	defer f.Close()
// 	err = testutil.CollectAndCompare(collector, f)
// 	assert.NoError(t, err)
// }

// TestGenerateJobsMetrics is only used to getnerate the .prom file for
// func TestGenerateJobsMetrics(t *testing.T) {
// 	reg := prometheus.NewRegistry()
// 	collector := NewJobsCollector(true, nil)
// 	err := reg.Register(collector)
// 	assert.NoError(t, err)
// 	gatherers := prometheus.Gatherers{reg}
// 	err = prometheus.WriteToTextfile(showJobsTestDataProm, gatherers)
// 	assert.NoError(t, err)
// }
