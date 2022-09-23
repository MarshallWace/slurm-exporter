/* Copyright 2017 Victor Penso, Matteo Dessalvi

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

// func TestNodesNodeMetrics(t *testing.T) {
// 	collector := NewNodesCollector(true, ".example.com")
// 	f, err := os.Open(showNodesDetailsTestDataProm)
// 	assert.NoError(t, err)
// 	defer f.Close()
// 	err = testutil.CollectAndCompare(collector, f)
// 	assert.NoError(t, err)
// }

// TestGenerateNodesNodeMetrics is only used to getnerate the .prom file for
// func TestGenerateNodesNodeMetrics(t *testing.T) {
// 	reg := prometheus.NewRegistry()
// 	collector := NewNodesCollector(true, ".example.com")
// 	err := reg.Register(collector)
// 	assert.NoError(t, err)
// 	err = reg.Register(ExporterErrors)
// 	assert.NoError(t, err)
// 	_, err = reg.Gather()
// 	assert.NoError(t, err)
// 	gatherers := prometheus.Gatherers{reg}
// 	err = prometheus.WriteToTextfile(showNodesDetailsTestDataProm, gatherers)
// 	assert.NoError(t, err)
// }
