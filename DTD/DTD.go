// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package DTD Represents main structs of a DTD
//
// Specifications: https://www.w3.org/TR/xml11/
//
// Found this reference very usefull: https://xmlwriter.net/xml_guide/attlist_declaration.shtml
//
// This is a simplified implementation
//
// This package offers one struct per DTD blocks
// Each struct will implements the IDTDBlock
// the IDTDBlock is necessary for the parser to populate and process its collection
package DTD

import (
	"strings"
)

const (

	// DTD Block type
	ATTRIBUTE       = 1
	COMMENT         = 3
	ELEMENT         = 4
	ENTITY          = 5
	PCDATA          = 6
	EXPORTED_ENTITY = 7
	ATTLIST         = 8

	// string type
	CDATA = 9

	// Tokenized Attribute type
	TOKEN_ID       = 10
	TOKEN_IDREF    = 11
	TOKEN_IDREFS   = 12
	TOKEN_ENTITY   = 13
	TOKEN_ENTITIES = 14
	TOKEN_NMTOKEN  = 15
	TOKEN_NMTOKENS = 16

	//Enumerated Attribute Type:	Attribute Description:
	ENUM_NOTATION = 17
	ENUM_ENUM     = 18
)

// IDTDBlock Interface for DTD block
type IDTDBlock interface {
	GetName() string
	Render() string
	SetExported(v bool)
	GetExported() bool
	GetSrc() string
	GetValue() string
	GetParameter() bool
	GetUrl() string
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
	Name  string
	Value string
}

// Attribute represent an attribute
type Attribute struct {
	Name     string
	Type     int
	Default  string
	Value    string
	Implied  bool
	Required bool
	Fixed    bool
}

// Attlist represent an attlist
type Attlist struct {
	Name       string
	Value      string
	Src        string
	Attributes []Attribute
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

// Translate convert block type constant to a name
func Translate(i int) string {
	switch i {
	case ATTRIBUTE:
		return "Attribute"
	case CDATA:
		return "CDATA"
	case COMMENT:
		return "Comment"
	case ELEMENT:
		return "Element"
	case ENTITY:
		return "Entity"
	case PCDATA:
		return "PCDATA"
	case EXPORTED_ENTITY:
		return "Exported"
	case ATTLIST:
		return "Attlist"
	default:
		panic("Unknown type" + string(i) + " requested")
	}
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
		exportedStr = join("\n%", e.Name, ";")
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

// GetExported Tells if the entity was exported
// implements IDTDBlock
func (e *Entity) GetExported() bool {
	return e.Exported
}

// GetSrc return the source filename where the entity was first found
// implements IDTDBlock
func (e *Entity) GetSrc() string {
	return e.Src
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

// GetValue Get the value
// implements IDTDBlock
func (c *Comment) GetValue() string {
	return c.Value
}

// GetParameter return parameter for entity only
// implements IDTDBlock
func (c *Comment) GetParameter() bool {
	panic("Comment have no Parameter")
}

// GetUrl the entity url
// implements IDTDBlock
func (c *Comment) GetUrl() string {
	panic("GetUrl not allowed for this block")
}

// GetExported Unused, tells if the comment was exported
// implements IDTDBlock
func (c *Comment) GetExported() bool {
	panic("Comment are not exported")
}

// IsCommentType check if the interface is a DTD.Comment
func IsCommentType(i interface{}) bool {
	switch i.(type) {
	case *Comment:
		return true
	default:
		return false
	}
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
	return e.Name
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

// GetValue Get the value
// implements IDTDBlock
func (e *ExportedEntity) GetValue() string {
	return e.Value
}

// GetParameter return parameter for entity only
// implements IDTDBlock
func (e *ExportedEntity) GetParameter() bool {
	panic("ExportedEntity have no Parameter")
}

// GetUrl the entity url
// implements IDTDBlock
func (e *ExportedEntity) GetUrl() string {
	panic("GetUrl not allowed for this block")
}

// GetExported Unused, tells if the comment was exported
// implements IDTDBlock
func (e *ExportedEntity) GetExported() bool {
	panic("ExportedEntity are not exported")
}

// IsExportedEntityType check if the interface is a DTD.ExportedEntity
func IsExportedEntityType(i interface{}) bool {
	switch i.(type) {
	case *ExportedEntity:
		return true
	default:
		return false
	}
}

/**
 * Attlist
 */

// Render an Attlist
// implements IDTDBlock
func (a *Attlist) Render() string {
	return join("<!ATTLIST", a.Name, " ", a.Value, ">", "\n")
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

// GetValue Get the value
// implements IDTDBlock
func (a *Attlist) GetValue() string {
	return a.Value
}

// GetParameter return parameter for entity only
// implements IDTDBlock
func (a *Attlist) GetParameter() bool {
	panic("Attlist have no Parameter")
}

// GetUrl the entity url
// implements IDTDBlock
func (a *Attlist) GetUrl() string {
	panic("GetUrl not allowed for this block")
}

// GetExported Unused, tells if the comment was exported
// implements IDTDBlock
func (a *Attlist) GetExported() bool {
	panic("Attlist are not exported")
}

// IsAttlistType check if the interface is a DTD.Comment
func IsAttlistType(i interface{}) bool {
	switch i.(type) {
	case *Attlist:
		return true
	default:
		return false
	}
}
