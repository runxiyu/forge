// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"github.com/dgraph-io/ristretto/v2"
	"go.lindenii.runxiyu.org/lindenii-common/clog"
)

// The key is the commit ID raw hash, followed by the file path.
var commitPathFileRawCache *ristretto.Cache[[]byte, string]

func init() {
	var err error
	commitPathFileRawCache, err = ristretto.NewCache(&ristretto.Config[[]byte, string]{
		NumCounters: 1e4,
		MaxCost:     1 << 60,
		BufferItems: 8192,
	})
	if err != nil {
		clog.Fatal(1, "Error initializing commitPathFileRawCache: "+err.Error())
	}
}
