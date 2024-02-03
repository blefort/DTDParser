// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package scanner allows to extract information from the DTD and create corresponding DTD structs
package formatter

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	DTDParser "github.com/blefort/DTDParser/parser"
	"go.uber.org/zap"
)

// IDTDBlock Interface for DTD block
type FormatterInterface interface {
	Render(f *Formatter, parentDir string)
	ValidateOptions(f *Formatter) bool
}

// Formatter type
type Formatter struct {
	Formatter      FormatterInterface
	Parser         *DTDParser.Parser
	OutputPath     string
	OutputFilename string
	Overwrite      bool
	Options        []map[string]interface{}
	Log            *zap.SugaredLogger
}

// NewFormatter create an instance of a formatter and return its reference
func NewFormatter(p *DTDParser.Parser, format FormatterInterface, outpath string, outfile string, log *zap.SugaredLogger) *Formatter {
	f := new(Formatter)
	f.Formatter = format
	f.Parser = p
	f.OutputPath = outpath
	f.OutputFilename = outfile
	f.Log = log
	return f
}

// SetOverwrite ask formatter to overwrite output
func (f *Formatter) SetOverwrite(o bool) {
	f.Overwrite = o
}

// SetOptions Provide a json string and pass options to the formatter
func (f *Formatter) SetOptions(jsonString *string) {
	jsonB := []byte(*jsonString)
	if err := json.Unmarshal(jsonB, &f.Options); err != nil {
		f.Log.DPanic(err)
		os.Exit(1)
	}
}

// createOutputFile Create output
func (f *Formatter) CreateOutputFile(filepath string) {

	file, err := os.Create(filepath)

	if err != nil {
		panic(err)
	}

	file.Close()
}

// Validate the data formatter options
func (f *Formatter) Validate() bool {

	if !f.Formatter.ValidateOptions(f) {
		f.Log.DPanic("Options are not valid")
		return false
	}

	exists := f.fileExists(f.OutputPath)

	if exists && !f.Overwrite {
		f.Log.DPanic("Output path alreay exists, you can use --overwrite option to force ")
		return false
	}

	return false
}

// RenderDTD Render a collection to a or a set of DTD files
func (f *Formatter) Render() {

	if f.Overwrite {
		err := os.RemoveAll(f.OutputPath)
		if err != nil {
			f.Log.DPanic(err)
		}
	}

	parentDir := ""

	if parentDir == "" {
		parentDir = filepath.Dir(f.CommonPrefix(*f.Parser.Filepaths))
		f.Log.Debugf("ParentDir from filepaths is: %s", parentDir)
	}

	f.Formatter.Render(f, parentDir)
}

// commonPrefix Given a slice of path (String) find the common shared directory path
// from https://www.rosettacode.org/wiki/Find_common_directory_path#Go
func (f *Formatter) CommonPrefix(paths []string) string {
	// Handle special cases.
	switch len(paths) {
	case 0:
		return ""
	case 1:
		return path.Clean(paths[0])
	}

	c := []byte(path.Clean(paths[0]))

	// Ignore the first path since it's already in c
	for _, v := range paths[1:] {
		// Clean up each path before testing it
		v = path.Clean(v)

		// Find the first non-common byte and truncate c
		if len(v) < len(c) {
			c = c[:len(v)]
		}
		for i := 0; i < len(c); i++ {
			if v[i] != c[i] {
				c = c[:i]
				break
			}
		}
	}

	return string(c)
}

// FinalDTDPath
func (f *Formatter) FinalDTDPath(parentDir string, i string) string {

	f.Log.Debugf("FinalDTDPath: source is: '%s'", i)

	// determine relative directory
	pDir := filepath.Dir(i)

	// relative dir remove parent directory from the path
	rDir := strings.TrimPrefix(pDir, parentDir)
	f.Log.Debugf("rDir %s", rDir)

	oDir := f.OutputPath

	// absolute from output dir
	if rDir != "" {
		oDir = f.OutputPath + "/" + rDir
	}

	if _, err := os.Stat(oDir); os.IsNotExist(err) {
		f.Log.Debugf("Create: %s", oDir)
		os.MkdirAll(oDir, 0770)
	}

	finalPath := oDir + "/" + filepath.Base(i)
	f.Log.Debugf("finalPath %s", finalPath)

	return finalPath
}

// fileExists Test if a file exists
func (f *Formatter) fileExists(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	}
	return false
}

// AvailaibleFormatters return list of formatter
func AvailaibleFormatters() []string {
	formatters := []string{"dtd", "go", "json"}
	return formatters
}

// Exists tells if a formatter is available
func Exists(formatter string) bool {
	return slices.Contains(AvailaibleFormatters(), formatter)
}
