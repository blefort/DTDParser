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

	"github.com/blefort/DTDParser/formatter"
	DTDformat "github.com/blefort/DTDParser/formatter/DTDformat"
	"github.com/blefort/DTDParser/formatter/GoFormat"
	"github.com/blefort/DTDParser/formatter/JsonFormat"
	DTDParser "github.com/blefort/DTDParser/parser"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

// main func
func main() {

	// Input file
	inputPath := flag.String("i", "", "Path to the DTD or catalog.xml")
	outputDir := flag.String("o", "", "Output directory")
	outputFilename := flag.String("f", "", "Output filename")
	formatOutput := flag.String("format", "go", "Choose the output format (go, DTD) ")
	optionsParam := flag.String("options", "{ package=\"MyPackage\"", "Formatter options, appends options as a json string")
	overwrite := flag.Bool("overwrite", false, "Overwrite output file")
	verbosity := flag.String("verbosity", "", "Verbose v, vv or vvv")
	ignoreExtRef := flag.Bool("ignore-external-dtd", false, "Process external DTD, but do not stop in case of errors")

	flag.Parse()

	// validate if input file is declared
	if *inputPath == "" {
		fmt.Println("Please provide a DTD or a catalog xml file")
		os.Exit(1)
	}

	if *outputDir == "" {
		fmt.Println("Please provide an output directory")
		os.Exit(1)
	}

	// setinput file
	inputPathAbs, _ := filepath.Abs(*inputPath)
	outputDirAbs, _ := filepath.Abs(*outputDir)

	// set logger
	log = setLogger(verbosity, inputPathAbs)

	log.Info("start")

	if _, err := os.Stat(inputPathAbs); os.IsNotExist(err) {
		log.Fatal("Input file does not exists")
		os.Exit(1)
	}

	if _, err := os.Stat(outputDirAbs); os.IsNotExist(err) {
		log.Fatal("Output directory does not exists")
		os.Exit(1)
	}

	empty, _ := IsEmptyDir(outputDirAbs)

	if !empty && !*overwrite {
		log.Fatal("Output directory is not empty, use -overwrite to overwrite files")
		os.Exit(1)
	}

	if !formatter.Exists(*formatOutput) {
		log.Fatalf("Formatter '%s' does not exists, current formatter available are '%s'", *formatOutput, strings.Join(formatter.AvailaibleFormatters(), ", "))
		os.Exit(1)
	}

	if *optionsParam != "" {

	}

	// log input
	log.Warnf("Starting DTD parser")
	log.Warnf(" - Option 'i': %s", *inputPath)
	log.Warnf(" - Option 'o' DTD: %s", *outputDir)
	log.Warnf(" - Option 'formater': %s", *formatOutput)
	log.Warnf(" - Option 'options': %s", *optionsParam)
	log.Warnf(" - Option 'verbosity': %s", *verbosity)
	log.Warnf(" - Option 'ignore-external-dtds': %t", *ignoreExtRef)

	log.Warnf("")

	if filepath.Ext(strings.ToLower(inputPathAbs)) == ".xml" {

		log.Debug("Attempt to parse a catalog")

	} else {

		var format formatter.FormatterInterface

		p := DTDParser.NewDTDParser(log)
		p.IgnoreExtRefIssue = *ignoreExtRef
		p.Parse(inputPathAbs)

		if strings.ToLower(*formatOutput) == "dtd" {
			format = DTDformat.New(log)
		}

		if strings.ToLower(*formatOutput) == "go" {
			format = GoFormat.New(log)
		}

		if strings.ToLower(*formatOutput) == "json" {
			format = JsonFormat.New(log)
		}

		formatter := formatter.NewFormatter(p, format, outputDirAbs, *outputFilename, log)
		formatter.Render()

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
