// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package scanner allows to extract information from the DTD and create corresponding DTD structs
package DTDformat

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/blefort/DTDParser/DTD"
	"github.com/blefort/DTDParser/formatter"
	DTDParser "github.com/blefort/DTDParser/parser"
	"go.uber.org/zap"
)

// IDTDBlock Interface for DTD block
type DTDFormat struct {
	delimitter string
	log        *zap.SugaredLogger
}

func New(log *zap.SugaredLogger) formatter.FormatterInterface {
	var f DTDFormat
	f.delimitter = "\t"
	f.log = log
	return &f
}

// setDelimitter Set delimitter
func (d *DTDFormat) SetDelimitter(delimitter string) {
	d.delimitter = delimitter
}

// printQuoted print the value with double quote if not empty
func (d *DTDFormat) renderQuoted(s string) string {
	if s == "" {
		return ""
	}
	return fmt.Sprintf("\"%s\"", s)
}

func (d *DTDFormat) ValidateOptions(f *formatter.Formatter) bool {
	return true
}

// RenderDTD Render a collection to a or a set of DTD files
func (d *DTDFormat) Render(f *formatter.Formatter, parentDir string) {

	// we process here all the file path of all DTD parsed
	// and determine the parent directory
	// the parentDir will be happened to the output dir
	//p.Log.Debugf("parent Dir is: %s, Filepaths are %+v", parentDir, *p.filepaths)

	finalPath := f.FinalDTDPath(parentDir, f.Parser.Filepath)
	f.CreateOutputFile(finalPath)
	d.RenderCollection(f.Parser, finalPath)

	//p.Log.Warnf("Render DTD '%s', %d blocks, %d nested parsers", finalPath, len(p.Collection), len(p.parsers))

	// process children parsers
	l := len(f.Parser.Parsers)
	for idx, parser := range f.Parser.Parsers {
		d.log.Warnf("Render DTD's child '%d/%d'", idx+1, l)
		d.RenderCollection(&parser, parentDir)
	}

}

// Render Render DTD blocks
func (d *DTDFormat) RenderCollection(parser *DTDParser.Parser, path string) {

	// export every blocks
	for _, block := range parser.Collection {
		//p.Log.Debugf("Exporting block: %#v ", block)
		switch block.(type) {
		case *DTD.Attlist:
			d.writeToFile(path, d.RenderAttlist(block))
		case *DTD.Element:
			d.writeToFile(path, d.RenderElement(block))
		case *DTD.Comment:
			d.writeToFile(path, d.RenderComment(block))
		case *DTD.Entity:
			d.writeToFile(path, d.RenderEntity(block))
		case *DTD.Notation:
			d.writeToFile(path, d.RenderNotation(block))
		default:
			panic("unidentified block")
		}
		//c := block.Render() + "\n"
		d.writeToFile(path, "\n\n")
	}

}

// RenderAttlist Render an ATTLIST
func (d *DTDFormat) RenderAttlist(b DTD.IDTDBlock) string {
	attributes := "\n"

	extra := b.GetExtra()

	for _, attr := range extra.Attributes {
		attributes += d.RenderAttribute(attr)
	}

	return join("<!ATTLIST ", b.GetName(), " ", attributes, ">")
}

// RenderAttlist Render an Element
func (d *DTDFormat) RenderElement(b DTD.IDTDBlock) string {
	return join("<!ELEMENT ", b.GetName(), " ", b.GetValue(), ">")
}

// RenderComment render a comment
func (d *DTDFormat) RenderComment(b DTD.IDTDBlock) string {
	return "<!--" + b.GetValue() + "-->"
}

// RenderEntity render an entity
func (d *DTDFormat) RenderEntity(b DTD.IDTDBlock) string {
	var m string
	var eType string
	var exportedStr string
	var url string

	extra := b.GetExtra()

	if extra.IsParameter {
		m = " % "
	} else {
		m = " "
	}

	if extra.IsPublic {
		eType += " PUBLIC "
	}
	if extra.IsSystem {
		eType += " SYSTEM "
	}

	if extra.IsExported {
		exportedStr = join("\n%", b.GetName(), ";")
	}

	if extra.Url != "" {
		url = d.renderQuoted(extra.Url)
	}

	return join("<!ENTITY", m, b.GetName(), " ", eType, "\"\n", d.delimitter, b.GetValue(), "\n\"", url, ">", exportedStr)
}

// RenderComment render a comment
func (d *DTDFormat) RenderNotation(b DTD.IDTDBlock) string {
	return b.Render()
}

// RenderAttribute Render an attribute
func (d *DTDFormat) RenderAttribute(a DTD.Attribute) string {
	s := d.delimitter

	if a.Name != "" {
		s += a.Name + d.delimitter
	}

	s += d.AttributeType(a.Type) + " "

	if a.Fixed {
		s += " #FIXED "
	}

	if a.Value != "" && !a.IsEntity && a.Fixed {
		s += d.renderQuoted(a.Value)
	} else if a.Value != "" && !a.IsEntity && !a.Fixed {
		s += a.Value
	} else if a.Value != "" && a.IsEntity {
		s += a.Value
	}

	if a.Implied {
		s += " #IMPLIED "
	}

	if a.Required {
		s += " #REQUIRED "
	}

	s += "\n"

	return s
}

// writeToFile write to a DTD file
func (d *DTDFormat) writeToFile(filepath string, s string) error {
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY, 0700)

	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.WriteString(f, s)

	if err != nil {
		return err
	}

	return f.Sync()
}

// AttributeType convert DTD Attribute type (int) to its corresponding string value
func (d *DTDFormat) AttributeType(a int) string {
	switch a {
	case DTD.CDATA:
		return "CDATA"
	case DTD.TOKEN_ID:
		return "ID"
	case DTD.TOKEN_IDREF:
		return "IDREF"
	case DTD.TOKEN_IDREFS:
		return "IDREFS"
	case DTD.TOKEN_ENTITY:
		return "ENTITY"
	case DTD.TOKEN_ENTITIES:
		return "ENTITIES"
	case DTD.TOKEN_NMTOKEN:
		return "NMTOKEN"
	case DTD.TOKEN_NMTOKENS:
		return "NMTOKENS"
	case DTD.ENUM_NOTATION:
		return "NOTATION"
	case DTD.ENUM_ENUM:
		return ""
	}
	d.log.Debugf("DTD formatter has not definition for attribute type '%s', this might be ok, an entity would not have any", a)
	return ""
}

// Helper to join strings
func join(strs ...string) string {
	var sb strings.Builder
	for _, str := range strs {
		sb.WriteString(str)
	}
	return sb.String()
}
