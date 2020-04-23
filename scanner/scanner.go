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

	log "github.com/sirupsen/logrus"

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
		return comment, nil
	}

	if nType == DTD.ENTITY {
		entStr := sc.SeekBlock()
		entity := ParseEntity(entStr)
		return entity, nil
	}

	if nType == DTD.ATTLIST {
		entStr := sc.SeekBlock()
		attlist := sc.ParseAttlist(entStr)
		return attlist, nil
	}

	if nType == DTD.ELEMENT {
		elStr := sc.SeekBlock()
		element := ParseElement(elStr)
		return element, nil
	}

	if nType == DTD.NOTATION {
		notStr := sc.SeekBlock()
		notation := ParseNotation(notStr)
		return notation, nil
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

// ParseEntity Parse a string and return pointer to a DTD.Element
// @ref https://www.w3.org/TR/xml11/#elemdecls
//
// Element Declaration
// [45]   	elementdecl	   ::=   	'<!ELEMENT' S Name S contentspec S? '>'	[VC: Unique Element Type Declaration]
// [46]   	contentspec	   ::=   	'EMPTY' | 'ANY' | Mixed | children
//
func ParseElement(s string) *DTD.Element {
	var e DTD.Element
	var s2 string

	s2 = normalizeSpace(s)
	parts := SeekWords(s2)

	e.Name = parts[0]
	parts = parts[1:len(parts)]

	e.Value = strings.Join(parts, " ")

	return &e
}

// ParseNotation Parse a string and return pointer to a DTD.Notation
// @ref https://www.w3.org/TR/xml11/#Notations
//
// Element Declaration
//
// [82]  NotationDec ::= '<!NOTATION' S Name S (ExternalID | PublicID) S? '>'  [VC: Unique Notation Name]
// [83]  PublicID    ::= 'PUBLIC' S PubidLiteral
//
func ParseNotation(s string) *DTD.Notation {
	var n DTD.Notation
	var s2 string

	s2 = normalizeSpace(s)
	parts := SeekWords(s2)
	l := len(parts)

	n.Name = parts[0]

	if isPublic(parts[1]) {
		n.Public = true
	}
	if isSystem(parts[1]) {
		n.System = true
	}

	if l == 3 {
		n.Value = trimQuotes(parts[2])
	} else if l == 4 {
		n.ID = trimQuotes(parts[2])
		n.Value = trimQuotes(parts[3])
	} else {
		panic("Unsupported Notation")
	}

	return &n
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
	parts := SeekWords(s2)

	// parameter
	if parts[0] == "%" {
		e.Parameter = true
		e.Name = parts[1]
	} else {
		e.Name = parts[0]
	}

	for _, part := range parts {

		log.Tracef("part is %v", part)

		if isPublic(part) {
			e.Public = true
		}
		if isSystem(part) {
			e.System = true
		}
	}

	if e.System || e.Public {
		e.ExternalDTD = true
		e.Url = strings.Trim(parts[len(parts)-1], "\"")
		assignIfEntityValue(&e, parts[len(parts)-2])
	} else {
		e.Value = parts[len(parts)-1]
	}

	return &e
}

// assignIfEntityValue test if v is not public system or empty before assigning it
func assignIfEntityValue(e *DTD.Entity, v string) {
	if v != "" && !isPublic(v) && !isSystem(v) {
		e.Value = v
	}
}

// SeekWords Walk a string and identify every words
func SeekWords(s string) []string {

	r2 := `"(.*?)"|\((.*)\)[\+|\?|\*]?|([^\s]+)`

	regex := regexp.MustCompile(r2)
	parts := regex.FindAllString(s, -1)

	log.Tracef("seekWords FindAllString found %#v", parts)
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
	log.Tracef("ParseAttlist received: '%s'", s)

	parts := SeekWords(s2)
	log.Tracef("parts are: %#v", parts)

	l := len(parts)

	if l == 0 {
		panic("Unable to scan Attlist")
	}

	attlist.Name = parts[0]
	log.Warnf("Attlist for Element name: '%s'", attlist.Name)

	for i = 1; i < l; i++ {

		// There are 3 to 4 values to test, we don't know

		var attr DTD.Attribute

		if strings.HasPrefix(parts[i], "%") {
			log.Tracef("Ref. to an entity found %s", parts[i])
			attr.Entities = append(attr.Entities, parts[i])
			attlist.Attributes = append(attlist.Attributes, attr)
			continue
		}

		// CASE 2
		// The others parts are processed by group of 3 to 4 depending the type
		// addToI will be the number of parts in the attribute definition,
		// min value is 3 (name, type, value)
		// this value is minus 1 because the for loop add a +1
		addToI := 1

		// Name is always the first
		attr.Name = parts[i]
		log.Tracef("* Processing attribute: '%s'", attr.Name)

		// Type is always in the second position
		attr.Type = DTD.SeekAttributeType(parts[i+1])

		if attr.Type == 0 {
			log.Fatalf("Could not identify attribute type, name: '%s', value: '%s'", attr.Name, parts[i+1])
		}

		// we have to test for 3 and 4
		log.Tracef("Type: '%s', DTD code: %d ('%s') (Empty means ENUM ENNUM)", parts[i+1], attr.Type, DTD.AttributeType(attr.Type))

		if attr.Type == DTD.ENUM_NOTATION {
			attr.Default = checkDefaultValue(trimQuotes(parts[i+2]))
			addToI++
		} else if attr.Type == DTD.CDATA {
			addToI += checkCDATADefaultValue(&attr, parts, i, 2)
		} else if attr.Type == DTD.ENUM_ENUM {
			addToI += checkEnumDefaultValue(&attr, parts, i, 1)
		} else {
			attr.Default = checkDefaultValue(parts[i+2])
		}

		addToI += checkDefaultDeclaration(&attr, parts, i, 2)
		addToI += checkDefaultDeclaration(&attr, parts, i, 3)

		attlist.Attributes = append(attlist.Attributes, attr)

		i = i + addToI
	}
	log.Tracef("%+v", attlist)
	return &attlist
}

// checkEnumDefaultValue
func checkEnumDefaultValue(attr *DTD.Attribute, parts []string, i int, position int) int {
	idx := i + position

	if idx >= len(parts) {
		log.Trace("Skipping checkEnumDefaultValue: end of parts reached")
		return 0
	}

	attr.Default = checkDefaultValue(parts[idx])
	log.Tracef("Enum Default Value: '%s'", attr.Default)
	return 1
}

// checkCDATADefaultValue
func checkCDATADefaultValue(attr *DTD.Attribute, parts []string, i int, position int) int {
	idx := i + position

	if idx >= len(parts) {
		log.Trace("Skipping checkCDATADefaultValue: end of parts reached")
		return 0
	}

	v := parts[idx]

	if isQuoted(v[0:1]) {
		log.Tracef("CDATA default value is: '%s'", v)
		attr.Default = v
		return 1
	}

	log.Trace("-CDATA default value not found")
	return 0
}

// checkCDATADefaultValue
func checkDefaultValue(v string) string {

	log.Tracef("testing default value'%s'", v)

	if isQuoted(v[0:1]) {
		log.Tracef("default value is '%s'", v)
		return v
	}

	if isRequired(v) {
		return ""
	}

	if isImplied(v) {
		return ""
	}

	if isFixed(v) {
		return ""
	}

	log.Tracef("Default value '%s' found", v)
	return v
}

// checkDefault Check if the default value if required, implied or fixed and reste Default property
func checkDefaultDeclaration(attr *DTD.Attribute, parts []string, i int, position int) int {
	idx := i + position

	if idx >= len(parts) {
		log.Trace("Skipping checkDefaultDeclaration: end of parts reached")
		return 0
	}

	v := parts[idx]

	if isRequired(v) {
		log.Tracef("'%s', position: '%d'", v, position)
		attr.Required = isRequired(v)
		return 1
	}
	if isImplied(v) {
		log.Tracef("'%s', position: '%d'", v, position)
		attr.Implied = isImplied(v)
		return 1
	}
	if isFixed(v) {
		log.Tracef("'%s', position: '%d'", v, position)

		idx2 := idx + 1

		if idx2 >= len(parts) {
			log.Trace("Skipping checkDefaultDeclaration: end of parts reached")
			return 1
		}

		attr.Fixed = isFixed(v)
		attr.Default = checkDefaultValue(trimQuotes(parts[idx2]))
		log.Tracef("Default value for FIXED is '%s'", attr.Default)
		return 2
	}

	log.Tracef("No default attribute value found in position '%d' (implied, required, fixed)", position)
	return 0
}

// isRequired parse string and return true if equals to #REQUIRED
func isRequired(s string) bool {
	if strings.Trim(strings.ToUpper(s), " ") == "#REQUIRED" {
		return true
	}
	return false
}

// isImplied parse string and return true if equals to #IMPLIED
func isImplied(s string) bool {
	if strings.Trim(strings.ToUpper(s), " ") == "#IMPLIED" {
		return true
	}
	return false
}

// isFixed parse string and return true if equals to #FIXED
func isFixed(s string) bool {
	if strings.Trim(strings.ToUpper(s), " ") == "#FIXED" {
		return true
	}
	return false
}

// isPublic parse string and return true if equals to PUBLIC
func isPublic(s string) bool {
	if strings.Trim(strings.ToUpper(s), " ") == "PUBLIC" {
		return true
	}
	return false
}

// isSystem parse string and return true if equals to SYSTEM
func isSystem(s string) bool {
	if strings.Trim(strings.ToUpper(s), " ") == "SYSTEM" {
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

		s += sc.Data.Text()
		log.Tracef("Character '%s', Word is '%s'", sc.Data.Text(), s)

		if s == "!--" {
			return DTD.COMMENT
		}

		// detecting the space below is important
		if s == "!ENTITY " {
			return DTD.ENTITY
		}
		if s == "!ELEMENT " {
			return DTD.ELEMENT
		}
		if s == "!ATTLIST " {
			return DTD.ATTLIST
		}
		if s == "!NOTATION " {
			return DTD.NOTATION
		}
		if sc.isWhitespace() {
			return 0
		}

	}

	return 0
}

// SeekBlock seek an entity
func (sc *DTDScanner) SeekBlock() string {
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
	space := regexp.MustCompile(`\s+|\t`)
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
