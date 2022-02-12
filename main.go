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

	// Input file
	DTDFullPath := flag.String("DTD", "", "Path to the DTD")
	DTDOutput := flag.String("output-dtd", "", "Output path to re-generate DTD")
	GoStructOutput := flag.String("output-struct", "", "Output path to generate go structs")
	stripComment := flag.Bool("srip-comments", false, "Strip comments")
	overwrite := flag.Bool("overwrite", false, "Overwrite output file")
	verbosity := flag.String("verbose", "v", "Verbose v, vv or trace")
	ignoreExtRef := flag.Bool("ignore-external-ref", false, "Do not process external references")

	flag.Parse()

	// Process DTD
	if *DTDFullPath == "" {
		panic("Please provide a DTD")
	}

	DTDFullPathAbs, err0 := filepath.Abs(*DTDFullPath)

	if err0 != nil {
		os.Exit(1)
	}

	logFile := DTDFullPathAbs + ".log"
	os.Remove(logFile)

	cfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
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

	// if *verbosity == "v" {
	// //	log.
	// }

	// if *verbosity == "vv" {
	// 	log.SetLevel(logrus.DebugLevel)
	// }

	// if *verbosity == "trace" {
	// 	log.SetLevel(logrus.TraceLevel)
	// }

	// log input
	log.Infof("Starting DTD parser")
	log.Infof(" - Option DTD: %s", *DTDFullPath)
	log.Infof(" - Option Output DTD: %s", *DTDOutput)
	log.Infof(" - Option Strip Comments: %t", *stripComment)
	log.Infof(" - Option Verbosity: %s", *verbosity)
	log.Infof(" - Option ignore external references: %t", *ignoreExtRef)
	//log.Infof(" - Option json log format: %s", *logFormat)
	log.Infof("")

	if _, err := os.Stat(DTDFullPathAbs); os.IsNotExist(err) {
		log.Fatal("Provide a valid path to a DTD file")
	}

	// New parser
	p := DTDParser.NewDTDParser(log)

	// Configure parser
	fmt.Printf("%v", p.Log)

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

	if *GoStructOutput != "" {

		outputPathAbs, err3 := filepath.Abs(*GoStructOutput)

		if err3 != nil {
			os.Exit(1)
		}

		p.SetStructPath(outputPathAbs)
	}

	// Parse & render
	t1 := time.Now()
	p.Parse(DTDFullPathAbs)
	t2 := time.Now()
	diff := t2.Sub(t1)
	fmt.Println(diff)
	//p.RenderDTD("")
	//p.RenderGoStructs()
}
