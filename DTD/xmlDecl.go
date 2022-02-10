// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package DTD Represents main structs of a DTD
package DTD

// Attlist represents an attlist
type XMLDecl struct {
	Version  string
	Encoding string
}

// Render an Attlist
// implements IDTDBlock
func (a *XMLDecl) Render() string {
	return join("")
}

// GetName Get the name
// implements IDTDBlock
func (a *XMLDecl) GetName() string {
	panic("A XML Declaration has no attribute name")
}

// SetExported set the current entity to exported
// implements IDTDBlock
func (a *XMLDecl) SetExported(v bool) {
	panic("A XML Declaration should never be set as exported")
}

// GetValue Get the value
// implements IDTDBlock
func (a *XMLDecl) GetValue() string {
	panic("A XML Declaration has no attribute value")
}

// GetParameter return parameter for entity only
// implements IDTDBlock
func (a *XMLDecl) GetParameter() bool {
	panic("A XML Declaration has no attribute parameter")
}

// GetUrl the entity url
// implements IDTDBlock
func (a *XMLDecl) GetUrl() string {
	panic("A XML Declaration has no attribute URL")
}

// GetExported Unused, tells if the comment was exported
// implements IDTDBlock
func (a *XMLDecl) GetExported() bool {
	panic("A XML Declaration can'T be exported")
}
