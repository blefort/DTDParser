// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package scanner allows to extract information from the DTD and create corresponding DTD structs
package scanner

import (
	"bufio"
	"errors"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/blefort/DTDParser/DTD"
)

// DTDScanner represents a DTD scanner
type DTDScanner struct {
	Data         *bufio.Scanner
	WithComments bool
	Filepath     string
}

// NewScanner returns a new DTD Scanner
func NewScanner(path string, s string) *DTDScanner {
	var scanner DTDScanner
	scanner.Data = bufio.NewScanner(strings.NewReader(s))
	scanner.Data.Split(bufio.ScanRunes)
	scanner.Filepath = path
	return &scanner
}

// Next Move to the next character
func (sc *DTDScanner) Next() bool {
	return sc.Data.Scan()
}

// Scan the string to find the next block
func (sc *DTDScanner) Scan() (DTD.IDTDBlock, error) {

	var nType int

	log.Tracef("Character '%s'", sc.Data.Text())

	// seek until a block it found
	if !sc.IsStartChar() {
		return nil, errors.New("no block found")
	}

	// determine DTD Block
	nType = sc.seekType()

	log.Tracef("Possible type is %v", nType)

	if nType != 0 {
		log.Tracef("Block %s found", DTD.Translate(nType))
	}

	// create struct depending DTD type
	if nType == DTD.COMMENT {
		commentStr := sc.SeekComment()
		comment := sc.ParseComment(commentStr)
		comment.Src = sc.Filepath
		return comment, nil
	}

	if nType == DTD.ENTITY {
		entStr := sc.SeekEntity()
		entity := ParseEntity(entStr)
		entity.Src = sc.Filepath
		return entity, nil
	}

	if nType == DTD.ATTLIST {
		entStr := sc.SeekEntity()
		attlist := sc.ParseAttlist(entStr)
		attlist.Src = sc.Filepath
		return attlist, nil
	}

	// this one update the entity struct in the collection
	// to see the Exported property to true
	// That way it could be rendered properly
	if nType == DTD.EXPORTED_ENTITY {
		var exported DTD.ExportedEntity
		exported.Name = sc.SeekExportedEntity()
		return &exported, nil
	}

	return nil, errors.New("no block found")
}

// ParseEntity Parse a string and return pointer to a DTD.Entity
// @ref https://www.w3.org/TR/xml11/#sec-entity-decl
//
// Entity Declaration
// [70]   	EntityDecl ::=   	GEDecl | PEDecl
// [71]   	GEDecl     ::=   	'<!ENTITY' S Name S EntityDef S? '>'
// [72]   	PEDecl     ::=   	'<!ENTITY' S '%' S Name S PEDef S? '>'
// [73]   	EntityDef  ::=   	EntityValue | (ExternalID NDataDecl?)
// [74]   	PEDef	   ::=   	EntityValue | ExternalID
//
func ParseEntity(s string) *DTD.Entity {
	var e DTD.Entity
	var s2 string

	s2 = normalizeSpace(s)
	parts := seekWords(s2)

	// parameter
	if parts[0] == "%" {
		e.Parameter = true
		e.Name = parts[1]
	} else {
		e.Name = parts[0]
	}

	for _, part := range parts {

		log.Tracef("part is %v", part)

		if part == "PUBLIC" {
			e.Public = true
		}
		if part == "SYSTEM" {
			e.System = true
		}
	}

	if (e.System || e.Public) && e.Parameter {
		e.ExternalDTD = true
	}

	if e.ExternalDTD {
		e.Url = strings.Trim(parts[len(parts)-1], "\"")
	}

	if e.ExternalDTD {
		e.Value = parts[len(parts)-2]
	} else {
		e.Value = parts[len(parts)-1]
	}

	return &e
}

// seekWords Walk a string and identify every words
func seekWords(s string) []string {
	regex := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)'`)
	parts := regex.FindAllString(s, -1)
	log.Tracef("seekWords found %+v", parts)
	return parts
}

// isQuoted returns true if a character is quote or a double quote
func isQuoted(s string) bool {
	return s == "\"" || s == "'"
}

// ParseAttlist Parse a string and return pointer to a DTD.Attlist
//
// [52]   	AttlistDecl	   ::=   	'<!ATTLIST' S Name AttDef* S? '>'
// [53]   	AttDef	   ::=   	S Name S AttType S DefaultDecl
//
func (sc *DTDScanner) ParseAttlist(s string) *DTD.Attlist {
	var attlist DTD.Attlist
	var s2 string
	var i int

	s2 = normalizeSpace(s)

	parts := seekWords(s2)
	log.Tracef("parts are: %#v", parts)

	l := len(parts)

	if l == 0 {
		panic("Unable to scan Attlist")
	}

	attlist.Name = parts[0]

	for i = 1; i < l; i++ {

		var attr DTD.Attribute

		// Entity could be used
		if strings.HasPrefix(parts[i], "%") {
			log.Tracef("Ref. to an entity found %s", parts[i])
			attr.Value = parts[i]
			attlist.Attributes = append(attlist.Attributes, attr)
			continue
		}

		// The others parts are processed by group of 2 or 3 depending the type
		attr.Name = parts[i]
		attr.Type = seekAttributeType(parts[i+1])

		log.Tracef("type is = %d", attr.Type)

		if attr.Type == 0 {
			log.Fatal("ParseAttlist: Could not identitfy attribute type")
		}

		if attr.Type == DTD.ENUM_NOTATION {
			attr.Default = trimQuotes(parts[i+3])
			checkDefault(&attr)
			attr.Default = trimQuotes(parts[i+2])
			i = i + 3
		} else if attr.Type == DTD.ENUM_ENUM {
			attr.Default = trimQuotes(parts[i+2])
			checkDefault(&attr)
			attr.Default = parts[i+1]
			i = i + 2
		} else if attr.Type == DTD.TOKEN_NMTOKEN {
			attr.Default = trimQuotes(parts[i+2])
			checkDefault(&attr)
			i = i + 3
		} else {
			attr.Default = trimQuotes(parts[i+2])
			checkDefault(&attr)
			i = i + 2
		}

		if attr.Fixed {
			attr.Default = trimQuotes(parts[i])
			i++
		}

		attlist.Attributes = append(attlist.Attributes, attr)
	}

	return &attlist
}

