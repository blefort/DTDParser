[![Build Status](https://app.travis-ci.com/blefort/DTDParser.svg?branch=master)](https://travis-ci.org/blefort/DTDParser) [![codecov](https://codecov.io/gh/blefort/DTDParser/branch/master/graph/badge.svg)](https://codecov.io/gh/blefort/DTDParser) [![Go Report Card](https://goreportcard.com/badge/github.com/blefort/DTDParser)](https://goreportcard.com/report/github.com/blefort/DTDParser)  [![GoDoc](https://godoc.org/github.com/blefort/DTDParser?status.svg)](https://godoc.org/github.com/blefort/DTDParser) 

# A DTD Parser 

Exploring Go language in a DTD parser. This is a personal project, I do it when I have time, if you interest in it, feel free to open an issue.
The goal of the project is to parse DTD and generate Structs to be used in others Go programs.

# How to run?

Simple example
```go run . -i sample/docbook.dtd -o out -format go -verbosity v```


# Roadmap

* [alpha] Parse DTD and generate corresponding structs in memory
* [alpha] Regenerate DTD - Alpha version
* [wip] Generate Structs to be used in other programs using Go prepare
* [ ] DTD Validation
   * [X] Missing external DTD

# License

Copyright 2019 Bertrand Lefort. All rights reserved.
Use of this source code is governed under MIT License
that can be found in the LICENSE file.
