// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package scanner allows to extract information from the DTD and create corresponding DTD structs
package formatter

import (
	"io"
	"os"
	"strings"

	"github.com/blefort/DTDParser/DTD"
	"go.uber.org/zap"
)

// IDTDBlock Interface for DTD block
type GoFormatter struct {
	delimitter  string
	log         *zap.SugaredLogger
	packageName string
}

// NewGoFormatter instantiate new GoFormatter struct
func NewGoFormatter(log *zap.SugaredLogger, packageName string) *GoFormatter {
	var f GoFormatter
	f.delimitter = "\t"
	f.log = log
	f.packageName = packageName
	return &f
}

// Render Render DTD blocks
func (ft *GoFormatter) Render(collection *[]DTD.IDTDBlock, path string) {

	ft.writeToFile(path, "package "+ft.packageName+"\n\n")

	// export every blocks
	for _, block := range *collection {
		//p.Log.Debugf("Exporting block: %#v ", block)
		switch block.(type) {

		case *DTD.Element:
			ft.writeToFile(path, ft.renderStruct(collection, block))
		default:
			continue
		}
		ft.writeToFile(path, "\n\n")
	}
}

// RenderAttlist Render an Element
func (ft *GoFormatter) renderStruct(collection *[]DTD.IDTDBlock, b DTD.IDTDBlock) string {
	return join("type ", strings.Title(strings.ToLower(b.GetName())), " struct {", ft.renderStructContent(collection, b), "}")
}

// RenderAttlist Render an Element
func (ft *GoFormatter) renderStructContent(collection *[]DTD.IDTDBlock, b DTD.IDTDBlock) string {
	content := ft.renderXMLName(b)
	content += ft.renderBlockElements(collection, b)
	return content
}

func (ft *GoFormatter) renderXMLName(b DTD.IDTDBlock) string {
	return join("\nXMLName xml.Name `xml:\"", b.GetName(), "\"`\n")
}

func (ft *GoFormatter) renderBlockElements(collection *[]DTD.IDTDBlock, b DTD.IDTDBlock) string {
	elements := ft.parseElementValue(b)
	content := ""
	for _, el := range *elements {
		content += "\n" + el + "\n"
	}
	return content
}

func (ft *GoFormatter) parseElementValue(b DTD.IDTDBlock) *[]string {
	var s []string
	return &s
}

// writeToFile write to a DTD file
func (ft *GoFormatter) writeToFile(filepath string, s string) error {
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
