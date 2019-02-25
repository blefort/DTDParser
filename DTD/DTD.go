// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package DTD Represents main structs of a DTD
//
// This package offer one struct which represents every blocks of the DTD
// Each struct will implements the IDTDBlock
// the IDTDBlock is necessary for the parser to populate and process its collection
package DTD

import "strings"

// Each constant represents a DTD block
const (
	ATTRIBUTE       = 1
	CDATA           = 2
	COMMENT         = 3
	ELEMENT         = 4
	ENTITY          = 5
	PCDATA          = 6
	EXPORTED_ENTITY = 7
	ATTLIST         = 8
)

// IDTDBlock Interface for DTD block
type IDTDBlock interface {
	GetName() string
	Render() string
	SetExported(v bool)
	GetSrc() string
}

// Entity represents a DTD Entity
type Entity struct {
	Parameter   bool
	ExternalDTD bool
	Name        string
	Value       string
	Public      bool
	System      bool
	Url         string
	Exported    bool
	Src         string
}

// ExportedEntity represent an exported entity
type ExportedEntity struct {
	Name string
}

// Attlist represent an attlist
type Attlist struct {
	Name  string
	Value string
	Src   string
}

// Comment represent a comment
type Comment struct {
	Value    string
	Exported bool
	Src      string
}

// Helper to join strings
func join(strs ...string) string {
	var sb strings.Builder
	for _, str := range strs {
		sb.WriteString(str)
	}
	return sb.String()
}

/**
 * Methods for entity struct
 */

// Render an entity
// implements IDTDBlock
func (e Entity) Render() string {

	var m string
	var eType string
	var exportedStr string
	var url string

	if e.ExternalDTD {
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
		exportedStr = join("%", e.Name, ";")
	}

	if e.Url == "" {
		url = ""
	} else {
		url = " \"" + e.Url + "\""
	}

	return join("<!ENTITY", m, e.Name, "\n", eType, e.Value, url, ">", exportedStr, "\n")
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

// GetSrc return the source filename where the entity was first found
// implements IDTDBlock
func (e *Entity) GetSrc() string {
	return e.Src
}

/**
 * Methods for comment struct
 */

// Render an entity
// implements IDTDBlock
func (c *Comment) Render() string {
	return "<!-- " + c.Value + " -->"
}

// GetName Get the name
// implements IDTDBlock
func (c *Comment) GetName() string {
	return "comment"
}

// SetExported set the current entity to exported
// implements IDTDBlock
func (c *Comment) SetExported(v bool) {
	panic("A comment should never be set as exported")
}

// GetSrc return the source filename where the entity was first found
// implements IDTDBlock
func (c *Comment) GetSrc() string {
	return c.Src
}

/**
 * Methods for ExportedEntity
 */

// Render an entity
// implements IDTDBlock
func (e *ExportedEntity) Render() string {
	return ""
}

// GetName Get the name
// implements IDTDBlock
func (e *ExportedEntity) GetName() string {
	return "exportedEntity"
}

// SetExported set the current entity to exported
// implements IDTDBlock
func (e *ExportedEntity) SetExported(v bool) {
	panic("A comment should never be set as exported")
}

// GetSrc return the source filename where the entity was first found
// implements IDTDBlock
func (e *ExportedEntity) GetSrc() string {
	panic("Am exported entity has no src")
}

/**
 * Attlist
 */

// Render an Attlist
// implements IDTDBlock
func (a *Attlist) Render() string {
	return "<!-- " + a.Value + " -->"
}

// GetName Get the name
// implements IDTDBlock
func (a *Attlist) GetName() string {
	return a.Name
}

// SetExported set the current entity to exported
// implements IDTDBlock
func (a *Attlist) SetExported(v bool) {
	panic("A comment should never be set as exported")
}

// GetSrc return the source filename where the entity was first found
// implements IDTDBlock
func (a *Attlist) GetSrc() string {
	return a.Src
}
