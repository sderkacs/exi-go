package core

import (
	"bufio"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/sderkacs/exi-go/utils"
)

type ProfileDisablingMechanism int

const (
	/** no disabling */
	ProfileDisablingMechanismNone ProfileDisablingMechanism = iota
	/**
	 * preferred mechanism: xsi:type attribute inserted. Note: current
	 * grammar changes & only possible right after SE
	 */
	ProfileDisablingMechanismXsiType
	/** 2nd mechanism: production has been inserted that is not usable */
	ProfileDisablingMechanismGhostProduction
)

type EXIBodyDecoder interface {
	// Sets the input stream and resets all internal states
	SetInputStream(reader *bufio.Reader) error

	// Sets input channel and resets all internal states
	SetInputChannel(channel DecoderChannel) error

	// Sets the input stream and does not reset internal states
	UpdateInputStream(reader *bufio.Reader) error

	// Sets input channel and does not reset internal states
	UpdateInputChannel(channel DecoderChannel) error

	// Reports the next available EXI event-type or 'nil' if no more
	// EXI event is available.
	Next() (EventType, bool, error)

	// Indicates the beginning of a set of XML events
	DecodeStartDocument() error

	// Indicates the end of a set of XML events
	DecodeEndDocument() error

	// Reads EXI start element and returns qualified name.
	DecodeStartElement() (*QNameContext, error)

	// Returns element prefix for current element as 'string'.
	GetElementPrefix() *string

	// Returns qualified name for element name as 'string'.
	//
	// QName ::= PrefixedName | UnprefixedName
	// PrefixedName ::= Prefix ':' LocalPart
	// UnprefixedName ::= LocalPart
	GetElementQNameAsString() string

	// Reads EXI a self-contained start element.
	DecodeStartSelfContainedFragment() error

	// Reads EXI end element and returns qualified name.
	DecodeEndElement() (*QNameContext, error)

	// Parses xsi:nil attribute
	DecodeAttributeXsiNil() (*QNameContext, error)

	// Parses xsi:type attribute
	DecodeAttributeXsiType() (*QNameContext, error)

	// Parses attribute and returns qualified name.
	DecodeAttribute() (*QNameContext, error)

	// Returns attribute prefix for (last) attribute as 'string'.
	GetAttributePrefix() *string

	// Returns qualified name for (last) attribute as 'string'.
	//
	// QName ::= PrefixedName | UnprefixedName
	// PrefixedName ::= Prefix ':' LocalPart
	// UnprefixedName ::= LocalPart
	GetAttributeQNameAsString() string

	// Provides attribute value
	GetAttributeValue() Value

	// Parses namespace declaration retrieving associated URI and prefix.
	DecodeNamespaceDeclaration() (*NamespaceDeclarationContainer, error)

	// Prefix declarations for current context (element)
	GetDeclaredPrefixDeclarations() []NamespaceDeclarationContainer

	// Decodes characters and reports them.
	DecodeCharacters() (Value, error)

	// Parses DOCTYPE with information items (name, publicID, systemID, text).
	DecodeDocType() (*DocTypeContainer, error)

	// Parses EntityReference and returns ER name.
	DecodeEntityReference() ([]rune, error)

	// Parses comment with associated characters and provides comment text.
	DecodeComment() ([]rune, error)

	// Parses processing instruction with associated target and data.
	DecodeProcessingInstruction() (ProcessingInstructionContainer, error)
}

type EXIBodyEncoder interface {
	SetOutputStream(writer bufio.Writer) error

	SetOutputChannel(channel EncoderChannel) error

	// Flushes (possibly) remaining bit(s) to output stream
	Flush() error

	SetErrorHandler(handler ErrorHandler)

	// Reports the beginning of a set of XML events
	EncodeStartDocument() error

	// Reports the end of a set of XML events.
	EncodeEndDocument() error

	// Supplies the start of an element.
	//
	// Provides access to the namespace URI, local name , and prefix
	// representation of the start tag.
	//
	// Element prefix can be null according to fidelity options.
	EncodeStartElement(uri, localName string, prefix *string) error
	EncodeStartElementByQName(se QName) error

	// Supplies the end tag of an element.
	EncodeEndElement() error

	// Supplies a list of namespace declarations, xsi:type and xsi:nil values
	// and the remaining attributes.
	EncodeAttributeList(attributes AttributeList) error

	// Supplies an attribute.
	EncodeAttribute(uri, localName string, prefix *string, value Value) error

	// Supplies an attribute with the according value.
	EncodeAttributeByQName(at QName, value Value) error

	// Namespaces are reported as a discrete Namespace event.
	EncodeNamespaceDeclaration(uri string, prefix *string) error

	// Supplies an xsi:nil attribute.
	EncodeAttributeXsiNil(nilValue Value, prefix *string) error

	// Supplies an xsi:type case.
	EncodeAttributeXsiType(typeValue Value, prefix *string) error

	// Supplies characters as Value.
	EncodeCharacters(chars Value) error

	// Supplies content items to represent a DOCTYPE definition
	EncodeDocType(name, publicID, systemID, text string) error

	// Supplies the name of an entity reference
	EncodeEntityReference(name string) error

	// Supplies the text of a comment.
	EncodeComment(ch []rune, start, length int) error

	// Supplies the target and data for an underlying processing instruction.
	EncodeProcessingInstruction(target, data string) error
}

type EXIStreamDecoder interface {
	GetBodyOnlyDecoder(reader *bufio.Reader) (EXIBodyDecoder, error)
	DecodeHeader(reader *bufio.Reader) (EXIBodyDecoder, error)
}

type EXIStreamEncoder interface {
	EncodeHeader(writer bufio.Writer) (EXIBodyEncoder, error)
}

/*
	ElementContext implementation
*/

type ElementContext struct {
	prefix             *string
	sqname             string
	gr                 Grammar
	nsDeclarations     []NamespaceDeclarationContainer
	isXMLSpacePreserve *bool
	qnc                *QNameContext
}

func NewElementContext(qnc *QNameContext, gr Grammar) *ElementContext {
	return &ElementContext{
		prefix:             nil,
		sqname:             "",
		gr:                 gr,
		nsDeclarations:     []NamespaceDeclarationContainer{},
		isXMLSpacePreserve: nil,
		qnc:                qnc,
	}
}

func (c *ElementContext) GetQNameAsString(preservePrefix bool) string {
	if c.sqname == "" {
		if preservePrefix {
			c.sqname = utils.GetQualifiedName(c.qnc.GetLocalName(), c.prefix)
		} else {
			c.sqname = c.qnc.GetDefaultQNameAsString()
		}
	}
	return c.sqname
}

func (c *ElementContext) SetPrefix(prefix *string) {
	c.prefix = prefix
}

func (c *ElementContext) GetPrefix() *string {
	return c.prefix
}

func (c *ElementContext) SetXMLSpacePreserve(isXMLSpacePreserve *bool) {
	c.isXMLSpacePreserve = isXMLSpacePreserve
}

func (c *ElementContext) IsXMLSpacePreserve() *bool {
	return c.isXMLSpacePreserve
}

/*
	RuntimeUriContext implementation
*/

type RuntimeUriContext struct {
	namespaceUriID int
	namespaceURI   string
	guc            *GrammarUriContext

	qnames   []*QNameContext
	prefixes []string
}

func NewRuntimeUriContext(namespaceUriID int, namespaceURI string) *RuntimeUriContext {
	return NewRuntimeUriContextWithContext(nil, namespaceUriID, namespaceURI)
}

func NewRuntimeUriContextWithContext(guc *GrammarUriContext, namespaceUriID int, namespaceURI string) *RuntimeUriContext {
	return &RuntimeUriContext{
		namespaceUriID: namespaceUriID,
		namespaceURI:   namespaceURI,
		guc:            guc,
		qnames:         []*QNameContext{},
		prefixes:       []string{},
	}
}

func RuntimeUriContextFromContext(guc *GrammarUriContext) *RuntimeUriContext {
	return NewRuntimeUriContextWithContext(guc, guc.GetNamespaceUriID(), guc.GetNamespaceUri())
}

func (c *RuntimeUriContext) clear(preservePrefix bool) {
	if c.guc == nil {
		c.namespaceURI = ""
	}

	// Note: re-use existing lists for subsequent runs
	if len(c.qnames) > 0 {
		c.qnames = []*QNameContext{}
	}
	if preservePrefix && len(c.prefixes) > 0 {
		c.prefixes = []string{}
	}
}

func (c *RuntimeUriContext) GetQNameContextByLocalName(localName string) *QNameContext {
	var qnc *QNameContext = nil
	if c.guc != nil {
		qnc = c.guc.GetQNameContextByLocalName(localName)
	}
	if qnc == nil {
		// check runtime qnames
		if len(c.qnames) != 0 {
			for i := len(c.qnames) - 1; i >= 0; i-- {
				qnc = c.qnames[i]
				if qnc.GetLocalName() == localName {
					return qnc
				}
			}
			qnc = nil // none found
		}
	}

	return qnc
}

func (c *RuntimeUriContext) GetQNameContextByLocalNameID(localNameID int) *QNameContext {
	var qnc *QNameContext = nil
	sub := 0
	if c.guc != nil {
		qnc = c.guc.GetQNameContextByLocalNameID(localNameID)
		sub = c.guc.GetNumberOfQNames()
	}
	if qnc == nil {
		// check runtime qnames
		if len(c.qnames) != 0 {
			localNameID -= sub
			if localNameID < 0 || localNameID >= len(c.qnames) {
				panic("index out of bounds")
			}
			qnc = c.qnames[localNameID]
		}
	}

	return qnc
}

func (c *RuntimeUriContext) GetNumberOfQNames() int {
	n := 0
	if c.guc != nil {
		n = c.guc.GetNumberOfQNames()
	}
	n += len(c.qnames)
	return n
}

func (c *RuntimeUriContext) AddQNameContext(localName string) *QNameContext {
	localNameID := c.GetNumberOfQNames()
	qName := QName{Space: c.namespaceURI, Local: localName}
	qnc := NewQNameContext(c.namespaceUriID, localNameID, qName)
	c.qnames = append(c.qnames, qnc)

	return qnc
}

func (c *RuntimeUriContext) GetNumberOfPrefixes() int {
	pfs := 0
	if c.guc != nil {
		pfs = c.guc.GetNumberOfPrefixes()
	}

	pfs += len(c.prefixes)

	return pfs
}

func (c *RuntimeUriContext) addPrefix(prefix string) {
	c.prefixes = append(c.prefixes, prefix)
}

func (c *RuntimeUriContext) getPrefixID(prefix string) int {
	id := NotFound
	sub := 0

	if c.guc != nil {
		id = c.guc.GetPrefixID(prefix)
		sub = c.guc.GetNumberOfPrefixes()
	}
	if id == NotFound {
		for i := range len(c.prefixes) {
			if c.prefixes[i] == prefix {
				return i + sub
			}
		}
	}

	return id
}

func (c *RuntimeUriContext) GetPrefix(prefixID int) *string {
	//TODO: checks for preservePrefix
	var prefix *string = nil
	sub := 0

	if c.guc != nil {
		prefix = c.guc.GetPrefix(prefixID)
		sub = c.guc.GetNumberOfPrefixes()
	}
	if prefix == nil {
		prefixID -= sub
		if prefixID < 0 || prefixID >= len(c.prefixes) {
			panic("index out of bounds")
		}
		prefix = &c.prefixes[prefixID]
	}

	return prefix
}

func (c *RuntimeUriContext) SetNamespaceUri(namespaceURI string) {
	c.namespaceURI = namespaceURI
}

func (c *RuntimeUriContext) GetNamespaceUri() string {
	return c.namespaceURI
}

func (c *RuntimeUriContext) GetNamespaceUriID() int {
	return c.namespaceUriID
}

/*
	AbstractEXIBodyCoder implementation
*/

const (
	ElementContextsInitialStackSize = 16
)

type AbstractEXIBodyCoder struct {
	exiFactory                EXIFactory
	grammar                   Grammars
	grammarContext            *GrammarContext
	fidelityOptions           *FidelityOptions
	preservePrefix            bool
	preserveLexicalValues     bool
	errorHandler              ErrorHandler
	booleanDatatype           *BooleanDatatype  // Boolean datatype (coder)
	elementContext            *ElementContext   // element-context and rule (stack) while traversing the EXI document
	elementContextStack       []*ElementContext // cached context to avoid heavy array lookup
	elementContextStackIndex  int
	runtimeGlobalElements     map[QNameContextMapKey]*StartElement // runtime global elements
	runtimeURIs               []*RuntimeUriContext
	xsiTypeContext            *QNameContext
	xsiNilContext             *QNameContext
	gURIs                     int // number of grammar uris
	nextUriID                 int
	limitGrammarLearning      bool
	maxBuiltInElementGrammars int
	maxBuiltInProductions     int
	learnedProductions        int
}

func NewAbstractEXIBodyCoder(exiFactory EXIFactory) (*AbstractEXIBodyCoder, error) {
	grammar := exiFactory.GetGrammars()
	grammarContext := grammar.GetGrammarContext()
	fidelityOptions := exiFactory.GetFidelityOptions()
	gURIs := grammarContext.GetNumberOfGrammarUriContexts()

	preservePrefix := fidelityOptions.IsFidelityEnabled(FeaturePrefix)
	preserveLexicalValues := fidelityOptions.IsFidelityEnabled(FeatureLexicalValue)

	runtimeURIs := make([]*RuntimeUriContext, gURIs)
	for i := range gURIs {
		ctx := grammarContext.GetGrammarUriContextByID(i)
		fmt.Printf("[DEBUG] ctx[%d] == %+v\n", i, ctx)
		runtimeURIs[i] = RuntimeUriContextFromContext(ctx)
	}

	var maxBuiltInElementGrammars int
	var maxBuiltInProductions int
	var limitGrammarLearning bool

	if grammar.IsSchemaInformed() {
		maxBuiltInElementGrammars = exiFactory.GetMaximumNumberOfBuiltInElementGrammars()
		maxBuiltInProductions = exiFactory.GetMaximumNumberOfBuiltInProductions()
		limitGrammarLearning = (maxBuiltInElementGrammars >= 0) || (maxBuiltInProductions >= 0)
	} else {
		maxBuiltInElementGrammars = -1
		maxBuiltInProductions = -1
		limitGrammarLearning = false
	}

	return &AbstractEXIBodyCoder{
		exiFactory:                exiFactory,
		grammar:                   grammar,
		grammarContext:            grammarContext,
		fidelityOptions:           fidelityOptions,
		preservePrefix:            preservePrefix,
		preserveLexicalValues:     preserveLexicalValues,
		errorHandler:              NewDefaultErrorHandler(),
		booleanDatatype:           NewBooleanDatatype(nil),
		elementContext:            nil,
		elementContextStack:       make([]*ElementContext, ElementContextsInitialStackSize),
		elementContextStackIndex:  0,
		runtimeGlobalElements:     map[QNameContextMapKey]*StartElement{},
		runtimeURIs:               runtimeURIs,
		xsiTypeContext:            nil,
		xsiNilContext:             nil,
		gURIs:                     gURIs,
		nextUriID:                 gURIs,
		limitGrammarLearning:      limitGrammarLearning,
		maxBuiltInElementGrammars: maxBuiltInElementGrammars,
		maxBuiltInProductions:     maxBuiltInProductions,
		learnedProductions:        0,
	}, nil
}

func (c *AbstractEXIBodyCoder) getXsiTypeContext() *QNameContext {
	if c.xsiTypeContext == nil {
		c.xsiTypeContext = c.grammarContext.GetGrammarUriContextByID(2).GetQNameContextByLocalNameID(1)
	}
	return c.xsiTypeContext
}

func (c *AbstractEXIBodyCoder) getXsiNilContext() *QNameContext {
	if c.xsiNilContext == nil {
		c.xsiNilContext = c.grammarContext.GetGrammarUriContextByID(2).GetQNameContextByLocalNameID(0)
	}
	return c.xsiNilContext
}

func (c *AbstractEXIBodyCoder) isBuiltInStartTagGrammarWithAtXsiTypeOnly(g Grammar) bool {
	if g.GetNumberOfEvents() == 1 {
		p0 := g.GetProductionByEventCode(0)
		ev0 := p0.GetEvent()

		if ev0.IsEventType(EventTypeAttribute) {
			at := ev0.(*Attribute)
			qn0 := at.GetQNameContext()

			if qn0.GetNamespaceUriID() == 2 && qn0.GetLocalNameID() == 1 {
				// AT type cast only
				return true
			}
		}
	}

	return false
}

func (c *AbstractEXIBodyCoder) getGlobalStartElement(qnc *QNameContext) *StartElement {
	se := qnc.GetGlobalStartElement()
	if se == nil {
		// no global StartElement stemming from schema-informed grammars
		// --> check for previous runtime SE
		se = c.runtimeGlobalElements[qnc.GetMapKey()]
		if se == nil {
			// no global runtime grammar yet
			se = NewStartElement(qnc)
			//TODO: which grammar to pick if no schema-information are available?
			if c.grammar.IsSchemaInformed() && c.exiFactory.IsUsingNonEvolvingGrammars() {
				sig := c.grammar.(*SchemaInformedGrammars)
				se.SetGrammar(sig.GetSchemaInformedElementFragmentGrammar())
			} else {
				se.SetGrammar(NewBuiltInStartTag())
			}

			c.runtimeGlobalElements[qnc.GetMapKey()] = se
		}
	}

	return se
}

func (c *AbstractEXIBodyCoder) getCurrentGrammar() Grammar {
	return c.elementContext.gr
}

