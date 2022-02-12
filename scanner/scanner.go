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

	"go.uber.org/zap"

	"github.com/blefort/DTDParser/DTD"
)

const (
	BLOCK_COMMENT = 0
	BLOCK_XML     = 1
	BLOCK_DTD     = 2
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
	Log          *zap.SugaredLogger
}

type parsedBlock struct {
	fullString string
	blockType  string
	name       string
	entity     string
	value      string
}

// NewScanner returns a new DTD Scanner
func NewScanner(path string, s string, log *zap.SugaredLogger) *DTDScanner {
	var scanner DTDScanner
	scanner.Data = bufio.NewScanner(strings.NewReader(s))
	scanner.Data.Split(bufio.ScanRunes)
	scanner.Filepath = path
	scanner.CharCount = 0
	scanner.LineCount = 1
	scanner.CRLF = false // assume Linux
	scanner.init = false
	scanner.scanResult = true
	scanner.Log = log
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
		sc.Log.Debugf("Line is %d", sc.LineCount)
	}
	return sc.scanResult
}

// Next Move to the next character
func (sc *DTDScanner) Previous() bool {
	sc.CharCount--
	ret := sc.Data.Scan()
	if sc.isEndOfLine() {
		sc.LineCount++
		sc.Log.Debugf("Line is %d", sc.LineCount)
	}
	return ret
}

// Scan the string to find the next block
func (sc *DTDScanner) Scan() (DTD.IDTDBlock, error) {

	sc.Log.Debug("Seeking for next block")
	s, declaration := sc.seekUntilNextBlock()

	sc.Log.Debugf("Declaration is %d", declaration)

	p, err := sc.extractDeclaration(s, declaration)

	if err != nil {
		return nil, errors.New("Unidentified block")
	}

	if p.blockType == "XMLDECL" {
		var xmldecl DTD.XMLDecl
		sc.Log.Debug("XMLDECL found at line '%d")
		return &xmldecl, nil
	}

	if p.blockType == "COMMENT" {
		comment := sc.ParseComment(p)
		sc.Log.Infof("Commment found at line '%d'", sc.CurrentLine)
		return comment, nil
	}

	if p.blockType == "ENTITY" {
		entity := sc.ParseEntity(p)
		sc.Log.Infof("ENTITY '%s' (line %d)", entity.GetName(), sc.CurrentLine)
		sc.logOutputAttributes(&entity.Attributes)
		return entity, nil
	}

	if p.blockType == "ATTLIST" {
		attlist := sc.ParseAttlist(p)
		sc.Log.Infof("ATTLIST '%s' (line %d)", attlist.GetName(), sc.CurrentLine)
		sc.logOutputAttributes(&attlist.Attributes)
		return attlist, nil
	}

	if p.blockType == "ELEMENT" {
		element := sc.ParseElement(p)
		sc.Log.Infof("ELEMENT '%s' (line %d)", element.GetName(), sc.CurrentLine)
		return element, nil
	}

	if p.blockType == "NOTATION" {
		notation := sc.ParseNotation(p)
		sc.Log.Infof("NOTATION '%s' (line %d)", notation.GetName(), sc.CurrentLine)
		return notation, nil
	}

	return nil, errors.New("Unidentified block")

}

