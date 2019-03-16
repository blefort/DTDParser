// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package DTDParser A DTD parser
package DTDParser

//
// Some Ideas taken from Ben Johnson
// https://blog.gopheracademy.com/advent-2014/parsers-lexers/
//
import (
	"encoding/xml"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"

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
	Overwrite         bool
}

// NewDTDParser returns a new DTD parser
func NewDTDParser() *Parser {
	p := Parser{}
	return &p
}

// SetOutputPath set the output path of the DTD
// to export the DTD
func (p *Parser) SetOutputPath(s string) {
	p.outputDirPath = s
}

// createOutputFile Create a DTD output file
func createOutputFile(filepath string, overwrite bool) {

	log.Infof("createOutputFile '%s', truncate will be '%t'", filepath, overwrite)

	if overwrite {
		removeFile(filepath)
	}

	f, err := os.Create(filepath)

	if err != nil {
		panic(err)
	}

	f.Close()
}

// removeFile Empty content of a file
func removeFile(filepath string) {
	log.Infof("Remove '%s'", filepath)
	err := os.Remove(filepath)
	if err != nil {
		log.Fatal(err)
	}
}

// Parse Parse a DTD using its path
func (p *Parser) Parse(filePath string) {

	if p.filepaths == nil {
		var filespaths []string
		p.filepaths = &filespaths
		log.Tracef("Parser filepaths was nil")
	}

	p.Filepath = filePath

	// Open file
	filebuffer, err := ioutil.ReadFile(p.Filepath)

	if err != nil {
		log.Fatal(err)
	}

	// Fileinfo
	stat, err := os.Stat(filePath)

	if err != nil {
		log.Fatal(err)
	}

	bytes := stat.Size()
	log.Infof("Parsing %s, %d bytes", p.Filepath, bytes/1024)

	*p.filepaths = append(*p.filepaths, p.Filepath)

	// use  bufio to read file rune by rune
	inputdata := string(filebuffer)

	log.Tracef("File content is: %s", inputdata)

	scanner := scanner.NewScanner(filePath, inputdata)

	// not sure if this is correct methodology
	// I tried to separate the DTD Scanner from the parser
	// the scanner should send DTD blocks that the parser
	// will put in a collection.

	for scanner.Next() {

		DTDBlock, err := scanner.Scan()

		if err != nil {
			log.Tracef("%v", err)
			continue
		}

		if DTD.IsExportedEntityType(DTDBlock) {
			p.SetExportEntity(DTDBlock.GetName())
		} else {
			p.Collection = append(p.Collection, DTDBlock)
		}
	}
	log.Infof("%d blocks found in this DTD", len(p.Collection))
}

// parseExternalEntity Parse an external DTD reference declared in an entity
func (p *Parser) parseExternalEntity(e *DTD.Entity) {

	if !e.ExternalDTD {
		return
	}

	base := filepath.Base(p.Filepath)
	parentDir := filepath.Dir(p.Filepath)
	path := parentDir + "/" + e.Url

	errMsg := "External DTD '" + e.Url + "' not found, declared in '" + base + "', entity '" + e.Name

	if _, err := os.Stat(path); os.IsNotExist(err) && !p.IgnoreExtRefIssue {
		log.Fatal(errMsg)
	} else if _, err := os.Stat(path); os.IsNotExist(err) && p.IgnoreExtRefIssue {
		log.Error(errMsg)
		return
	} else {
		log.Infof("*** New parser *** for external entity %s", path)
		extP := NewDTDParser()
		extP.outputDirPath = p.outputDirPath
		extP.filepaths = p.filepaths
		extP.WithComments = p.WithComments
		extP.IgnoreExtRefIssue = p.IgnoreExtRefIssue
		extP.Overwrite = p.Overwrite
		extP.Parse(path)

		p.parsers = append(p.parsers, *extP)
	}

}

// SetExportEntity Mark an entity block are exported in the collection
func (p *Parser) SetExportEntity(name string) {
	for _, block := range p.Collection {
		if block.GetName() == name {
			log.Tracef("Marking %s as exported", name)
			block.SetExported(true)

			if DTD.IsEntityType(block) {
				p.parseExternalEntity(block.(*DTD.Entity))
			} else {
				log.Fatal("A none entity was marked to be exported")
			}
			return
		}
	}
}

// RenderDTD Render a collection to a or a set of DTD files
func (p *Parser) RenderDTD(parentDir string) {

	log.Infof("Render DTD,  %d nested parsers", len(p.parsers))

	// we process here all the file path of all DTD parsed
	// and determine the parent directory
	// the parentDir will be happened to the output dir
	log.Tracef("parent Dir is: %s, Filepaths are %+v", parentDir, *p.filepaths)

	if parentDir == "" {
		parentDir = commonPrefix(*p.filepaths)
		log.Tracef("ParentDir from filepaths is: %s", parentDir)
	}

	// create output files,  check if they exists before and if they should be overwritten
	for _, filename := range *p.filepaths {

		finalPath := p.determineFinalDTDPath(parentDir, filename)

		if _, err := os.Stat(finalPath); os.IsNotExist(err) {
			log.Infof("Create DTD: '%s'", finalPath)
			createOutputFile(finalPath, false)
			renderHead(finalPath)
		} else if p.Overwrite {
			log.Infof("Overwrite DTD: '%s'", finalPath)
			createOutputFile(finalPath, true)
			renderHead(finalPath)
		} else {
			log.Fatalf("Output DTD: '%s' already exists, please remove it before or use flag -overwrite", finalPath)
		}
	}

	// export every blocks
	for _, block := range p.Collection {

		log.Tracef("Exporting block: %#v ", block)
		finalPath := p.determineFinalDTDPath(parentDir, block.GetSrc())
		writeToFile(finalPath, block.Render())

	}

	// process children parsers
	for _, parser := range p.parsers {
		parser.RenderDTD(parentDir)
	}
}

func (p *Parser) determineFinalDTDPath(parentDir string, i string) string {

	log.Tracef("determineFinalDTDPath: source is: '%s'", i)

	// determine relative directory
	pDir := filepath.Dir(i)

	// relative dir remove parent directory from the path
	rDir := strings.TrimPrefix(pDir, parentDir)
	log.Tracef("rDir %s", rDir)

	oDir := p.outputDirPath

	// absolute from output dir
	if rDir != "" {
		oDir = p.outputDirPath + "/" + rDir
	}

	if _, err := os.Stat(oDir); os.IsNotExist(err) {
		log.Tracef("Create: %s", oDir)
		os.MkdirAll(oDir, 0770)
	}

	finalPath := oDir + "/" + filepath.Base(i)
	log.Tracef("finalPath %s", finalPath)

	return finalPath
}

// writeToFile write to a DTD file
func writeToFile(filepath string, s string) error {
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY, 0700)

	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.WriteString(f, s+"\n")

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
		v = filepath.Dir(path.Clean(v))

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

// renderHead Render the head of the DTD
func renderHead(filepath string) {
	writeToFile(filepath, xml.Header)
}
