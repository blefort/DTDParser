// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package scanner allows to extract information from the DTD and create corresponding DTD structs
package formatter

import (
	"io"
	"os"

	"github.com/blefort/DTDParser/DTD"
	"go.uber.org/zap"
)

// IDTDBlock Interface for DTD block
type DTDFormatter struct {
	delimitter string
	log        *zap.SugaredLogger
}

func NewDTDFormatter(log *zap.SugaredLogger) *DTDFormatter {
	var f DTDFormatter
	f.delimitter = "\t"
	f.log = log
	return &f
}

// Render Render DTD blocks
func (ft *DTDFormatter) Render(collection *[]DTD.IDTDBlock, path string) {

	// export every blocks
	for _, block := range *collection {
		//p.Log.Debugf("Exporting block: %#v ", block)
		switch block.(type) {
		case *DTD.Attlist:
			ft.writeToFile(path, ft.RenderAttlist(block))
		case *DTD.Element:
			ft.writeToFile(path, ft.RenderElement(block))
		case *DTD.Comment:
			ft.writeToFile(path, ft.RenderComment(block))
		case *DTD.Entity:
			ft.writeToFile(path, ft.RenderEntity(block))
		case *DTD.Notation:
			ft.writeToFile(path, ft.RenderNotation(block))
		default:
			panic("unidentified block")
		}
		//c := block.Render() + "\n"
		ft.writeToFile(path, "\n\n")
	}

}

// RenderAttlist Render an ATTLIST
func (ft *DTDFormatter) RenderAttlist(b DTD.IDTDBlock) string {
	attributes := "\n"

	extra := b.GetExtra()

	for _, attr := range extra.Attributes {
		attributes += ft.RenderAttribute(attr)
	}

	return join("<!ATTLIST ", b.GetName(), " ", attributes, ">")
}

// RenderAttlist Render an Element
func (ft *DTDFormatter) RenderElement(b DTD.IDTDBlock) string {
	return join("<!ELEMENT ", b.GetName(), " ", b.GetValue(), ">")
}

// RenderComment render a comment
func (ft *DTDFormatter) RenderComment(b DTD.IDTDBlock) string {
	return "<!--" + b.GetValue() + "-->"
}

// RenderEntity render an entity
func (ft *DTDFormatter) RenderEntity(b DTD.IDTDBlock) string {
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
		url = renderQuoted(extra.Url)
	}

	return join("<!ENTITY", m, b.GetName(), " ", eType, "\"\n", ft.delimitter, b.GetValue(), "\n\"", url, ">", exportedStr)
}

// RenderComment render a comment
func (ft *DTDFormatter) RenderNotation(b DTD.IDTDBlock) string {
	return b.Render()
}

// RenderAttribute Render an attribute
func (ft *DTDFormatter) RenderAttribute(a DTD.Attribute) string {
	s := ft.delimitter

	if a.Name != "" {
		s += a.Name + ft.delimitter
	}

	s += ft.AttributeType(a.Type) + " "

	if a.Fixed {
		s += " #FIXED "
	}

	if a.Value != "" && !a.IsEntity && a.Fixed {
		s += renderQuoted(a.Value)
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
func (ft *DTDFormatter) writeToFile(filepath string, s string) error {
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
func (ft *DTDFormatter) AttributeType(a int) string {
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
	ft.log.Debugf("DTD formatter has not definition for attribute type '%s', this might be ok, an entity would not have any", a)
	return ""
}
