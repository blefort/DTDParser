// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package DTDParser A DTD parser
package DTDParser

//
// Some Ideas taken from Ben Johnson
// https://bp.Log.gopheracademy.com/advent-2014/parsers-lexers/
//
import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"github.com/blefort/DTDParser/DTD"
	"github.com/blefort/DTDParser/scanner"
)

// Parser is a structure that represents the parser
// it can manage multiple DTD parsers
type Parser struct {
	WithComments      bool
	IgnoreExtRefIssue bool
	Filepath          string
	Collection        []DTD.IDTDBlock
	parsers           []Parser
	filepaths         *[]string
	outputDirPath     string
	outputStructPath  string
	Overwrite         bool
	Log               *zap.SugaredLogger
}

// NewDTDParser returns a new DTD parser
func NewDTDParser(Log *zap.SugaredLogger) *Parser {
	p := Parser{}
	p.Log = Log
	return &p
}

// SetOutputPath set the output path of the DTD
// to export the DTD
func (p *Parser) SetOutputPath(s string) {
	p.outputDirPath = s
}

// SetStructPath set the output path to export the DTD a go Struct
func (p *Parser) SetStructPath(s string) {
	p.outputStructPath = s
}

// createOutputFile Create a DTD output file
func (p *Parser) createOutputFile(filepath string, overwrite bool) {

	exists := p.fileExists(filepath)

	if exists && overwrite {
		p.removeFile(filepath)
	} else if exists {
		p.Log.Debugf("createOutputFile '%s' exists and can't be overwritten", filepath)
		return
	}

	p.Log.Debugf("createOutputFile '%s', truncate will be '%t'", filepath, overwrite)

	f, err := os.Create(filepath)

	if err != nil {
		panic(err)
	}

	f.Close()
}

// fileExists Test if a file exists
func (p *Parser) fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else {
		return false
	}
}

// removeFile Empty content of a file
func (p *Parser) removeFile(filepath string) {
	p.Log.Debugf("Remove '%s'", filepath)
	err := os.Remove(filepath)
	if err != nil {
		p.Log.Fatal(err)
	}
}

// Parse Parse a DTD using its path
func (p *Parser) Parse(filePath string) {
	var filespaths []string

	if p.filepaths == nil {

		p.filepaths = &filespaths
		p.Log.Debugf("Parser filepaths was nil")
	}
	p.Filepath = filePath

	// Open file

	filebuffer, err := ioutil.ReadFile(p.Filepath)

	if err != nil {
		p.Log.Fatal(err)
	}

	// Fileinfo
	stat, err := os.Stat(filePath)
	if err != nil {

		p.Log.Fatal(err)
	}

	bytes := stat.Size()
	p.Log.Debugf("Parsing '%s', %d bytes", p.Filepath, bytes)

	*p.filepaths = append(*p.filepaths, p.Filepath)

	// use  bufio to read file rune by rune
	inputdata := string(filebuffer)

	//p.Log.Debugf("File content is: %s", inputdata)
	scanner := scanner.NewScanner(filePath, inputdata, p.Log)

	// not sure if this is correct methodology
	// I tried to separate the DTD Scanner from the parser
	// the scanner should send DTD blocks that the parser
	// will put in a collection.
	for scanner.NextBlock() {

		DTDBlock, err := scanner.Scan()

		if err != nil {
			p.Log.Debugf("%v", err)
			continue
		}

		if DTD.IsExportedEntityType(DTDBlock) {
			p.SetExportEntity(DTDBlock.GetName())
		} else {
			p.Collection = append(p.Collection, DTDBlock)
		}

		if DTD.IsEntityType(DTDBlock) {
			p.parseExternalEntity(DTDBlock.(*DTD.Entity))
		}

	}
	p.Log.Infof("%d blocks found in DTD '%s'", len(p.Collection), p.Filepath)
}

// parseExternalEntity Parse an external DTD reference declared in an entity
func (p *Parser) parseExternalEntity(e *DTD.Entity) {

	p.Log.Debugf("Check entity '%s' for external reference", e.Name)

	if !e.IsExternal {
		p.Log.Debugf("No external DTD in entity %s", e.Name)
		return
	}

	base := filepath.Base(p.Filepath)
	parentDir := filepath.Dir(p.Filepath)
	path := parentDir + "/" + e.Url

	errMsg := "External DTD '" + e.Url + "' not found, declared in '" + base + "', entity '" + e.Name

	if _, err := os.Stat(path); os.IsNotExist(err) && !p.IgnoreExtRefIssue {
		p.Log.Fatal(errMsg)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) && p.IgnoreExtRefIssue {
		p.Log.Warnf(errMsg)
		return
	}

	p.Log.Warnf("*** New parser *** for external entity %s", path)

	extP := NewDTDParser(p.Log)
	extP.outputDirPath = p.outputDirPath
	extP.filepaths = p.filepaths
	extP.WithComments = p.WithComments
	extP.IgnoreExtRefIssue = p.IgnoreExtRefIssue
	extP.Overwrite = p.Overwrite
	extP.Parse(path)

	p.Log.Warnf("*** /end of New parser %s", path)
	p.parsers = append(p.parsers, *extP)

}

