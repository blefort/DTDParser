// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package DTD Represents main structs of a DTD
package DTD

// Element represents a DTD element
type Element struct {
	Name    string
	Value   string
	Content *ElementGroup
}

// Render an Element
// implements IDTDBlock
func (e *Element) Render() string {
	return join("<!ELEMENT ", e.Name, " ", e.Value, ">", "\n")
}

// GetName Get the name
// implements IDTDBlock
func (e *Element) GetName() string {
	return e.Name
}

// SetExported set the current entity to exported
// implements IDTDBlock
func (e *Element) SetExported(v bool) {
	panic("An element should never be set as exported")
}

// GetValue Get the value
// implements IDTDBlock
func (e *Element) GetValue() string {
	return e.Value
}

// GetExtra Get extrainformation
func (e *Element) GetExtra() *DTDExtra {
	var extra DTDExtra
	return &extra
}

// IsElementType check if the interface is a DTD.Element
func IsElementType(i interface{}) bool {
	switch i.(type) {
	case *Element:
		return true
	default:
		return false
	}
}
