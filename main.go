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
	"time"

	DTDParser "github.com/blefort/DTDParser/parser"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// main func
func main() {

	var level zap.AtomicLevel

	// Input file
	DTDFullPath := flag.String("DTD", "", "Path to the DTD")
	DTDOutput := flag.String("output", "", "Output path to re-generate DTD")
	GoStructOutput := flag.String("type", "", "Output path to generate go structs")
	formatter := flag.String("format", "go", "Choose the output format (go, DTD) ")
	packageName := flag.String("package", "", "Package name")
	overwrite := flag.Bool("overwrite", false, "Overwrite output file")
	verbosity := flag.String("verbosity", "", "Verbose v, vv or vvv")
	ignoreExtRef := flag.Bool("ignore-external-dtd", false, "Do not process external DTD")

	flag.Parse()

	// Process DTD
	if *DTDFullPath == "" {
		panic("Please provide a DTD")
	}

	if *formatter == "go" && *packageName == "" {
		panic("Please provide a package name")
	}

	DTDFullPathAbs, err0 := filepath.Abs(*DTDFullPath)

	if err0 != nil {
		os.Exit(1)
	}

	logFile := DTDFullPathAbs + ".log"
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
	log := logger.Sugar()

	// log input
	log.Warnf("Starting DTD parser")
	log.Warnf(" - Option DTD: %s", *DTDFullPath)
	log.Warnf(" - Option Output DTD: %s", *DTDOutput)
	log.Warnf(" - Option Formater: %s", *formatter)
	log.Warnf(" - Option Verbosity: %s", *verbosity)
	log.Warnf(" - Option ignore external references: %t", *ignoreExtRef)

	log.Warnf("")

	if _, err := os.Stat(DTDFullPathAbs); os.IsNotExist(err) {
		log.Fatal("Provide a valid path to a DTD file")
	}

	// New parser
	p := DTDParser.NewDTDParser(log)
	p.IgnoreExtRefIssue = *ignoreExtRef
	p.SetFormatter(*formatter)
	p.Package = *packageName

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

	if *GoStructOutput != "" {

		outputPathAbs, err3 := filepath.Abs(*GoStructOutput)

		if err3 != nil {
			os.Exit(1)
		}

		p.SetStructPath(outputPathAbs)
	}

	// Parse & render
	t1 := time.Now().Unix()
	p.Parse(DTDFullPathAbs)
	t2 := time.Now().Unix()
	diff := t2 - t1
	log.Warnf(fmt.Sprintf("Parsed in %d ms", diff))
	p.Render("")
}
