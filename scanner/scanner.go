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
	CurrentLine  int
	CurrentChar  int
	CharCount    int
	LineCount    int
	CRLF         bool // 0 linux / 1 dos
	init         bool
	scanResult   bool
}

type parsedBlock struct {
	fullString string
	blockType  string
	name       string
	entity     string
	value      string
}

// NewScanner returns a new DTD Scanner
func NewScanner(path string, s string) *DTDScanner {
	var scanner DTDScanner
	scanner.Data = bufio.NewScanner(strings.NewReader(s))
	scanner.Data.Split(bufio.ScanRunes)
	scanner.Filepath = path
	scanner.CharCount = 0
	scanner.LineCount = 1
	scanner.CRLF = false // assume Linux
	scanner.init = false
	scanner.scanResult = true
	return &scanner
}

// Next Move to the next Block
func (sc *DTDScanner) NextBlock() bool {
	return sc.scanResult
}

// Next Move to the next character
func (sc *DTDScanner) next() bool {
	sc.CharCount++
	sc.scanResult = sc.Data.Scan()
	if sc.isEndOfLine() {
		sc.LineCount++
		log.Tracef("Line is %d", sc.LineCount)
	}
	return sc.scanResult
}

// Next Move to the next character
func (sc *DTDScanner) Previous() bool {
	sc.CharCount--
	ret := sc.Data.Scan()
	if sc.isEndOfLine() {
		sc.LineCount++
		log.Tracef("Line is %d", sc.LineCount)
	}
	return ret
}

// Scan the string to find the next block
func (sc *DTDScanner) Scan() (DTD.IDTDBlock, error) {

	s := sc.seekUntilNextBlock()

	//log.Warnf("Block line: '%s'", s)

	p, err := SeekBlockParts(s)

	if err != nil {
		return nil, errors.New("Unidentified block")
	}

	// 	// create struct depending DTD type
	// 	if nType == DTD.COMMENT {
	// 		commentStr := sc.SeekComment()
	// 		comment := sc.ParseComment(commentStr)
	// 		log.Warnf("Commment '%s' found at line '%d", comment.GetName(), sc.CurrentLine)
	// 		return comment, nil
	// 	}

	if p.blockType == "ENTITY" {
		entity := sc.ParseEntity(p)
		log.Infof("ENTITY '%s' (line %d)", entity.GetName(), sc.CurrentLine)
		sc.logOutputAttributes(&entity.Attributes)
		return entity, nil
	}

	if p.blockType == "ATTLIST" {
		attlist := sc.ParseAttlist(p)
		log.Infof("ATTLIST '%s' (line %d)", attlist.GetName(), sc.CurrentLine)
		sc.logOutputAttributes(&attlist.Attributes)
		return attlist, nil
	}

	if p.blockType == "ELEMENT" {
		element := sc.ParseElement(p)
		log.Infof("ELEMENT '%s' (line %d), value: '%s'", element.GetName(), sc.CurrentLine, element.Value)
		return element, nil
	}

	if p.blockType == "NOTATION" {
		notation := sc.ParseNotation(p)
		log.Infof("NOTATION '%s' (line %d)", notation.GetName(), sc.CurrentLine)
		return notation, nil
	}

	return nil, errors.New("Unidentified block")

}