// SetExportEntity Mark an entity block are exported in the collection
func (p *Parser) SetExportEntity(name string) {
	for _, block := range p.Collection {
		if block.GetName() == name {
			p.Log.Debugf("Marking %s as exported", name)
			block.SetExported(true)
			return
		}
	}
}

// RenderDTD Render a collection to a or a set of DTD files
func (p *Parser) RenderDTD(parentDir string) {

	// we process here all the file path of all DTD parsed
	// and determine the parent directory
	// the parentDir will be happened to the output dir
	//p.Log.Debugf("parent Dir is: %s, Filepaths are %+v", parentDir, *p.filepaths)

	if parentDir == "" {
		parentDir = filepath.Dir(commonPrefix(*p.filepaths))
		p.Log.Debugf("ParentDir from filepaths is: %s", parentDir)
	}

	finalPath := p.determineFinalDTDPath(parentDir, p.Filepath)

	if _, err := os.Stat(finalPath); os.IsNotExist(err) {
		p.Log.Infof("Create DTD: '%s'", finalPath)
		p.createOutputFile(finalPath, false)
	} else if p.Overwrite {
		p.Log.Infof("Overwrite DTD: '%s'", finalPath)
		p.createOutputFile(finalPath, true)
	} else {
		p.Log.Fatalf("Output DTD: '%s' already exists, please remove it before or use flag -overwrite", finalPath)
	}

	p.Log.Warnf("Render DTD '%s', %d blocks, %d nested parsers", finalPath, len(p.Collection), len(p.parsers))

	// export every blocks
	for _, block := range p.Collection {
		p.Log.Debugf("Exporting block: %#v ", block)
		c := block.Render() + "\n"
		writeToFile(finalPath, c)
	}

	// process children parsers
	for idx, parser := range p.parsers {
		p.Log.Warnf("Render DTD's child '%d/%d'", idx+1, len(p.parsers))
		parser.RenderDTD(parentDir)
	}
}

// RenderGoStructs Render a collection to a or a file containing go structs
func (p *Parser) RenderGoStructs() {

	finalPath := p.outputStructPath + "/structs.go"

	if _, err := os.Stat(finalPath); os.IsNotExist(err) {
		p.Log.Infof("Create Go struct: '%s'", finalPath)
		p.createOutputFile(finalPath, false)
	} else if p.Overwrite {
		p.Log.Infof("Overwrite: '%s'", finalPath)
		p.createOutputFile(finalPath, true)
	} else {
		p.Log.Fatalf("Output Go struct: '%s' already exists, please remove it before or use flag -overwrite", finalPath)
	}

	p.Log.Warnf("Render DTD '%s', %d blocks, %d nested parsers", finalPath, len(p.Collection), len(p.parsers))

	// export every blocks
	// for _, block := range p.Collection {
	// 	p.Log.Debugf("Exporting block: %#v ", block)
	// 	c := block.Render() + "\n"
	// 	writeToFile(finalPath, c)
	// }

	// // process children parsers
	// for _, parser := range p.parsers {
	// 	parser.RenderDTD(parentDir)
	// }
}

func (p *Parser) determineFinalDTDPath(parentDir string, i string) string {

	p.Log.Debugf("determineFinalDTDPath: source is: '%s'", i)

	// determine relative directory
	pDir := filepath.Dir(i)

	// relative dir remove parent directory from the path
	rDir := strings.TrimPrefix(pDir, parentDir)
	p.Log.Debugf("rDir %s", rDir)

	oDir := p.outputDirPath

	// absolute from output dir
	if rDir != "" {
		oDir = p.outputDirPath + "/" + rDir
	}

	if _, err := os.Stat(oDir); os.IsNotExist(err) {
		p.Log.Debugf("Create: %s", oDir)
		os.MkdirAll(oDir, 0770)
	}

	finalPath := oDir + "/" + filepath.Base(i)
	p.Log.Debugf("finalPath %s", finalPath)

	return finalPath
}

// writeToFile write to a DTD file
func writeToFile(filepath string, s string) error {
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

// commonPrefix Given a slice of path (String) find the common shared directory path
// from https://www.rosettacode.org/wiki/Find_common_directory_path#Go
func commonPrefix(paths []string) string {
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