func (c *AbstractEXIBodyCoder) updateCurrentRule(newCurrentGrammar Grammar) {
	c.elementContext.gr = newCurrentGrammar
}

func (c *AbstractEXIBodyCoder) getElementContext() *ElementContext {
	return c.elementContext
}

func (c *AbstractEXIBodyCoder) updateElementContext(elementContext *ElementContext) {
	c.elementContext = elementContext
}

func (c *AbstractEXIBodyCoder) SetErrorHandler(handler ErrorHandler) {
	c.errorHandler = handler
}

// re-init (rule stack etc)
func (c *AbstractEXIBodyCoder) InitForEachRun() error {
	// clear runtime data
	c.runtimeGlobalElements = map[QNameContextMapKey]*StartElement{}
	for i := range c.nextUriID {
		c.runtimeURIs[i].clear(c.preservePrefix)
	}

	// re-set schema-informed grammar IDs
	c.nextUriID = c.gURIs

	// possible document/fragment grammar
	var startRule Grammar
	if c.exiFactory.IsFragment() {
		startRule = c.grammar.GetFragmentGrammar()
	} else {
		startRule = c.grammar.GetDocumentGrammar()
	}

	// (core) context
	ec := NewElementContext(nil, startRule)

	c.elementContextStackIndex = 0
	c.elementContextStack[0] = ec
	c.elementContext = ec

	return nil
}

func (c *AbstractEXIBodyCoder) declarePrefix(prefix *string, uri string) {
	c.declarePrefixWithNamespaceDeclaraion(NewNamespaceDeclarationContainer(uri, prefix))
}

func (c *AbstractEXIBodyCoder) declarePrefixWithNamespaceDeclaraion(nsDecl NamespaceDeclarationContainer) {
	if slices.Contains(c.elementContext.nsDeclarations, nsDecl) {
		panic("multiple equal namespace declarations")
	}

	c.elementContext.nsDeclarations = append(c.elementContext.nsDeclarations, nsDecl)
}

func (c *AbstractEXIBodyCoder) getURI(prefix *string) *string {
	for i := c.elementContextStackIndex; i > 0; i-- {
		ec := c.elementContextStack[i]

		for k := range len(ec.nsDeclarations) {
			ns := ec.nsDeclarations[k]
			if ns.prefix == prefix || (ns.prefix != nil && prefix != nil && *ns.prefix == *prefix) {
				return &ns.namespaceURI
			}
		}
	}

	if prefix == nil || len(*prefix) == 0 {
		return utils.AsPtr(XMLNullNS_URI)
	} else {
		return nil
	}
}

func (c *AbstractEXIBodyCoder) getPrefix(uri string) *string {
	for i := c.elementContextStackIndex; i > 0; i-- {
		ec := c.elementContextStack[i]

		for k := range len(ec.nsDeclarations) {
			ns := ec.nsDeclarations[k]
			if ns.namespaceURI == uri {
				return ns.prefix
			}
		}
	}

	return nil
}

func (c *AbstractEXIBodyCoder) pushElement(updContextGrammar Grammar, se *StartElement) {
	// update "rule" item of current peak (for popElement() later on)
	c.elementContext.gr = updContextGrammar

	// check element context array size
	c.elementContextStackIndex++
	if len(c.elementContextStack) == c.elementContextStackIndex {
		elementContextStackNew := make([]*ElementContext, len(c.elementContextStack)<<2)
		copy(elementContextStackNew, c.elementContextStack)
		c.elementContextStack = elementContextStackNew
	}

	// create new stack item & push it
	c.elementContext = NewElementContext(se.GetQNameContext(), se.GetGrammar())
	c.elementContextStack[c.elementContextStackIndex] = c.elementContext
}

func (c *AbstractEXIBodyCoder) popElement() *ElementContext {
	if c.elementContextStackIndex < 0 {
		panic("index out of bounds")
	}

	// pop element from stack
	poppedEC := c.elementContextStack[c.elementContextStackIndex]
	c.elementContextStack[c.elementContextStackIndex] = nil
	c.elementContextStackIndex--
	c.elementContext = c.elementContextStack[c.elementContextStackIndex]

	return poppedEC
}

func (c *AbstractEXIBodyCoder) addUri(uri string) *RuntimeUriContext {
	var ruc *RuntimeUriContext
	uriID := c.nextUriID
	c.nextUriID++

	if uriID < len(c.runtimeURIs) {
		// re-use existing entry
		ruc = c.runtimeURIs[uriID]
		// Update namespace uri (ID is already ok)
		ruc.SetNamespaceUri(uri)
	} else {
		// create new uri entry
		ruc = NewRuntimeUriContext(uriID, uri)
		c.runtimeURIs = append(c.runtimeURIs, ruc)
	}

	return ruc
}

func (c *AbstractEXIBodyCoder) GetNumberOfURIs() int {
	return c.nextUriID
}

func (c *AbstractEXIBodyCoder) GetURI(namespaceURI string) *RuntimeUriContext {
	for i := 0; i < c.nextUriID && i < len(c.runtimeURIs); i++ {
		ruc := c.runtimeURIs[i]
		if ruc.namespaceURI == namespaceURI {
			return ruc
		}
	}

	return nil
}

func (c *AbstractEXIBodyCoder) GetURIByNamespaceID(namespaceUriID int) *RuntimeUriContext {
	if namespaceUriID < 0 || namespaceUriID >= len(c.runtimeURIs) {
		panic("index out of bounds")
	}
	return c.runtimeURIs[namespaceUriID]
}

func (c *AbstractEXIBodyCoder) emitWarning(message string) {
	c.errorHandler.Warning(fmt.Errorf("%s, options = %+v", message, c.fidelityOptions))
}

/*
	AbstractEXIBodyDecoder implementation
*/

type AbstractEXIBodyDecoder struct {
	*AbstractEXIBodyCoder
	nextEvent             Event // next event
	nextGrammar           Grammar
	nextEventType         EventType
	channel               DecoderChannel // decoder stream
	numberOfUriContexts   int            // namespaces/prefixes
	typeDecoder           TypeDecoder    // type decoder
	stringDecoder         StringDecoder  // string decoder
	attributeQNameContext *QNameContext
	attributePrefix       *string
	attributeValue        Value
}

func NewAbstractEXIBodyDecoder(exiFactory EXIFactory) (*AbstractEXIBodyDecoder, error) {
	bc, err := NewAbstractEXIBodyCoder(exiFactory)
	if err != nil {
		return nil, err
	}
	decoder, err := exiFactory.CreateTypeDecoder()
	if err != nil {
		return nil, err
	}

	return &AbstractEXIBodyDecoder{
		AbstractEXIBodyCoder:  bc,
		nextEvent:             nil,
		nextGrammar:           nil,
		nextEventType:         -1,
		channel:               nil,
		numberOfUriContexts:   bc.grammar.GetGrammarContext().GetNumberOfGrammarUriContexts(),
		typeDecoder:           decoder,
		stringDecoder:         exiFactory.CreateStringDecoder(),
		attributeQNameContext: nil,
		attributePrefix:       nil,
		attributeValue:        nil,
	}, nil
}

func (d *AbstractEXIBodyDecoder) pushElement(updContextGrammar Grammar, se *StartElement) {
	d.AbstractEXIBodyCoder.pushElement(updContextGrammar, se)

	if !d.preservePrefix && d.elementContextStackIndex == 1 {
		// Note: can be done several times due to multiple root elements in fragments.
		gc := d.grammar.GetGrammarContext()
		for i := 2; i < gc.GetNumberOfGrammarUriContexts(); i++ {
			guc := gc.GetGrammarUriContextByID(i)
			prefix := guc.GetDefaultPrefix()
			d.declarePrefix(&prefix, guc.GetNamespaceUri())
		}
	}
}

func (d *AbstractEXIBodyDecoder) InitForEachRun() error {
	if err := d.AbstractEXIBodyCoder.InitForEachRun(); err != nil {
		return err
	}

	d.stringDecoder.Clear()
	if d.exiFactory.GetSharedStrings() != nil {
		d.stringDecoder.SetSharedStrings(*d.exiFactory.GetSharedStrings())
	}

	return nil
}

func (d *AbstractEXIBodyDecoder) decodeQName(channel DecoderChannel) (*QNameContext, error) {
	// decode uri & local-name
	ruc, err := d.decodeURI(channel)
	if err != nil {
		return nil, err
	}
	return d.decodeLocalName(ruc, channel)
}

func (d *AbstractEXIBodyDecoder) decodeURI(channel DecoderChannel) (*RuntimeUriContext, error) {
	numberBitsURI := utils.GetCodingLength(d.GetNumberOfURIs() + 1)
	uriID, err := channel.DecodeNBitUnsignedInteger(numberBitsURI)
	if err != nil {
		return nil, err
	}

	var ruc *RuntimeUriContext
	if uriID == 0 {
		// string value was not found
		// ==> zero (0) as an n-nit unsigned integer
		// followed by uri encoded as string
		uriRunes, err := channel.DecodeString()
		if err != nil {
			return nil, err
		}
		ruc = d.addUri(string(uriRunes))
	} else {
		// string value found
		// ==> value(i+1) is encoded as n-bit unsigned integer
		uriID--
		ruc = d.GetURIByNamespaceID(uriID)
	}

	return ruc, nil
}

func (d *AbstractEXIBodyDecoder) decodeLocalName(ruc *RuntimeUriContext, channel DecoderChannel) (*QNameContext, error) {
	length, err := channel.DecodeUnsignedInteger()
	if err != nil {
		return nil, err
	}

	var qnc *QNameContext
	if length > 0 {
		// string value was not found in local partition
		// ==> string literal is encoded as a String
		// with the length of the string incremented by one
		runes, err := channel.DecodeStringOnly(length - 1)
		if err != nil {
			return nil, err
		}
		// After encoding the string value, it is added to the string table
		// partition and assigned the next available compact identifier.
		qnc = ruc.AddQNameContext(string(runes))
	} else {
		// string value found in local partition
		// ==> string value is represented as zero (0) encoded as an
		// Unsigned Integer followed by an the compact identifier of the
		// string value as an n-bit unsigned integer n is log2 m and m is
		// the number of entries in the string table partition.
		n := utils.GetCodingLength(ruc.GetNumberOfQNames())
		localNameID, err := channel.DecodeNBitUnsignedInteger(n)
		if err != nil {
			return nil, err
		}
		qnc = ruc.GetQNameContextByLocalNameID(localNameID)
	}

	return qnc, nil
}

func (d *AbstractEXIBodyDecoder) decodeQNamePrefix(ruc *RuntimeUriContext, channel DecoderChannel) (*string, error) {
	var prefix *string = nil

	if ruc.namespaceUriID == 0 {
		prefix = utils.AsPtr(XMLNullNS_URI)
	} else {
		numberOfPrefixes := ruc.GetNumberOfPrefixes()
		if numberOfPrefixes > 0 {
			id := 0
			if numberOfPrefixes > 1 {
				tmp, err := channel.DecodeNBitUnsignedInteger(utils.GetCodingLength(numberOfPrefixes))
				if err != nil {
					return nil, err
				}
				id = tmp
			}

			prefix = ruc.GetPrefix(id)
		} else {
			// no previous NS mapping in charge
			// Note: should only happen for SE events where NS appears afterwards.
		}
	}

	return prefix, nil
}

func (d *AbstractEXIBodyDecoder) decodeNamespacePrefix(ruc *RuntimeUriContext, channel DecoderChannel) (*string, error) {
	var prefix *string

	nPfx := utils.GetCodingLength(ruc.GetNumberOfPrefixes() + 1)
	pfxID, err := channel.DecodeNBitUnsignedInteger(nPfx)
	if err != nil {
		return nil, err
	}

	if pfxID == 0 {
		// string value was not found
		// ==> zero (0) as an n-nit unsigned integer
		// followed by pfx encoded as string
		runes, err := channel.DecodeString()
		if err != nil {
			return nil, err
		}
		prefix = utils.AsPtr(string(runes))

		ruc.addPrefix(string(runes))
	} else {
		// string value found
		// ==> value(i+1) is encoded as n-bit unsigned integer
		prefix = ruc.GetPrefix(pfxID - 1)
	}

	return prefix, nil
}

func (d *AbstractEXIBodyDecoder) decodeEventCode() (EventType, error) {
	// 1st level
	currentGrammar := d.getCurrentGrammar()
	codeLength := d.fidelityOptions.Get1stLevelEventCodeLength(currentGrammar)
	ec, err := d.channel.DecodeNBitUnsignedInteger(codeLength)
	if err != nil {
		return -1, err
	}

	if ec < 0 {
		return -1, fmt.Errorf("invalid 1st level error code: %d", ec)
	}

	if ec < currentGrammar.GetNumberOfEvents() {
		// 1st level
		ei := currentGrammar.GetProductionByEventCode(ec)
		d.nextEvent = ei.GetEvent()
		d.nextGrammar = ei.GetNextGrammar()
		d.nextEventType = d.nextEvent.GetEventType()
	} else {
		// 2nd level ?
		ec2, err := d.decode2ndLevelEventCode()
		if err != nil {
			return -1, err
		}

		if ec2 == NotFound {
			// 3rd level
			ec3, err := d.decode3rdLevelEventCode()
			if err != nil {
				return -1, err
			}
			d.nextEventType = d.fidelityOptions.Get3rdLevelEventType(ec3)

			// unset events
			d.nextEvent = nil
			d.nextGrammar = nil
		} else {
			d.nextEventType = d.fidelityOptions.Get2ndLevelEventType(ec2, currentGrammar)

			if d.nextEventType == EventTypeAttributeInvalidValue {
				if err := d.updateInvalidValueAttribute(ec); err != nil {
					return -1, err
				}
			} else {
				// unset events
				d.nextEvent = nil
				d.nextGrammar = nil
			}
		}
	}

	return d.nextEventType, nil
}

func (d *AbstractEXIBodyDecoder) GetAttributePrefix() *string {
	return d.attributePrefix
}

func (d *AbstractEXIBodyDecoder) GetAttributeQNameAsString() string {
	if d.preservePrefix {
		return utils.GetQualifiedName(d.attributeQNameContext.GetLocalName(), d.attributePrefix)
	} else {
		return d.attributeQNameContext.GetDefaultQNameAsString()
	}
}

func (d *AbstractEXIBodyDecoder) GetAttributeValue() Value {
	return d.attributeValue
}

func (d *AbstractEXIBodyDecoder) updateInvalidValueAttribute(ec int) error {
	sir := d.getCurrentGrammar().(SchemaInformedGrammar)

	ec3AT, err := d.channel.DecodeNBitUnsignedInteger(utils.GetCodingLength(sir.GetNumberOfDeclaredAttributes() + 1))
	if err != nil {
		return err
	}

	if ec3AT < sir.GetNumberOfDeclaredAttributes() {
		// deviated attribute
		ec = ec3AT + sir.GetLeastAttributeEventCode()
		ei := sir.GetProductionByEventCode(ec)

		d.nextEvent = ei.GetEvent()
		d.nextGrammar = ei.GetNextGrammar()
	} else if ec3AT == sir.GetNumberOfDeclaredAttributes() {
		// ANY deviated attribute (no qname present)
		d.nextEventType = EventTypeAttributeAnyInvalidValue
	} else {
		return errors.New("error occured while decoding deviated attribute")
	}

	return nil
}

func (d *AbstractEXIBodyDecoder) decode2ndLevelEventCode() (int, error) {
	currentGrammar := d.getCurrentGrammar()
	ch2 := d.fidelityOptions.Get2ndLevelCharacteristics(currentGrammar)

	level2, err := d.channel.DecodeNBitUnsignedInteger(utils.GetCodingLength(ch2))
	if err != nil {
		return -1, err
	}

	ch3 := d.fidelityOptions.Get3rdLevelCharacteristics()

	if ch3 > 0 {
		if level2 < ch2-1 {
			return level2, nil
		} else {
			return NotFound, nil
		}
	} else {
		if level2 < ch2 {
			return level2, nil
		} else {
			return NotFound, nil
		}
	}
}

func (d *AbstractEXIBodyDecoder) decode3rdLevelEventCode() (int, error) {
	ch3 := d.fidelityOptions.Get3rdLevelCharacteristics()
	return d.channel.DecodeNBitUnsignedInteger(utils.GetCodingLength(ch3))
}

func (d *AbstractEXIBodyDecoder) decodeStartDocumentStructure() error {
	d.updateCurrentRule(d.getCurrentGrammar().GetProductionByEventCode(0).GetNextGrammar())
	return nil
}

func (d *AbstractEXIBodyDecoder) decodeEndDocumentStructure() error {
	if d.limitGrammarLearning {
		if d.maxBuiltInElementGrammars != -1 {
			evolvedGrs := 0

			for _, se := range d.runtimeGlobalElements {
				stg := se.GetGrammar()
				if stg.GetGrammarType() != GrammarTypeBuiltInStartTagContent {
					return fmt.Errorf("invalid start element grammar type: %d", stg.GetGrammarType())
				}
				ecg := stg.GetElementContentGrammar()
				if ecg.GetGrammarType() != GrammarTypeBuiltInElementContent {
					return fmt.Errorf("invalid built-in element content grammar type: %d", ecg.GetGrammarType())
				}

				if ecg.GetNumberOfEvents() != 1 {
					// BuiltIn Element Content grammar has EE per default
					evolvedGrs++
				} else {
					if stg.GetNumberOfEvents() > 1 {
						evolvedGrs++
					} else if stg.GetNumberOfEvents() == 1 {
						// check for AT(xsi:type)
						if !d.isBuiltInStartTagGrammarWithAtXsiTypeOnly(stg) {
							evolvedGrs++
						}
					}
				}
			}

			if evolvedGrs > d.maxBuiltInElementGrammars {
				return fmt.Errorf("EXI profile stream does not respect parameter maxBuiltInElementGrammars. Expected %d but was %d", d.maxBuiltInElementGrammars, evolvedGrs)
			}
		}
	}

	return nil
}

