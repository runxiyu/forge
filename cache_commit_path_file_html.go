// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"html/template"

	"github.com/dgraph-io/ristretto/v2"
	"go.lindenii.runxiyu.org/lindenii-common/clog"
)

var commitPathFileHTMLCache *ristretto.Cache[[]byte, template.HTML]

func init() {
	var err error
	commitPathFileHTMLCache, err = ristretto.NewCache(&ristretto.Config[[]byte, template.HTML]{
		NumCounters: 1e4,
		MaxCost:     1 << 60,
		BufferItems: 8192,
	})
	if err != nil {
		clog.Fatal(1, "Error initializing commitPathFileHTMLCache: "+err.Error())
	}
}
