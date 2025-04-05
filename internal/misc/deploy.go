// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package misc

import (
	"io"
	"io/fs"
	"os"
)

func DeployBinary(src fs.File, dst string) (err error) {
	var dstFile *os.File
	if dstFile, err = os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755); err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, src)
	return err
}
