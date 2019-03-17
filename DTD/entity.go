// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package DTD Represents main structs of a DTD
package DTD

// Entity representss a DTD Entity
type Entity struct {
	Parameter   bool
	ExternalDTD bool
	Name        string
	Value       string
	Public      bool
	System      bool
	Url         string
	Exported    bool
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
	} else {
		eType = " "
	}

	if e.Exported {
		exportedStr = join("\n%", e.Name, ";")
	}

	if e.Url == "" {
		url = ""
	} else {
		url = " \"" + e.Url + "\""
	}

	el := join("<!ENTITY", m, e.Name, "\n", eType, e.Value, url, ">", exportedStr, "\n")

	return el
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

// GetExported Tells if the entity was exported
// implements IDTDBlock
func (e *Entity) GetExported() bool {
	return e.Exported
}

// GetValue Get the value
// implements IDTDBlock
func (e *Entity) GetValue() string {
	return e.Value
}

// GetParameter return parameter for entity only
// implements IDTDBlock
func (e *Entity) GetParameter() bool {
	return e.Parameter
}

// GetUrl the entity url
// implements IDTDBlock
func (e *Entity) GetUrl() string {
	return e.Url
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
