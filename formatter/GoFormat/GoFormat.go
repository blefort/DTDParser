// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package scanner allows to extract information from the DTD and create corresponding DTD structs
package GoFormat

import (
	"io"
	"os"
	"strings"

	"github.com/blefort/DTDParser/DTD"
	"github.com/blefort/DTDParser/formatter"
	DTDParser "github.com/blefort/DTDParser/parser"
	"go.uber.org/zap"
)

// IDTDBlock Interface for DTD block
type GoFormat struct {
	delimitter  string
	log         *zap.SugaredLogger
	packageName string
}

// NewGoFormat instantiate new GoFormat struct
func New(log *zap.SugaredLogger) formatter.FormatterInterface {
	var f GoFormat
	f.delimitter = "\t"
	f.log = log
	return &f
}

func (d *GoFormat) ValidateOptions(f *formatter.Formatter) bool {
	return true
}

// RenderGoStructs Render a collection to a or a file containing go structs
func (d *GoFormat) Render(f *formatter.Formatter, parentDir string) {

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

func (d *GoFormat) RenderCollection(parser *DTDParser.Parser, path string) {

	// export every blocks
	for _, block := range parser.Collection {
		//p.Log.Debugf("Exporting block: %#v ", block)
		switch block.(type) {

		case *DTD.Element:
			d.writeToFile(path, d.renderStruct(&parser.Collection, block))
		default:
			continue
		}
		d.writeToFile(path, "\n\n")
	}
}

// RenderAttlist Render an Element
func (ft *GoFormat) renderStruct(collection *[]DTD.IDTDBlock, b DTD.IDTDBlock) string {
	return join("type ", strings.Title(strings.ToLower(b.GetName())), " struct {", ft.renderStructContent(collection, b), "}")
}

// RenderAttlist Render an Element
func (ft *GoFormat) renderStructContent(collection *[]DTD.IDTDBlock, b DTD.IDTDBlock) string {
	content := ft.renderXMLName(b)
	content += ft.renderBlockElements(collection, b)
	return content
}

func (ft *GoFormat) renderXMLName(b DTD.IDTDBlock) string {
	return join("\nXMLName xml.Name `xml:\"", b.GetName(), "\"`\n")
}

func (ft *GoFormat) renderBlockElements(collection *[]DTD.IDTDBlock, b DTD.IDTDBlock) string {
	elements := ft.parseElementValue(b)
	content := ""
	for _, el := range *elements {
		content += "\n" + el + "\n"
	}
	return content
}

func (ft *GoFormat) parseElementValue(b DTD.IDTDBlock) *[]string {
	var s []string
	return &s
}

// writeToFile write to a DTD file
func (d *GoFormat) writeToFile(filepath string, s string) error {
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

// Helper to join strings
func join(strs ...string) string {
	var sb strings.Builder
	for _, str := range strs {
		sb.WriteString(str)
	}
	return sb.String()
}
