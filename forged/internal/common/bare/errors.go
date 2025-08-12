// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright (c) 2025 Drew Devault <https://drewdevault.com>

package bare

import (
	"errors"
	"fmt"
	"reflect"
)

var ErrInvalidStr = errors.New("string contains invalid UTF-8 sequences")

type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return fmt.Sprintf("unsupported type for marshaling: %s\n", e.Type.String())
}
