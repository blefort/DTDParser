// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package DTD Represents main structs of a DTD
package DTD

// Attlist represents an attlist
type Attlist struct {
	Name       string
	Value      string
	Attributes []Attribute
}

// Render an Attlist
// implements IDTDBlock
func (a *Attlist) Render() string {
	attributes := ""

	for _, attr := range a.Attributes {
		attributes += attr.Render()
	}
	return join("<!ATTLIST ", a.Name, " ", attributes, ">")
}

// GetName Get the name
// implements IDTDBlock
func (a *Attlist) GetName() string {
	return a.Name
}

// SetExported set the current entity to exported
// implements IDTDBlock
func (a *Attlist) SetExported(v bool) {
	panic("A comment should never be set as exported")
}

// GetValue Get the value
// implements IDTDBlock
func (a *Attlist) GetValue() string {
	return a.Value
}

// GetParameter return parameter for entity only
// implements IDTDBlock
func (a *Attlist) GetParameter() bool {
	panic("Attlist have no Parameter")
}

// GetUrl the entity url
// implements IDTDBlock
func (a *Attlist) GetUrl() string {
	panic("GetUrl not allowed for this block")
}

// GetExported Unused, tells if the comment was exported
// implements IDTDBlock
func (a *Attlist) GetExported() bool {
	panic("Attlist are not exported")
}

// IsAttlistType check if the interface is a DTD.Comment
func IsAttlistType(i interface{}) bool {
	switch i.(type) {
	case *Attlist:
		return true
	default:
		return false
	}
}
