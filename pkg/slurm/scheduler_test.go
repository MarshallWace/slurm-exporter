// Copyright 2017 Victor Penso, Matteo Dessalvi
// SPDX-FileCopyrightText: 2022 2022 Marshall Wace <opensource@mwam.com>
//
// SPDX-License-Identifier: GPL3

package slurm

import (
	"testing"
)

func TestSchedulerGetMetrics(t *testing.T) {
	coll := NewSchedulerCollector(true)
	t.Logf("%+v", coll.SchedulerGetMetrics())
}
