package DTD

type ElementGroup struct {
	IsAnd      bool
	ZeroOrMore bool
	OneOrMore  bool
	Optional   bool
	Elements   []string        `json:"Elements,omitempty"`
	Children   []*ElementGroup `json:"Children,omitempty"`
	parent     *ElementGroup
}

func (e *ElementGroup) AddChild() *ElementGroup {
	c := new(ElementGroup)
	c.parent = e
	e.Children = append(e.Children, c)
	return c
}

func (e *ElementGroup) SetRoot() {
	e.parent = e
}

func (e *ElementGroup) GetParent() *ElementGroup {
	return e.parent
}

func (e *ElementGroup) AddChildElement(el *string) {
	if *el == "" {
		return
	}
	e.Elements = append(e.Elements, *el)
	*el = ""
}