// ParseEntity Parse a string and return pointer to a DTD.Element
// @ref https://www.w3.org/TR/xml11/#elemdecls
//
// Element Declaration
// [45]   	elementdecl	   ::=   	'<!ELEMENT' S Name S contentspec S? '>'	[VC: Unique Element Type Declaration]
// [46]   	contentspec	   ::=   	'EMPTY' | 'ANY' | Mixed | children
//
func (sc *DTDScanner) ParseElement(p *parsedBlock) *DTD.Element {
	var e DTD.Element
	e.Name = p.name
	e.Value = p.value
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
func (sc *DTDScanner) ParseNotation(p *parsedBlock) *DTD.Notation {
	var n DTD.Notation
	// var s2 string

	// s2 = normalizeSpace(s)
	// parts := SeekWords(s2)
	// l := len(parts)

	n.Name = p.name

	// if isPublic(parts[1]) {
	// 	n.Public = true
	// }
	// if isSystem(parts[1]) {
	// 	n.System = true
	// }

	// if l == 3 {
	// 	n.Value = trimQuotes(parts[2])
	// } else if l == 4 {
	// 	n.ID = trimQuotes(parts[2])
	// 	n.Value = trimQuotes(parts[3])
	// } else {
	// 	panic("Unsupported Notation")
	// }

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
func (sc *DTDScanner) ParseEntity(p *parsedBlock) *DTD.Entity {
	var e DTD.Entity
	// //var s2 string

	// //s2 = normalizeSpace(s)
	// parts := SeekWords(s)
	e.Name = p.name
	// // parameter
	if p.entity == "%" {
		e.Parameter = true
	}

	e.Value = p.value

	return &e
}

// ParseAttlist Parse a string and return pointer to a DTD.Attlist
//
// [52]   	AttlistDecl	   ::=   	'<!ATTLIST' S Name AttDef* S? '>'
// [53]   	AttDef	   ::=   	S Name S AttType S DefaultDecl
//
func (sc *DTDScanner) ParseAttlist(p *parsedBlock) *DTD.Attlist {
	var attlist DTD.Attlist
	attlist.Name = p.name
	sc.parseAttributes(p.value, &attlist.Attributes)
	return &attlist
}

func (sc *DTDScanner) parseAttributes(s string, attributes *[]DTD.Attribute) {

	var i int

	i = -1
	s2 := normalizeSpace(s)
	parts := SeekWords(s2)
	l := len(parts)

	log.Tracef("parse Attr: parts %v, s was '%s'", parts, s)

	// for idx, part := range parts {
	// 	log.Warnf("part %d: %s", idx, part)
	// }

	if l == 0 {
		panic("Unable to scan Attlist")
	}

	nextWord := func(i *int, l int) bool {
		if *i+1 < l {
			*i++
			log.Tracef("Next word, i: %d, %s ", *i, parts[*i])
			return true
		}
		return false
	}

	for nextWord(&i, l) {

		var attr DTD.Attribute

		// empty string
		if normalizeSpace(parts[i]) == "" {
			continue
		}

		// instruction varies depending the type, we don't know in advance

		// reference to an entity
		if strings.HasPrefix(parts[i], "%") {
			attr.Value = parts[i]
			log.Tracef("- Ref. to an entity found: %s", attr.Render())
			attr.IsEntity = true
			*attributes = append(*attributes, attr)
			continue
		}

		// first word is always the attribute name
		attr.Name = normalizeSpace(parts[i])
		log.Tracef("Processing attribute: '%s'", attr.Name)

		if !nextWord(&i, l) {
			log.Fatalf("Not enough arguments to loop through attributes i:%d", i)

		}

		// CASE 2
		// The others parts are processed by group of 3 to 4 depending the type

		// Type is always in the second position
		attr.Type = DTD.SeekAttributeType(parts[i])

		log.Tracef("attribute type is %d", attr.Type)

		if attr.Type == DTD.CDATA { // 20
			nextWord(&i, l)
			checkNextTwoArguments(parts, &i, &attr)
		}

		if attr.Type == DTD.TOKEN_ID ||
			attr.Type == DTD.TOKEN_IDREF ||
			attr.Type == DTD.TOKEN_IDREFS ||
			attr.Type == DTD.TOKEN_ENTITY ||
			attr.Type == DTD.TOKEN_ENTITIES ||
			attr.Type == DTD.TOKEN_NMTOKEN ||
			attr.Type == DTD.TOKEN_NMTOKENS {
			nextWord(&i, l)
			checkNextTwoArguments(parts, &i, &attr)
		}
		if attr.Type == DTD.ENUM_NOTATION {
			nextWord(&i, l)
			attr.Value = parts[i]
			nextWord(&i, l)
			checkDefaultDeclaration(&attr, parts[i])
		}
		if attr.Type == DTD.ENUM_ENUM {
			attr.Value = parts[i]
			nextWord(&i, l)
			checkDefaultDeclaration(&attr, parts[i])
		}

		*attributes = append(*attributes, attr)
		log.Warnf("*Attr rendered: %s", attr.Render())

		if attr.Type == 0 {
			log.Fatalf("Could not identify attribute type at line %d, name: '%s', value: '%s'", sc.LineCount, attr.Name, parts[i])
		}

	}
}

func checkNextTwoArguments(parts []string, i *int, attr *DTD.Attribute) {

	var s2 string
	s1 := parts[*i]

	if *i+1 < len(parts) {
		s2 = parts[*i+1]
	} else {
		s2 = ""
	}

	log.Tracef("checkNextTwoArguments is (i: %d) testing 2 args %s and %s", *i, s1, s2)

	if isQuoted(s1[0:1]) {
		*&attr.Value = s1
		log.Tracef("value is %s", s1)
		if checkDefaultDeclaration(attr, s2) {
			*i++
		}
	}

	if checkDefaultDeclaration(attr, s1) {
		log.Tracef("check if s2 is quoted: '%s'", s2)
		if s2 != "" && isQuoted(s2[0:1]) {
			log.Tracef("s2 is quoted: '%s'", s2)
			log.Tracef("value is %s", s2)
			*i++
			*&attr.Value = s2

		}
	}
	log.Tracef("checkNextTwoArguments i: %d", *i)
}

// output attributes in the log
func (sc *DTDScanner) logOutputAttributes(attributes *[]DTD.Attribute) {
	for _, attr := range *attributes {
		log.Warnf("- Attribute: %s", attr.Render())
	}
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

// SeekWords Walk a string and identify every words
func SeekBlockParts(s string) (*parsedBlock, error) {

	var p parsedBlock

	r2 := `<\!(ENTITY|ELEMENT|ATTLIST|COMMENT|NOTATION)\s*(\%)?\s*(\S+)?\s*([^>]+)?>\s*(%[^>\s]+)?`

	regex := regexp.MustCompile(r2)
	parts := regex.FindAllStringSubmatch(s, -1)

	if len(parts) == 0 {
		return nil, errors.New("Could not find any DTD block in " + s)
	}

	p.fullString = parts[0][0]
	p.blockType = parts[0][1]
	p.entity = parts[0][2]
	p.name = parts[0][3]
	p.value = parts[0][4]

	log.Tracef("SeekBlockParts, parsed: \n-name: %s\n-type:%s\n-entity:%s\n-value:%s, s was '%s'", p.name, p.blockType, p.entity, p.value, s)

	return &p, nil
}

// isQuoted returns true if a character is quote or a double quote
func isQuoted(s string) bool {
	return s == "\"" || s == "'"
}

// checkEnumDefaultValue
func checkEnumDefaultValue(attr *DTD.Attribute, parts []string) {

	attr.Default = checkDefaultValue(parts[0])
	log.Tracef("Enum Default Value: '%s'", attr.Default)
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
func checkDefaultDeclaration(attr *DTD.Attribute, s string) bool {

	if isRequired(s) {
		log.Trace("REQUIRED detected")
		attr.Required = isRequired(s)
		return true
	}
	if isImplied(s) {
		log.Trace("IMPLIED detected")
		attr.Implied = isImplied(s)
		return true
	}
	if isFixed(s) {
		log.Trace("FIXED detected")
		attr.Fixed = isFixed(s)
		return true
	}
	return false
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
	ret := sc.Data.Text() == "<"
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
	for sc.next() {
		if sc.Data.Text() == ";" {
			return s
		}
		s += sc.Data.Text()
	}
	return s
}

// isWhitespace Determine if a string is a whitespace
func (sc *DTDScanner) isWhitespace() bool {
	return sc.Data.Text() == " " || sc.Data.Text() == "\t" || sc.isEndOfLine()
}

// isEndOfLine Identitfy a carriage return
func (sc *DTDScanner) isEndOfLine() bool {

	if sc.Data.Text() == "\n" {
		return true
	}
	if sc.Data.Text() == "\r" {
		sc.CRLF = true
		return false
	}
	return false
}

// seekUntilNextBlock return string until next block is found
func (sc *DTDScanner) seekUntilNextBlock() string {

	var s string

	sc.CurrentLine = sc.LineCount
	s += sc.Data.Text()

	for sc.next() {

		if sc.IsStartChar() && sc.init {
			return s
		}

		if sc.IsStartChar() && !sc.init {
			sc.init = true
		}

		s += sc.Data.Text()
		log.Tracef("seekUntilNextBlock: Character '%s', Word is '%s'", sc.Data.Text(), s)

	}

	return s

}

// seekType Seek the type of the next DTD block
func (sc *DTDScanner) getBlockTypeCode(DTDType string) int {

	if DTDType == "ENTITY" {
		return DTD.ENTITY
	}

	if DTDType == "!ELEMENT " {
		return DTD.ELEMENT
	}

	if DTDType == "!ATTLIST " {
		return DTD.ATTLIST
	}

	if DTDType == "!NOTATION " {
		return DTD.NOTATION
	}
	return 0
}

// SeekBlock seek an entity
func (sc *DTDScanner) SeekBlock() string {
	var s string

	for sc.next() {

		c := sc.Data.Text()

		if c == "\r" || c == "\n" {
			log.Trace("Character '\\n' (new line)")
		} else {
			log.Tracef("Character '%s'", c)
		}

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

	for sc.next() {
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
