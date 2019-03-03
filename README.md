[![Go Report Card](https://goreportcard.com/badge/github.com/blefort/DTDParser)](https://goreportcard.com/report/github.com/blefort/DTDParser) [![Build Status](https://travis-ci.org/blefort/DTDParser.svg?branch=master)](https://travis-ci.org/blefort/DTDParser)

# A DTDParser

This work is in progress. This is also my first project in Go.

The goal of the project is to parse DTD and generate Structs to be used in others Go programs


# Roadmap

* [ ] Parse DTD and generate corresponding structs in memory
    * [X] Comments struct
    * [X] Entity Struct + exported entity
    * [ ] Element
    * [ ] Attlist
    * [X] External DTD
* [X] Regenerate DTD
* [ ] Generate Structs to be used in other programs using Go prepare
* [ ] Validation
   * [X] Missing external DTD


# Documentation

* [GoDoc](https://godoc.org/github.com/blefort/DTDParser)

# License

Copyright 2019 Bertrand Lefort. All rights reserved.
Use of this source code is governed under MIT License
that can be found in the LICENSE file.