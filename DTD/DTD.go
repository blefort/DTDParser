// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package DTD Representss main structs of a DTD
//
// Specifications: https://www.w3.org/TR/xml11/
//
// Found this reference very usefull: https://xmlwriter.net/xml_guide/attlist_declaration.shtml
//
// # This is a simplified implementation
//
// This package offers one struct per DTD blocks
// Each struct will implements the IDTDBlock
// the IDTDBlock is necessary for the parser to populate and process its collection
package DTD

import (
	"fmt"
	"strings"
)

const (
	// DTD Block type
	UNIDENTIFIED    = -1
	XMLDECL         = 0
	ATTRIBUTE       = 1
	COMMENT         = 3
	ELEMENT         = 4
	ENTITY          = 5
	PCDATA          = 6
	EXPORTED_ENTITY = 7
	ATTLIST         = 8
	NOTATION        = 9

	// string type
	CDATA = 20

	// Tokenized Attribute type
	TOKEN_ID       = 30
	TOKEN_IDREF    = 31
	TOKEN_IDREFS   = 32
	TOKEN_ENTITY   = 33
	TOKEN_ENTITIES = 34
	TOKEN_NMTOKEN  = 35
	TOKEN_NMTOKENS = 36

	//Enumerated Attribute Type:	Attribute Description:
	ENUM_NOTATION = 37
	ENUM_ENUM     = 38
)

type DTDExtra struct {
	IsPublic    bool
	IsSystem    bool
	IsExported  bool
	IsParameter bool
	Attributes  []Attribute
	Url         string
	PublicID    string
	SystemID    string
}

// IDTDBlock Interface for DTD block
type IDTDBlock interface {
	GetName() string
	Render() string
	SetExported(v bool)
	GetValue() string
	GetExtra() *DTDExtra
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
	case COMMENT:
		return "Comment"
	case ELEMENT:
		return "Element"
	case ENTITY:
		return "Entity"
	case EXPORTED_ENTITY:
		return "Exported"
	case ATTLIST:
		return "Attlist"
	case NOTATION:
		return "Notation"
	default:
		panic("Unknown type" + fmt.Sprintf("%d", i) + " requested")
	}
}

// renderSystem Render SYSTEM NOTATION
func renderSystem(isSystem bool) string {
	if isSystem {
		return " SYSTEM "
	}
	return ""
}

// renderSystem Render PUBLIC NOTATION
func renderPublic(isPublic bool) string {
	if isPublic {
		return " PUBLIC "
	}
	return ""
}
