// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"github.com/dgraph-io/ristretto/v2"
	"go.lindenii.runxiyu.org/lindenii-common/clog"
)

var indexCommitsDisplayCache *ristretto.Cache[[]byte, []commitDisplay]

func init() {
	var err error
	indexCommitsDisplayCache, err = ristretto.NewCache(&ristretto.Config[[]byte, []commitDisplay]{
		NumCounters: 1e4,
		MaxCost:     1 << 60,
		BufferItems: 8192,
	})
	if err != nil {
		clog.Fatal(1, "Error initializing indexCommitsCache: "+err.Error())
	}
}
