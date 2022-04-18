// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package DTD Represents main structs of a DTD
package DTD

// ExportedEntity represents an exported entity
type ExportedEntity struct {
	Name  string
	Value string
}

// Render an entity
// implements IDTDBlock
func (e *ExportedEntity) Render() string {
	return ""
}

// GetName Get the name
// implements IDTDBlock
func (e *ExportedEntity) GetName() string {
	return e.Name
}

// SetExported set the current entity to exported
// implements IDTDBlock
func (e *ExportedEntity) SetExported(v bool) {
	panic("A comment should never be set as exported")
}

// GetValue Get the value
// implements IDTDBlock
func (e *ExportedEntity) GetValue() string {
	return e.Value
}

// GetParameter return parameter for entity only
// implements IDTDBlock
func (e *ExportedEntity) GetParameter() bool {
	panic("ExportedEntity have no Parameter")
}

// GetUrl the entity url
// implements IDTDBlock
func (e *ExportedEntity) GetUrl() string {
	panic("GetUrl not allowed for this block")
}

// GetExported Unused, tells if the comment was exported
// implements IDTDBlock
func (e *ExportedEntity) GetExported() bool {
	panic("ExportedEntity are not exported")
}

// GetAttributes return a list of attributes
func (e *ExportedEntity) GetAttributes() []Attribute {
	panic("Comment have no attributes")
}

// IsExportedEntityType check if the interface is a DTD.ExportedEntity
func IsExportedEntityType(i interface{}) bool {
	switch i.(type) {
	case *ExportedEntity:
		return true
	default:
		return false
	}
}
