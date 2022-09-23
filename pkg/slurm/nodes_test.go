// Copyright 2022 Marshall Wace Asset Management
// SPDX-FileCopyrightText: 2022 2022 Marshall Wace <opensource@mwam.com>
//
// SPDX-License-Identifier: GPL3

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
