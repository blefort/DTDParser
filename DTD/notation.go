// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package DTD Represents main structs of a DTD
package DTD

// Notation reprensents a notation
type Notation struct {
	Name   string
	Public bool
	System bool
	Url    string
	Value  string
	ID     string
}

// Render an Notation
// implements IDTDBlock
func (n *Notation) Render() string {
	return join("<!NOTATION", n.Name, ">", "\n")
}

// GetName Get the name
// implements IDTDBlock
func (n *Notation) GetName() string {
	return n.Name
}

// SetExported set the current entity to exported
// implements IDTDBlock
func (n *Notation) SetExported(v bool) {
	panic("An Notation should never be set as exported")
}

// GetValue Get the value
// implements IDTDBlock
func (n *Notation) GetValue() string {
	return n.Value
}

// GetParameter return parameter for entity only
// implements IDTDBlock
func (n *Notation) GetParameter() bool {
	panic("Notation have no Parameter")
}

// GetUrl the entity url
// implements IDTDBlock
func (n *Notation) GetUrl() string {
	panic("GetUrl not allowed for this block")
}

// GetExported Unused, tells if the comment was exported
// implements IDTDBlock
func (n *Notation) GetExported() bool {
	panic("Notation are not exported")
}

// IsNotationType check if the interface is a DTD.Notation
func IsNotationType(i interface{}) bool {
	switch i.(type) {
	case *Notation:
		return true
	default:
		return false
	}
}
