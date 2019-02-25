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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/blefort/DTDParser/DTD"
	"github.com/blefort/DTDParser/scanner"
)

// LOG LEVELS
const (
	LogMute    = 0 // Mute logs (default)
	LogVerbose = 1 // Steps are presented in the log
	LogFull    = 2 // Every single operations
)

var (
	filepaths         []string
	logVerbosityLevel int
	outputDirPath     string
)

// Parser is a structure that represents the parser
// it can manage multiple DTD
type Parser struct {
	WithComments      bool
	IgnoreExtRefIssue bool
	Filepath          string
	Collection        []DTD.IDTDBlock
	parsers           []Parser
}

// NewDTDParser returns a new DTD parser
func NewDTDParser() *Parser {
	p := Parser{}
	return &p
}

// SetOutputPath set the output path of the DTD
// to export the DTD
func (p *Parser) SetOutputPath(s string) {
	outputDirPath = s
}

// createOutputFile Create a DTD output file
func createOutputFile(filename string) {
	f, err := os.Create(filename)

	if err != nil {
		panic(err)
	}

	f.Close()
}

// SetVerbose allows to set the log verbosity level
// DTDParser.LogMute default log are muted
// DTDParser.LogVerbose general processing steps are output to the log
// DTDParser.LogVerbose all processing steps are output to the log
func (p *Parser) SetVerbose(b int) {
	logVerbosityLevel = b
}

// Parse Parse a DTD using its path
func (p *Parser) Parse(filePath string) {

	p.Filepath = filePath

	// Open file
	filebuffer, err := ioutil.ReadFile(p.Filepath)

	if err != nil {
		panic(err)
	}

	// Fileinfo
	stat, err := os.Stat(filePath)

	if err != nil {
		panic(err)
	}

	bytes := stat.Size()
	log(LogVerbose, "Parsing %s, %d bytes", p.Filepath, bytes/1024)

	filepaths = append(filepaths, filepath.Dir(p.Filepath))

	// use  bufio to read file rune by rune
	inputdata := string(filebuffer)
	scanner := scanner.NewScanner(filePath, inputdata)

	// not sure if this is correct methodology
	// I tried to separate the DTD Scanner from the parser
	// the scanner should send DTD blocks that the parser
	// will put in a collection.

	for scanner.Next() {

		DTDBlock, err := scanner.Scan()

		if err != nil {
			log(LogFull, "%v", err)
			continue
		}

		p.Collection = append(p.Collection, DTDBlock)

		// if export == "" {

		// } else {
		//   p.SetExportEntity(p.Collection, export)
		// }

	}
	log(LogVerbose, "%s blocks found in this DTD %s", len(p.Collection), p.Filepath)

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
		panic(errMsg)
	} else if _, err := os.Stat(path); os.IsNotExist(err) && p.IgnoreExtRefIssue {
		log(LogVerbose, "-> ERROR:"+errMsg)
		return
	} else {

		extP := NewDTDParser()
		extP.WithComments = p.WithComments
		extP.IgnoreExtRefIssue = p.IgnoreExtRefIssue
		extP.SetVerbose(logVerbosityLevel)
		extP.Parse(path)

		p.parsers = append(p.parsers, *extP)
	}

}

// SetExportEntity Mark an entity block are exported in the collection
func (p *Parser) SetExportEntity(C []DTD.IDTDBlock, name string) {
	for idx := range C {
		if C[idx].GetName() == name {
			var block DTD.IDTDBlock
			block = C[idx]
			block.SetExported(true)
			return
		}
	}
}

// Render a collection to a or a set of DTD files
func (p *Parser) Render(parentDir string) {

	if outputDirPath == "" {
		log(LogVerbose, "No rendering requested")
		return
	}

	log(LogFull, "Rendering starts")

	// we process here all the file path of all DTD parsed
	// and determine the parent directory
	// the parentDir will be happened to the output dir
	if parentDir == "" {
		parentDir = commonPrefix(filepaths)
		log(LogVerbose, "ParentDir: %s", parentDir)
	}

	for _, block := range p.Collection {

		// extract origin location of the file
		src := block.GetSrc()
		log(LogVerbose, "src %s", src)

		// determine relative directory
		pDir := filepath.Dir(src)

		// relative dir remove parent directory from the path
		rDir := strings.TrimPrefix(pDir, parentDir)
		log(LogVerbose, "rDir %s", rDir)

		// absolute from output dir
		oDir := outputDirPath + "/" + rDir

		if _, err := os.Stat(oDir); os.IsNotExist(err) {
			log(LogVerbose, "Create: %s", oDir)
			os.MkdirAll(oDir, 0770)
		}

		finalPath := oDir + "/" + filepath.Base(src)
		log(LogVerbose, "finalPath %s", finalPath)

		if _, err := os.Stat(finalPath); os.IsNotExist(err) {
			log(LogVerbose, "Create DTD: %s", finalPath)
			createOutputFile(finalPath)
			renderHead(finalPath)
		}

		//log(LogFull, "Write %s", reflect.TypeOf(block), block.GetName(), //)
		writeToFile(finalPath, block.Render())

	}

	for _, parser := range p.parsers {
		parser.Render(parentDir)
	}

	PrintMemUsage()
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

// renderHead Render the head of the DTD
func renderHead(filepath string) {
	writeToFile(filepath, xml.Header)
}

// log log
// will be moved in a separate file in a next version
func log(level int, format string, a ...interface{}) {
	if logVerbosityLevel > 0 && logVerbosityLevel >= level {
		fmt.Fprintf(os.Stdout, "-"+format+"\n", a...)
	}
}

// PrintMemUsage Print memory usage
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log(LogFull, "Alloc = %v MiB, TotalAlloc = %v MiB, tSys = %v MiB, tNumGC = %v", bToMb(m.Alloc), bToMb(m.TotalAlloc), bToMb(m.Sys), m.NumGC)
}

// bToMb helper to convert bytes to megabytes, used for development
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
