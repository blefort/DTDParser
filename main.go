// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Cli to parse a DTD
// The goal is to parse DTD and create corresponding Go structs
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	DTDParser "github.com/blefort/DTDParser/parser"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

// main func
func main() {

	// Input file
	inputPath := flag.String("i", "", "Path to the DTD or catalog.xml")
	outputPath := flag.String("o", "", "Output path to re-generate DTD")
	formatter := flag.String("format", "go", "Choose the output format (go, DTD) ")
	packageName := flag.String("package", "", "Package name")
	overwrite := flag.Bool("overwrite", false, "Overwrite output file")
	verbosity := flag.String("verbosity", "", "Verbose v, vv or vvv")
	ignoreExtRef := flag.Bool("ignore-external-dtd", false, "Do not process external DTD")

	flag.Parse()

	// validate if input file is declared
	if *inputPath == "" {
		panic("Please provide a DTD or a catalog xml file")
	}

	if *outputPath == "" {
		panic("Please provide an output directory")
	}

	// setinput file
	inputPathAbs, _ := filepath.Abs(*inputPath)
	outputPathAbs, _ := filepath.Abs(*outputPath)

	// set logger
	log = setLogger(verbosity, inputPathAbs)

	if _, err := os.Stat(inputPathAbs); os.IsNotExist(err) {
		log.Fatal("Input file does not exists")
	}

	if _, err := os.Stat(outputPathAbs); os.IsNotExist(err) {
		log.Fatal("Output directory does not exists")
	}

	empty, _ := IsEmptyDir(outputPathAbs)
	if !empty {
		log.Fatal("Output directory does not exists")
	}

	if *formatter == "go" && *packageName == "" {
		log.Fatal("Please provide a package name to format the output")
	}

	// log input
	log.Warnf("Starting DTD parser")
	log.Warnf(" - Option 'i': %s", *inputPath)
	log.Warnf(" - Option 'o' DTD: %s", *outputPath)
	log.Warnf(" - Option 'formater': %s", *formatter)
	log.Warnf(" - Option 'verbosity': %s", *verbosity)
	log.Warnf(" - Option 'ignore-external-dtds': %t", *ignoreExtRef)

	log.Warnf("")

	if filepath.Ext(strings.ToLower(inputPathAbs)) == ".xml" {

		log.Debug("Attempt to parse a catalog")

	} else {

		p := DTDParser.NewDTDParser(log)
		p.SetOutputPath(outputPathAbs)
		p.SetStructPath(outputPathAbs)
		p.IgnoreExtRefIssue = *ignoreExtRef
		p.SetFormatter(*formatter)
		p.Package = *packageName

		if *overwrite {
			p.Overwrite = true
		}

		// Parse & render
		t1 := time.Now().Unix()
		p.Parse(inputPathAbs)
		t2 := time.Now().Unix()
		diff := t2 - t1
		log.Warnf(fmt.Sprintf("Parsed in %d ms", diff))
		p.Render("")

	}

}

// setLogger return a logger based on requested verbosity
func setLogger(verbosity *string, inputPathAbs string) *zap.SugaredLogger {
	var level zap.AtomicLevel

	logFile := inputPathAbs + ".log"
	os.Remove(logFile)

	if *verbosity == "v" {
		level = zap.NewAtomicLevelAt(zap.WarnLevel)
	} else if *verbosity == "vv" {
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	} else if *verbosity == "vvv" {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		level = zap.NewAtomicLevelAt(zap.FatalLevel)
	}

	cfg := zap.Config{
		Level:             level,
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: true,
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "m",
			LevelKey:       "",
			TimeKey:        "",
			NameKey:        "",
			CallerKey:      "",
			StacktraceKey:  "stack",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeName:     zapcore.FullNameEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
	}

	cfg.EncoderConfig.TimeKey = zapcore.OmitKey

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync() // flushes buffer, if any
	return logger.Sugar()
}

// IsEmptyDir Test if a directory is empty
// from https://stackoverflow.com/questions/30697324/how-to-check-if-directory-on-path-is-empty
func IsEmptyDir(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