func (d *AbstractEXIBodyDecoder) decodeStartElementStructure() (*QNameContext, error) {
	if d.nextEventType != EventTypeStartElement {
		return nil, fmt.Errorf("next event type is not start element: %d", d.nextEventType)
	}
	se := d.nextEvent.(*StartElement)
	// push element
	d.pushElement(d.nextGrammar, se)
	// handle element prefix
	qnc := se.GetQNameContext()
	if err := d.handleElementPrefix(qnc); err != nil {
		return nil, err
	}

	return qnc, nil
}

func (d *AbstractEXIBodyDecoder) decodeStartElementNSStructure() (*QNameContext, error) {
	if d.nextEventType != EventTypeStartElementNS {
		return nil, fmt.Errorf("next event type is not start element NS: %d", d.nextEventType)
	}

	seNS := d.nextEvent.(*StartElementNS)

	// decode local-name
	ruc := d.GetURIByNamespaceID(seNS.GetNamespaceUriID())
	qnc, err := d.decodeLocalName(ruc, d.channel)
	if err != nil {
		return nil, err
	}

	nextSE := d.getGlobalStartElement(qnc)

	// push element
	d.pushElement(d.nextGrammar, nextSE)
	// handle element prefix
	if err := d.handleElementPrefix(qnc); err != nil {
		return nil, err
	}

	return qnc, nil
}

func (d *AbstractEXIBodyDecoder) decodeStartElementGenericStructure() (*QNameContext, error) {
	if d.nextEventType != EventTypeStartElementGeneric {
		return nil, fmt.Errorf("next event type is not start element generic: %d", d.nextEventType)
	}

	qnc, err := d.decodeQName(d.channel)
	if err != nil {
		return nil, err
	}

	nextSE := d.getGlobalStartElement(qnc)

	// learn start-element, necessary for FragmentContent grammar
	d.getCurrentGrammar().LearnStartElement(nextSE)
	// push element
	d.pushElement(d.nextGrammar.GetElementContentGrammar(), nextSE)

	// handle element prefix
	if err := d.handleElementPrefix(qnc); err != nil {
		return nil, err
	}

	return qnc, nil
}

func (d *AbstractEXIBodyDecoder) decodeStartElementGenericUndeclaredStructure() (*QNameContext, error) {
	if d.nextEventType != EventTypeStartElementGenericUndeclared {
		return nil, fmt.Errorf("next event type is not start element generic undeclared: %d", d.nextEventType)
	}

	qnc, err := d.decodeQName(d.channel)
	if err != nil {
		return nil, err
	}

	nextSE := d.getGlobalStartElement(qnc)

	// learn start-element ?
	currentGrammar := d.getCurrentGrammar()
	currentGrammar.LearnStartElement(nextSE)

	// push element
	d.pushElement(d.nextGrammar.GetElementContentGrammar(), nextSE)

	// handle element prefix
	if err := d.handleElementPrefix(qnc); err != nil {
		return nil, err
	}

	return qnc, nil
}

func (d *AbstractEXIBodyDecoder) decodeEndElementStructure() (*ElementContext, error) {
	return d.popElement(), nil
}

func (d *AbstractEXIBodyDecoder) decodeEndElementUndeclaredStructure() (*ElementContext, error) {
	d.getCurrentGrammar().LearnEndElement()
	return d.popElement(), nil
}

// Handles and xsi:nil attributes
func (d *AbstractEXIBodyDecoder) decodeAttributeXsiNilStructure() error {
	d.attributeQNameContext = d.getXsiNilContext()
	// handle AT prefix
	if err := d.handleAttributePrefix(d.attributeQNameContext); err != nil {
		return err
	}

	if d.preserveLexicalValues {
		// as string
		value, err := d.typeDecoder.ReadValue(d.booleanDatatype, d.getXsiNilContext(), d.channel, d.stringDecoder)
		if err != nil {
			return err
		}
		d.attributeValue = value
	} else {
		// as boolean
		value, err := d.channel.DecodeBooleanValue()
		if err != nil {
			return err
		}
		d.attributeValue = value
	}

	xsiNil := false

	bv, ok := d.attributeValue.(*BooleanValue)
	if ok {
		xsiNil = bv.ToBoolean()
	} else {
		// parse string value again (lexical value mode)
		// note: Java source also have this weird code...
		bv, ok = d.attributeValue.(*BooleanValue)
		if ok {
			xsiNil = bv.ToBoolean()
		} else {
			s, err := d.attributeValue.ToString()
			if err != nil {
				return err
			}
			bv = BooleanValueParse(s)
			if bv != nil {
				xsiNil = bv.ToBoolean()
			}
		}
	}

	currentGrammar := d.getCurrentGrammar()
	if xsiNil && currentGrammar.IsSchemaInformed() {
		// jump to typeEmpty
		te, err := currentGrammar.(SchemaInformedFirstStartTagGrammar).GetTypeEmpty()
		if err != nil {
			return err
		}
		d.updateCurrentRule(te)
	}

	return nil
}

// Handles and xsi:type attributes
func (d *AbstractEXIBodyDecoder) decodeAttributeXsiTypeStructure() error {
	d.attributeQNameContext = d.getXsiTypeContext()
	// handle AT prefix
	if err := d.handleAttributePrefix(d.attributeQNameContext); err != nil {
		return err
	}

	var qnc *QNameContext = nil

	// read xsi:type content
	if d.preserveLexicalValues {
		// assert(preservePrefix); // Note: requirement
		tmp, err := d.typeDecoder.ReadValue(BuiltInGetDefaultDatatype(), d.getXsiTypeContext(), d.channel, d.stringDecoder)
		if err != nil {
			return err
		}
		d.attributeValue = tmp

		sType, err := d.attributeValue.ToString()
		if err != nil {
			return err
		}

		// extract prefix
		qncTypePrefix := utils.GetPrefixPart(sType)

		// URI
		qnameURI := d.getURI(&qncTypePrefix)
		ruc := d.GetURI(*qnameURI)
		if ruc != nil {
			// local-name
			qnameLocalName := utils.GetLocalPart(sType)
			qnc = ruc.GetQNameContextByLocalName(qnameLocalName)
		}
	} else {
		// typed
		tmp, err := d.decodeQName(d.channel)
		if err != nil {
			return err
		}
		qnc = tmp

		var qncTypePrefix *string
		if d.preservePrefix {
			tmp, err := d.decodeQNamePrefix(d.GetURIByNamespaceID(qnc.GetNamespaceUriID()), d.channel)
			if err != nil {
				return err
			}
			qncTypePrefix = tmp
		} else {
			d.checkDefaultPrefixNamespaceDeclaration(qnc)
			qncTypePrefix = utils.AsPtr(qnc.GetDefaultPrefix())
		}
		d.attributeValue = NewQNameValue(qnc.GetNamespaceUri(), qnc.GetLocalName(), qncTypePrefix)
	}

	if qnc != nil && qnc.GetTypeGrammar() != nil {
		// update current rule
		d.updateCurrentRule(qnc.GetTypeGrammar())
	}

	return nil
}

func (d *AbstractEXIBodyDecoder) handleElementPrefix(qnc *QNameContext) error {
	var pfx *string

	if d.preservePrefix {
		tmp, err := d.decodeQNamePrefix(d.GetURIByNamespaceID(qnc.GetNamespaceUriID()), d.channel)
		if err != nil {
			return err
		}
		pfx = tmp
		// Note: IF elementPrefix is still null it will be determined by a
		// subsequently following NS event
	} else {
		// element prefix
		d.checkDefaultPrefixNamespaceDeclaration(qnc)
		pfx = utils.AsPtr(qnc.GetDefaultPrefix())
	}

	d.getElementContext().SetPrefix(pfx)

	return nil
}

func (d *AbstractEXIBodyDecoder) handleAttributePrefix(qnc *QNameContext) error {
	if d.preservePrefix {
		tmp, err := d.decodeQNamePrefix(d.GetURIByNamespaceID(qnc.GetNamespaceUriID()), d.channel)
		if err != nil {
			return err
		}
		d.attributePrefix = tmp
	} else {
		d.checkDefaultPrefixNamespaceDeclaration(qnc)
		d.attributePrefix = utils.AsPtr(qnc.GetDefaultPrefix())
	}
	return nil
}

func (d *AbstractEXIBodyDecoder) checkDefaultPrefixNamespaceDeclaration(qnc *QNameContext) {
	if d.preservePrefix {
		panic("preserve prefix is not permitted")
	}

	if qnc.GetNamespaceUriID() < d.numberOfUriContexts {
		// schema-known grammar uris/prefixes have been declared in root element
	} else {
		uri := qnc.GetNamespaceUri()
		pfx := d.getPrefix(uri)

		if pfx != nil {
			pfx = utils.AsPtr(qnc.GetDefaultPrefix())
			d.declarePrefix(pfx, uri)
		}
	}
}

func (d *AbstractEXIBodyDecoder) decodeAttributeStructure() (Datatype, error) {
	at := d.nextEvent.(*Attribute)

	// qname
	d.attributeQNameContext = at.GetQNameContext()
	// handle attribute prefix
	if err := d.handleAttributePrefix(d.attributeQNameContext); err != nil {
		return nil, err
	}

	// update current rule
	d.updateCurrentRule(d.nextGrammar)

	return at.datatype, nil
}

func (d *AbstractEXIBodyDecoder) decodeAttributeNSStructure() error {
	atNS := d.nextEvent.(*AttributeNS)
	ruc := d.GetURIByNamespaceID(atNS.GetNamespaceUriID())

	tmp, err := d.decodeLocalName(ruc, d.channel)
	if err != nil {
		return err
	}
	d.attributeQNameContext = tmp

	// handle attribute prefix
	if err := d.handleAttributePrefix(d.attributeQNameContext); err != nil {
		return err
	}

	// update current rule
	d.updateCurrentRule(d.nextGrammar)

	return nil
}

func (d *AbstractEXIBodyDecoder) decodeAttributeAnyInvalidValueStructure() error {
	return d.decodeAttributeGenericStructureOnly()
}

func (d *AbstractEXIBodyDecoder) decodeAttributeGenericStructure() error {
	// decode structure
	if err := d.decodeAttributeGenericStructureOnly(); err != nil {
		return err
	}

	// update current rule
	d.updateCurrentRule(d.nextGrammar)

	return nil
}

func (d *AbstractEXIBodyDecoder) decodeAttributeGenericUndeclaredStructure() error {
	if err := d.decodeAttributeGenericStructureOnly(); err != nil {
		return err
	}
	if err := d.getCurrentGrammar().LearnAttribute(NewAttribute(d.attributeQNameContext)); err != nil {
		return err
	}
	return nil
}

func (d *AbstractEXIBodyDecoder) decodeAttributeGenericStructureOnly() error {
	// decode uri & local-name
	tmp, err := d.decodeQName(d.channel)
	if err != nil {
		return err
	}
	d.attributeQNameContext = tmp

	// handle attribute prefix
	if err := d.handleAttributePrefix(d.attributeQNameContext); err != nil {
		return err
	}

	return nil
}

func (d *AbstractEXIBodyDecoder) decodeCharactersStructure() (Datatype, error) {
	if d.nextEventType != EventTypeCharacters {
		return nil, fmt.Errorf("next event type is not characters: %d", d.nextEventType)
	}

	// update current rule
	d.updateCurrentRule(d.nextGrammar)
	return d.nextEvent.(*Characters).GetDataType(), nil
}

func (d *AbstractEXIBodyDecoder) decodeCharactersGenericStructure() error {
	if d.nextEventType != EventTypeCharactersGeneric {
		return fmt.Errorf("next event type is not characters generic: %d", d.nextEventType)
	}

	// update current rule
	d.updateCurrentRule(d.nextGrammar)
	return nil
}

func (d *AbstractEXIBodyDecoder) decodeCharactersGenericUndeclaredStructure() error {
	if d.nextEventType != EventTypeCharactersGenericUndeclared {
		return fmt.Errorf("next event type is not characters generic undeclared: %d", d.nextEventType)
	}

	// learn character event ?
	currentGrammar := d.getCurrentGrammar()
	currentGrammar.LearnCharacters()

	// update current rule
	d.updateCurrentRule(currentGrammar.GetElementContentGrammar())
	return nil
}

func (d *AbstractEXIBodyDecoder) decodeNamespaceDeclarationStructure() (*NamespaceDeclarationContainer, error) {
	// prefix mapping
	ruc, err := d.decodeURI(d.channel)
	if err != nil {
		return nil, err
	}
	nsPrefix, err := d.decodeNamespacePrefix(ruc, d.channel)
	if err != nil {
		return nil, err
	}

	localElementNS, err := d.channel.DecodeBoolean()
	if err != nil {
		return nil, err
	}
	if localElementNS {
		d.getElementContext().SetPrefix(nsPrefix)
	}

	// NS
	nsDecl := NewNamespaceDeclarationContainer(ruc.GetNamespaceUri(), nsPrefix)
	d.declarePrefixWithNamespaceDeclaraion(nsDecl)

	return &nsDecl, nil
}

func (d *AbstractEXIBodyDecoder) decodeEntityReferenceStructure() ([]rune, error) {
	// decode name AS string
	runes, err := d.channel.DecodeString()
	if err != nil {
		return []rune{}, err
	}

	// update current rule
	d.updateCurrentRule(d.getCurrentGrammar().GetElementContentGrammar())

	return runes, nil
}

func (d *AbstractEXIBodyDecoder) decodeCommentStructure() ([]rune, error) {
	runes, err := d.channel.DecodeString()
	if err != nil {
		return []rune{}, err
	}

	// update current rule
	d.updateCurrentRule(d.getCurrentGrammar().GetElementContentGrammar())

	return runes, nil
}

func (d *AbstractEXIBodyDecoder) decodeProcessingInstructionStructure() (ProcessingInstructionContainer, error) {
	// target & data
	runes, err := d.channel.DecodeString()
	if err != nil {
		return ProcessingInstructionContainer{}, err
	}
	target := string(runes)

	runes, err = d.channel.DecodeString()
	if err != nil {
		return ProcessingInstructionContainer{}, err
	}
	data := string(runes)

	// update current rule
	d.updateCurrentRule(d.getCurrentGrammar().GetElementContentGrammar())

	return ProcessingInstructionContainer{
		Target: target,
		Data:   data,
	}, nil
}

func (d *AbstractEXIBodyDecoder) decodeDocTypeStructure() (*DocTypeContainer, error) {
	name, err := d.channel.DecodeString()
	if err != nil {
		return nil, err
	}
	publicID, err := d.channel.DecodeString()
	if err != nil {
		return nil, err
	}
	systemID, err := d.channel.DecodeString()
	if err != nil {
		return nil, err
	}
	text, err := d.channel.DecodeString()
	if err != nil {
		return nil, err
	}

	return &DocTypeContainer{
		Name:     name,
		PublicID: publicID,
		SystemID: systemID,
		Text:     text,
	}, nil
}

func (d *AbstractEXIBodyDecoder) DecodeStartSelfContainedFragment() error {
	return fmt.Errorf("exi self contained")
}

/*
	AbstractEXIBodyEncoder implementation
*/

const (
	MisuseOfPreservePrefixes string = "A prefix with value null cannot be used in Preserve.Prefixes mode. Report prefix or set your XML reader to do so. e.g., SAX xmlReader.setFeature(\"http://xml.org/sax/features/namespaces\", true); and xmlReader.setFeature(\"http://xml.org/sax/features/namespace-prefixes\", false)"
)

type AbstractEXIBodyEncoder struct {
	*AbstractEXIBodyCoder
	exiHeader          EXIHeaderEncoder
	sePrefix           *string // prefix of previous start element (relevant for preserving prefixes)
	seUri              *string // URI of previous start element (relevant for preserving prefixes)
	channel            EncoderChannel
	typeEncoder        TypeEncoder
	stringEncoder      StringEncoder
	encodingOptions    EncodingOptions
	bChars             []Value // buffers character values before flushing them out
	isXMLSpacePreserve bool
	lastEvent          EventType
	cbuffer            []rune // character buffer for CH trimming, replacing, collapsing
}

func NewAbstractEXIBodyEncoder(exiFactory EXIFactory) (*AbstractEXIBodyEncoder, error) {
	aec, err := NewAbstractEXIBodyCoder(exiFactory)
	if err != nil {
		return nil, err
	}
	typeEncoder, err := exiFactory.CreateTypeEncoder()
	if err != nil {
		return nil, err
	}

	return &AbstractEXIBodyEncoder{
		AbstractEXIBodyCoder: aec,
		exiHeader:            EXIHeaderEncoder{}, //TODO: IMPLEMENTATION!!!
		sePrefix:             nil,
		seUri:                nil,
		channel:              nil,
		typeEncoder:          typeEncoder,
		stringEncoder:        exiFactory.CreateStringEncoder(),
		encodingOptions:      *exiFactory.GetEncodingOptions(),
		bChars:               []Value{},
		isXMLSpacePreserve:   false,
		lastEvent:            -1,
		cbuffer:              []rune{},
	}, nil
}

