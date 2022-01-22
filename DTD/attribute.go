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
	Entities []string
}

// Render an Attribute
func (a *Attribute) Render() string {
	s := ""

	for _, ent := range a.Entities {
		s = s + ent + " "
	}

	if len(s) > 0 {
		return s
	}

	s += a.Name + " "
	s += AttributeType(a.Type) + " "

	switch a.Type {
	case ENUM_NOTATION:
		s += a.Default
		break
	case ENUM_ENUM:
		s += a.Default
		break
	case TOKEN_NMTOKEN:
		break
	default:
		s += a.Default
	}

	if a.Implied {
		s += " #IMPLIED "
	}

	if a.Required {
		s += " #REQUIRED "
	}

	if a.Fixed {
		s += " #FIXED "
	}

	if a.Type == TOKEN_NMTOKEN {
		s += a.Default
	}

	if a.Value != "" {
		s += " " + a.Value
	}

	return s
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
	log.Tracef("No attribute type conversion possible for '%d'", a)
	return ""
}
