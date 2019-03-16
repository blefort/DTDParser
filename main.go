// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Cli to parse a DTD
// The goal is to parse DTD and create corresponding Go structs
package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	log "github.com/Sirupsen/logrus"
	DTDParser "github.com/blefort/DTDParser/parser"
)

// main func
func main() {

	// Input file
	DTDFullPath := flag.String("DTD", "", "Path to the DTD")
	DTDOutput := flag.String("output-dtd", "", "Output path to regenetate DTD")
	stripComment := flag.Bool("srip-comments", false, "Strip comments")
	overwrite := flag.Bool("overwrite", false, "Overwrite output file")
	verbosity := flag.String("verbose", "v", "Verbose v, vv or trace")
	logFormat := flag.String("log-format", "default", "Log format, <json> or <default>")
	ignoreExtRef := flag.Bool("ignore-external-ref", false, "Do not process external references")

	flag.Parse()

	// configure logger
	if *logFormat == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	}

	if *verbosity == "v" {
		log.SetLevel(logrus.InfoLevel)
	}

	if *verbosity == "vv" {
		log.SetLevel(logrus.DebugLevel)
	}

	if *verbosity == "trace" {
		log.SetLevel(logrus.TraceLevel)
	}

	// log input
	log.Infof("Starting DTD parser")
	log.Infof(" - Option DTD: %s", *DTDFullPath)
	log.Infof(" - Option Output DTD: %s", *DTDOutput)
	log.Infof(" - Option Strip Comments: %t", *stripComment)
	log.Infof(" - Option Verbosity: %s", *verbosity)
	log.Infof(" - Option ignore external references: %t", *ignoreExtRef)
	log.Infof(" - Option json log format: %s", *logFormat)
	log.Infof("")

	// Process DTD
	if *DTDFullPath == "" {
		panic("Please provide a DTD")
	}

	DTDFullPathAbs, err0 := filepath.Abs(*DTDFullPath)

	if err0 != nil {
		os.Exit(1)
	}

	if _, err := os.Stat(DTDFullPathAbs); os.IsNotExist(err) {
		log.Fatal("Provide a valid path to a DTD file")
	}

	// New parser
	p := DTDParser.NewDTDParser()

	// Configure parser
	p.WithComments = !*stripComment
	p.IgnoreExtRefIssue = *ignoreExtRef

	if *overwrite {
		p.Overwrite = true
	}

	if *DTDOutput != "" {

		outputPathAbs, err3 := filepath.Abs(*DTDOutput)

		if err3 != nil {
			os.Exit(1)
		}

		p.SetOutputPath(outputPathAbs)
	}

	// Parse & render
	p.Parse(DTDFullPathAbs)
	p.RenderDTD("")
}
