// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"log/slog"

	"github.com/emersion/go-message"
)

func lmtpHandlePatch(groupPath []string, repoName string, email *message.Entity) (err error) {
	slog.Info("Pretend like I'm handling a patch!")
	return nil
}