func (e *AbstractEXIBodyEncoder) InitForEachRun() error {
	if err := e.AbstractEXIBodyCoder.InitForEachRun(); err != nil {
		return err
	}

	e.learnedProductions = 0
	e.stringEncoder.Clear()
	if e.exiFactory.GetSharedStrings() != nil {
		if err := e.stringEncoder.SetSharedStrings(*e.exiFactory.GetSharedStrings()); err != nil {
			return err
		}
	}
	e.bChars = []Value{}
	e.isXMLSpacePreserve = false

	return nil
}

func (e *AbstractEXIBodyEncoder) encodeQName(namespaceURI, localName string, channel EncoderChannel) (*QNameContext, error) {
	// uri
	ruc, err := e.encodeURI(namespaceURI, channel)
	if err != nil {
		return nil, err
	}

	// local name
	return e.encodeLocalName(localName, ruc, channel)
}

func (e *AbstractEXIBodyEncoder) encodeURI(namespaceURI string, channel EncoderChannel) (*RuntimeUriContext, error) {
	numberBitsURI := utils.GetCodingLength(e.GetNumberOfURIs() + 1)
	ruc := e.GetURI(namespaceURI)

	if ruc == nil {
		// uri string value was not found
		// ==> zero (0) as an n-nit unsigned integer
		// followed by uri encoded as string
		if err := channel.EncodeNBitUnsignedInteger(0, numberBitsURI); err != nil {
			return nil, err
		}
		if err := channel.EncodeString(namespaceURI); err != nil {
			return nil, err
		}
		// after encoding string value is added to table
		ruc = e.addUri(namespaceURI)
	} else {
		// string value found
		// ==> value(i+1) is encoded as n-bit unsigned integer
		if err := channel.EncodeNBitUnsignedInteger(ruc.GetNamespaceUriID()+1, numberBitsURI); err != nil {
			return nil, err
		}
	}

	return ruc, nil
}

func (e *AbstractEXIBodyEncoder) encodeQNamePrefix(qnc *QNameContext, prefix *string, channel EncoderChannel) error {
	if prefix == nil {
		e.emitWarning(MisuseOfPreservePrefixes)
	}

	namespaceUriID := qnc.GetNamespaceUriID()

	if namespaceUriID == 0 {
		// XMLConstants.NULL_NS_URI
		// default namespace --> DEFAULT_NS_PREFIX
	} else {
		ruc := e.GetURIByNamespaceID(namespaceUriID)
		numberOfPrefixes := ruc.GetNumberOfPrefixes()

		switch numberOfPrefixes {
		case 0:
			// If there are no prefixes specified for the
			// URI of the QName by preceding NS events in the EXI stream,
			// the prefix is undefined. An undefined prefix is represented
			// using zero bits (i.e., omitted).
			// --> requires following NS
		case 1:
			// If there is only one prefix, the prefix is implicit
		default:
			pfxID := ruc.getPrefixID(*prefix)
			if pfxID == NotFound {
				// choose *one* prefix which gets modified by
				// local-element-ns anyway ?
				pfxID = 0
			}

			// overlapping URIs
			return channel.EncodeNBitUnsignedInteger(pfxID, utils.GetCodingLength(numberOfPrefixes))
		}
	}

	return nil
}

func (e *AbstractEXIBodyEncoder) encodeLocalName(localName string, ruc *RuntimeUriContext, channel EncoderChannel) (*QNameContext, error) {
	// look for localNameID
	qnc := ruc.GetQNameContextByLocalName(localName)

	if qnc == nil {
		// string value was not found in local partition
		// ==> string literal is encoded as a String
		// with the length of the string incremented by one
		if err := channel.EncodeUnsignedInteger(len(localName) + 1); err != nil {
			return nil, err
		}
		if err := channel.EncodeStringOnly(localName); err != nil {
			return nil, err
		}
		// After encoding the string value, it is added to the string
		// table partition and assigned the next available compact
		// identifier.
		qnc = ruc.AddQNameContext(localName)
	} else {
		// string value found in local partition
		// ==> string value is represented as zero (0) encoded as an
		// Unsigned Integer followed by an the compact identifier of the
		// string value as an n-bit unsigned integer n is log2 m and m is
		// the number of entries in the string table partition
		if err := channel.EncodeUnsignedInteger(0); err != nil {
			return nil, err
		}
		n := utils.GetCodingLength(ruc.GetNumberOfQNames())
		if err := channel.EncodeNBitUnsignedInteger(qnc.GetLocalNameID(), n); err != nil {
			return nil, err
		}
	}

	return qnc, nil
}

func (e *AbstractEXIBodyEncoder) encodeNamespacePrefix(ruc *RuntimeUriContext, prefix *string, channel EncoderChannel) error {
	nPfx := utils.GetCodingLength(ruc.GetNumberOfPrefixes() + 1)
	pfxID := ruc.getPrefixID(*prefix)

	if pfxID == NotFound {
		// string value was not found
		// ==> zero (0) as an n-bit unsigned integer
		// followed by pfx encoded as string
		if err := channel.EncodeNBitUnsignedInteger(0, nPfx); err != nil {
			return err
		}
		if err := channel.EncodeStringOnly(*prefix); err != nil {
			return err
		}
		// after encoding string value is added to table
		ruc.addPrefix(*prefix)
	} else {
		// string value found
		// ==> value(i+1) is encoded as n-bit unsigned integer
		if err := channel.EncodeNBitUnsignedInteger(pfxID+1, nPfx); err != nil {
			return err
		}
	}

	return nil
}

func (e *AbstractEXIBodyEncoder) Flush() error {
	return e.channel.Flush()
}

func (e *AbstractEXIBodyEncoder) writeString(text string) error {
	return e.channel.EncodeString(text)
}

func (e *AbstractEXIBodyEncoder) isTypeValid(datatype Datatype, value Value) (bool, error) {
	return e.typeEncoder.IsValid(datatype, value)
}

func (e *AbstractEXIBodyEncoder) writeValue(qnc *QNameContext) error {
	panic("abstract")
}

func (e *AbstractEXIBodyEncoder) encode1stLevelEventCode(pos int) error {
	codeLength := e.fidelityOptions.Get1stLevelEventCodeLength(e.getCurrentGrammar())
	if codeLength > 0 {
		return e.channel.EncodeNBitUnsignedInteger(pos, codeLength)
	}
	return nil
}

func (e *AbstractEXIBodyEncoder) encode2ndLevelEventCode(pos int) error {
	// 1st level
	currentGrammar := e.getCurrentGrammar()
	if err := e.channel.EncodeNBitUnsignedInteger(currentGrammar.GetNumberOfEvents(), e.fidelityOptions.Get1stLevelEventCodeLength(currentGrammar)); err != nil {
		return err
	}

	// 2nd level
	ch2 := e.fidelityOptions.Get2ndLevelCharacteristics(currentGrammar)
	if pos >= ch2 {
		return errors.New("position is less than 2nd level characteristics")
	}

	return e.channel.EncodeNBitUnsignedInteger(pos, utils.GetCodingLength(ch2))
}

func (e *AbstractEXIBodyEncoder) encode3rdLevelEventCode(pos int) error {
	// 1st level
	currentGrammar := e.getCurrentGrammar()
	if err := e.channel.EncodeNBitUnsignedInteger(currentGrammar.GetNumberOfEvents(), e.fidelityOptions.Get1stLevelEventCodeLength(currentGrammar)); err != nil {
		return err
	}

	// 2nd level
	ch2 := e.fidelityOptions.Get2ndLevelCharacteristics(currentGrammar)
	ec2 := 0
	if ch2 > 0 {
		ec2 = ch2 - 1
	}
	if err := e.channel.EncodeNBitUnsignedInteger(ec2, utils.GetCodingLength(ch2)); err != nil {
		return err
	}

	// 3rd level
	ch3 := e.fidelityOptions.Get3rdLevelCharacteristics()
	if pos >= ch3 {
		return errors.New("position is less than 3rd level characteristics")
	}

	return e.channel.EncodeNBitUnsignedInteger(pos, utils.GetCodingLength(ch3))
}

func (e *AbstractEXIBodyEncoder) EncodeStartDocument() error {
	if e.channel == nil {
		return errors.New("no valid EXI OutputStream set for encoding. Please use SetOutput( ... )")
	}
	if err := e.InitForEachRun(); err != nil {
		return err
	}

	ei := e.getCurrentGrammar().GetProduction(EventTypeStartDocument)

	// Note: no EventCode needs to be written since there is only
	// one choice
	if ei == nil {
		return errors.New("no EXI Event found for startDocument")
	}

	e.updateCurrentRule(ei.GetNextGrammar())
	e.lastEvent = EventTypeStartDocument

	return nil
}

func (e *AbstractEXIBodyEncoder) EncodeEndDocument() error {
	if err := e.checkPendingCharacters(EventTypeEndDocument); err != nil {
		return err
	}

	ei := e.getCurrentGrammar().GetProduction(EventTypeEndDocument)

	if ei != nil {
		if err := e.encode1stLevelEventCode(ei.GetEventCode()); err != nil {
			return err
		}
	} else {
		return errors.New("no EXI Event found for endDocument")
	}

	e.lastEvent = EventTypeEndDocument

	return nil
}

func (e *AbstractEXIBodyEncoder) EncodeStartElementByQName(se QName) error {
	return e.EncodeStartElement(se.Space, se.Local, se.Prefix)
}

func (e *AbstractEXIBodyEncoder) EncodeStartElement(uri, localName string, prefix *string) error {
	if err := e.checkPendingCharacters(EventTypeStartElement); err != nil {
		return err
	}

	e.sePrefix = prefix
	e.seUri = &uri

	var ei Production
	var updContextRule Grammar
	var nextSE *StartElement

	currentGrammar := e.getCurrentGrammar()

	ei = currentGrammar.GetStartElementProduction(uri, localName)
	if ei != nil {
		if !ei.GetEvent().IsEventType(EventTypeStartElement) {
			return errors.New("wrong event type for start element")
		}

		// encode 1st level EventCode
		if err := e.encode1stLevelEventCode(ei.GetEventCode()); err != nil {
			return err
		}
		// nextSE ...
		nextSE = ei.GetEvent().(*StartElement)
		// qname implicit by SE(qname) event, prefix only missing
		if e.preservePrefix {
			if err := e.encodeQNamePrefix(nextSE.GetQNameContext(), prefix, e.channel); err != nil {
				return err
			}
		}
		// next context rule
		updContextRule = ei.GetNextGrammar()
	} else {
		ei = currentGrammar.GetStartElementNSProduction(uri)
		if ei != nil {
			if !ei.GetEvent().IsEventType(EventTypeStartElementNS) {
				return errors.New("wrong event type for start element NS")
			}

			// encode 1st level EventCode
			if err := e.encode1stLevelEventCode(ei.GetEventCode()); err != nil {
				return err
			}

			seNS := ei.GetEvent().(*StartElementNS)
			ruc := e.GetURIByNamespaceID(seNS.GetNamespaceUriID())

			// encode local-name (and prefix)
			qnc, err := e.encodeLocalName(localName, ruc, e.channel)
			if err != nil {
				return err
			}

			if e.preservePrefix {
				if err := e.encodeQNamePrefix(qnc, prefix, e.channel); err != nil {
					return err
				}
			}

			// next context rule
			updContextRule = ei.GetNextGrammar()
			// next SE ...
			nextSE = e.getGlobalStartElement(qnc)
		} else {
			// try SE(*), generic SE on first level
			ei = currentGrammar.GetProduction(EventTypeStartElementGeneric)
			if ei != nil {
				if !ei.GetEvent().IsEventType(EventTypeStartElementGeneric) {
					return errors.New("wrong event type for start element generic")
				}

				// encode 1st level EventCode
				if err := e.encode1stLevelEventCode(ei.GetEventCode()); err != nil {
					return err
				}

				// next context rule
				updContextRule = ei.GetNextGrammar()
			} else {
				// Undeclared SE(*) can be found on 2nd level
				ecSEUndeclared := e.fidelityOptions.Get2ndLevelEventCode(EventTypeStartElementGenericUndeclared, currentGrammar)

				if ecSEUndeclared == NotFound {
					// Note: should never happen except in strict mode
					return fmt.Errorf("unexpected SE {%s}%s, %+v", uri, localName, e.exiFactory)
				}

				// limit grammar learning ?
				switch e.limitGrammars() {
				case ProfileDisablingMechanismXsiType:
					if err := e.insertXsiTypeAnyType(); err != nil {
						return err
					}
					currentGrammar = e.getCurrentGrammar()
					// encode 1st level EventCode
					ei = currentGrammar.GetProduction(EventTypeStartElementGeneric)
					if ei == nil {
						return errors.New("ei == nil")
					}
					if err := e.encode1stLevelEventCode(ei.GetEventCode()); err != nil {
						return err
					}
					updContextRule = ei.GetNextGrammar()
				case ProfileDisablingMechanismGhostProduction:
					fallthrough
				default:
					// encode [undeclared] event-code
					if err := e.encode2ndLevelEventCode(ecSEUndeclared); err != nil {
						return err
					}
					// next context rule
					updContextRule = currentGrammar.GetElementContentGrammar()
				}
			}

			// encode entire qualified name
			qnc, err := e.encodeQName(uri, localName, e.channel)
			if err != nil {
				return err
			}
			if e.preservePrefix {
				if err := e.encodeQNamePrefix(qnc, prefix, e.channel); err != nil {
					return err
				}
			}

			// next SE ...
			nextSE = e.getGlobalStartElement(qnc)

			// learning for built-in grammar (here and not as part of
			// SE_Undecl(*) because of FragmentContent!)
			currentGrammar.LearnStartElement(nextSE)
			e.productionLearningCounting(currentGrammar)
		}
	}

	e.pushElement(updContextRule, nextSE)
	e.lastEvent = EventTypeStartElement

	return nil
}

func (e *AbstractEXIBodyEncoder) productionLearningCounting(g Grammar) {
	if e.limitGrammarLearning {
		// Note: no counting for schema-informed grammars and
		// BuiltInFragmentGrammar
		if e.maxBuiltInProductions >= 0 && !g.IsSchemaInformed() && g.GetGrammarType() != GrammarTypeBuiltInFragmentContent {
			e.learnedProductions++
		}
	}
}

func (e *AbstractEXIBodyEncoder) limitGrammars() ProfileDisablingMechanism {
	retVal := ProfileDisablingMechanismNone
	currentGrammar := e.getCurrentGrammar()

	if e.limitGrammarLearning && e.grammar.IsSchemaInformed() && !currentGrammar.IsSchemaInformed() {
		// number of built-in grammars reached
		if e.maxBuiltInElementGrammars != -1 {
			csize := len(e.runtimeGlobalElements)
			if csize > e.maxBuiltInElementGrammars {
				if currentGrammar.GetNumberOfEvents() == 0 {
					// new grammar that hits bound
					retVal = ProfileDisablingMechanismXsiType
				} else if e.isBuiltInStartTagGrammarWithAtXsiTypeOnly(currentGrammar) {
					// previous type cast
					retVal = ProfileDisablingMechanismXsiType
				}
			}
		}

		// number of productions reached?
		if e.maxBuiltInProductions != -1 && retVal == ProfileDisablingMechanismNone && e.learnedProductions >= e.maxBuiltInProductions {
			// bound reached
			if e.lastEvent == EventTypeStartElement || e.lastEvent == EventTypeNamespaceDeclaration {
				// First mean possible: Insert xsi:type
				retVal = ProfileDisablingMechanismXsiType
			} else {
				// Only 2nd mean possible: use ghost productions
				retVal = ProfileDisablingMechanismGhostProduction
				currentGrammar.StopLearning()
			}
		}
	}

	return retVal
}

func (e *AbstractEXIBodyEncoder) insertXsiTypeAnyType() error {
	var pfx *string = nil
	if e.preservePrefix {
		// XMLConstants.W3C_XML_SCHEMA_NS_URI ==
		// "http://www.w3.org/2001/XMLSchema"
		pfx = e.getPrefix(XMLSchemaNS_URI)
		if pfx == nil {
			// no prefixes for XSD have been declared so far.
			pfx = utils.AsPtr("xsdP")
			if err := e.EncodeNamespaceDeclaration(XMLSchemaNS_URI, pfx); err != nil {
				return err
			}
		}
	}
	qnv := NewQNameValue(XMLSchemaNS_URI, "anyType", pfx)

	// needed to avoid grammar learning
	return e.encodeAttributeXsiTypeWithForce2ndLP(qnv, pfx, true)
}

func (e *AbstractEXIBodyEncoder) EncodeNamespaceDeclaration(uri string, prefix *string) error {
	e.declarePrefix(prefix, uri)

	if e.preservePrefix {
		// event code
		currentGrammar := e.getCurrentGrammar()
		ec2 := e.fidelityOptions.Get2ndLevelEventCode(EventTypeNamespaceDeclaration, currentGrammar)

		if e.fidelityOptions.Get2ndLevelEventType(ec2, currentGrammar) != EventTypeNamespaceDeclaration {
			return errors.New("2nd level event code do not match event type EventTypeNamespaceDeclaration")
		}
		if err := e.encode2ndLevelEventCode(ec2); err != nil {
			return err
		}

		// prefix mapping
		euc, err := e.encodeURI(uri, e.channel)
		if err != nil {
			return err
		}
		if err := e.encodeNamespacePrefix(euc, prefix, e.channel); err != nil {
			return err
		}

		// local-element-ns
		if e.sePrefix == nil {
			// the prefix was not properly reported
			e.emitWarning(MisuseOfPreservePrefixes)
			// try to fix that issue by checking URI
			if err := e.channel.EncodeBoolean(e.seUri != nil && *e.seUri == uri); err != nil {
				return err
			}
		} else {
			if err := e.channel.EncodeBoolean(*prefix == *e.sePrefix); err != nil {
				return err
			}
		}

		e.lastEvent = EventTypeNamespaceDeclaration
	}

	return nil
}

