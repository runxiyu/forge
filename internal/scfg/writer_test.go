// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2020 Simon Ser <https://emersion.fr>

package scfg

import (
	"bytes"
	"testing"
)

func TestWrite(t *testing.T) {
	for _, tc := range []struct {
		src  Block
		want string
		err  error
	}{
		{
			src:  Block{},
			want: "",
		},
		{
			src: Block{{
				Name: "dir",
				Children: Block{{
					Name:   "blk1",
					Params: []string{"p1", `"p2"`},
					Children: Block{
						{
							Name:   "sub1",
							Params: []string{"arg11", "arg12"},
						},
						{
							Name:   "sub2",
							Params: []string{"arg21", "arg22"},
						},
						{
							Name:   "sub3",
							Params: []string{"arg31", "arg32"},
							Children: Block{
								{
									Name: "sub-sub1",
								},
								{
									Name:   "sub-sub2",
									Params: []string{"arg321", "arg322"},
								},
							},
						},
					},
				}},
			}},
			want: `dir {
	blk1 p1 "\"p2\"" {
		sub1 arg11 arg12
		sub2 arg21 arg22
		sub3 arg31 arg32 {
			sub-sub1
			sub-sub2 arg321 arg322
		}
	}
}
`,
		},
		{
			src:  Block{{Name: "dir1"}},
			want: "dir1\n",
		},
		{
			src:  Block{{Name: "dir\"1"}},
			want: "\"dir\\\"1\"\n",
		},
		{
			src:  Block{{Name: "dir'1"}},
			want: "\"dir'1\"\n",
		},
		{
			src:  Block{{Name: "dir:}"}},
			want: "\"dir:}\"\n",
		},
		{
			src:  Block{{Name: "dir:{"}},
			want: "\"dir:{\"\n",
		},
		{
			src:  Block{{Name: "dir\t1"}},
			want: `"dir` + "\t" + `1"` + "\n",
		},
		{
			src:  Block{{Name: "dir 1"}},
			want: "\"dir 1\"\n",
		},
		{
			src:  Block{{Name: "dir1", Params: []string{"arg1", "arg2", `"arg3"`}}},
			want: "dir1 arg1 arg2 " + `"\"arg3\""` + "\n",
		},
		{
			src:  Block{{Name: "dir1", Params: []string{"arg1", "arg 2", "arg'3"}}},
			want: "dir1 arg1 \"arg 2\" \"arg'3\"\n",
		},
		{
			src:  Block{{Name: "dir1", Params: []string{"arg1", "", "arg3"}}},
			want: "dir1 arg1 \"\" arg3\n",
		},
		{
			src:  Block{{Name: "dir1", Params: []string{"arg1", `"` + "\"\"" + `"`, "arg3"}}},
			want: "dir1 arg1 " + `"\"\"\"\""` + " arg3\n",
		},
		{
			src: Block{{
				Name: "dir1",
				Children: Block{
					{Name: "sub1"},
					{Name: "sub2", Params: []string{"arg1", "arg2"}},
				},
			}},
			want: `dir1 {
	sub1
	sub2 arg1 arg2
}
`,
		},
		{
			src: Block{{Name: ""}},
			err: errDirEmptyName,
		},
		{
			src: Block{{
				Name: "dir",
				Children: Block{
					{Name: "sub1"},
					{Name: "", Children: Block{{Name: "sub21"}}},
				},
			}},
			err: errDirEmptyName,
		},
	} {
		t.Run("", func(t *testing.T) {
			var buf bytes.Buffer
			err := Write(&buf, tc.src)
			switch {
			case err != nil && tc.err != nil:
				if got, want := err.Error(), tc.err.Error(); got != want {
					t.Fatalf("invalid error:\ngot= %q\nwant=%q", got, want)
				}
				return
			case err == nil && tc.err != nil:
				t.Fatalf("got err=nil, want=%q", tc.err.Error())
			case err != nil && tc.err == nil:
				t.Fatalf("could not marshal: %+v", err)
			case err == nil && tc.err == nil:
				// ok.
			}
			if got, want := buf.String(), tc.want; got != want {
				t.Fatalf(
					"invalid marshal representation:\ngot:\n%s\nwant:\n%s\n---",
					got, want,
				)
			}
		})
	}
}
