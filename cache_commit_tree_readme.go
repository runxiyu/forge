// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"html/template"

	"github.com/dgraph-io/ristretto/v2"
	"go.lindenii.runxiyu.org/lindenii-common/clog"
)

type treeReadmeCacheEntry struct {
	DisplayTree    []displayTreeEntry
	ReadmeFilename string
	ReadmeRendered template.HTML
}

var treeReadmeCache *ristretto.Cache[[]byte, treeReadmeCacheEntry]

func init() {
	var err error
	treeReadmeCache, err = ristretto.NewCache(&ristretto.Config[[]byte, treeReadmeCacheEntry]{
		NumCounters: 1e4,
		MaxCost:     1 << 60,
		BufferItems: 8192,
	})
	if err != nil {
		clog.Fatal(1, "Error initializing indexPageCache: "+err.Error())
	}
}
