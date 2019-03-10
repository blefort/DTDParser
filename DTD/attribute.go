// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package DTD Represents main structs of a DTD
package DTD

// Attribute represents an attribute
type Attribute struct {
	Name     string
	Type     int
	Default  string
	Value    string
	Implied  bool
	Required bool
	Fixed    bool
}
