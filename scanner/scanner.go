// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// package scanner allows to extract information from the DTD and create corresponding DTD structs
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
// Use Cases:
//
// <!ENTITY  name value>
// <!ENTITY % name PUBLIC REFERENCE URL>
// <!ENTITY % name SYSTEM URL">
// <!ENTITY % name PUBLIC REFERENCE URL>
//
func ParseEntity(s string) *DTD.Entity {
	var e DTD.Entity
	var s2 string

	s2 = normalizeSpace(s)

	regex := regexp.MustCompile(`\s?(%)\s([^\s]+)|^([^\s]+)|(PUBLIC|SYSTEM)|"([^"]+)"`)
	parts := regex.FindAllString(s2, -1)

	// determine if ExternalDTD
	nameParts := strings.Split(parts[0], " ")
	if len(nameParts) == 2 {
		e.Parameter = true
		e.Name = nameParts[1]
	} else {
		e.Name = parts[0]
	}

	for _, part := range parts {
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

// ParseAttlist Parse a string and return pointer to a DTD.Attlist
// Declaration Types
// #REQUIRED
// <!ATTLIST ElementName AttributeName AttributeType #REQUIRED>
// #IMPLIED
// <!ATTLIST ElementName AttributeName AttributeType #IMPLED>
//
// #FIXED
// <!ATTLIST ElementName AttributeName AttributeType #FIXED AttributeValue>
//
// Attribute type can't be determined whem parsing because some entities can be part of the content
//
func (sc *DTDScanner) ParseAttlist(s string) *DTD.Attlist {
	var a DTD.Attlist
	var s2 string

	s2 = normalizeSpace(s)

	regex := regexp.MustCompile(`([%?\w\-]+)`)
	parts := regex.FindAllString(s2, -1)

	// set the name
	a.Name = parts[0]
	a.Value = s2[len(a.Name) : len(s2)-1]

	return &a
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
	return space.ReplaceAllString(s1, " ")
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
