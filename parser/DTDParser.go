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
	"io/ioutil"
	"os"
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
	Elements          map[string]*Element
	Parsers           []Parser
	Filepaths         *[]string
	Log               *zap.SugaredLogger
}

type Element struct {
	Element    *DTD.Element
	Attributes *DTD.Attlist
	Comment    *DTD.Comment
}

// NewDTDParser returns a new DTD parser
func NewDTDParser(Log *zap.SugaredLogger) *Parser {
	p := Parser{}
	p.Log = Log
	return &p
}

// Parse Parse a DTD using its path
func (p *Parser) Parse(filePath string) {
	var filespaths []string

	p.Log.Infof("parsing '%s'", filePath)

	if p.Filepaths == nil {
		p.Filepaths = &filespaths
		p.Log.Debugf("Parser filepaths was nil")
	}
	p.Filepath = filePath
	p.Elements = make(map[string]*Element)

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

	*p.Filepaths = append(*p.Filepaths, p.Filepath)

	// use  bufio to read file rune by rune
	inputdata := string(filebuffer)

	//p.Log.Debugf("File content is: %s", inputdata)
	scanner := scanner.NewScanner(filePath, inputdata, p.Log)

	// not sure if this is correct methodology
	// I tried to separate the DTD Scanner from the parser
	// the scanner should send DTD blocks that the parser
	// will put in a collection.
	for scanner.NextBlock() {

		DTDBlock, extraWords, err := scanner.Scan()

		if err != nil {
			p.Log.Debugf("%v", err)
			continue
		}

		for _, word := range extraWords {
			entityName := strings.Trim(word.Read(), "%; ")
			p.Log.Warnf("Exporting entity: '" + entityName + "'")
			p.SetExportEntity(entityName)
		}

		DTDElement, ok := DTDBlock.(*DTD.Element)

		if ok {
			el := p.setElement(DTDElement.Name)
			el.Element = DTDElement
		}

		DTDAttr, ok := DTDBlock.(*DTD.Attlist)

		if ok {
			el := p.setElement(DTDAttr.Name)
			el.Attributes = DTDAttr
		}

		p.Collection = append(p.Collection, DTDBlock.(DTD.IDTDBlock))

		if DTD.IsEntityType(DTDBlock) {
			p.parseExternalEntity(DTDBlock.(*DTD.Entity))
		}

	}
	p.Log.Infof("%d blocks found in DTD '%s'", len(p.Collection), p.Filepath)
}

func (p *Parser) setElement(name string) *Element {
	val, ok := p.Elements[name]
	if ok {
		return val
	}
	var El Element
	p.Elements[name] = &El
	return &El
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
	extP.Filepaths = p.Filepaths
	extP.WithComments = p.WithComments
	extP.IgnoreExtRefIssue = p.IgnoreExtRefIssue

	extP.Parse(path)

	p.Log.Warnf("*** /end of New parser %s", path)
	p.Parsers = append(p.Parsers, *extP)

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
	p.Log.Warnf("could not find ", name, " in the current collection")
}