func (e *AbstractEXIBodyEncoder) EncodeEndElement() error {
	if err := e.checkPendingCharacters(EventTypeStartElement); err != nil {
		return err
	}

	currentGrammar := e.getCurrentGrammar()
	ei := currentGrammar.GetProduction(EventTypeEndElement)

	if ei != nil {
		// encode EventCode (common case)
		if err := e.encode1stLevelEventCode(ei.GetEventCode()); err != nil {
			return err
		}
	} else {
		// Undeclared EE can be found on 2nd level
		ecEEUndeclared := e.fidelityOptions.Get2ndLevelEventCode(EventTypeEndElementUndeclared, currentGrammar)

		if ecEEUndeclared == NotFound {
			// Should only happen in STRICT mode
			// Special case: SAX does not inform about empty ("") CH events

			if err := e.encodeCharactersForce(EmptyStringValue); err != nil {
				return err
			}
			currentGrammar = e.getCurrentGrammar()
			ei = currentGrammar.GetProduction(EventTypeEndElement)
			if err := e.encode1stLevelEventCode(ei.GetEventCode()); err != nil {
				return err
			}

			//TODO: More informative errors
		} else {
			// limit grammar learning ?
			switch e.limitGrammars() {
			case ProfileDisablingMechanismXsiType:
				if err := e.insertXsiTypeAnyType(); err != nil {
					return err
				}
				currentGrammar = e.getCurrentGrammar()
				// encode 1st level EventCode
				ei = currentGrammar.GetProduction(EventTypeEndElement)
				if ei == nil {
					return errors.New("production is nil")
				}
				if err := e.encode1stLevelEventCode(ei.GetEventCode()); err != nil {
					return err
				}
			case ProfileDisablingMechanismGhostProduction:
				fallthrough
			default:
				// encode [undeclared] event-code
				if err := e.encode2ndLevelEventCode(ecEEUndeclared); err != nil {
					return err
				}
				// learn end-element event ?
				currentGrammar.LearnEndElement()
				e.productionLearningCounting(currentGrammar)
			}
		}
	}

	// pop element from stack
	ec := e.popElement()

	// make sure to adapt xml:space behavior
	if ec.IsXMLSpacePreserve() != nil {
		// check in the hierarchy whether there is xml:space present OR
		// "default"
		isOtherPreserve := false
		for i := e.elementContextStackIndex; i >= 0; i-- {
			isP := e.elementContextStack[i].IsXMLSpacePreserve()
			if isP != nil {
				isOtherPreserve = *isP
				break
			}
		}
		e.isXMLSpacePreserve = isOtherPreserve
	}

	e.lastEvent = EventTypeEndElement

	return nil
}

func (e *AbstractEXIBodyEncoder) EncodeAttributeList(attributes AttributeList) error {
	// 1. NS
	for i := range attributes.GetNumberOfNamespaceDeclarations() {
		ns := attributes.GetNamespaceDeclaration(i)
		if err := e.EncodeNamespaceDeclaration(ns.namespaceURI, ns.prefix); err != nil {
			return err
		}
	}

	// 2. XSI-Type
	if attributes.HasXsiType() {
		if err := e.EncodeAttributeXsiType(NewStringValueFromString(*attributes.GetXsiTypeRaw()), attributes.GetXsiTypePrefix()); err != nil {
			return err
		}
	}

	// 2. XSI-Nil
	if attributes.HasXsiNil() {
		if err := e.EncodeAttributeXsiNil(NewStringValueFromString(*attributes.GetXsiNil()), attributes.GetXsiNilPrefix()); err != nil {
			return err
		}
	}

	// 4. Remaining Attributes
	for i := range attributes.GetNumberOfAttributes() {
		if err := e.EncodeAttribute(*attributes.GetAttributeURI(i), *attributes.GetAttributeLocalName(i),
			attributes.GetAttributePrefix(i), NewStringValueFromString(*attributes.GetAttributeValue(i))); err != nil {
			return err
		}
	}

	return nil
}

func (e *AbstractEXIBodyEncoder) EncodeAttributeXsiType(kind Value, pfx *string) error {
	force2ndLevelProduction := false
	if e.limitGrammars() == ProfileDisablingMechanismXsiType {
		force2ndLevelProduction = true
	}
	return e.encodeAttributeXsiTypeWithForce2ndLP(kind, pfx, force2ndLevelProduction)
}

func (e *AbstractEXIBodyEncoder) encodeAttributeXsiTypeWithForce2ndLP(kind Value, pfx *string, force2ndLevelProduction bool) error {
	/*
	 * The value of each AT (xsi:type) event is represented as a QName.
	 */
	var qnamePrefix *string
	var qnameURI *string
	var qnameLocalName string

	qv, ok := kind.(*QNameValue)
	if ok {
		qnameURI = utils.AsPtr(qv.GetNamespaceURI())
		qnamePrefix = qv.GetPrefix()
		qnameLocalName = qv.GetLocalName()
	} else {
		sType, err := kind.ToString()
		if err != nil {
			return err
		}
		// extract prefix
		qnamePrefix = utils.AsPtr(utils.GetPrefixPart(sType))
		// String
		qnameURI = e.getURI(qnamePrefix)

		/*
		 * If there is no namespace in scope for the specified qname prefix,
		 * the QName uri is set to empty ("") and the QName localName is set
		 * to the full lexical value of the QName, including the prefix.
		 */
		if qnameURI == nil {
			/* uri in scope for prefix */
			qnameURI = utils.AsPtr(XMLNullNS_URI)
			qnameLocalName = sType
		} else {
			qnameLocalName = utils.GetLocalPart(sType)
		}
	}

	currentGrammar := e.getCurrentGrammar()
	ec2 := e.fidelityOptions.Get2ndLevelEventCode(EventTypeAttributeXsiType, currentGrammar)

	if ec2 != NotFound {
		if e.fidelityOptions.Get2ndLevelEventType(ec2, currentGrammar) == EventTypeAttributeXsiType {
			return errors.New("2nd level event code do not match event type EventTypeAttributeXsiType")
		}

		// encode event-code, AT(xsi:type)
		if err := e.encode2ndLevelEventCode(ec2); err != nil {
			return err
		}
		// prefix
		if e.preservePrefix {
			if err := e.encodeQNamePrefix(e.getXsiTypeContext(), pfx, e.channel); err != nil {
				return err
			}
		}
	} else {
		// Note: cannot be encoded as any other attribute due to the
		// different channels in compression mode

		// try first (learned) xsi:type attribute
		var ei Production
		if force2ndLevelProduction {
			// only 2nd level of interest
			ei = nil
		} else {
			ei = currentGrammar.GetAttributeProduction(XMLSchemaInstanceNS_URI, XSIType)
		}

		if ei != nil {
			if err := e.encode1stLevelEventCode(ei.GetEventCode()); err != nil {
				return err
			}
		} else {
			// try generic attribute
			if !force2ndLevelProduction {
				ei = currentGrammar.GetProduction(EventTypeAttributeGeneric)
			}

			if ei != nil {
				if err := e.encode1stLevelEventCode(ei.GetEventCode()); err != nil {
					return err
				}
			} else {
				ec2 = e.fidelityOptions.Get2ndLevelEventCode(EventTypeAttributeGenericUndeclared, currentGrammar)

				if ec2 != NotFound {
					if err := e.encode2ndLevelEventCode(ec2); err != nil {
						return err
					}
					qncType := e.getXsiTypeContext()

					if e.limitGrammarLearning {
						// 3.2 Grammar Learning Disabling Parameters
						// - In particular, the AT(xsi:type) productions that would be inserted in grammars
						// that would be instantiated after the maximumNumberOfBuiltInElementGrammars
						// threshold are not counted.
						if len(e.runtimeGlobalElements) > e.maxBuiltInElementGrammars && currentGrammar.GetNumberOfEvents() == 0 {
							// can't evolve anymore
							currentGrammar.StopLearning()
						} else {
							e.productionLearningCounting(currentGrammar)
						}
					}
					if err := currentGrammar.LearnAttribute(NewAttribute(qncType)); err != nil {
						return err
					}
				} else {
					return fmt.Errorf("type cast is not encodable")
				}
			}

			// xsi:type as qname
			qncType := e.getXsiTypeContext()
			if _, err := e.encodeQName(qncType.GetNamespaceUri(), qncType.GetLocalName(), e.channel); err != nil {
				return err
			}

			if e.preservePrefix {
				if err := e.encodeQNamePrefix(qncType, pfx, e.channel); err != nil {
					return err
				}
			}
		}
	}

	// write xsi:type value "content" as qname
	var qncType *QNameContext
	if e.preserveLexicalValues {
		// Note: IF xsi:type values are encoded as String, prefixes need to
		// be preserved as well!
		if len(*qnamePrefix) > 0 && !e.preservePrefix {
			return fmt.Errorf("xsi:type is not encodable, preserve lexicalValues requires prefixes preserved as well")
		}

		// as string
		if _, err := e.typeEncoder.IsValid(BuiltInGetDefaultDatatype(), kind); err != nil {
			return err
		}
		if err := e.typeEncoder.WriteValue(e.getXsiTypeContext(), e.channel, e.stringEncoder); err != nil {
			return err
		}

		ruc := e.GetURI(*qnameURI)
		if ruc != nil {
			qncType = ruc.GetQNameContextByLocalName(*qnameURI)
		} else {
			qncType = nil
		}
	} else {
		// typed
		qnc, err := e.encodeQName(*qnameURI, qnameLocalName, e.channel)
		if err != nil {
			return err
		}
		qncType = qnc

		if e.preservePrefix {
			if err := e.encodeQNamePrefix(qncType, qnamePrefix, e.channel); err != nil {
				return err
			}
		}
	}

	// grammar exists ?
	if qncType != nil && qncType.GetTypeGrammar() != nil {
		// update grammar according to given xsi:type
		e.updateCurrentRule(qncType.GetTypeGrammar())
	}

	e.lastEvent = EventTypeAttributeXsiType

	return nil
}

func (e *AbstractEXIBodyEncoder) EncodeAttributeXsiNil(nilValue Value, pfx *string) error {
	currentGrammar := e.getCurrentGrammar()
	if currentGrammar.IsSchemaInformed() {
		siCurrentRule := currentGrammar.(SchemaInformedGrammar)

		validNil := false
		validNilValue := false

		nv, ok := nilValue.(*BooleanValue)
		if ok {
			validNil = true
			validNilValue = nv.ToBoolean()
		} else {
			nilValueS, err := nilValue.ToString()
			if err != nil {
				return err
			}
			nv = BooleanValueParse(nilValueS)
			if nv != nil {
				validNil = true
				validNilValue = nv.ToBoolean()
			}
		}

		if validNil {
			// Note: in some cases we can simply skip the xsi:nil event
			if !e.preserveLexicalValues && !validNilValue && !e.encodingOptions.IsOptionEnabled(OptionIncludeInsignificanXsiNil) {
				return nil
			}

			// schema-valid boolean
			ec2 := e.fidelityOptions.Get2ndLevelEventCode(EventTypeAttributeXsiNil, siCurrentRule)
			if ec2 != NotFound {
				// encode event-code only
				if err := e.encode2ndLevelEventCode(ec2); err != nil {
					return err
				}
				// prefix
				if e.preservePrefix {
					if err := e.encodeQNamePrefix(e.getXsiNilContext(), pfx, e.channel); err != nil {
						return err
					}
				}

				// encode nil value "content" as Boolean
				if e.preserveLexicalValues {
					// as string
					if _, err := e.typeEncoder.IsValid(e.booleanDatatype, nilValue); err != nil {
						return err
					}
					if err := e.typeEncoder.WriteValue(e.getXsiTypeContext(), e.channel, e.stringEncoder); err != nil {
						return err
					}
				} else {
					// typed
					if err := e.channel.EncodeBoolean(validNilValue); err != nil {
						return err
					}
				}

				if validNilValue { // jump to typeEmpty
					// update current rule
					grammar, err := siCurrentRule.(SchemaInformedFirstStartTagGrammar).GetTypeEmpty()
					if err != nil {
						return err
					}
					e.updateCurrentRule(grammar)
				}
			} else {
				ei := currentGrammar.GetProduction(EventTypeAttributeGeneric)
				if ei != nil {
					if err := e.encode1stLevelEventCode(ei.GetEventCode()); err != nil {
						return err
					}
					// qname & prefix
					euc, err := e.encodeURI(XMLSchemaInstanceNS_URI, e.channel)
					if err != nil {
						return err
					}
					if _, err := e.encodeLocalName(XSINil, euc, e.channel); err != nil {
						return err
					}
					if e.preservePrefix {
						if err := e.encodeQNamePrefix(e.getXsiNilContext(), pfx, e.channel); err != nil {
							return err
						}
					}

					// encode nil value "content" as Boolean
					if e.preserveLexicalValues {
						// as string
						if _, err := e.typeEncoder.IsValid(e.booleanDatatype, nilValue); err != nil {
							return err
						}
						if err := e.typeEncoder.WriteValue(e.getXsiNilContext(), e.channel, e.stringEncoder); err != nil {
							return err
						}
					} else {
						// typed
						if err := e.channel.EncodeBoolean(validNilValue); err != nil {
							return err
						}
					}

					if validNilValue { // jump to typeEmpty
						// update current rule
						grammar, err := siCurrentRule.(SchemaInformedFirstStartTagGrammar).GetTypeEmpty()
						if err != nil {
							return err
						}
						e.updateCurrentRule(grammar)
					}
				} else {
					return errors.New("attribute xsi=nil cannot be encoded")
				}
			}
		} else {
			// If the value is not a schema-valid Boolean, the
			// AT (xsi:nil) event is represented by
			// the AT (*) [untyped value] terminal
			sig := currentGrammar.(SchemaInformedGrammar)
			if err := e.encodeSchemaInvalidAttributeEventCode(sig.GetNumberOfDeclaredAttributes()); err != nil {
				return err
			}
			euc, err := e.encodeURI(XMLSchemaInstanceNS_URI, e.channel)
			if err != nil {
				return err
			}
			if _, err := e.encodeLocalName(XSINil, euc, e.channel); err != nil {
				return err
			}
			if e.preservePrefix {
				if err := e.encodeQNamePrefix(e.getXsiNilContext(), pfx, e.channel); err != nil {
					return err
				}
			}

			datatype := BuiltInGetDefaultDatatype()
			if _, err := e.isTypeValid(datatype, nilValue); err != nil {
				return err
			}
			//TODO: Check!!!
			if err := e.writeValue(e.getXsiTypeContext()); err != nil {
				return err
			}
		}
	} else {
		// encode as any other attribute
		if err := e.EncodeAttribute(XMLSchemaInstanceNS_URI, XSINil, pfx, nilValue); err != nil {
			return err
		}
	}

	e.lastEvent = EventTypeAttributeXsiNil

	return nil
}

func (e *AbstractEXIBodyEncoder) encodeSchemaInvalidAttributeEventCode(eventCode3 int) error {
	currentGrammar := e.getCurrentGrammar()
	// schema-invalid AT
	ec2ATDeviated := e.fidelityOptions.Get2ndLevelEventCode(EventTypeAttributeInvalidValue, currentGrammar)
	if err := e.encode2ndLevelEventCode(ec2ATDeviated); err != nil {
		return err
	}
	// encode 3rd level event-code
	// AT specialty: calculate 3rd level attribute event-code
	// int eventCode3 = ei.getEventCode() - currentRule.getLeastAttributeEventCode();
	sig := currentGrammar.(SchemaInformedGrammar)
	if err := e.channel.EncodeNBitUnsignedInteger(eventCode3, utils.GetCodingLength(sig.GetNumberOfDeclaredAttributes()+1)); err != nil {
		return err
	}

	return nil
}

func (e *AbstractEXIBodyEncoder) EncodeAttributeByQName(at QName, value Value) error {
	return e.EncodeAttribute(at.Space, at.Local, at.Prefix, value)
}

