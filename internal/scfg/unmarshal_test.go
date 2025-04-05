// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2020 Simon Ser <https://emersion.fr>

package scfg_test

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"

	"go.lindenii.runxiyu.org/forge/internal/scfg"
)

func ExampleDecoder() {
	var data struct {
		Foo int `scfg:"foo"`
		Bar struct {
			Param string `scfg:",param"`
			Baz   string `scfg:"baz"`
		} `scfg:"bar"`
	}

	raw := `foo 42
bar asdf {
	baz hello
}
`

	r := strings.NewReader(raw)
	if err := scfg.NewDecoder(r).Decode(&data); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Foo = %v\n", data.Foo)
	fmt.Printf("Bar.Param = %v\n", data.Bar.Param)
	fmt.Printf("Bar.Baz = %v\n", data.Bar.Baz)

	// Output:
	// Foo = 42
	// Bar.Param = asdf
	// Bar.Baz = hello
}

type nestedStructInner struct {
	Bar string `scfg:"bar"`
}

type structParams struct {
	Params []string `scfg:",param"`
	Bar    string
}

type textUnmarshaler struct {
	text string
}

func (tu *textUnmarshaler) UnmarshalText(text []byte) error {
	tu.text = string(text)
	return nil
}

type textUnmarshalerParams struct {
	Params []textUnmarshaler `scfg:",param"`
}

var barStr = "bar"

var unmarshalTests = []struct {
	name string
	raw  string
	want interface{}
}{
	{
		name: "stringMap",
		raw: `hello world
foo bar`,
		want: map[string]string{
			"hello": "world",
			"foo":   "bar",
		},
	},
	{
		name: "simpleStruct",
		raw: `MyString asdf
MyBool true
MyInt -42
MyUint 42
MyFloat 3.14`,
		want: struct {
			MyString string
			MyBool   bool
			MyInt    int
			MyUint   uint
			MyFloat  float32
		}{
			MyString: "asdf",
			MyBool:   true,
			MyInt:    -42,
			MyUint:   42,
			MyFloat:  3.14,
		},
	},
	{
		name: "simpleStructTag",
		raw:  `foo bar`,
		want: struct {
			Foo string `scfg:"foo"`
		}{
			Foo: "bar",
		},
	},
	{
		name: "sliceParams",
		raw:  `Foo a s d f`,
		want: struct {
			Foo []string
		}{
			Foo: []string{"a", "s", "d", "f"},
		},
	},
	{
		name: "arrayParams",
		raw:  `Foo a s d f`,
		want: struct {
			Foo [4]string
		}{
			Foo: [4]string{"a", "s", "d", "f"},
		},
	},
	{
		name: "pointers",
		raw:  `Foo bar`,
		want: struct {
			Foo *string
		}{
			Foo: &barStr,
		},
	},
	{
		name: "nestedMap",
		raw: `foo {
	bar baz
}`,
		want: struct {
			Foo map[string]string `scfg:"foo"`
		}{
			Foo: map[string]string{"bar": "baz"},
		},
	},
	{
		name: "nestedStruct",
		raw: `foo {
	bar baz
}`,
		want: struct {
			Foo nestedStructInner `scfg:"foo"`
		}{
			Foo: nestedStructInner{
				Bar: "baz",
			},
		},
	},
	{
		name: "structParams",
		raw: `Foo param1 param2 {
	Bar baz
}`,
		want: struct {
			Foo structParams
		}{
			Foo: structParams{
				Params: []string{"param1", "param2"},
				Bar:    "baz",
			},
		},
	},
	{
		name: "textUnmarshaler",
		raw: `Foo param1
Bar param2
Baz param3`,
		want: struct {
			Foo []textUnmarshaler
			Bar *textUnmarshaler
			Baz textUnmarshalerParams
		}{
			Foo: []textUnmarshaler{{"param1"}},
			Bar: &textUnmarshaler{"param2"},
			Baz: textUnmarshalerParams{
				Params: []textUnmarshaler{{"param3"}},
			},
		},
	},
	{
		name: "directiveStructSlice",
		raw: `Foo param1 param2 {
	Bar baz
}
Foo param3 param4`,
		want: struct {
			Foo []structParams
		}{
			Foo: []structParams{
				{
					Params: []string{"param1", "param2"},
					Bar:    "baz",
				},
				{
					Params: []string{"param3", "param4"},
				},
			},
		},
	},
	{
		name: "directiveMapSlice",
		raw: `Foo {
	key1 param1
}
Foo {
	key2 param2
}`,
		want: struct {
			Foo []map[string]string
		}{
			Foo: []map[string]string{
				{"key1": "param1"},
				{"key2": "param2"},
			},
		},
	},
}

func TestUnmarshal(t *testing.T) {
	for _, tc := range unmarshalTests {
		tc := tc // capture variable
		t.Run(tc.name, func(t *testing.T) {
			testUnmarshal(t, tc.raw, tc.want)
		})
	}
}

func testUnmarshal(t *testing.T, raw string, want interface{}) {
	out := reflect.New(reflect.TypeOf(want))
	r := strings.NewReader(raw)
	if err := scfg.NewDecoder(r).Decode(out.Interface()); err != nil {
		t.Fatalf("Decode() = %v", err)
	}
	got := out.Elem().Interface()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Decode() = \n%#v\n but want \n%#v", got, want)
	}
}
