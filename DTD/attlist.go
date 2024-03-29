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
	Entities   []string
}

// Render an Attlist
// implements IDTDBlock
func (a *Attlist) Render() string {
	attributes := "\n"

	for _, attr := range a.Attributes {
		attributes += attr.Render()
	}

	return join("<!ATTLIST ", a.Name, " ", attributes, ">\n")
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

// GetExported Unused, tells if the comment was exported
// implements IDTDBlock
func (a *Attlist) GetExported() bool {
	panic("Attlist are not exported")
}

// GetExtra Get extrainformation
func (a *Attlist) GetExtra() *DTDExtra {
	var extra DTDExtra
	extra.Attributes = a.Attributes
	return &extra
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