func (e *AbstractEXIBodyEncoder) EncodeAttribute(uri, localName string, prefix *string, value Value) error {
	var ei Production
	var qnc *QNameContext
	var next Grammar

	currentGrammar := e.getCurrentGrammar()
	ei = currentGrammar.GetAttributeProduction(uri, localName)
	if ei != nil {
		// declared AT(uri:localName)
		at := ei.GetEvent().(*Attribute)
		qnc = at.GetQNameContext()

		valid, err := e.isTypeValid(at.GetDataType(), value)
		if err != nil {
			return err
		}
		if valid {
			if err := e.encode1stLevelEventCode(ei.GetEventCode()); err != nil {
				return err
			}
		} else {
			sig := currentGrammar.(SchemaInformedGrammar)
			eventCode3 := ei.GetEventCode() - sig.GetLeastAttributeEventCode()
			if err := e.encodeSchemaInvalidAttributeEventCode(eventCode3); err != nil {
				return err
			}
			if _, err := e.isTypeValid(BuiltInGetDefaultDatatype(), value); err != nil {
				return err
			}
			next = ei.GetNextGrammar()
		}
	} else {
		switch e.limitGrammars() {
		case ProfileDisablingMechanismXsiType:
			if err := e.insertXsiTypeAnyType(); err != nil {
				return err
			}
			currentGrammar = e.getCurrentGrammar()
		case ProfileDisablingMechanismGhostProduction:
			fallthrough
		default:
			// no special action
		}

		ei = currentGrammar.GetAttributeNSProduction(uri)
		if ei == nil {
			ei = currentGrammar.GetProduction(EventTypeAttributeGeneric)
			if ei == nil {
				// Undeclared AT(*) can be found on 2nd level
			}
		}

		globalAT, err := e.getGlobalAttribute(uri, localName)
		if err != nil {
			return err
		}
		if currentGrammar.IsSchemaInformed() && globalAT != nil {
			/*
			 * In a schema-informed grammar, all productions of the form
			 * LeftHandSide : AT (*) are evaluated as follows:
			 *
			 * Let qname be the qname of the attribute matched by AT (*) If
			 * a global attribute definition exists for qname, let
			 * global-type be the datatype of the global attribute.
			 */
			valid, err := e.isTypeValid(globalAT.GetDataType(), value)
			if err != nil {
				return err
			}
			if valid {
				/*
				 * If the attribute value can be represented using the
				 * datatype representation associated with global-type, it
				 * SHOULD be represented using the datatype representation
				 * associated with global-type (see 7. Representing Event
				 * Content).
				 */
				if ei == nil {
					if err := e.encodeAttributeEventCodeUndeclared(currentGrammar, localName); err != nil {
						return err
					}
				} else {
					if err := e.encode1stLevelEventCode(ei.GetEventCode()); err != nil {
						return err
					}
				}
			} else {
				/*
				 * If the attribute value is not represented using the
				 * datatype representation associated with global-type,
				 * represent the attribute event using the AT (*) [untyped
				 * value] terminal (see 8.5.4.4 Undeclared Productions).
				 */
				/*
				 * AT (*) [untyped value] Element i, j n.(m+1).(x) x
				 * represents the number of attributes declared in the
				 * schema for this context
				 */
				sig := currentGrammar.(SchemaInformedGrammar)
				if err := e.encodeSchemaInvalidAttributeEventCode(sig.GetNumberOfDeclaredAttributes()); err != nil {
					return err
				}
				if _, err := e.isTypeValid(BuiltInGetDefaultDatatype(), value); err != nil {
					return err
				}
			}

			if ei == nil || ei.GetEvent().IsEventType(EventTypeAttributeGeneric) {
				// (un)declared AT(*)
				qnc, err = e.encodeQName(uri, localName, e.channel)
				if err != nil {
					return err
				}
				if ei == nil {
					next = currentGrammar
				} else {
					next = ei.GetNextGrammar()
				}
			} else {
				// declared AT(uri:*)
				atNS := ei.GetEvent().(*AttributeNS)
				// localname only
				ruc := e.GetURIByNamespaceID(atNS.GetNamespaceUriID())
				qnc, err = e.encodeLocalName(localName, ruc, e.channel)
				if err != nil {
					return err
				}
				next = ei.GetNextGrammar()
			}
		} else {
			// no schema-informed grammar --> default datatype in any case
			// NO global attribute --> default datatype
			if _, err := e.isTypeValid(BuiltInGetDefaultDatatype(), value); err != nil {
				return err
			}

			var err error

			if ei == nil {
				// Undeclared AT(*), 2nd level
				qnc, err = e.encodeUndeclaredAT(currentGrammar, uri, localName)
				if err != nil {
					return err
				}
				next = currentGrammar
			} else {
				// Declared AT(uri:*) or AT(*) on 1st level
				qnc, err = e.encodeDeclaredAT(ei, uri, localName)
				if err != nil {
					return err
				}
				next = ei.GetNextGrammar()
			}
		}
	}

	if qnc == nil {
		return fmt.Errorf("qname context is nil")
	}

	if e.preservePrefix {
		if err := e.encodeQNamePrefix(qnc, prefix, e.channel); err != nil {
			return err
		}
	}

	// so far: event-code has been written & datatype is settled
	// the actual value is still missing
	//TODO: Check!!!
	if err := e.writeValue(qnc); err != nil {
		return err
	}

	// update current rule
	if next == nil {
		return fmt.Errorf("next rule is nil")
	}
	e.updateCurrentRule(next)

	if value.GetValueType() == ValueTypeString && XML_NS_URI == uri {
		ec := e.getElementContext()
		valueS, err := value.ToString()
		if err != nil {
			return err
		}
		if valueS == "preserve" {
			e.isXMLSpacePreserve = true
			ec.SetXMLSpacePreserve(utils.AsPtr(true))
		} else if valueS == "default" {
			e.isXMLSpacePreserve = false
			ec.SetXMLSpacePreserve(utils.AsPtr(false))
		}
	}

	e.lastEvent = EventTypeAttribute

	return nil
}

func (e *AbstractEXIBodyEncoder) encodeAttributeEventCodeUndeclared(currentGrammar Grammar, localName string) error {
	ecATUndeclared := e.fidelityOptions.Get2ndLevelEventCode(EventTypeAttributeGenericUndeclared, currentGrammar)

	if ecATUndeclared == NotFound {
		if !e.fidelityOptions.isStrict {
			return errors.New("fidelity options are not strict")
		}
		return fmt.Errorf("attribute '%s' cannot be encoded", localName)
	}

	// encode event-code
	if err := e.encode2ndLevelEventCode(ecATUndeclared); err != nil {
		return err
	}

	return nil
}

func (e *AbstractEXIBodyEncoder) encodeDeclaredAT(ei Production, uri, localName string) (*QNameContext, error) {
	// eventCode
	if err := e.encode1stLevelEventCode(ei.GetEventCode()); err != nil {
		return nil, err
	}

	var qnc *QNameContext
	var err error
	if ei.GetEvent().IsEventType(EventTypeAttributeNS) {
		// declared AT(uri:*)
		atNS := ei.GetEvent().(*AttributeNS)
		// localname only
		ruc := e.GetURIByNamespaceID(atNS.GetNamespaceUriID())
		qnc, err = e.encodeLocalName(localName, ruc, e.channel)
		if err != nil {
			return nil, err
		}
	} else {
		// declared AT(*)
		qnc, err = e.encodeQName(uri, localName, e.channel)
		if err != nil {
			return nil, err
		}
	}

	return qnc, nil
}

func (e *AbstractEXIBodyEncoder) encodeUndeclaredAT(currentGrammar Grammar, uri, localName string) (*QNameContext, error) {
	// event-code
	if err := e.encodeAttributeEventCodeUndeclared(currentGrammar, localName); err != nil {
		return nil, err
	}

	// qualified name
	qnc, err := e.encodeQName(uri, localName, e.channel)
	if err != nil {
		return nil, err
	}

	// learn attribute event
	if err := currentGrammar.LearnAttribute(NewAttribute(qnc)); err != nil {
		return nil, err
	}
	e.productionLearningCounting(currentGrammar)

	return qnc, nil
}

func (e *AbstractEXIBodyEncoder) getGlobalAttribute(uri, localName string) (*Attribute, error) {
	ruc := e.GetURI(uri)
	if ruc != nil {
		return e.getGlobalAttributeWithRuntimeUriContext(ruc, localName)
	}

	return nil, nil
}

func (e *AbstractEXIBodyEncoder) getGlobalAttributeWithRuntimeUriContext(ruc *RuntimeUriContext, localName string) (*Attribute, error) {
	if ruc == nil {
		return nil, errors.New("runtime uri context is nil")
	}
	qnc := ruc.GetQNameContextByLocalName(localName)
	if qnc != nil {
		return qnc.GetGlobalAttribute(), nil
	}

	return nil, nil
}

// returns false if no CH datatype is available or schema-less
func (e *AbstractEXIBodyEncoder) getDatatypeWhiteSpace() (WhiteSpace, bool) {
	currentGrammar := e.getCurrentGrammar()
	if currentGrammar.IsSchemaInformed() && currentGrammar.GetNumberOfEvents() > 0 {
		prod := currentGrammar.GetProduction(0)
		if prod.GetEvent().GetEventType() == EventTypeCharacters {
			ch := prod.GetEvent().(*Characters)
			return ch.GetDataType().GetWhiteSpace(), true
		}
	}

	return -1, false
}

func (e *AbstractEXIBodyEncoder) replace(runes []rune, len int) {
	// replace
	// All occurrences of #x9 (tab), #xA (line feed) and #xD (carriage
	// return) are replaced with #x20 (space)
	for i := range len {
		if runes[i] == '\t' || runes[i] == '\n' || runes[i] == '\r' {
			runes[i] = ' '
		}
	}
}

func (e *AbstractEXIBodyEncoder) shiftLeft(runes []rune, pos, len int) {
	copy(runes[pos:], runes[pos+1:pos+len])
}

func (e *AbstractEXIBodyEncoder) collapse(runes []rune, len int) int {
	// collapse
	// After the processing implied by replace, contiguous sequences of
	// #x20's are collapsed to a single #x20, and leading and trailing
	// #x20's are removed.
	e.replace(runes, len)

	newLen := len

	// contiguous sequences of #x20's are collapsed to a single #x20
	if newLen > 1 {
		i := 0
		for i < (newLen - 1) {
			thisRune := runes[i]
			nextRune := runes[i+1]

			if thisRune == ' ' && nextRune == ' ' {
				// eliminate one space
				e.shiftLeft(runes, i, newLen-1)
				newLen--
			} else {
				i++
			}
		}
	}

	// leading and trailing #x20's are removed
	newLen = e.trimSpaces(runes, newLen)
	return newLen
}

func (e *AbstractEXIBodyEncoder) trimSpaces(runes []rune, len int) int {
	// leading and trailing #x20's are removed
	newLen := len

	// leading
	for newLen > 0 && runes[0] == ' ' {
		// eliminate one leading space
		e.shiftLeft(runes, 0, newLen)
		newLen--
	}

	// trailing
	i := newLen - 1
	for i >= 0 && runes[i] == ' ' {
		// eliminate one trailing space
		newLen--
		i--
	}

	return newLen
}

func (e *AbstractEXIBodyEncoder) isSolelyWS(runes []rune, len int) bool {
	for _, r := range runes {
		if !unicode.IsSpace(r) {
			return false
		}
	}

	return true
}

func (e *AbstractEXIBodyEncoder) trimWS(runes []rune, len int) int {
	// leading and trailing whitespaces are removed
	newLen := len

	// leading
	for newLen > 0 && unicode.IsSpace(runes[0]) {
		// eliminate one leading space
		e.shiftLeft(runes, 0, newLen)
		newLen--
	}

	// trailing
	i := newLen - 1
	for i >= 0 && unicode.IsSpace(runes[i]) {
		// eliminate one trailing space
		newLen--
		i--
	}

	return newLen
}

func (e *AbstractEXIBodyEncoder) modeValuesToCBuffer() (int, error) {
	length := 0
	for i := range len(e.bChars) {
		clen, err := e.bChars[i].GetCharactersLength()
		if err != nil {
			return -1, err
		}
		length += clen
	}
	if len(e.cbuffer) < length {
		e.cbuffer = make([]rune, length)
	}
	pos := 0
	for i := range len(e.bChars) {
		v := e.bChars[i]
		if err := v.FillCharactersBuffer(e.cbuffer, pos); err != nil {
			return -1, err
		}
		clen, err := v.GetCharactersLength()
		if err != nil {
			return -1, err
		}
		pos += clen
	}
	if length != pos {
		return -1, errors.New("length != pos")
	}

	return length, nil
}

func (e *AbstractEXIBodyEncoder) checkPendingCharacters(nextEvent EventType) error {
	numberOfValues := len(e.bChars)
	if numberOfValues > 0 {
		if numberOfValues == 1 && e.bChars[0].GetValueType() != ValueTypeString {
			// typed data uses its own whitespace rules
			return e.encodeCharactersForce(e.bChars[0])
		} else {
			// else: string or multiple typed values
			ws, ok := e.getDatatypeWhiteSpace()
			// Don't we want to prune insignificant whitespace characters
			wsEQ := ok && ws == WhiteSpacePreserve
			if !(e.preserveLexicalValues || e.isXMLSpacePreserve || wsEQ) {
				cbufLen, err := e.modeValuesToCBuffer()
				if err != nil {
					return err
				}
				if ok && ws == WhiteSpaceReplace {
					// replace
					// All occurrences of #x9 (tab), #xA (line feed) and #xD
					// (carriage return) are replaced with #x20 (space)
					e.replace(e.cbuffer, cbufLen)
				} else if ok && ws == WhiteSpaceCollapse {
					// collapse
					// After the processing implied by replace, contiguous
					// sequences of #x20's are collapsed to a single #x20,
					// and leading and trailing #x20's are removed.
					e.replace(e.cbuffer, cbufLen)
					cbufLen = e.collapse(e.cbuffer, cbufLen)
				} else {
					// schema-less, no datatype
					// https://lists.w3.org/Archives/Public/public-exi/2015Oct/0008.html
					// If it is schema-less:
					// - Simple data (data between s+e) are all preserved.
					// - For complex data (data between s+s, e+s, e+e), it
					// is same as schema-informed case.
					if (e.lastEvent == EventTypeStartElement || e.lastEvent == EventTypeAttribute || e.lastEvent == EventTypeAttributeXsiNil ||
						e.lastEvent == EventTypeAttributeXsiType || e.lastEvent == EventTypeNamespaceDeclaration) &&
						(nextEvent == EventTypeEndElement || nextEvent == EventTypeComment || nextEvent == EventTypeProcessingInstruction ||
							nextEvent == EventTypeDocType) {
						// simple data --> preserve
					} else {
						// For complex data (data between s+s, e+s, e+e),
						// whitespaces nodes (i.e.
						// strings that consist solely of whitespaces) are
						// removed
						if e.isSolelyWS(e.cbuffer, cbufLen) {
							cbufLen = 0
						}
					}
				}

				if cbufLen == 0 {
					// --> omit empty string
				} else {
					sv := NewStringValueFromString(string(e.cbuffer[:cbufLen]))
					if err := e.encodeCharactersForce(sv); err != nil {
						return err
					}
				}
			} else {
				// preserve data as is
				if numberOfValues == 1 {
					if err := e.encodeCharactersForce(e.bChars[0]); err != nil {
						return err
					}
				} else {
					// collapse all events to a single one (not very
					// efficient in most of the cases)
					length, err := e.modeValuesToCBuffer()
					if err != nil {
						return err
					}
					sv := NewStringValueFromString(string(e.cbuffer[:length]))
					if err := e.encodeCharactersForce(sv); err != nil {
						return err
					}
				}
			}
		}
		e.bChars = []Value{}
	}

	return nil
}

func (e *AbstractEXIBodyEncoder) EncodeCharacters(chars Value) error {
	e.bChars = append(e.bChars, chars)
	return nil
}

func (e *AbstractEXIBodyEncoder) encodeCharactersForce(chars Value) error {
	currentGrammar := e.getCurrentGrammar()
	ei := currentGrammar.GetProduction(EventTypeCharacters)

	// valid value and valid event-code ?
	valid, err := e.isTypeValid((ei.GetEvent().(DatatypeEvent)).GetDatatype(), chars)
	if err != nil {
		return err
	}
	if ei != nil && valid {
		// right characters event found & data type-valid
		// --> encode EventCode, schema-valid content plus grammar moves
		// on
		if err := e.encode1stLevelEventCode(ei.GetEventCode()); err != nil {
			return err
		}
		//TODO: Check!!!!
		if err := e.writeValue(e.getElementContext().qnc); err != nil {
			return err
		}
		e.updateCurrentRule(ei.GetNextGrammar())
	} else {
		// generic CH (on first level)
		ei = currentGrammar.GetProduction(EventTypeCharactersGeneric)

		if ei != nil {
			// encode EventCode
			if err := e.encode1stLevelEventCode(ei.GetEventCode()); err != nil {
				return err
			}
			// encode schema-invalid content as string
			if _, err := e.isTypeValid(BuiltInGetDefaultDatatype(), chars); err != nil {
				return err
			}
			//TODO: Check!!!
			if err := e.writeValue(e.getElementContext().qnc); err != nil {
				return err
			}
			// update current rule
			e.updateCurrentRule(ei.GetNextGrammar())
		} else {
			// Undeclared CH can be found on 2nd level
			ecCHUndeclared := e.fidelityOptions.Get2ndLevelEventCode(EventTypeCharactersGenericUndeclared, currentGrammar)

			if ecCHUndeclared == NotFound {
				if e.exiFactory.IsFragment() {
					// characters in "outer" fragment element
					e.emitWarning("skip ch")
				} else if !e.isXMLSpacePreserve && e.fidelityOptions.IsStrict() {
					charsS, err := chars.ToString()
					if err != nil {
						return err
					}
					if len(strings.TrimSpace(charsS)) == 0 {
						e.emitWarning("skip ch: " + charsS)
					}
				} else {
					return errors.New("characters cannot be encoded")
				}
			} else {
				var updContextRule Grammar

				switch e.limitGrammars() {
				case ProfileDisablingMechanismXsiType:
					if err := e.insertXsiTypeAnyType(); err != nil {
						return err
					}
					currentGrammar = e.getCurrentGrammar()
					// encode 1st level EventCode
					ei = currentGrammar.GetProduction(EventTypeCharactersGeneric)
					if ei == nil {
						return errors.New("ei == nil")
					}
					if err := e.encode1stLevelEventCode(ei.GetEventCode()); err != nil {
						return err
					}
					// next rule
					updContextRule = ei.GetNextGrammar()
				case ProfileDisablingMechanismGhostProduction:
					fallthrough
				default:
					// encode [undeclared] event-code
					if err := e.encode2ndLevelEventCode(ecCHUndeclared); err != nil {
						return err
					}
					// learn characters event ?
					currentGrammar.LearnCharacters()
					e.productionLearningCounting(currentGrammar)
					// next rule
					updContextRule = currentGrammar.GetElementContentGrammar()
				}

				// content as string
				if _, err := e.isTypeValid(BuiltInGetDefaultDatatype(), chars); err != nil {
					return err
				}
				//TODO: Check!!!
				if err := e.writeValue(e.getElementContext().qnc); err != nil {
					return err
				}
				// update current rule
				e.updateCurrentRule(updContextRule)
			}
		}
	}

	return nil
}

