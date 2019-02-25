// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Cli to parse a DTD
// The goal is to parse DTD and create corresponding Go structs
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	DTDParser "github.com/blefort/DTDParser/parser"
)

// main func
func main() {

	var DTDFullPath string
	var withComments bool
	var verbose bool
	var verboseExtra bool
	var ignoreExtRefIssue bool
	var output string

	// declare some flags
	flag.StringVar(&DTDFullPath, "DTD", "", "Path to the DTD")
	flag.StringVar(&output, "o", "", "Output path to regenetate DTD")
	flag.BoolVar(&withComments, "keep-comments", false, "Keep comments while parsing")
	flag.BoolVar(&verbose, "v", false, "Verbose")
	flag.BoolVar(&verboseExtra, "vv", false, "Extra verbose")
	flag.BoolVar(&ignoreExtRefIssue, "ignore-ext-ref-issue", false, "Keep comments while parsing")
	flag.Parse()

	DTDFullPathAbs, err0 := filepath.Abs(DTDFullPath)

	if err0 != nil {
		os.Exit(1)
	}

	if verbose || verboseExtra {
		fmt.Printf("Starting DTD parser with options:\n")
		fmt.Printf(" * DTD: %s\n", DTDFullPath)
		fmt.Printf(" * o: %s\n", output)
		fmt.Printf(" * withComments: %t\n", withComments)
		fmt.Printf(" * verbose: %t\n", (verbose || verboseExtra))
		fmt.Printf(" * Extra verbose: %t\n", verboseExtra)
		fmt.Printf(" * ignoreExtRefIssue: %t\n", ignoreExtRefIssue)
		fmt.Printf("")
	}

	if _, err := os.Stat(DTDFullPathAbs); os.IsNotExist(err) {
		fmt.Printf("Provide a valid path to a DTD file")
		os.Exit(1)
	}

	// New parser
	p := DTDParser.NewDTDParser()

	// Configure parser
	p.WithComments = withComments
	p.IgnoreExtRefIssue = ignoreExtRefIssue

	if verbose {
		p.SetVerbose(DTDParser.LogVerbose)
	}

	if verboseExtra {
		p.SetVerbose(DTDParser.LogFull)
	}

	if output != "" {
		outputPathAbs, err3 := filepath.Abs(output)

		if err3 != nil {
			os.Exit(1)
		}

		p.SetOutputPath(outputPathAbs)
	}

	// Parse
	p.Parse(DTDFullPathAbs)
	p.Render("")
}
