// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2020 Simon Ser <https://emersion.fr>
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package scfg

import (
	"encoding"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

// Decoder reads and decodes an scfg document from an input stream.
type Decoder struct {
	r                 io.Reader
	unknownDirectives []*Directive
}

// NewDecoder returns a new decoder which reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// UnknownDirectives returns a slice of all unknown directives encountered
// during Decode.
func (dec *Decoder) UnknownDirectives() []*Directive {
	return dec.unknownDirectives
}

// Decode reads scfg document from the input and stores it in the value pointed
// to by v.
//
// If v is nil or not a pointer, Decode returns an error.
//
// Blocks can be unmarshaled to:
//
//   - Maps. Each directive is unmarshaled into a map entry. The map key must
//     be a string.
//   - Structs. Each directive is unmarshaled into a struct field.
//
// Duplicate directives are not allowed, unless the struct field or map value
// is a slice of values representing a directive: structs or maps.
//
// Directives can be unmarshaled to:
//
//   - Maps. The children block is unmarshaled into the map. Parameters are not
//     allowed.
//   - Structs. The children block is unmarshaled into the struct. Parameters
//     are allowed if one of the struct fields contains the "param" option in
//     its tag.
//   - Slices. Parameters are unmarshaled into the slice. Children blocks are
//     not allowed.
//   - Arrays. Parameters are unmarshaled into the array. The number of
//     parameters must match exactly the length of the array. Children blocks
//     are not allowed.
//   - Strings, booleans, integers, floating-point values, values implementing
//     encoding.TextUnmarshaler. Only a single parameter is allowed and is
//     unmarshaled into the value. Children blocks are not allowed.
//
// The decoding of each struct field can be customized by the format string
// stored under the "scfg" key in the struct field's tag. The tag contains the
// name of the field possibly followed by a comma-separated list of options.
// The name may be empty in order to specify options without overriding the
// default field name. As a special case, if the field name is "-", the field
// is ignored. The "param" option specifies that directive parameters are
// stored in this field (the name must be empty).
func (dec *Decoder) Decode(v interface{}) error {
	block, err := Read(dec.r)
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("scfg: invalid value for unmarshaling")
	}

	return dec.unmarshalBlock(block, rv)
}

func (dec *Decoder) unmarshalBlock(block Block, v reflect.Value) error {
	v = unwrapPointers(v)
	t := v.Type()

	dirsByName := make(map[string][]*Directive, len(block))
	for _, dir := range block {
		dirsByName[dir.Name] = append(dirsByName[dir.Name], dir)
	}

	switch v.Kind() {
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			return fmt.Errorf("scfg: map key type must be string")
		}
		if v.IsNil() {
			v.Set(reflect.MakeMap(t))
		} else if v.Len() > 0 {
			clearMap(v)
		}

		for name, dirs := range dirsByName {
			mv := reflect.New(t.Elem()).Elem()
			if err := dec.unmarshalDirectiveList(dirs, mv); err != nil {
				return err
			}
			v.SetMapIndex(reflect.ValueOf(name), mv)
		}
	case reflect.Struct:
		si, err := getStructInfo(t)
		if err != nil {
			return err
		}

		for name, dirs := range dirsByName {
			fieldIndex, ok := si.children[name]
			if !ok {
				dec.unknownDirectives = append(dec.unknownDirectives, dirs...)
				continue
			}
			fv := v.Field(fieldIndex)
			if err := dec.unmarshalDirectiveList(dirs, fv); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("scfg: unsupported type for unmarshaling blocks: %v", t)
	}

	return nil
}

func (dec *Decoder) unmarshalDirectiveList(dirs []*Directive, v reflect.Value) error {
	v = unwrapPointers(v)
	t := v.Type()

	if v.Kind() != reflect.Slice || !isDirectiveType(t.Elem()) {
		if len(dirs) > 1 {
			return newUnmarshalDirectiveError(dirs[1], "directive must not be specified more than once")
		}
		return dec.unmarshalDirective(dirs[0], v)
	}

	sv := reflect.MakeSlice(t, len(dirs), len(dirs))
	for i, dir := range dirs {
		if err := dec.unmarshalDirective(dir, sv.Index(i)); err != nil {
			return err
		}
	}
	v.Set(sv)
	return nil
}

// isDirectiveType checks whether a type can only be unmarshaled as a
// directive, not as a parameter. Accepting too many types here would result in
// ambiguities, see:
// https://lists.sr.ht/~emersion/public-inbox/%3C20230629132458.152205-1-contact%40emersion.fr%3E#%3Ch4Y2peS_YBqY3ar4XlmPDPiNBFpYGns3EBYUx3_6zWEhV2o8_-fBQveRujGADWYhVVCucHBEryFGoPtpC3d3mQ-x10pWnFogfprbQTSvtxc=@emersion.fr%3E
func isDirectiveType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	textUnmarshalerType := reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	if reflect.PtrTo(t).Implements(textUnmarshalerType) {
		return false
	}

	switch t.Kind() {
	case reflect.Struct, reflect.Map:
		return true
	default:
		return false
	}
}