func (e *AbstractEXIBodyEncoder) EncodeDocType(name, publicID, systemID, text string) error {
	if e.fidelityOptions.IsFidelityEnabled(FeatureDTD) {
		if err := e.checkPendingCharacters(EventTypeDocType); err != nil {
			return err
		}

		// DOCTYPE can be found on 2nd level
		ec2 := e.fidelityOptions.Get2ndLevelEventCode(EventTypeDocType, e.getCurrentGrammar())
		if err := e.encode2ndLevelEventCode(ec2); err != nil {
			return err
		}

		// name, public, system, text AS string
		if err := e.writeString(name); err != nil {
			return err
		}
		if err := e.writeString(publicID); err != nil {
			return err
		}
		if err := e.writeString(systemID); err != nil {
			return err
		}
		if err := e.writeString(text); err != nil {
			return err
		}
	}

	return nil
}

func (e *AbstractEXIBodyEncoder) doLimitGrammarLearningForErCmPi() error {
	switch e.limitGrammars() {
	case ProfileDisablingMechanismXsiType:
		if err := e.insertXsiTypeAnyType(); err != nil {
			return err
		}
	default:
		// no special action
	}

	return nil
}

func (e *AbstractEXIBodyEncoder) EncodeEntityReference(name string) error {
	if e.fidelityOptions.IsFidelityEnabled(FeatureDTD) {
		if err := e.checkPendingCharacters(EventTypeEntityReference); err != nil {
			return err
		}

		// grammar learning restricting (if necessary)
		if err := e.doLimitGrammarLearningForErCmPi(); err != nil {
			return err
		}

		// EntityReference can be found on 2nd level
		currentGrammar := e.getCurrentGrammar()
		ec2 := e.fidelityOptions.Get2ndLevelEventCode(EventTypeEntityReference, currentGrammar)
		if err := e.encode2ndLevelEventCode(ec2); err != nil {
			return err
		}

		// name as string
		if err := e.writeString(name); err != nil {
			return err
		}

		// update current rule
		e.updateCurrentRule(currentGrammar.GetElementContentGrammar())
	}

	return nil
}

func (e *AbstractEXIBodyEncoder) EncodeComment(ch []rune, start, length int) error {
	if e.fidelityOptions.IsFidelityEnabled(FeatureComment) {
		if err := e.checkPendingCharacters(EventTypeComment); err != nil {
			return err
		}

		// grammar learning restricting (if necessary)
		if err := e.doLimitGrammarLearningForErCmPi(); err != nil {
			return err
		}

		// comments can be found on 3rd level
		currentGrammar := e.getCurrentGrammar()
		ec3 := e.fidelityOptions.Get3rdLevelEventCode(EventTypeComment)
		if err := e.encode3rdLevelEventCode(ec3); err != nil {
			return err
		}

		// encode CM content
		if err := e.writeString(string(ch[start : start+length])); err != nil {
			return err
		}

		// update current rule
		e.updateCurrentRule(currentGrammar.GetElementContentGrammar())
	}

	return nil
}

func (e *AbstractEXIBodyEncoder) EncodeProcessingInstruction(target, data string) error {
	if e.fidelityOptions.IsFidelityEnabled(FeaturePI) {
		if err := e.checkPendingCharacters(EventTypeProcessingInstruction); err != nil {
			return err
		}

		// grammar learning restricting (if necessary)
		if err := e.doLimitGrammarLearningForErCmPi(); err != nil {
			return err
		}

		// processing instructions can be found on 3rd level
		currentGrammar := e.getCurrentGrammar()
		ec3 := e.fidelityOptions.Get3rdLevelEventCode(EventTypeProcessingInstruction)
		if err := e.encode3rdLevelEventCode(ec3); err != nil {
			return err
		}

		// encode PI content
		if err := e.writeString(target); err != nil {
			return err
		}
		if err := e.writeString(data); err != nil {
			return err
		}

		// update current rule
		e.updateCurrentRule(currentGrammar.GetElementContentGrammar())
	}

	return nil
}

/*
	EXIStreamDecoderImpl implementation
*/

type EXIStreamDecoderImpl struct {
	EXIStreamDecoder
	exiHeader        *EXIHeaderDecoder
	exiBody          EXIBodyDecoder
	noOptionsFactory EXIFactory
}

func NewEXIStreamDecoderImpl(noOptionsFactory EXIFactory) (*EXIStreamDecoderImpl, error) {
	exiBody, err := noOptionsFactory.CreateEXIBodyDecoder()
	if err != nil {
		return nil, err
	}
	return &EXIStreamDecoderImpl{
		exiHeader:        NewEXIHeaderDecoder(),
		exiBody:          exiBody,
		noOptionsFactory: noOptionsFactory,
	}, nil
}

func (d *EXIStreamDecoderImpl) GetBodyOnlyDecoder(reader *bufio.Reader) (EXIBodyDecoder, error) {
	if err := d.exiBody.SetInputStream(reader); err != nil {
		return nil, err
	}
	return d.exiBody, nil
}

func (d *EXIStreamDecoderImpl) DecodeHeader(reader *bufio.Reader) (EXIBodyDecoder, error) {
	headerChannel := NewBitDecoderChannel(reader)
	exiFactory, err := d.exiHeader.Parse(headerChannel, d.noOptionsFactory)
	if err != nil {
		return nil, err
	}

	// update body decoder if EXI options tell to do so
	if exiFactory != d.noOptionsFactory {
		d.exiBody, err = exiFactory.CreateEXIBodyDecoder()
		if err != nil {
			return nil, err
		}
	}
	if exiFactory.GetCodingMode() == CodingModeBitPacked {
		if err := d.exiBody.SetInputChannel(headerChannel); err != nil {
			return nil, err
		}
	} else {
		if err := d.exiBody.SetInputStream(reader); err != nil {
			return nil, err
		}
	}

	return d.exiBody, nil
}

/*
	EXIStreamEncoderImpl implementation
*/

type EXIStreamEncoderImpl struct {
	EXIStreamEncoder
	exiHeader *EXIHeaderEncoder
	exiBody   *EXIBodyEncoder
	noFactory EXIFactory
}

func NewEXIStreamEncoderImpl(exiFactory EXIFactory) (*EXIStreamEncoderImpl, error) {
	return nil, nil
}

/*
	EXIBodyDecoderInOrder implementation
*/

type EXIBodyDecoderInOrder struct {
	*AbstractEXIBodyDecoder
}

func NewEXIBodyDecoderInOrder(exiFactory EXIFactory) (*EXIBodyDecoderInOrder, error) {
	abd, err := NewAbstractEXIBodyDecoder(exiFactory)
	if err != nil {
		return nil, err
	}

	return &EXIBodyDecoderInOrder{
		AbstractEXIBodyDecoder: abd,
	}, nil
}

func (d *EXIBodyDecoderInOrder) SetInputStream(reader *bufio.Reader) error {
	if err := d.UpdateInputStream(reader); err != nil {
		return err
	}
	return d.InitForEachRun()
}

func (d *EXIBodyDecoderInOrder) SetInputChannel(channel DecoderChannel) error {
	if err := d.UpdateInputChannel(channel); err != nil {
		return err
	}
	return d.InitForEachRun()
}

func (d *EXIBodyDecoderInOrder) UpdateInputStream(reader *bufio.Reader) error {
	codingMode := d.exiFactory.GetCodingMode()

	// setup data-stream only
	if codingMode == CodingModeBitPacked {
		// create new bit-aligned channel
		if err := d.UpdateInputChannel(NewBitDecoderChannel(reader)); err != nil {
			return err
		}
	} else {
		if codingMode != CodingModeBytePacked {
			return fmt.Errorf("unexpected coding mode: %d", codingMode)
		}
		// create new byte-aligned channel
		if err := d.UpdateInputChannel(NewByteDecoderChannel(reader)); err != nil {
			return err
		}
	}

	return nil
}

func (d *EXIBodyDecoderInOrder) UpdateInputChannel(channel DecoderChannel) error {
	d.channel = channel
	return nil
}

func (d *EXIBodyDecoderInOrder) GetChannel() DecoderChannel {
	return d.channel
}

func (d *EXIBodyDecoderInOrder) InitForEachRun() error {
	if err := d.AbstractEXIBodyDecoder.InitForEachRun(); err != nil {
		return err
	}

	d.nextEvent = nil
	d.nextEventType = EventTypeStartDocument

	return nil
}

func (d *EXIBodyDecoderInOrder) Next() (EventType, bool, error) {
	if d.nextEventType == EventTypeEndDocument {
		return -1, false, nil
	} else {
		ec, err := d.decodeEventCode()
		if err != nil {
			return -1, false, err
		}
		return ec, true, nil
	}
}

func (d *EXIBodyDecoderInOrder) DecodeStartDocument() error {
	return d.decodeStartDocumentStructure()
}

func (d *EXIBodyDecoderInOrder) DecodeEndDocument() error {
	return d.decodeEndDocumentStructure()
}

func (d *EXIBodyDecoderInOrder) DecodeStartElement() (*QNameContext, error) {
	switch d.nextEventType {
	case EventTypeStartElement:
		return d.decodeStartElementStructure()
	case EventTypeStartElementNS:
		return d.decodeStartElementNSStructure()
	case EventTypeStartElementGeneric:
		return d.decodeStartElementGenericStructure()
	case EventTypeStartElementGenericUndeclared:
		return d.decodeStartElementGenericUndeclaredStructure()
	default:
		return nil, fmt.Errorf("invalid decode state: %d", d.nextEventType)
	}
}

func (d *EXIBodyDecoderInOrder) GetElementPrefix() *string {
	return d.getElementContext().GetPrefix()
}

func (d *EXIBodyDecoderInOrder) GetElementQNameAsString() string {
	return d.getElementContext().GetQNameAsString(d.preservePrefix)
}

func (d *EXIBodyDecoderInOrder) DecodeEndElement() (*QNameContext, error) {
	var ec *ElementContext
	var err error
	switch d.nextEventType {
	case EventTypeEndElement:
		ec, err = d.decodeEndElementStructure()
		if err != nil {
			return nil, err
		}
	case EventTypeEndElementUndeclared:
		ec, err = d.decodeEndElementUndeclaredStructure()
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid decode state: %d", d.nextEventType)
	}

	return ec.qnc, nil
}

func (d *EXIBodyDecoderInOrder) DecodeAttributeXsiNil() (*QNameContext, error) {
	if d.nextEventType != EventTypeAttributeXsiNil {
		return nil, fmt.Errorf("next event type != Attribute xsi:nil")
	}
	if err := d.decodeAttributeXsiNilStructure(); err != nil {
		return nil, err
	}

	return d.attributeQNameContext, nil
}

func (d *EXIBodyDecoderInOrder) DecodeAttributeXsiType() (*QNameContext, error) {
	if d.nextEventType != EventTypeAttributeXsiType {
		return nil, fmt.Errorf("next event type != Attribute xsi:type")
	}
	if err := d.decodeAttributeXsiTypeStructure(); err != nil {
		return nil, err
	}

	return d.attributeQNameContext, nil

}

func (d *EXIBodyDecoderInOrder) readAttributeContentWithDatatype(dt Datatype) error {
	value, err := d.typeDecoder.ReadValue(dt, d.attributeQNameContext, d.channel, d.stringDecoder)
	if err != nil {
		return err
	}
	d.attributeValue = value
	return nil
}

func (d *EXIBodyDecoderInOrder) readAttributeContent() error {
	if d.attributeQNameContext.GetNamespaceUriID() == d.getXsiTypeContext().GetNamespaceUriID() {
		localNameID := d.attributeQNameContext.GetLocalNameID()
		if localNameID == d.getXsiTypeContext().GetLocalNameID() {
			if err := d.decodeAttributeXsiTypeStructure(); err != nil {
				return err
			}
		} else if localNameID == d.getXsiTypeContext().GetLocalNameID() && d.getCurrentGrammar().IsSchemaInformed() { //TODO: Weird condition. XSI:NIL?
			if err := d.decodeAttributeXsiNilStructure(); err != nil {
				return err
			}
		} else {
			if err := d.readAttributeContentWithDatatype(BuiltInGetDefaultDatatype()); err != nil {
				return err
			}
		}
	} else {
		// Attribute globalAT;
		dt := BuiltInGetDefaultDatatype()

		if d.getCurrentGrammar().IsSchemaInformed() && d.attributeQNameContext.GetGlobalAttribute() != nil {
			dt = d.attributeQNameContext.GetGlobalAttribute().GetDataType()
		}

		if err := d.readAttributeContentWithDatatype(dt); err != nil {
			return err
		}
	}

	return nil
}

func (d *EXIBodyDecoderInOrder) DecodeAttribute() (*QNameContext, error) {
	switch d.nextEventType {
	case EventTypeAttribute:
		dt, err := d.decodeAttributeStructure()
		if err != nil {
			return nil, err
		}
		if d.attributeQNameContext.Equals(d.getXsiTypeContext()) {
			if err := d.decodeAttributeXsiTypeStructure(); err != nil {
				return nil, err
			}
		} else {
			if err := d.readAttributeContentWithDatatype(dt); err != nil {
				return nil, err
			}
		}
	case EventTypeAttributeNS:
		if err := d.decodeAttributeNSStructure(); err != nil {
			return nil, err
		}
		if err := d.readAttributeContent(); err != nil {
			return nil, err
		}
	case EventTypeAttributeGeneric:
		if err := d.decodeAttributeGenericStructure(); err != nil {
			return nil, err
		}
		if err := d.readAttributeContent(); err != nil {
			return nil, err
		}
	case EventTypeAttributeGenericUndeclared:
		if err := d.decodeAttributeGenericUndeclaredStructure(); err != nil {
			return nil, err
		}
		if err := d.readAttributeContent(); err != nil {
			return nil, err
		}
	case EventTypeAttributeInvalidValue:
		if _, err := d.decodeAttributeStructure(); err != nil {
			return nil, err
		}
		// Note: attribute content datatype is not the right one (invalid)
		if err := d.readAttributeContentWithDatatype(BuiltInGetDefaultDatatype()); err != nil {
			return nil, err
		}
	case EventTypeAttributeAnyInvalidValue:
		if err := d.decodeAttributeAnyInvalidValueStructure(); err != nil {
			return nil, err
		}
		if err := d.readAttributeContentWithDatatype(BuiltInGetDefaultDatatype()); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid decode state: %d", d.nextEventType)
	}

	return d.attributeQNameContext, nil
}

func (d *EXIBodyDecoderInOrder) DecodeNamespaceDeclaration() (*NamespaceDeclarationContainer, error) {
	return d.decodeNamespaceDeclarationStructure()
}

func (d *EXIBodyDecoderInOrder) GetDeclaredPrefixDeclarations() []NamespaceDeclarationContainer {
	return d.getElementContext().nsDeclarations
}

func (d *EXIBodyDecoderInOrder) DecodeCharacters() (Value, error) {
	var dt Datatype
	var err error

	switch d.nextEventType {
	case EventTypeCharacters:
		dt, err = d.decodeCharactersStructure()
		if err != nil {
			return nil, err
		}
	case EventTypeCharactersGeneric:
		err = d.decodeCharactersGenericStructure()
		if err != nil {
			return nil, err
		}
		dt = BuiltInGetDefaultDatatype()
	case EventTypeCharactersGenericUndeclared:
		err = d.decodeCharactersGenericUndeclaredStructure()
		if err != nil {
			return nil, err
		}
		dt = BuiltInGetDefaultDatatype()
	default:
		return nil, fmt.Errorf("invalid decode state: %d", d.nextEventType)
	}

	return d.typeDecoder.ReadValue(dt, d.getElementContext().qnc, d.channel, d.stringDecoder)
}

func (d *EXIBodyDecoderInOrder) DecodeDocType() (*DocTypeContainer, error) {
	return d.decodeDocTypeStructure()
}

func (d *EXIBodyDecoderInOrder) DecodeEntityReference() ([]rune, error) {
	return d.decodeEntityReferenceStructure()
}

func (d *EXIBodyDecoderInOrder) DecodeComment() ([]rune, error) {
	return d.decodeCommentStructure()
}

func (d *EXIBodyDecoderInOrder) DecodeProcessingInstruction() (ProcessingInstructionContainer, error) {
	return d.decodeProcessingInstructionStructure()
}

/*
	EXIBodyDecoderInOrder implementation
*/

type EXIBodyEncoderInOrder struct {
	*AbstractEXIBodyEncoder
}

func NewEXIBodyEncoderInOrder(exiFactory EXIFactory) (*EXIBodyEncoderInOrder, error) {
	abe, err := NewAbstractEXIBodyEncoder(exiFactory)
	if err != nil {
		return nil, err
	}
	return &EXIBodyEncoderInOrder{
		AbstractEXIBodyEncoder: abe,
	}, nil
}

