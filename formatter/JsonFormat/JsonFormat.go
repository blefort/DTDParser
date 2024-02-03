// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package scanner allows to extract information from the DTD and create corresponding DTD structs
package JsonFormat

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/blefort/DTDParser/formatter"
	DTDParser "github.com/blefort/DTDParser/parser"
	"go.uber.org/zap"
)

// IDTDBlock Interface for DTD block
type JsonFormat struct {
	log *zap.SugaredLogger
}

func New(log *zap.SugaredLogger) formatter.FormatterInterface {
	var f JsonFormat
	f.log = log
	return &f
}

func (d *JsonFormat) ValidateOptions(f *formatter.Formatter) bool {
	return true
}

// RenderDTD Render a collection to a or a set of DTD files
func (d *JsonFormat) Render(f *formatter.Formatter, parentDir string) {

	// we process here all the file path of all DTD parsed
	// and determine the parent directory
	// the parentDir will be happened to the output dir
	//p.Log.Debugf("parent Dir is: %s, Filepaths are %+v", parentDir, *p.filepaths)

	finalPath := f.OutputPath + "/data.json"
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
func (d *JsonFormat) RenderCollection(parser *DTDParser.Parser, path string) {

	b, err := json.MarshalIndent(parser.Collection, "", "    ")
	if err != nil {
		fmt.Println(err)
		return
	}

	d.writeToFile(path, string(b))

}

// writeToFile write to a DTD file
func (d *JsonFormat) writeToFile(filepath string, s string) error {
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