func (dec *Decoder) unmarshalDirective(dir *Directive, v reflect.Value) error {
	v = unwrapPointers(v)
	t := v.Type()

	if v.CanAddr() {
		if _, ok := v.Addr().Interface().(encoding.TextUnmarshaler); ok {
			if len(dir.Children) != 0 {
				return newUnmarshalDirectiveError(dir, "directive requires zero children")
			}
			return unmarshalParamList(dir, v)
		}
	}

	switch v.Kind() {
	case reflect.Map:
		if len(dir.Params) > 0 {
			return newUnmarshalDirectiveError(dir, "directive requires zero parameters")
		}
		if err := dec.unmarshalBlock(dir.Children, v); err != nil {
			return err
		}
	case reflect.Struct:
		si, err := getStructInfo(t)
		if err != nil {
			return err
		}

		if si.param >= 0 {
			if err := unmarshalParamList(dir, v.Field(si.param)); err != nil {
				return err
			}
		} else {
			if len(dir.Params) > 0 {
				return newUnmarshalDirectiveError(dir, "directive requires zero parameters")
			}
		}

		if err := dec.unmarshalBlock(dir.Children, v); err != nil {
			return err
		}
	default:
		if len(dir.Children) != 0 {
			return newUnmarshalDirectiveError(dir, "directive requires zero children")
		}
		if err := unmarshalParamList(dir, v); err != nil {
			return err
		}
	}
	return nil
}

func unmarshalParamList(dir *Directive, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Slice:
		t := v.Type()
		sv := reflect.MakeSlice(t, len(dir.Params), len(dir.Params))
		for i, param := range dir.Params {
			if err := unmarshalParam(param, sv.Index(i)); err != nil {
				return newUnmarshalParamError(dir, i, err)
			}
		}
		v.Set(sv)
	case reflect.Array:
		if len(dir.Params) != v.Len() {
			return newUnmarshalDirectiveError(dir, fmt.Sprintf("directive requires exactly %v parameters", v.Len()))
		}
		for i, param := range dir.Params {
			if err := unmarshalParam(param, v.Index(i)); err != nil {
				return newUnmarshalParamError(dir, i, err)
			}
		}
	default:
		if len(dir.Params) != 1 {
			return newUnmarshalDirectiveError(dir, "directive requires exactly one parameter")
		}
		if err := unmarshalParam(dir.Params[0], v); err != nil {
			return newUnmarshalParamError(dir, 0, err)
		}
	}

	return nil
}

func unmarshalParam(param string, v reflect.Value) error {
	v = unwrapPointers(v)
	t := v.Type()

	// TODO: improve our logic following:
	// https://cs.opensource.google/go/go/+/refs/tags/go1.21.5:src/encoding/json/decode.go;drc=b9b8cecbfc72168ca03ad586cc2ed52b0e8db409;l=421
	if v.CanAddr() {
		if v, ok := v.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return v.UnmarshalText([]byte(param))
		}
	}

	switch v.Kind() {
	case reflect.String:
		v.Set(reflect.ValueOf(param))
	case reflect.Bool:
		switch param {
		case "true":
			v.Set(reflect.ValueOf(true))
		case "false":
			v.Set(reflect.ValueOf(false))
		default:
			return fmt.Errorf("invalid bool parameter %q", param)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(param, 10, t.Bits())
		if err != nil {
			return fmt.Errorf("invalid %v parameter: %v", t, err)
		}
		v.Set(reflect.ValueOf(i).Convert(t))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(param, 10, t.Bits())
		if err != nil {
			return fmt.Errorf("invalid %v parameter: %v", t, err)
		}
		v.Set(reflect.ValueOf(u).Convert(t))
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(param, t.Bits())
		if err != nil {
			return fmt.Errorf("invalid %v parameter: %v", t, err)
		}
		v.Set(reflect.ValueOf(f).Convert(t))
	default:
		return fmt.Errorf("unsupported type for unmarshaling parameter: %v", t)
	}

	return nil
}

func unwrapPointers(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

func clearMap(v reflect.Value) {
	for _, k := range v.MapKeys() {
		v.SetMapIndex(k, reflect.Value{})
	}
}

type unmarshalDirectiveError struct {
	lineno int
	name   string
	msg    string
}

func newUnmarshalDirectiveError(dir *Directive, msg string) *unmarshalDirectiveError {
	return &unmarshalDirectiveError{
		name:   dir.Name,
		lineno: dir.lineno,
		msg:    msg,
	}
}

func (err *unmarshalDirectiveError) Error() string {
	return fmt.Sprintf("line %v, directive %q: %v", err.lineno, err.name, err.msg)
}

type unmarshalParamError struct {
	lineno     int
	directive  string
	paramIndex int
	err        error
}

func newUnmarshalParamError(dir *Directive, paramIndex int, err error) *unmarshalParamError {
	return &unmarshalParamError{
		directive:  dir.Name,
		lineno:     dir.lineno,
		paramIndex: paramIndex,
		err:        err,
	}
}

func (err *unmarshalParamError) Error() string {
	return fmt.Sprintf("line %v, directive %q, parameter %v: %v", err.lineno, err.directive, err.paramIndex+1, err.err)
}
