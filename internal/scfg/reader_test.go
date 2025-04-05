// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2020 Simon Ser <https://emersion.fr>

package scfg

import (
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

var readTests = []struct {
	name string
	src  string
	want Block
}{
	{
		name: "flat",
		src: `dir1 param1 param2 "" param3
dir2
dir3 param1

# comment
dir4 "param 1" 'param 2'`,
		want: Block{
			{Name: "dir1", Params: []string{"param1", "param2", "", "param3"}},
			{Name: "dir2", Params: []string{}},
			{Name: "dir3", Params: []string{"param1"}},
			{Name: "dir4", Params: []string{"param 1", "param 2"}},
		},
	},
	{
		name: "simpleBlocks",
		src: `block1 {
	dir1 param1 param2
	dir2 param1
}

block2 {
}

block3 {
	# comment
}

block4 param1 "param2" {
	dir1
}`,
		want: Block{
			{
				Name:   "block1",
				Params: []string{},
				Children: Block{
					{Name: "dir1", Params: []string{"param1", "param2"}},
					{Name: "dir2", Params: []string{"param1"}},
				},
			},
			{Name: "block2", Params: []string{}, Children: Block{}},
			{Name: "block3", Params: []string{}, Children: Block{}},
			{
				Name:   "block4",
				Params: []string{"param1", "param2"},
				Children: Block{
					{Name: "dir1", Params: []string{}},
				},
			},
		},
	},
	{
		name: "nested",
		src: `block1 {
	block2 {
		dir1 param1
	}

	block3 {
	}
}

block4 {
	block5 {
		block6 param1 {
			dir1
		}
	}

	dir1
}`,
		want: Block{
			{
				Name:   "block1",
				Params: []string{},
				Children: Block{
					{
						Name:   "block2",
						Params: []string{},
						Children: Block{
							{Name: "dir1", Params: []string{"param1"}},
						},
					},
					{
						Name:     "block3",
						Params:   []string{},
						Children: Block{},
					},
				},
			},
			{
				Name:   "block4",
				Params: []string{},
				Children: Block{
					{
						Name:   "block5",
						Params: []string{},
						Children: Block{{
							Name:   "block6",
							Params: []string{"param1"},
							Children: Block{{
								Name:   "dir1",
								Params: []string{},
							}},
						}},
					},
					{
						Name:   "dir1",
						Params: []string{},
					},
				},
			},
		},
	},
	{
		name: "quotes",
		src:  `"a \b ' \" c" 'd \e \' " f' a\"b`,
		want: Block{
			{Name: "a b ' \" c", Params: []string{"d e ' \" f", "a\"b"}},
		},
	},
	{
		name: "quotes-2",
		src:  `dir arg1 "arg2" ` + `\"\"`,
		want: Block{
			{Name: "dir", Params: []string{"arg1", "arg2", "\"\""}},
		},
	},
	{
		name: "quotes-3",
		src:  `dir arg1 "\"\"\"\"" arg3`,
		want: Block{
			{Name: "dir", Params: []string{"arg1", `"` + "\"\"" + `"`, "arg3"}},
		},
	},
}

func TestRead(t *testing.T) {
	for _, test := range readTests {
		t.Run(test.name, func(t *testing.T) {
			r := strings.NewReader(test.src)
			got, err := Read(r)
			if err != nil {
				t.Fatalf("Read() = %v", err)
			}
			stripLineno(got)
			if !reflect.DeepEqual(got, test.want) {
				t.Error(spew.Sprintf("Read() = \n %v \n but want \n %v", got, test.want))
			}
		})
	}
}

func stripLineno(block Block) {
	for _, dir := range block {
		dir.lineno = 0
		stripLineno(dir.Children)
	}
}
