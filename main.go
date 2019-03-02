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

	var DTDFullPath string
	var withComments bool
	var verbose bool
	var verboseExtra bool
	var verboseTrace bool
	var ignoreExtRefIssue bool
	var jsonLogFormat bool
	var DTDoutput string

	// declare some flags
	flag.StringVar(&DTDFullPath, "DTD", "", "Path to the DTD")
	flag.StringVar(&DTDoutput, "o", "", "Output path to regenetate DTD")
	flag.BoolVar(&withComments, "keep-comments", false, "Keep comments while parsing")
	flag.BoolVar(&verbose, "v", false, "Verbose")
	flag.BoolVar(&verboseExtra, "vv", false, "Extra verbose")
	flag.BoolVar(&verboseTrace, "trace", false, "Trace")
	flag.BoolVar(&jsonLogFormat, "log-json", false, "Log in json format")
	flag.BoolVar(&ignoreExtRefIssue, "ignore-ext-ref-issue", false, "Keep comments while parsing")
	flag.Parse()

	DTDFullPathAbs, err0 := filepath.Abs(DTDFullPath)

	if err0 != nil {
		os.Exit(1)
	}

	// configure logger
	if jsonLogFormat {
		log.SetFormatter(&log.JSONFormatter{})
	}

	if verbose {
		log.SetLevel(logrus.InfoLevel)
	}

	if verboseExtra {
		log.SetLevel(logrus.DebugLevel)
	}

	if verboseTrace {
		log.SetLevel(logrus.TraceLevel)
	}

	log.Infof("Starting DTD parser")
	log.Infof(" - Option DTD: %s", DTDFullPath)
	log.Infof(" - Option o: %s", DTDoutput)
	log.Infof(" - Option withComments: %t", withComments)
	log.Infof(" - Option verbose: %t", (verbose || verboseExtra))
	log.Infof(" - Option Extra verbose: %t", verboseExtra)
	log.Infof(" - Option ignoreExtRefIssue: %t", ignoreExtRefIssue)
	log.Infof(" - Option json log format: %t", jsonLogFormat)
	log.Infof("")

	if _, err := os.Stat(DTDFullPathAbs); os.IsNotExist(err) {
		log.Fatal("Provide a valid path to a DTD file")
	}

	// New parser
	p := DTDParser.NewDTDParser()

	// Configure parser
	p.WithComments = withComments
	p.IgnoreExtRefIssue = ignoreExtRefIssue

	if DTDoutput != "" {
		outputPathAbs, err3 := filepath.Abs(DTDoutput)

		if err3 != nil {
			os.Exit(1)
		}

		p.SetOutputPath(outputPathAbs)
	}

	// Parse
	p.Parse(DTDFullPathAbs)

	log.Info("Rendering DTD")
	p.Render("")
}