//
func checkDefault(attr *DTD.Attribute) {
	attr.Required = isRequired(attr.Default)
	attr.Implied = isImplied(attr.Default)
	attr.Fixed = isFixed(attr.Default)

	if attr.Required || attr.Implied || attr.Fixed {
		attr.Default = ""
	}
}

// isRequired parse string and return true if equals to #REQUIRED
func isRequired(s string) bool {
	if strings.ToUpper(s) == "#REQUIRED" {
		return true
	}
	return false
}

// isImplied parse string and return true if equals to #IMPLIED
func isImplied(s string) bool {
	if strings.ToUpper(s) == "#IMPLIED" {
		return true
	}
	return false
}

// isFixed parse string and return true if equals to #FIXED
func isFixed(s string) bool {
	if strings.ToUpper(s) == "#FIXED" {
		return true
	}
	return false
}

// trimQuotes Trim surrounding quotes
func trimQuotes(s string) string {
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}
	return s
}

// seekAttributeType Attempt to identify attribute type
func seekAttributeType(s string) int {
	switch strings.ToUpper(s) {
	case "CDATA":
		return DTD.CDATA
	case "ID":
		return DTD.TOKEN_ID
	case "IDREF":
		return DTD.TOKEN_IDREF
	case "IDREFS":
		return DTD.TOKEN_IDREFS
	case "ENTITY":
		return DTD.TOKEN_ENTITY
	case "ENTITIES":
		return DTD.TOKEN_ENTITIES
	case "NMTOKEN":
		return DTD.TOKEN_NMTOKEN
	case "NMTOKENS":
		return DTD.TOKEN_NMTOKENS
	case "NOTATION":
		return DTD.ENUM_NOTATION
	}
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		return DTD.ENUM_ENUM
	}
	return 0
}

// ParseComment Parse a string and return pointer to DTD.Comment
func (sc *DTDScanner) ParseComment(s string) *DTD.Comment {
	var c DTD.Comment
	s = strings.TrimRight(s, "-")
	c.Value = s
	return &c
}

// IsStartChar Determine if a character is the beginning of a DTD block
func (sc *DTDScanner) IsStartChar() bool {
	ret := sc.Data.Text() == "<" || sc.Data.Text() == "%"
	log.Tracef("IsStartChar: %t", ret)
	return ret
}

// SeekExportedEntity Seek an exported DTD entity
// For example:
//  <!ENTITY % concept-dec
//           PUBLIC "-//OASIS//ENTITIES DITA 1.2 Concept//EN"
//           "concept.ent">%concept-dec;
//
// %concept-dec means that the entity is exported
func (sc *DTDScanner) SeekExportedEntity() string {
	var s string
	for sc.Data.Scan() {
		if sc.Data.Text() == ";" {
			return s
		}
		s += sc.Data.Text()
	}
	return s
}

// isWhitespace Determine if a string is a whitespace
func (sc *DTDScanner) isWhitespace() bool {
	return sc.Data.Text() == " " || sc.Data.Text() == "\t" || sc.Data.Text() == "\n"
}

// seekType Seek the type of the next DTD block
func (sc *DTDScanner) seekType() int {

	var s string

	if sc.Data.Text() == "%" {
		return DTD.EXPORTED_ENTITY
	}

	for sc.Data.Scan() {

		log.Tracef("Character '%s'", sc.Data.Text())

		s += sc.Data.Text()

		log.Tracef("Word is '%s'", s)

		if sc.isWhitespace() {
			return 0
		}

		if s == "!--" {
			return DTD.COMMENT
		}
		if s == "!ENTITY" {
			return DTD.ENTITY
		}
		if s == "!ATTLIST" {
			return DTD.ATTLIST
		}

	}

	return 0
}

// SeekEntity seek an entity
func (sc *DTDScanner) SeekEntity() string {
	var s string

	for sc.Data.Scan() {

		log.Tracef("Character '%s'", sc.Data.Text())

		if sc.Data.Text() == ">" {
			break
		}
		s += sc.Data.Text()
	}
	return s
}

// normalizeSpace Convert Line breaks, multiple space into a single space
func normalizeSpace(s string) string {
	regexLineBreak := regexp.MustCompile(`(?s)(\r?\n)|\t`)
	s1 := regexLineBreak.ReplaceAllString(s, " ")
	space := regexp.MustCompile(`\s+`)
	nm := strings.Trim(space.ReplaceAllString(s1, " "), " ")
	log.Tracef("Normalized string is '%s'", nm)
	return nm
}

// SeekComment Seek a comment
func (sc *DTDScanner) SeekComment() string {

	var s string

	for sc.Data.Scan() {
		var last string

		log.Tracef("Character '%s'", sc.Data.Text())

		// last 2 character of a string
		if len(s) > 2 {
			last = s[len(s)-2:]
		}

		if sc.Data.Text() == ">" && last == "--" {
			break
		}
		s += sc.Data.Text()
	}
	return s
}
