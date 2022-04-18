// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package DTD Represents main structs of a DTD
package DTD

import "go.uber.org/zap/zapcore"

// Entity representss a DTD Entity
type Entity struct {
	Parameter  bool
	IsExternal bool
	IsInternal bool
	Name       string
	Value      string
	Public     bool
	System     bool
	Url        string
	Exported   bool
	Attributes []Attribute
}

func (e Entity) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("Rendered Entity", e.Render())
	return nil
}

// Render an entity
// implements IDTDBlock
func (e Entity) Render() string {

	var m string
	var eType string
	var exportedStr string
	var url string

	if e.Parameter {
		m = " % "
	} else {
		m = " "
	}

	if e.Public {
		eType += " PUBLIC "
	} else if e.System {
		eType += " SYSTEM "
	}

	if e.Exported {
		exportedStr = join("\n%", e.Name, ";")
	}

	if e.Url != "" {
		url = "\"" + e.Url + "\""
	}

	return join("<!ENTITY", m, e.Name, " ", eType, "\"", e.Value, "\"", url, "\n>", exportedStr, "\n")
}

// GetName Get the name
// implements IDTDBlock
func (e *Entity) GetName() string {
	return e.Name
}

// SetExported set the current entity to exported
// implements IDTDBlock
func (e *Entity) SetExported(v bool) {
	e.Exported = v
}

// GetValue Get the value
// implements IDTDBlock
func (e *Entity) GetValue() string {
	return e.Value
}

// GetExtra Get extrainformation
func (e *Entity) GetExtra() *DTDExtra {
	var extra DTDExtra
	extra.IsExported = e.Exported
	extra.IsParameter = e.Parameter
	extra.IsPublic = e.Public
	extra.IsSystem = e.System
	extra.Url = e.Url
	return &extra
}

// IsEntityType check if the interface is a DTD.ExportedEntity
func IsEntityType(i interface{}) bool {
	switch i.(type) {
	case *Entity:
		return true
	default:
		return false
	}
}
