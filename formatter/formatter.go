// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package scanner allows to extract information from the DTD and create corresponding DTD structs
package formatter

import (
	"strings"

	"github.com/blefort/DTDParser/DTD"
)

// IDTDBlock Interface for DTD block
type iDTDFormatter interface {
	Render(Collection []DTD.IDTDBlock) string
}

// Helper to join strings
func join(strs ...string) string {
	var sb strings.Builder
	for _, str := range strs {
		sb.WriteString(str)
	}
	return sb.String()
}

// printQuoted print the value with double quote if not empty
func renderQuoted(s string) string {
	if s == "" {
		return ""
	}
	return "\"" + s + "\""
}
