package core

/*
	NamespaceDeclarationContainer implementation
*/

type NamespaceDeclarationContainer struct {
	NamespaceURI string
	Prefix       *string
}

func NewNamespaceDeclarationContainer(namespaceURI string, prefix *string) NamespaceDeclarationContainer {
	return NamespaceDeclarationContainer{
		NamespaceURI: namespaceURI,
		Prefix:       prefix,
	}
}

func (c *NamespaceDeclarationContainer) Equals(o any) bool {
	if o == nil {
		return false
	}
	nsc, ok := o.(*NamespaceDeclarationContainer)
	if ok {
		return c == nsc
	}
	return false
}

/*
	DocTypeContainer implementation
*/

type DocTypeContainer struct {
	Name     []rune
	PublicID []rune
	SystemID []rune
	Text     []rune
}

/*
	ProcessingInstructionContainer implementation
*/

type ProcessingInstructionContainer struct {
	Target string
	Data   string
}
