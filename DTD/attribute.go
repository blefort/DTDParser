// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package DTD Represents main structs of a DTD
package DTD

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

// Attribute represents an attribute
type Attribute struct {
	Name     string
	Type     int
	Default  string
	Value    string
	Implied  bool
	Required bool
	Fixed    bool
	IsEntity bool
}

// Render an Attribute
func (a *Attribute) Render(withLineBreak bool) string {
	s := ""
	if withLineBreak {
		s = "\t"
	}

	if a.Name != "" {
		s += a.Name
		if withLineBreak {
			s += "\t"
		}
	}

	s += AttributeType(a.Type) + " "

	if a.Fixed {
		s += " #FIXED "
	}

	if a.Value != "" && !a.IsEntity {
		s += "\"" + a.Value + "\""
	} else if a.Value != "" && a.IsEntity {
		s += a.Value
	}

	if a.Implied {
		s += " #IMPLIED "
	}

	if a.Required {
		s += " #REQUIRED "
	}

	if withLineBreak {
		s += "\n"
	}

	return s
}

// GetEntityValue return the entity name that was referenced
// since the reference contains start with a '%' and finish with a ';'
// it must be removed
func (a *Attribute) GetEntityValue() string {
	if a.IsEntity {
		return a.Value[1 : len(a.Value)-1]
	}
	return ""
}

// GetExported Unused, tells if the comment was exported
// implements IDTDBlock
func (a *Attribute) GetName() string {
	return "attribute"
}

// SeekAttributeType Attempt to identify attribute type
func SeekAttributeType(s string) int {
	switch strings.ToUpper(s) {
	case "CDATA":
		return CDATA
	case "ID":
		return TOKEN_ID
	case "IDREF":
		return TOKEN_IDREF
	case "IDREFS":
		return TOKEN_IDREFS
	case "ENTITY":
		return TOKEN_ENTITY
	case "ENTITIES":
		return TOKEN_ENTITIES
	case "NMTOKEN":
		return TOKEN_NMTOKEN
	case "NMTOKENS":
		return TOKEN_NMTOKENS
	case "NOTATION":
		return ENUM_NOTATION
	}
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		return ENUM_ENUM
	}
	return 0
}

// AttributeType convert DTD Attribute type (int) to its corresponding string value
func AttributeType(a int) string {
	switch a {
	case CDATA:
		return "CDATA"
	case TOKEN_ID:
		return "ID"
	case TOKEN_IDREF:
		return "IDREF"
	case TOKEN_IDREFS:
		return "IDREFS"
	case TOKEN_ENTITY:
		return "ENTITY"
	case TOKEN_ENTITIES:
		return "ENTITIES"
	case TOKEN_NMTOKEN:
		return "NMTOKEN"
	case TOKEN_NMTOKENS:
		return "NMTOKENS"
	case ENUM_NOTATION:
		return "NOTATION"
	case ENUM_ENUM:
		return ""
	}
	log.Debugf("No attribute type conversion possible for '%d'", a)
	return ""
}
