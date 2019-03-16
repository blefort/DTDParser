// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package DTD Represents main structs of a DTD
package DTD

// Comment represents a comment
type Comment struct {
	Value    string
	Exported bool
	Src      string
}

// Render an entity
// implements IDTDBlock
func (c *Comment) Render() string {
	return "<!--" + c.Value + "-->"
}

// GetName Get the name
// implements IDTDBlock
func (c *Comment) GetName() string {
	return "comment"
}

// SetExported set the current entity to exported
// implements IDTDBlock
func (c *Comment) SetExported(v bool) {
	panic("A comment should never be set as exported")
}

// GetSrc return the source filename where the entity was first found
// implements IDTDBlock
func (c *Comment) GetSrc() string {
	return c.Src
}

// GetValue Get the value
// implements IDTDBlock
func (c *Comment) GetValue() string {
	return c.Value
}

// GetParameter return parameter for entity only
// implements IDTDBlock
func (c *Comment) GetParameter() bool {
	panic("Comment have no Parameter")
}

// GetUrl the entity url
// implements IDTDBlock
func (c *Comment) GetUrl() string {
	panic("GetUrl not allowed for this block")
}

// GetExported Unused, tells if the comment was exported
// implements IDTDBlock
func (c *Comment) GetExported() bool {
	panic("Comment are not exported")
}

// IsCommentType check if the interface is a DTD.Comment
func IsCommentType(i interface{}) bool {
	switch i.(type) {
	case *Comment:
		return true
	default:
		return false
	}
}
