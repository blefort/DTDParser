package DTD

type ElementGroup struct {
	IsGroup    bool
	IsAnd      bool
	ZeroOrMore bool
	OneOrMore  bool
	Optional   bool
	Name       string
	Children   []*ElementGroup
	Parent     *ElementGroup
}

func (e *ElementGroup) AddChild() *ElementGroup {
	c := new(ElementGroup)
	c.Parent = e
	e.Children = append(e.Children, c)
	return c
}

func (e *ElementGroup) GetParent() *ElementGroup {
	return e.Parent
}
