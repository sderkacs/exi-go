package core

import (
	"fmt"
	"sync"

	"github.com/sderkacs/exi-go/utils"
)

type Grammars interface {
	IsSchemaInformed() bool
	GetSchemaID() *string
	SetSchemaID(schemaID *string) error
	IsBuiltInXMLSchemaTypesOnly() bool
	GetDocumentGrammar() Grammar
	GetFragmentGrammar() Grammar
	GetGrammarContext() *GrammarContext
}

/*
	AbstractGrammars implementation
*/

type AbstractGrammars struct {
	Grammars
	documentGrammar  Grammar
	fragmentGrammar  Grammar
	grammarContext   *GrammarContext
	isSchemaInformed bool
}

func NewAbstractGrammars(isSchemaInformed bool, grammarContext *GrammarContext) *AbstractGrammars {
	return &AbstractGrammars{
		documentGrammar:  nil,
		fragmentGrammar:  nil,
		grammarContext:   grammarContext,
		isSchemaInformed: isSchemaInformed,
	}
}

func (g *AbstractGrammars) GetGrammarContext() *GrammarContext {
	return g.grammarContext
}

func (g *AbstractGrammars) IsSchemaInformed() bool {
	return g.isSchemaInformed
}

func (g *AbstractGrammars) GetDocumentGrammar() Grammar {
	return g.documentGrammar
}

/*
	SchemaInformedGrammars implementation
*/

type SchemaInformedGrammars struct {
	*AbstractGrammars
	builtInXMLSchemaTypesOnly bool
	schemaID                  *string
	elementFragmentGrammar    SchemaInformedGrammar
}

func NewSchemaInformedGrammars(
	grammarContext *GrammarContext,
	document *Document,
	fragment *Fragment,
	elementFragmentGrammar SchemaInformedGrammar,
) *SchemaInformedGrammars {
	return &SchemaInformedGrammars{
		AbstractGrammars: &AbstractGrammars{
			documentGrammar:  document,
			fragmentGrammar:  fragment,
			grammarContext:   grammarContext,
			isSchemaInformed: true,
		},
		builtInXMLSchemaTypesOnly: false,
		schemaID:                  utils.AsPtr(EmptyString),
		elementFragmentGrammar:    elementFragmentGrammar,
	}
}

func (g *SchemaInformedGrammars) SetBuiltInXMLSchemaTypesOnly(builtInXMLSchemaTypesOnly bool) {
	g.builtInXMLSchemaTypesOnly = builtInXMLSchemaTypesOnly
	g.schemaID = utils.AsPtr(EmptyString)
}

func (g *SchemaInformedGrammars) GetSchemaID() *string {
	return g.schemaID
}

func (g *SchemaInformedGrammars) SetSchemaID(schemaID *string) error {
	if g.builtInXMLSchemaTypesOnly && (schemaID == nil || *schemaID != EmptyString) {
		return fmt.Errorf("xml schema types only grammars do have schema ID == '' associated with it")
	} else {
		if schemaID == nil || *schemaID == EmptyString {
			return fmt.Errorf("schema-informed grammars do have schemaId != '' && schemaId != null associated with it")
		}
	}
	g.schemaID = schemaID

	return nil
}

func (g *SchemaInformedGrammars) IsBuiltInXMLSchemaTypesOnly() bool {
	return g.builtInXMLSchemaTypesOnly
}

func (g *SchemaInformedGrammars) GetFragmentGrammar() Grammar {
	return g.fragmentGrammar
}

func (g *SchemaInformedGrammars) GetSchemaInformedElementFragmentGrammar() SchemaInformedGrammar {
	return g.elementFragmentGrammar
}

/*
	SchemaLessGrammars implementation
*/

var (
	schemaLessGrammarContext *GrammarContext
	schemaLessInit           sync.Once
)

type SchemaLessGrammars struct {
	*AbstractGrammars
}

func initSchemaLessGrammarContext() {
	contexts := [3]*GrammarUriContext{}
	qNameID := 0

	qncs := make([]*QNameContext, len(LocalNamesEmpty))
	contexts[0] = NewGrammarUriContext(0, EmptyString, qncs, PrefixesEmpty)

	qncs = make([]*QNameContext, len(LocalNamesXML))
	for i := 0; i < len(qncs); i++ {
		qncs[i] = NewQNameContext(1, i, utils.QName{Space: XML_NS_URI, Local: LocalNamesXML[i]})
		qNameID++
	}
	contexts[1] = NewGrammarUriContext(1, XML_NS_URI, qncs, PrefixesXML)

	qncs = make([]*QNameContext, len(LocalNamesXSI))
	for i := 0; i < len(qncs); i++ {
		qncs[i] = NewQNameContext(1, i, utils.QName{Space: XMLSchemaInstanceNS_URI, Local: LocalNamesXSI[i]})
		qNameID++
	}
	contexts[2] = NewGrammarUriContext(1, XMLSchemaInstanceNS_URI, qncs, PrefixesXSI)

	schemaLessGrammarContext = NewGrammarContext(contexts[:], qNameID)
}

func NewSchemaLessGrammars() *SchemaLessGrammars {
	schemaLessInit.Do(initSchemaLessGrammarContext)
	ag := NewAbstractGrammars(false, schemaLessGrammarContext)
	g := &SchemaLessGrammars{
		AbstractGrammars: ag,
	}
	g.Grammars = ag

	//builtInDocEndGrammar := NewDocEndWithLabel(utils.AsPtr("DocEnd"))
	//builtInDocEndGrammar.AddTerminalProduction(NewEndDocument())
	//buildInDocContentGrammar := NewBuiltInDocContentWithLabel(builtInDocEndGrammar, utils.AsPtr("DocContent"))
	//g.documentGrammar = NewDocumentWithLabel(utils.AsPtr("Document"))
	//g.documentGrammar.AddProduction(NewStartDocument(), buildInDocContentGrammar)

	builtInDocEndGrammar := NewDocEndWithLabel("DocEnd")
	builtInDocEndGrammar.AddTerminalProduction(NewEndDocument())
	buildInDocContentGrammar := NewBuiltInDocContentWithLabel(builtInDocEndGrammar, "DocContent")
	g.documentGrammar = NewDocumentWithLabel("Document")
	g.documentGrammar.AddProduction(NewStartDocument(), buildInDocContentGrammar)

	return g
}

func (g *SchemaLessGrammars) GetSchemaID() *string {
	return nil
}

func (g *SchemaLessGrammars) IsBuiltInXMLSchemaTypesOnly() bool {
	return false
}

func (g *SchemaLessGrammars) SetSchemaID(schemaID *string) error {
	if schemaID != nil {
		return fmt.Errorf("schema-less grammars do have schemaId == null associated with it")
	}

	return nil
}

/*
 * Note: create new instance since fragment content grammar may have been
 * changed over time
 */
func (g *SchemaLessGrammars) GetFragmentGrammar() Grammar {
	builtInFragmentContentGrammar := NewBuiltInFragmentContent()

	//g.fragmentGrammar = NewFragmentWithLabel(utils.AsPtr("Fragment"))
	g.fragmentGrammar = NewFragmentWithLabel("Fragment")
	g.fragmentGrammar.AddProduction(NewStartDocument(), builtInFragmentContentGrammar)

	return g.fragmentGrammar
}