func (e *EXIBodyEncoderInOrder) SetOutputStream(writer bufio.Writer) error {
	codingMode := e.exiFactory.GetCodingMode()

	// setup data-stream only
	if codingMode == CodingModeBitPacked {
		// create new bit-aligned channel
		e.SetOutputChannel(NewBitEncoderChannel(writer))
	} else {
		if codingMode != CodingModeBytePacked {
			return errors.New("coding mode != bype packed")
		}
		// create new byte-aligned channel
		e.SetOutputChannel(NewByteEncoderChannel(writer))
	}

	return nil
}

func (e *EXIBodyEncoderInOrder) SetOutputChannel(channel EncoderChannel) error {
	e.channel = channel
	return nil
}

func (e *EXIBodyEncoderInOrder) WriteValue(qnc *QNameContext) error {
	return e.typeEncoder.WriteValue(qnc, e.channel, e.stringEncoder)
}

/*
	EXIBodyDecoderInOrderSC implementation
*/

type EXIBodyDecoderInOrderSC struct {
	*EXIBodyDecoderInOrder
	scDecoder *EXIBodyDecoderInOrderSC
}

func NewEXIBodyDecoderInOrderSC(exiFactory EXIFactory) (*EXIBodyDecoderInOrderSC, error) {
	ebdio, err := NewEXIBodyDecoderInOrder(exiFactory)
	if err != nil {
		return nil, err
	}
	if !ebdio.fidelityOptions.IsFidelityEnabled(FeatureSC) {
		return nil, errors.New("self contained feature is not enabled")
	}
	return &EXIBodyDecoderInOrderSC{
		EXIBodyDecoderInOrder: ebdio,
		scDecoder:             nil,
	}, nil
}

func (d *EXIBodyDecoderInOrderSC) InitForEachRun() error {
	if err := d.EXIBodyDecoderInOrder.InitForEachRun(); err != nil {
		return err
	}
	// clear possibly remaining decoder
	d.scDecoder = nil
	return nil
}

func (d *EXIBodyDecoderInOrderSC) SkipSCElement(skip int64) error {
	// Note: Bytes to be skipped need to be known
	if d.nextEventType != EventTypeSelfContained {
		return errors.New("next event type is not self contained element")
	}
	if err := d.channel.Align(); err != nil {
		return err
	}
	for range skip {
		if _, err := d.channel.Decode(); err != nil {
			return err
		}
	}
	d.popElement()
	return nil
}

func (d *EXIBodyDecoderInOrderSC) Next() (EventType, bool, error) {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.Next()
	} else {
		et, exists, err := d.scDecoder.Next()
		if err != nil {
			return -1, false, err
		}
		if !exists {
			return -1, false, errors.New("no next events")
		}
		if et == EventTypeEndDocument {
			if err := d.scDecoder.DecodeEndDocument(); err != nil {
				return -1, false, err
			}
			// Skip to the next byte-aligned boundary in the stream if it is
			// not already at such a boundary
			if err := d.channel.Align(); err != nil {
				return -1, false, err
			}
			// indicate that SC portion is over
			d.scDecoder = nil
			d.popElement()

			et, exists, err = d.EXIBodyDecoderInOrder.Next()
			if err != nil {
				return -1, false, err
			}
		}

		return et, exists, nil
	}
}

func (d *EXIBodyDecoderInOrderSC) DecodeStartDocument() error {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.DecodeStartDocument()
	} else {
		return d.scDecoder.DecodeStartDocument()
	}
}

func (d *EXIBodyDecoderInOrderSC) DecodeEndDocument() error {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.DecodeEndDocument()
	} else {
		return errors.New("self contained element is not closed properly?")
	}
}

func (d *EXIBodyDecoderInOrderSC) DecodeStartElement() (*QNameContext, error) {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.DecodeStartElement()
	} else {
		return d.scDecoder.DecodeStartElement()
	}
}

func (d *EXIBodyDecoderInOrderSC) DecodeStartSelfContainedFragment() error {
	if d.scDecoder == nil {
		// SC Factory & Decoder
		scEXIFactory := d.exiFactory.Clone()
		scEXIFactory.SetFragment(true)
		decoder, err := scEXIFactory.CreateEXIBodyDecoder()
		if err != nil {
			return err
		}
		d.scDecoder = decoder.(*EXIBodyDecoderInOrderSC)
		d.scDecoder.channel = d.channel
		d.scDecoder.SetErrorHandler(d.errorHandler)
		if err := d.scDecoder.InitForEachRun(); err != nil {
			return err
		}

		// Skip to the next byte-aligned boundary in the stream if it is not
		// already at such a boundary
		if err := d.channel.Align(); err != nil {
			return err
		}

		// Evaluate the sequence of events (SD, SE(qname), content, ED)
		// according to the Fragment grammar
		if err := d.scDecoder.DecodeStartDocument(); err != nil {
			return err
		}
		et, exists, err := d.Next()
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("no next event")
		}
		switch et {
		case EventTypeStartElement, EventTypeStartElementNS, EventTypeStartElementGeneric, EventTypeStartElementGenericUndeclared:
			if _, err := d.scDecoder.DecodeStartElement(); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported EventType %d in SelfContained element", et)
		}
	} else {
		if err := d.scDecoder.DecodeStartSelfContainedFragment(); err != nil {
			return err
		}
	}

	return nil
}

func (d *EXIBodyDecoderInOrderSC) DecodeEndElement() (*QNameContext, error) {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.DecodeEndElement()
	} else {
		return d.scDecoder.DecodeEndElement()
	}
}

func (d *EXIBodyDecoderInOrderSC) GetElementPrefix() *string {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.GetElementPrefix()
	} else {
		return d.scDecoder.GetElementPrefix()
	}
}

func (d *EXIBodyDecoderInOrderSC) GetElementQNameAsString() string {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.GetElementQNameAsString()
	} else {
		return d.scDecoder.GetElementQNameAsString()
	}
}

func (d *EXIBodyDecoderInOrderSC) DecodeAttributeXsiNil() (*QNameContext, error) {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.DecodeAttributeXsiNil()
	} else {
		return d.scDecoder.DecodeAttributeXsiNil()
	}
}

func (d *EXIBodyDecoderInOrderSC) DecodeAttributeXsiType() (*QNameContext, error) {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.DecodeAttributeXsiType()
	} else {
		return d.scDecoder.DecodeAttributeXsiType()
	}
}

func (d *EXIBodyDecoderInOrderSC) DecodeAttribute() (*QNameContext, error) {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.DecodeAttribute()
	} else {
		return d.scDecoder.DecodeAttribute()
	}
}

func (d *EXIBodyDecoderInOrderSC) GetAttributePrefix() *string {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.GetAttributePrefix()
	} else {
		return d.scDecoder.GetAttributePrefix()
	}
}

func (d *EXIBodyDecoderInOrderSC) GetAttributeQNameAsString() string {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.GetAttributeQNameAsString()
	} else {
		return d.scDecoder.GetAttributeQNameAsString()
	}
}

func (d *EXIBodyDecoderInOrderSC) GetAttributeValue() Value {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.GetAttributeValue()
	} else {
		return d.scDecoder.GetAttributeValue()
	}
}

func (d *EXIBodyDecoderInOrderSC) GetDeclaredPrefixDeclarations() []NamespaceDeclarationContainer {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.GetDeclaredPrefixDeclarations()
	} else {
		return d.scDecoder.GetDeclaredPrefixDeclarations()
	}
}

func (d *EXIBodyDecoderInOrderSC) DecodeNamespaceDeclaration() (*NamespaceDeclarationContainer, error) {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.DecodeNamespaceDeclaration()
	} else {
		return d.scDecoder.DecodeNamespaceDeclaration()
	}
}

func (d *EXIBodyDecoderInOrderSC) DecodeCharacters() (Value, error) {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.DecodeCharacters()
	} else {
		return d.scDecoder.DecodeCharacters()
	}
}

func (d *EXIBodyDecoderInOrderSC) DecodeDocType() (*DocTypeContainer, error) {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.DecodeDocType()
	} else {
		return d.scDecoder.DecodeDocType()
	}
}

func (d *EXIBodyDecoderInOrderSC) DecodeEntityReference() ([]rune, error) {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.DecodeEntityReference()
	} else {
		return d.scDecoder.DecodeEntityReference()
	}
}

func (d *EXIBodyDecoderInOrderSC) DecodeComment() ([]rune, error) {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.DecodeComment()
	} else {
		return d.scDecoder.DecodeComment()
	}
}

func (d *EXIBodyDecoderInOrderSC) DecodeProcessingInstruction() (ProcessingInstructionContainer, error) {
	if d.scDecoder == nil {
		return d.EXIBodyDecoderInOrder.DecodeProcessingInstruction()
	} else {
		return d.scDecoder.DecodeProcessingInstruction()
	}
}

/*
	EXIBodyEncoderInOrderSC implementation
*/

type EXIBodyEncoderInOrderSC struct {
	*EXIBodyEncoderInOrder
	scEncoder *EXIBodyEncoderInOrderSC
}

func NewEXIBodyEncoderInOrderSC(exiFactory EXIFactory) (*EXIBodyEncoderInOrderSC, error) {
	ebeio, err := NewEXIBodyEncoderInOrder(exiFactory)
	if err != nil {
		return nil, err
	}
	return &EXIBodyEncoderInOrderSC{
		EXIBodyEncoderInOrder: ebeio,
		scEncoder:             nil,
	}, nil
}

func (e *EXIBodyEncoderInOrderSC) InitForEachRun() error {
	if err := e.EXIBodyEncoderInOrder.InitForEachRun(); err != nil {
		return err
	}
	e.scEncoder = nil
	return nil
}

func (e *EXIBodyEncoderInOrderSC) SetErrorHandler(errorHandler ErrorHandler) {
	if e.scEncoder == nil {
		e.EXIBodyEncoderInOrder.SetErrorHandler(errorHandler)
	} else {
		e.scEncoder.SetErrorHandler(errorHandler)
	}
}

func (e *EXIBodyEncoderInOrderSC) EncodeStartDocument() error {
	if e.scEncoder == nil {
		return e.EXIBodyEncoderInOrder.EncodeStartDocument()
	} else {
		return e.scEncoder.EncodeStartDocument()
	}
}

func (e *EXIBodyEncoderInOrderSC) EncodeEndDocument() error {
	if e.scEncoder == nil {
		return e.EXIBodyEncoderInOrder.EncodeEndDocument()
	} else {
		return e.scEncoder.EncodeEndDocument()
	}
}

func (e *EXIBodyEncoderInOrderSC) encodeEndSC() error {
	// end SC fragment
	if err := e.scEncoder.EncodeEndDocument(); err != nil {
		return err
	}
	// Skip to the next byte-aligned boundary in the stream if it is
	// not already at such a boundary
	if err := e.channel.Align(); err != nil {
		return err
	}
	// indicate that SC portion is over
	e.scEncoder = nil
	e.EXIBodyEncoderInOrder.popElement()

	// NOTE: NO outer EE
	// Spec says
	// "Evaluate the sequence of events (SD, SE(qname), content, ED) .."
	// e.g., "sc" is self-Contained element
	// Sequence: <sc>foo</sc>
	// --> SE(sc) --> SC --> SD --> SE(sc) --> CH --> EE --> ED
	// content == SE(sc) --> CH --> EE

	return nil
}

func (e *EXIBodyEncoderInOrderSC) EncodeStartElement(uri, localName string, prefix *string) error {
	if e.scEncoder == nil {
		if err := e.EXIBodyEncoderInOrder.EncodeStartElement(uri, localName, prefix); err != nil {
			return err
		}
		qname := e.getElementContext().qnc.GetQName()

		// start SC fragment?
		if e.exiFactory.IsSelfContainedElement(qname) {
			ec2 := e.fidelityOptions.Get2ndLevelEventCode(EventTypeSelfContained, e.getCurrentGrammar())
			if err := e.encode2ndLevelEventCode(ec2); err != nil {
				return err
			}

			// Skip to the next byte-aligned boundary in the stream if it is
			// not already at such a boundary
			if err := e.channel.Align(); err != nil {
				return err
			}

			// infor
			if e.exiFactory.GetSelfContainedHandler() != nil {
				if err := e.exiFactory.GetSelfContainedHandler().ScElement(&uri, &localName, e.channel); err != nil {
					return err
				}
			}

			// start SC element
			if err := e.encodeStartSC(uri, localName, prefix); err != nil {
				return err
			}
		}
	} else {
		if err := e.scEncoder.EncodeStartElement(uri, localName, prefix); err != nil {
			return err
		}
	}

	return nil
}

func (e *EXIBodyEncoderInOrderSC) encodeStartSC(uri, localName string, prefix *string) error {
	// SC Factory & Encoder
	scEXIFactory := e.exiFactory.Clone()
	scEXIFactory.SetFragment(true)
	encoder, err := scEXIFactory.CreateEXIBodyEncoder()
	if err != nil {
		return err
	}
	e.scEncoder = encoder.(*EXIBodyEncoderInOrderSC)
	e.scEncoder.channel = e.channel
	e.scEncoder.SetErrorHandler(e.errorHandler)

	// Evaluate the sequence of events (SD, SE(qname), content, ED)
	// according to the Fragment grammar
	if err := e.scEncoder.EncodeStartDocument(); err != nil {
		return err
	}
	// NO SC again
	if err := e.scEncoder.encodeStartElementNoSC(uri, localName, prefix); err != nil {
		return err
	}
	// from now on events are forwarded to the scEncoder
	if e.preservePrefix {
		// encode NS inner declaration for SE
		if err := e.scEncoder.EncodeNamespaceDeclaration(uri, prefix); err != nil {
			return err
		}
	}

	return nil
}

func (e *EXIBodyEncoderInOrderSC) encodeStartElementNoSC(uri, localName string, prefix *string) error {
	return e.EXIBodyEncoderInOrder.EncodeStartElement(uri, localName, prefix)
}

func (e *EXIBodyEncoderInOrderSC) EncodeEndElement() error {
	if e.scEncoder == nil {
		return e.EXIBodyEncoderInOrder.EncodeEndElement()
	} else {
		// fetch qname before EE
		qname := e.scEncoder.getElementContext().qnc.GetQName()
		// EE
		if err := e.scEncoder.EncodeEndElement(); err != nil {
			return err
		}

		if e.getElementContext().qnc.GetQName() == qname &&
			e.scEncoder.getCurrentGrammar().GetProduction(EventTypeEndDocument) != nil {
			return e.encodeEndSC()
		}
	}

	return nil
}

func (e *EXIBodyEncoderInOrderSC) EncodeAttribute(uri, localName string, prefix *string, value Value) error {
	if e.scEncoder == nil {
		return e.EXIBodyEncoderInOrder.EncodeAttribute(uri, localName, prefix, value)
	} else {
		return e.scEncoder.EncodeAttribute(uri, localName, prefix, value)
	}
}

func (e *EXIBodyEncoderInOrderSC) EncodeAttributeByQName(at QName, value Value) error {
	if e.scEncoder == nil {
		return e.EXIBodyEncoderInOrder.EncodeAttributeByQName(at, value)
	} else {
		return e.scEncoder.EncodeAttributeByQName(at, value)
	}
}

func (e *EXIBodyEncoderInOrderSC) EncodeNamespaceDeclaration(uri string, prefix *string) error {
	if e.scEncoder == nil {
		return e.EXIBodyEncoderInOrder.EncodeNamespaceDeclaration(uri, prefix)
	} else {
		return e.scEncoder.EncodeNamespaceDeclaration(uri, prefix)
	}
}

func (e *EXIBodyEncoderInOrderSC) EncodeAttributeXsiNil(nilValue Value, prefix *string) error {
	if e.scEncoder == nil {
		return e.EXIBodyEncoderInOrder.EncodeAttributeXsiNil(nilValue, prefix)
	} else {
		return e.scEncoder.EncodeAttributeXsiNil(nilValue, prefix)
	}
}

func (e *EXIBodyEncoderInOrderSC) EncodeAttributeXsiType(value Value, prefix *string) error {
	if e.scEncoder == nil {
		return e.EXIBodyEncoderInOrder.EncodeAttributeXsiType(value, prefix)
	} else {
		return e.scEncoder.EncodeAttributeXsiType(value, prefix)
	}
}

func (e *EXIBodyEncoderInOrderSC) EncodeCharacters(chars Value) error {
	if e.scEncoder == nil {
		return e.EXIBodyEncoderInOrder.EncodeCharacters(chars)
	} else {
		return e.scEncoder.EncodeCharacters(chars)
	}
}

func (e *EXIBodyEncoderInOrderSC) EncodeDocType(name, publicID, systemID, text string) error {
	if e.scEncoder == nil {
		return e.EXIBodyEncoderInOrder.EncodeDocType(name, publicID, systemID, text)
	} else {
		return e.scEncoder.EncodeDocType(name, publicID, systemID, text)
	}
}

func (e *EXIBodyEncoderInOrderSC) EncodeEntityReference(name string) error {
	if e.scEncoder == nil {
		return e.EXIBodyEncoderInOrder.EncodeEntityReference(name)
	} else {
		return e.scEncoder.EncodeEntityReference(name)
	}
}

func (e *EXIBodyEncoderInOrderSC) EncodeComment(runes []rune, start, length int) error {
	if e.scEncoder == nil {
		return e.EXIBodyEncoderInOrder.EncodeComment(runes, start, length)
	} else {
		return e.scEncoder.EncodeComment(runes, start, length)
	}
}

func (e *EXIBodyEncoderInOrderSC) EncodeProcessingInstruction(target, data string) error {
	if e.scEncoder == nil {
		return e.EXIBodyEncoderInOrder.EncodeProcessingInstruction(target, data)
	} else {
		return e.scEncoder.EncodeProcessingInstruction(target, data)
	}
}
