// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2020 Simon Ser <https://emersion.fr>

package scfg

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// structInfo contains scfg metadata for structs.
type structInfo struct {
	param    int            // index of field storing parameters
	children map[string]int // indices of fields storing child directives
}

var (
	structCacheMutex sync.Mutex
	structCache      = make(map[reflect.Type]*structInfo)
)

func getStructInfo(t reflect.Type) (*structInfo, error) {
	structCacheMutex.Lock()
	defer structCacheMutex.Unlock()

	if info := structCache[t]; info != nil {
		return info, nil
	}

	info := &structInfo{
		param:    -1,
		children: make(map[string]int),
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous {
			return nil, fmt.Errorf("scfg: anonymous struct fields are not supported")
		} else if !f.IsExported() {
			continue
		}

		tag := f.Tag.Get("scfg")
		parts := strings.Split(tag, ",")
		k, options := parts[0], parts[1:]
		if k == "-" {
			continue
		} else if k == "" {
			k = f.Name
		}

		isParam := false
		for _, opt := range options {
			switch opt {
			case "param":
				isParam = true
			default:
				return nil, fmt.Errorf("scfg: invalid option %q in struct tag", opt)
			}
		}

		if isParam {
			if info.param >= 0 {
				return nil, fmt.Errorf("scfg: param option specified multiple times in struct tag in %v", t)
			}
			if parts[0] != "" {
				return nil, fmt.Errorf("scfg: name must be empty when param option is specified in struct tag in %v", t)
			}
			info.param = i
		} else {
			if _, ok := info.children[k]; ok {
				return nil, fmt.Errorf("scfg: key %q specified multiple times in struct tag in %v", k, t)
			}
			info.children[k] = i
		}
	}

	structCache[t] = info
	return info, nil
}