// ParseComment Parse a string and return pointer to DTD.Comment
func (sc *DTDScanner) ParseComment(p *parsedBlock) *DTD.Comment {
	var c DTD.Comment
	//s = strings.TrimRight(s, "-")
	sc.Log.Debugf("comment stre receive '%s' and '%s'", p.name, p.value)
	c.Value = p.value
	return &c
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

	parts := sc.SeekWords(p.value)
	//l := len(parts)
	sc.Log.Debugf("%v", parts)

	n.Name = sc.normalizeSpace(p.name)

	if isPublic(parts[0]) {
		n.Public = true
		n.ID = parts[1]
		if len(parts) >= 3 {
			n.Value = parts[2]
		}

	}
	if isSystem(parts[0]) {
		n.System = true
		n.Value = parts[1]
		n.Url = parts[1]
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
func (sc *DTDScanner) ParseEntity(p *parsedBlock) *DTD.Entity {
	var e DTD.Entity
	e.Name = p.name

	// parameter
	if p.entity == "%" {
		e.Parameter = true
	}

	parts := sc.SeekWords(p.value)
	l := len(parts)

	sc.Log.Debugf("parts %v", parts)

	sc.Log.Debugf("l: %d", l)

	if l == 1 {
		e.Value = sc.normalizeSpace(parts[0])
		e.IsInternal = true
		e.IsExternal = false
		sc.Log.Debugf("value is %s", e.Value)
	}

	if l > 1 {
		if isPublic(parts[0]) {
			e.Public = true
		} else {
			e.Public = false
		}

		if isSystem(parts[0]) {
			e.System = true
		} else {
			e.System = false
		}
	}

	if e.System {
		e.Url = sc.normalizeSpace(parts[1])
	}

	if e.Public {
		e.Value = sc.normalizeSpace(parts[1])
		e.Url = sc.normalizeSpace(parts[2])
	}

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
	s2 := sc.normalizeSpace(s)
	parts := sc.SeekWords(s2)
	l := len(parts)

	sc.Log.Debugf("parse Attr: parts %v, s was '%s'", parts, s)

	// for idx, part := range parts {
	// 	sc.Log.Debugf("part %d: %s", idx, part)
	// }

	if l == 0 {
		panic("Unable to scan Attlist")
	}

	nextWord := func(i *int, l int) bool {
		if *i+1 < l {
			*i++
			sc.Log.Debugf("Next word, i: %d, %s ", *i, parts[*i])
			return true
		}
		return false
	}

	for nextWord(&i, l) {

		var attr DTD.Attribute

		// empty string
		if sc.normalizeSpace(parts[i]) == "" {
			continue
		}

		// instruction varies depending the type, we don't know in advance

		// reference to an entity
		if strings.HasPrefix(parts[i], "%") {
			attr.Value = parts[i]
			sc.Log.Debugf("- Ref. to an entity found: %s", attr.Render())
			attr.IsEntity = true
			*attributes = append(*attributes, attr)
			continue
		}

		// first word is always the attribute name
		attr.Name = sc.normalizeSpace(parts[i])
		sc.Log.Debugf("Processing attribute: '%s'", attr.Name)

		if !nextWord(&i, l) {
			sc.Log.Fatalf("Not enough arguments to loop through attributes i:%d", i)

		}

		// CASE 2
		// The others parts are processed by group of 3 to 4 depending the type

		// Type is always in the second position
		attr.Type = DTD.SeekAttributeType(parts[i])

		sc.Log.Debugf("attribute type is %d", attr.Type)

		if attr.Type == DTD.CDATA { // 20
			nextWord(&i, l)
			sc.checkNextTwoArguments(parts, &i, &attr)
		}

		if attr.Type == DTD.TOKEN_ID ||
			attr.Type == DTD.TOKEN_IDREF ||
			attr.Type == DTD.TOKEN_IDREFS ||
			attr.Type == DTD.TOKEN_ENTITY ||
			attr.Type == DTD.TOKEN_ENTITIES ||
			attr.Type == DTD.TOKEN_NMTOKEN ||
			attr.Type == DTD.TOKEN_NMTOKENS {
			nextWord(&i, l)
			sc.checkNextTwoArguments(parts, &i, &attr)
		}
		if attr.Type == DTD.ENUM_NOTATION {
			nextWord(&i, l)
			attr.Value = parts[i]
			nextWord(&i, l)
			sc.checkDefaultDeclaration(&attr, parts[i])
		}
		if attr.Type == DTD.ENUM_ENUM {
			attr.Value = parts[i]
			nextWord(&i, l)
			sc.checkDefaultDeclaration(&attr, parts[i])
		}

		*attributes = append(*attributes, attr)
		sc.Log.Debugf("*Attr rendered: %s", attr.Render())

		if attr.Type == 0 {
			sc.Log.Fatalf("Could not identify attribute type at line %d, name: '%s', value: '%s'", sc.LineCount, attr.Name, parts[i])
		}

	}
}

func (sc *DTDScanner) checkNextTwoArguments(parts []string, i *int, attr *DTD.Attribute) {

	var s2 string
	s1 := parts[*i]

	if *i+1 < len(parts) {
		s2 = parts[*i+1]
	} else {
		s2 = ""
	}

	sc.Log.Debugf("checkNextTwoArguments is (i: %d) testing 2 args %s and %s", *i, s1, s2)

	if isQuoted(s1[0:1]) {
		*&attr.Value = s1
		sc.Log.Debugf("value is %s", s1)
		if sc.checkDefaultDeclaration(attr, s2) {
			*i++
		}
	}

	if sc.checkDefaultDeclaration(attr, s1) {
		sc.Log.Debugf("check if s2 is quoted: '%s'", s2)
		if s2 != "" && isQuoted(s2[0:1]) {
			sc.Log.Debugf("s2 is quoted: '%s'", s2)
			sc.Log.Debugf("value is %s", s2)
			*i++
			*&attr.Value = s2

		}
	}
	sc.Log.Debugf("checkNextTwoArguments i: %d", *i)
}

// output attributes in the log
func (sc *DTDScanner) logOutputAttributes(attributes *[]DTD.Attribute) {
	for _, attr := range *attributes {
		sc.Log.Debugf("- Attribute: %s", attr.Render())
	}
}

// assignIfEntityValue test if v is not public system or empty before assigning it
func assignIfEntityValue(e *DTD.Entity, v string) {
	if v != "" && !isPublic(v) && !isSystem(v) {
		e.Value = v
	}
}

// SeekWords Walk a string and identify every words
func (sc *DTDScanner) SeekWords(s string) []string {

	r2 := `"([^"]+)"|\((.*)\)[\+|\?|\*]?|([^\s]+)`

	regex := regexp.MustCompile(r2)
	parts := regex.FindAllString(s, -1)

	sc.Log.Debugf("seekWords FindAllString found %#v", parts)
	return parts
}

// extractDeclaration call the rigt sekk method depending the declaration
func (sc *DTDScanner) extractDeclaration(s string, declaration int) (*parsedBlock, error) {

	if declaration == BLOCK_XML {
		return sc.SeekXMLParts(s)
	}

	//sc.Log.Debugf("Block line: '%s'", s)
	if declaration == BLOCK_COMMENT {
		return sc.SeekCommentParts(s)
	}

	if sc.normalizeSpace(s) == "" {
		return nil, errors.New("End of DTD")
	}

	return sc.SeekBlockParts(s)
}

//SeekXMLParts extract Comment information from string using a Regex
func (sc *DTDScanner) SeekXMLParts(s string) (*parsedBlock, error) {
	var p parsedBlock
	p.blockType = "XMLDECL"
	p.value = s
	return &p, nil
}

// SeekCommentParts extract Comment information from string using a Regex
func (sc *DTDScanner) SeekCommentParts(s string) (*parsedBlock, error) {

	var p parsedBlock
	var r string

	r = `<!--([\s\S\n]*?)-->`
	regex := regexp.MustCompile(r)
	parts := regex.FindAllStringSubmatch(s, -1)

	p.value = strings.TrimSpace(parts[0][1])
	p.blockType = "COMMENT"

	sc.Log.Debugf("SeekCommentPart, parsed: , name: [%s], type: [%s], entity: [%s], value: [%s], s was [%s]", p.name, p.blockType, p.entity, p.value, s)

	return &p, nil
}

// SeekBlockParts extract DTD information from string using a Regex
func (sc *DTDScanner) SeekBlockParts(s string) (*parsedBlock, error) {

	var p parsedBlock
	var r string

	sc.Log.Debugf("SeekBlockParts received %s", s)

	r = `<\!(ENTITY|ELEMENT|ATTLIST|COMMENT|NOTATION)\s*(\%)?\s*(\S+)?\s*([^>]+)?>\s*(%[^>\s]+)?`

	regex := regexp.MustCompile(r)
	parts := regex.FindAllStringSubmatch(s, -1)

	p.fullString = parts[0][0]
	p.blockType = parts[0][1]
	p.entity = parts[0][2]
	p.name = parts[0][3]
	p.value = parts[0][4]

	sc.Log.Debugf("SeekBlockParts, parsed: \n-name: %s\n-type:%s\n-entity:%s\n-value:%s, s was '%s'", p.name, p.blockType, p.entity, p.value, s)

	return &p, nil
}

// isQuoted returns true if a character is quote or a double quote
func isQuoted(s string) bool {
	return s == "\"" || s == "'"
}

// checkEnumDefaultValue
func (sc *DTDScanner) checkEnumDefaultValue(attr *DTD.Attribute, parts []string) {

	attr.Default = sc.checkDefaultValue(parts[0])
	sc.Log.Debugf("Enum Default Value: '%s'", attr.Default)
}

// checkCDATADefaultValue
func (sc *DTDScanner) checkDefaultValue(v string) string {

	sc.Log.Debugf("testing default value'%s'", v)

	if isQuoted(v[0:1]) {
		sc.Log.Debugf("default value is '%s'", v)
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

	sc.Log.Debugf("Default value '%s' found", v)
	return v
}

// checkDefault Check if the default value if required, implied or fixed and reste Default property
func (sc *DTDScanner) checkDefaultDeclaration(attr *DTD.Attribute, s string) bool {

	if isRequired(s) {
		sc.Log.Debug("REQUIRED detected")
		attr.Required = isRequired(s)
		return true
	}
	if isImplied(s) {
		sc.Log.Debug("IMPLIED detected")
		attr.Implied = isImplied(s)
		return true
	}
	if isFixed(s) {
		sc.Log.Debug("FIXED detected")
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

// IsStartChar Determine if a character is the beginning of a DTD block
func (sc *DTDScanner) IsStartChar() bool {
	ret := sc.Data.Text() == "<"
	sc.Log.Debugf("IsStartChar: %t", ret)
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
func (sc *DTDScanner) seekUntilNextBlock() (string, int) {

	var s string
	isComment := false
	isXmlDecl := false

	sc.CurrentLine = sc.LineCount
	s += sc.Data.Text()

	for sc.next() {

		empty := sc.normalizeSpace(s)

		if sc.IsStartChar() && sc.init && !isComment && !isXmlDecl && empty != " " {
			sc.Log.Debugf("DTD block '%s'", s)
			return s, BLOCK_DTD
		}

		if isComment && s[len(s)-3:] == "-->" {
			sc.Log.Debugf("Comment block %s", s)
			sc.init = false
			return s, BLOCK_COMMENT
		}

		if isXmlDecl && s[len(s)-2:] == "?>" {
			sc.Log.Debugf("XML block %s", s)
			sc.init = false
			return s, BLOCK_XML
		}

		if sc.IsStartChar() && !sc.init {
			sc.init = true
		}

		s += sc.Data.Text()

		if sc.normalizeSpace(s) == "<!--" {
			sc.Log.Debug("Comment is detected\n")
			isComment = true
		}

		if sc.normalizeSpace(s) == "<?xml" {
			sc.Log.Debug("XML is detected\n")
			isXmlDecl = true
		}

		sc.Log.Debugf("seekUntilNextBlock: Character '%s', Word is '%s'", sc.Data.Text(), s)

	}

	if isComment {
		return s, BLOCK_COMMENT
	}

	if isXmlDecl {
		return s, BLOCK_XML
	}

	return s, BLOCK_DTD

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
			sc.Log.Debug("Character '\\n' (new line)")
		} else {
			sc.Log.Debugf("Character '%s'", c)
		}

		if sc.Data.Text() == ">" {
			break
		}

		s += sc.Data.Text()
	}
	return s
}

// normalizeSpace Convert Line breaks, multiple space into a single space
func (sc *DTDScanner) normalizeSpace(s string) string {
	regexLineBreak := regexp.MustCompile(`(?s)(\r?\n)|\t`)
	s1 := regexLineBreak.ReplaceAllString(s, " ")
	space := regexp.MustCompile(`\s+|\t`)
	nm := strings.Trim(space.ReplaceAllString(s1, " "), " ")
	sc.Log.Debugf("normalizeSpace: string is '%s'", nm)
	return nm
}
