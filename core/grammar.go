package core

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"

	"github.com/sderkacs/exi-go/utils"
)

type GrammarType int

const (
	/* Root grammars */
	GrammarTypeDocument GrammarType = iota
	GrammarTypeFragment
	GrammarTypeDocEnd
	/* Schema-informed Document and Fragment Grammars */
	GrammarTypeSchemaInformedDocContent
	GrammarTypeSchemaInformedFragmentContent
	/* Schema-informed Element and Type Grammars */
	GrammarTypeSchemaInformedFirstStartTagContent
	GrammarTypeSchemaInformedStartTagContent
	GrammarTypeSchemaInformedElementContent
	/* Built-in Document and Fragment Grammars */
	GrammarTypeBuiltInDocConent
	GrammarTypeBuiltInFragmentContent
	/* Built-in Element Grammars */
	GrammarTypeBuiltInStartTagContent
	GrammarTypeBuiltInElementContent
)

var (
	endRule SchemaInformedGrammar = &SchemaInformedElement{
		AbstractSchemaInformedContent: &AbstractSchemaInformedContent{
			AbstractSchemaInformedGrammar: NewAbstractSchemaInformedGrammarWithLabel(utils.AsPtr("<END>")),
		},
	}
	startElementGeneric Event = NewStartElementGeneric()
	endElement          Event = NewEndElement()
)

type Grammar interface {
	IsSchemaInformed() bool
	HasEndElement() bool
	GetGrammarType() GrammarType
	GetNumberOfEvents() int
	AddProduction(event Event, grammar Grammar) error
	LearnStartElement(se *StartElement)
	LearnEndElement()
	LearnAttribute(at *Attribute) error
	LearnCharacters()
	StopLearning()
	LearningStopped() int
	GetElementContentGrammar() Grammar
	GetProduction(eventType EventType) Production
	GetStartElementProduction(namespaceUri, localName string) Production
	GetStartElementNSProduction(namespaceUri string) Production
	GetAttributeProduction(namespaceUri, localName string) Production
	GetAttributeNSProduction(namespaceUri string) Production
	GetProductionByEventCode(eventCode int) Production
}

type SchemaInformedGrammar interface {
	Grammar
	AddTerminalProduction(event Event)
	GetNumberOfDeclaredAttributes() int
	GetLeastAttributeEventCode() int

	// Label
	SetLabel(label string)
	GetLabel() string

	// clones schema-informed rule
	Duplicate() SchemaInformedGrammar
}

type BuiltInGrammar interface{}

type SchemaInformedStartTagGrammar interface {
	SchemaInformedGrammar
	SetElementContentGrammar(elementContent2 Grammar)
}

type SchemaInformedFirstStartTagGrammar interface {
	SchemaInformedStartTagGrammar
	SetTypeCastable(hasNamedSubtypes bool)
	IsTypeCastable() bool
	SetNillable(nillable bool)
	IsNillable() bool
	SetTypeEmpty(typeEmpty SchemaInformedFirstStartTagGrammar)
	GetTypeEmpty() (SchemaInformedFirstStartTagGrammar, error)
}

/*
	AbstractGrammar implementation
*/

type AbstractGrammar struct {
	Grammar
	label                     *string
	stopLearningContainerSize int
}

func NewAbstractGrammar() *AbstractGrammar {
	return &AbstractGrammar{
		label:                     nil,
		stopLearningContainerSize: NotFound,
	}
}

func NewAbstractGrammarWithLabel(label *string) *AbstractGrammar {
	return &AbstractGrammar{
		label:                     label,
		stopLearningContainerSize: NotFound,
	}
}

func (g *AbstractGrammar) AddTerminalProduction(event Event) {
	if !(event.IsEventType(EventTypeEndElement) || event.IsEventType(EventTypeEndDocument)) {
		panic("not a terminal production") //TODO: Return error?
	}
	g.AddProduction(event, endRule)
}

func (g *AbstractGrammar) LearnStartElement(se *StartElement) {}

func (g *AbstractGrammar) LearnEndElement() {}

func (g *AbstractGrammar) LearnAttribute(at *Attribute) error {
	return nil
}

func (g *AbstractGrammar) LearnCharacters() {}

func (g *AbstractGrammar) StopLearning() {}

func (g *AbstractGrammar) LearningStopped() int {
	return g.stopLearningContainerSize
}

func (g *AbstractGrammar) SetLabel(label string) {
	g.label = &label
}

func (g *AbstractGrammar) GetLabel() string {
	if g.label != nil && *g.label != "" {
		return *g.label
	} else {
		return "AbstractGrammar" //TODO: Add some hash?
	}
}

func (g *AbstractGrammar) GetEventCode(eventType EventType, events []EventType) int {
	return slices.Index(events, eventType)
}

func (g *AbstractGrammar) GetElementContentGrammar() Grammar {
	return g
}

func (g *AbstractGrammar) checkQualifiedName(c QName, namespaceUri, localName string) bool {
	return c.Local == localName && c.Space == namespaceUri
}

/*
	AbstractSchemaInformedGrammar implementation
*/

type AbstractSchemaInformedGrammar struct {
	*AbstractGrammar
	containers                 []Production
	codeLengthA                int
	codeLengthB                int
	hasEndElement              bool
	leastAttributeEventCode    int
	numberOfDeclaredAttributes int
}

func NewAbstractSchemaInformedGrammar() *AbstractSchemaInformedGrammar {
	return NewAbstractSchemaInformedGrammarWithLabel(nil)
}

func NewAbstractSchemaInformedGrammarWithLabel(label *string) *AbstractSchemaInformedGrammar {
	return &AbstractSchemaInformedGrammar{
		AbstractGrammar:            NewAbstractGrammarWithLabel(label),
		containers:                 []Production{},
		codeLengthA:                0,
		codeLengthB:                0,
		hasEndElement:              false,
		leastAttributeEventCode:    NotFound,
		numberOfDeclaredAttributes: 0,
	}
}

func (g *AbstractSchemaInformedGrammar) HasEndElement() bool {
	return g.hasEndElement
}

func (g *AbstractSchemaInformedGrammar) isTerminalRule() bool {
	return false //TODO: ???
}

func (g *AbstractSchemaInformedGrammar) IsSchemaInformed() bool {
	return true
}

func (g *AbstractSchemaInformedGrammar) GetNumberOfDeclaredAttributes() int {
	return g.numberOfDeclaredAttributes
}

func (g *AbstractSchemaInformedGrammar) GetLeastAttributeEventCode() int {
	return g.leastAttributeEventCode
}

func (g *AbstractSchemaInformedGrammar) GetNumberOfEvents() int {
	return len(g.containers)
}

func (g *AbstractSchemaInformedGrammar) AddProduction(event Event, grammar Grammar) error {
	if g.isTerminalRule() {
		return fmt.Errorf("EndGrammar can not have events attached")
	}

	if (event.IsEventType(EventTypeEndElement) ||
		event.IsEventType(EventTypeAttributeGeneric) ||
		event.IsEventType(EventTypeStartElementGeneric)) && g.GetProduction(event.GetEventType()) != nil {
		log.Printf("Event %d is already preset", event.GetEventType())
	} else {
		if event.IsEventType(EventTypeEndElement) {
			g.hasEndElement = true
		}

		for _, prod := range g.containers {
			if prod.GetEvent().Equals(event) {
				if prod.GetNextGrammar() != grammar {
					return fmt.Errorf("same event %d with indistinguishable 'next' grammar", event.GetEventType())
				}
			}
		}
	}

	return g.updateSortedEvents(event, grammar)
}

func (g *AbstractSchemaInformedGrammar) updateSortedEvents(newEvent Event, newGrammar Grammar) error {
	sortedEvents := []Event{}

	// see http://www.w3.org/TR/exi/#eventCodeAssignment
	o2 := newEvent
	added := false

	for _, prod := range g.containers {
		o1 := prod.GetEvent()

		if !added {
			diff := o1.GetEventType() - o2.GetEventType()
			if diff == 0 {
				switch o1.GetEventType() {
				case EventTypeAttribute:
					cmpA := AttributeCompareFunc(o1.(*Attribute), o2.(*Attribute))
					if cmpA < 0 {
						// comes after
					} else if cmpA > 0 {
						// comes now
						sortedEvents = append(sortedEvents, o2)
						added = true
					} else {
						return fmt.Errorf("twice the same attribute name when sorting")
					}
				case EventTypeAttributeNS:
					atNS1 := o1.(*AttributeNS)
					atNS2 := o2.(*AttributeNS)

					cmpNS := strings.Compare(atNS1.GetNamespaceUri(), atNS2.GetNamespaceUri())
					if cmpNS < 0 {
						// comes after
					} else if cmpNS > 0 {
						sortedEvents = append(sortedEvents, o2)
						added = true
					} else {
						return fmt.Errorf("twice the same attribute uri in AT(*uri) when sorting")
					}
				case EventTypeStartElement, EventTypeStartElementNS:
					// sorted in schema order --> new event comes after
				default:
					return fmt.Errorf("no valid event type for sorting")
				}
			} else if diff < 0 {
				// new event type comes after
			} else {
				sortedEvents = append(sortedEvents, o2)
				added = true
			}
		}

		// add old event
		sortedEvents = append(sortedEvents, o1)
	}

	if !added {
		sortedEvents = append(sortedEvents, o2)
	}

	newContainers := make([]Production, len(sortedEvents))
	eventCode := 0
	newOneAdded := false

	for _, ev := range sortedEvents {
		if ev == newEvent {
			newContainers[eventCode] = NewSchemaInformedProduction(newGrammar, newEvent, eventCode)
			newOneAdded = true
		} else {
			i := eventCode
			if newOneAdded {
				i = eventCode - 1
			}
			oldEI := g.containers[i]
			newContainers[eventCode] = NewSchemaInformedProduction(oldEI.GetNextGrammar(), oldEI.GetEvent(), eventCode)
		}
		eventCode++
	}

	// re-set *old* array
	g.containers = newContainers

	// calculate ahead of time two different first level code lengths
	g.codeLengthA = utils.GetCodingLength(g.GetNumberOfEvents())
	g.codeLengthB = utils.GetCodingLength(g.GetNumberOfEvents() + 1)

	g.leastAttributeEventCode = NotFound
	g.numberOfDeclaredAttributes = 0

	for idx, er := range g.containers {
		if er.GetEvent().IsEventType(EventTypeAttribute) {
			if g.leastAttributeEventCode == NotFound {
				g.leastAttributeEventCode = idx
			}
			g.numberOfDeclaredAttributes++
		}
	}

	return nil
}

func (g *AbstractSchemaInformedGrammar) JoinGrammars(rule Grammar) {
	for i := 0; i < rule.GetNumberOfEvents(); i++ {
		ei := rule.GetProduction(EventType(i))
		g.AddProduction(ei.GetEvent(), ei.GetNextGrammar())
	}
}

func (g *AbstractSchemaInformedGrammar) Duplicate() SchemaInformedGrammar {
	o := *g
	return &o
}

func (g *AbstractSchemaInformedGrammar) GetProduction(eventType EventType) Production {
	for _, ei := range g.containers {
		if ei.GetEvent().IsEventType(eventType) {
			return ei
		}
	}

	return nil
}

func (g *AbstractSchemaInformedGrammar) GetStartElementProduction(namespaceUri, localName string) Production {
	for _, ei := range g.containers {
		if ei.GetEvent().IsEventType(EventTypeStartElement) {
			seEI := ei.GetEvent().(StartElement)
			if g.checkQualifiedName(seEI.qname, namespaceUri, localName) {
				return ei
			}
		}
	}

	return nil
}

func (g *AbstractSchemaInformedGrammar) GetStartElementNSProduction(namespaceUri string) Production {
	for _, ei := range g.containers {
		if ei.GetEvent().IsEventType(EventTypeStartElementNS) {
			seEI := ei.GetEvent().(StartElementNS)
			if seEI.GetNamespaceUri() == namespaceUri {
				return ei
			}
		}
	}

	return nil
}

func (g *AbstractSchemaInformedGrammar) GetAttributeProduction(namespaceUri, localName string) Production {
	for _, ei := range g.containers {
		if ei.GetEvent().IsEventType(EventTypeAttribute) {
			atEI := ei.GetEvent().(*Attribute)
			if g.checkQualifiedName(atEI.qname, namespaceUri, localName) {
				return ei
			}
		}
	}

	return nil
}

func (g *AbstractSchemaInformedGrammar) GetAttributeNSProduction(namespaceUri string) Production {
	for _, ei := range g.containers {
		if ei.GetEvent().IsEventType(EventTypeAttributeNS) {
			atEI := ei.GetEvent().(AttributeNS)
			if atEI.GetNamespaceUri() == namespaceUri {
				return ei
			}
		}
	}

	return nil
}

func (g *AbstractSchemaInformedGrammar) GetProductionByEventCode(eventCode int) Production {
	return g.containers[eventCode]
}

/*
	AbstractSchemaInformedContent implementation
*/

type AbstractSchemaInformedContent struct {
	*AbstractSchemaInformedGrammar
}

func NewAbstractSchemaInformedContent() *AbstractSchemaInformedContent {
	return &AbstractSchemaInformedContent{
		AbstractSchemaInformedGrammar: NewAbstractSchemaInformedGrammar(),
	}
}

/*
	SchemaInformedElement implementation
*/

type SchemaInformedElement struct {
	*AbstractSchemaInformedContent
}

func NewSchemaInformedElement() *SchemaInformedElement {
	asic := NewAbstractSchemaInformedContent()
	se := &SchemaInformedElement{
		AbstractSchemaInformedContent: asic,
	}
	asic.Grammar = se

	return se
}

func (e *SchemaInformedElement) GetGrammarType() GrammarType {
	return GrammarTypeSchemaInformedElementContent
}

/*
	AbstractBuiltInGrammar implementation
*/

type AbstractBuiltInGrammar struct {
	BuiltInGrammar
	*AbstractGrammar
	containers []Production
	ec1Length  int
}

func NewBuiltInGrammar() *AbstractBuiltInGrammar {
	return &AbstractBuiltInGrammar{
		AbstractGrammar: NewAbstractGrammar(),
		containers:      []Production{},
		ec1Length:       0,
	}
}

func (g *AbstractBuiltInGrammar) HasEndElement() bool {
	return false
}

func (g *AbstractBuiltInGrammar) StopLearning() {
	if g.stopLearningContainerSize == NotFound {
		g.stopLearningContainerSize = len(g.containers)
	}
}

func (g *AbstractBuiltInGrammar) IsSchemaInformed() bool {
	return false
}

func (g *AbstractBuiltInGrammar) GetTypeEmpty() Grammar {
	return g
}

func (g *AbstractBuiltInGrammar) GetNumberOfEvents() int {
	return len(g.containers)
}

func (g *AbstractBuiltInGrammar) AddProduction(event Event, grammar Grammar) error {
	g.containers = append(g.containers, NewSchemaLessProduction(g, grammar, event, g.GetNumberOfEvents()))
	g.ec1Length = utils.GetCodingLength(len(g.containers) + 1)
	return nil
}

func (g *AbstractBuiltInGrammar) Contains(event Event) bool {
	idx := slices.IndexFunc(g.containers, func(prod Production) bool {
		return prod.GetEvent().Equals(event)
	})

	return idx != NotFound
}

func (g *AbstractBuiltInGrammar) GetProduction(eventType EventType) Production {
	for _, ei := range g.containers {
		if ei.GetEvent().IsEventType(eventType) {
			if !g.isExiProfileGhostNode(ei) {
				return ei
			}
		}
	}

	return nil
}

func (g *AbstractBuiltInGrammar) isExiProfileGhostNode(ei Production) bool {
	if g.stopLearningContainerSize == NotFound {
		return false
	} else {
		return ei.GetEventCode() < (g.GetNumberOfEvents() - g.stopLearningContainerSize)
	}
}

func (g *AbstractBuiltInGrammar) GetStartElementProduction(namespaceUri, localName string) Production {
	for _, ei := range g.containers {
		if ei.GetEvent().IsEventType(EventTypeStartElement) {
			seEI := ei.GetEvent().(StartElement)
			if g.checkQualifiedName(seEI.GetQName(), namespaceUri, localName) {
				if !g.isExiProfileGhostNode(ei) {
					return ei
				}
			}
		}
	}

	return nil
}

func (g *AbstractBuiltInGrammar) GetStartElementNSProduction(namespaceUri string) Production {
	return nil
}

func (g *AbstractBuiltInGrammar) GetAttributeProduction(namespaceUri, localName string) Production {
	for _, ei := range g.containers {
		if ei.GetEvent().IsEventType(EventTypeAttribute) {
			atEI := ei.GetEvent().(*Attribute)
			if g.checkQualifiedName(atEI.GetQName(), namespaceUri, localName) {
				if !g.isExiProfileGhostNode(ei) {
					return ei
				}
			}
		}
	}

	return nil
}

func (g *AbstractBuiltInGrammar) GetAttributeNSProduction(namespaceUri string) Production {
	return nil
}

func (g *AbstractBuiltInGrammar) GetProductionByEventCode(eventCode int) Production {
	return g.containers[g.GetNumberOfEvents()-1-eventCode]
}

/*
	AbstractBuiltInContent implementation
*/

// TODO: Possible key comparsion issues
var (
	optionsStartTag     map[*FidelityOptions][]EventType = map[*FidelityOptions][]EventType{}
	optionsChildContent map[*FidelityOptions][]EventType = map[*FidelityOptions][]EventType{}
)

func BuiltInContentGet2ndLevelEventsStartTagItems(fidelityOptions *FidelityOptions) []EventType {
	_, exists := optionsStartTag[fidelityOptions]
	if !exists {
		events := []EventType{EventTypeEndElementUndeclared, EventTypeAttributeGenericUndeclared}
		if fidelityOptions.IsFidelityEnabled(FeaturePrefix) {
			events = append(events, EventTypeNamespaceDeclaration)
		}
		if fidelityOptions.IsFidelityEnabled(FeatureSC) {
			events = append(events, EventTypeSelfContained)
		}

		optionsStartTag[fidelityOptions] = events
	}

	return optionsStartTag[fidelityOptions]
}

func BuiltInContentGet2ndLevelEventsChildContentItems(fidelityOptions *FidelityOptions) []EventType {
	_, exists := optionsStartTag[fidelityOptions]
	if !exists {
		events := []EventType{EventTypeStartElementGenericUndeclared, EventTypeCharactersGenericUndeclared}
		if fidelityOptions.IsFidelityEnabled(FeatureDTD) {
			events = append(events, EventTypeEntityReference)
		}

		optionsChildContent[fidelityOptions] = events
	}

	return optionsChildContent[fidelityOptions]
}

type AbstractBuiltInContent struct {
	*AbstractBuiltInGrammar
	learnedCH bool
}

func NewAbstractBuiltInContent() *AbstractBuiltInContent {
	return &AbstractBuiltInContent{
		AbstractBuiltInGrammar: NewBuiltInGrammar(),
		learnedCH:              false,
	}
}

func (c *AbstractBuiltInContent) LearnCharacters() {
	if !c.learnedCH {
		c.AddProduction(NewCharacters(BuiltInGetDefaultDatatype()), c.GetElementContentGrammar())
		c.learnedCH = true
	}
}

/*
	BuiltInDocContent implementation
*/

type BuiltInDocContent struct {
	*AbstractBuiltInGrammar
	docEnd Grammar
}

func NewBuiltInDocContent(docEnd Grammar) *BuiltInDocContent {
	return &BuiltInDocContent{
		AbstractBuiltInGrammar: NewBuiltInGrammar(),
		docEnd:                 docEnd,
	}
}

func NewBuiltInDocContentWithLabel(docEnd Grammar, label string) *BuiltInDocContent {
	c := &BuiltInDocContent{
		AbstractBuiltInGrammar: NewBuiltInGrammar(),
		docEnd:                 docEnd,
	}
	c.SetLabel(label)

	return c
}

func (c *BuiltInDocContent) GetGrammarType() GrammarType {
	return GrammarTypeBuiltInDocConent
}

func (c *BuiltInDocContent) AddProduction(event Event, grammar Grammar) error {
	if !event.IsEventType(EventTypeStartElementGeneric) || c.GetNumberOfEvents() > 0 {
		return fmt.Errorf("mis-use of BuiltInDocContent grammar")
	}
	c.AbstractBuiltInGrammar.AddProduction(event, grammar)

	return nil
}

/*
	BuiltInElement implementation
*/

type BuiltInElement struct {
	*AbstractBuiltInContent
}

func NewBuiltInElement() *BuiltInElement {
	e := &BuiltInElement{
		AbstractBuiltInContent: NewAbstractBuiltInContent(),
	}
	e.AddProduction(endElement, endRule)

	return e
}

func (e *BuiltInElement) HasEndElement() bool {
	return true
}

func (e *BuiltInElement) GetGrammarType() GrammarType {
	return GrammarTypeBuiltInElementContent
}

func (e *BuiltInElement) LearnStartElement(se *StartElement) {
	e.AddProduction(se, e)
}

func (e *BuiltInElement) LearnAttribute(at *Attribute) error {
	return fmt.Errorf("element content rule cannot learn AT events")
}

/*
	BuiltInFragmentContent implementation
*/

type BuiltInFragmentContent struct {
	*AbstractBuiltInGrammar
}

func NewBuiltInFragmentContent() *BuiltInFragmentContent {
	c := &BuiltInFragmentContent{
		AbstractBuiltInGrammar: NewBuiltInGrammar(),
	}
	c.AddTerminalProduction(NewEndDocument())
	c.AddProduction(startElementGeneric, c)

	return c
}

func (c *BuiltInFragmentContent) GetGrammarType() GrammarType {
	return GrammarTypeBuiltInFragmentContent
}

func (c *BuiltInFragmentContent) LearnStartElement(se *StartElement) {
	if !c.Contains(se) {
		c.AddProduction(se, c)
	}
}

/*
	BuiltInStartTag implementation
*/

type BuiltInStartTag struct {
	*AbstractBuiltInContent
	elementContent *BuiltInElement
	learnedEE      bool
	learnedXsiType bool
}

func NewBuiltInStartTag() *BuiltInStartTag {
	return &BuiltInStartTag{
		AbstractBuiltInContent: NewAbstractBuiltInContent(),
		elementContent:         NewBuiltInElement(),
	}
}

func (t *BuiltInStartTag) HasEndElement() bool {
	return t.learnedEE
}

func (t *BuiltInStartTag) GetGrammarType() GrammarType {
	return GrammarTypeBuiltInStartTagContent
}

func (t *BuiltInStartTag) GetElementContentGrammar() Grammar {
	return t.elementContent
}

func (t *BuiltInStartTag) LearnStartElement(se *StartElement) {
	t.AddProduction(se, t.GetElementContentGrammar())
}

func (t *BuiltInStartTag) LearnEndElement() {
	if !t.learnedEE {
		t.AddTerminalProduction(endElement)
		t.learnedEE = true
	}
}

func (t *BuiltInStartTag) LearnAttribute(at *Attribute) error {
	qnc := at.GetQNameContext()
	if qnc.GetNamespaceUriID() == 2 && qnc.GetLocalNameID() == 1 {
		if !t.learnedXsiType {
			t.AddProduction(at, t)
			t.learnedXsiType = true
		}
	} else {
		t.AddProduction(at, t)
	}

	return nil
}

/*
	DocEnd implementation
*/

type DocEnd struct {
	*AbstractSchemaInformedGrammar
}

func NewDocEnd() *DocEnd {
	return &DocEnd{
		AbstractSchemaInformedGrammar: NewAbstractSchemaInformedGrammar(),
	}
}

func NewDocEndWithLabel(label string) *DocEnd {
	return &DocEnd{
		AbstractSchemaInformedGrammar: NewAbstractSchemaInformedGrammarWithLabel(&label),
	}
}

func (e *DocEnd) GetGrammarType() GrammarType {
	return GrammarTypeDocEnd
}

/*
	Document implementation
*/

type Document struct {
	*AbstractSchemaInformedGrammar
}

func NewDocument() *Document {
	return &Document{
		AbstractSchemaInformedGrammar: NewAbstractSchemaInformedGrammar(),
	}
}

func NewDocumentWithLabel(label string) *Document {
	return &Document{
		AbstractSchemaInformedGrammar: NewAbstractSchemaInformedGrammarWithLabel(&label),
	}
}

func (d *Document) GetGrammarType() GrammarType {
	return GrammarTypeDocument
}

/*
	Fragment implementation
*/

type Fragment struct {
	*AbstractSchemaInformedGrammar
}

func NewFragment() *Fragment {
	return &Fragment{
		AbstractSchemaInformedGrammar: NewAbstractSchemaInformedGrammar(),
	}
}

func NewFragmentWithLabel(label string) *Fragment {
	return &Fragment{
		AbstractSchemaInformedGrammar: NewAbstractSchemaInformedGrammarWithLabel(&label),
	}
}

func (f *Fragment) GetGrammarType() GrammarType {
	return GrammarTypeFragment
}

/*
	SchemaInformedDocContent implementation
*/

type SchemaInformedDocContent struct {
	*AbstractSchemaInformedGrammar
}

func NewSchemaInformedDocContent() *SchemaInformedDocContent {
	return &SchemaInformedDocContent{
		AbstractSchemaInformedGrammar: NewAbstractSchemaInformedGrammar(),
	}
}

func NewSchemaInformedDocContentWithLabel(label string) *SchemaInformedDocContent {
	return &SchemaInformedDocContent{
		AbstractSchemaInformedGrammar: NewAbstractSchemaInformedGrammarWithLabel(&label),
	}
}

func (c *SchemaInformedDocContent) GetGrammarType() GrammarType {
	return GrammarTypeSchemaInformedDocContent
}

/*
	SchemaInformedFirstStartTag implementation
*/

var (
	sistElementContent2Empty SchemaInformedGrammar = NewSchemaInformedElement()
	sistInit                 sync.Once
)

// Must be initialized by NewSchemaInformedStartTag*() functions only!
type SchemaInformedStartTag struct {
	*AbstractSchemaInformedContent
	elementContent2 Grammar
	sifst           SchemaInformedStartTagGrammar
}

func NewSchemaInformedStartTag() *SchemaInformedStartTag {
	sistInit.Do(func() {
		sistElementContent2Empty.AddTerminalProduction(endElement)
	})
	st := &SchemaInformedStartTag{
		AbstractSchemaInformedContent: NewAbstractSchemaInformedContent(),
		elementContent2:               nil,
		sifst:                         nil,
	}

	return st
}

func NewSchemaInformedStartTagWithEC2(elementContent2 SchemaInformedGrammar) *SchemaInformedStartTag {
	st := NewSchemaInformedStartTag()
	st.elementContent2 = elementContent2

	return st
}

func (t *SchemaInformedStartTag) GetGrammarType() GrammarType {
	return GrammarTypeSchemaInformedStartTagContent
}

func (t *SchemaInformedStartTag) SetElementContentGrammar(elementContent2 Grammar) {
	t.elementContent2 = elementContent2
}

func (t *SchemaInformedStartTag) Clone() *SchemaInformedStartTag {
	clone := *t

	// remove self-references
	for idx, ei := range clone.containers {
		if ei.GetNextGrammar() == t {
			clone.containers[idx] = NewSchemaInformedProduction(&clone, ei.GetEvent(), idx)
		}
	}

	return &clone
}

func (t *SchemaInformedStartTag) GetTypeEmptyInterval() (SchemaInformedStartTagGrammar, error) {
	if t.sifst == nil {
		switch t.GetGrammarType() {
		case GrammarTypeSchemaInformedFirstStartTagContent:
			t.sifst = NewSchemaInformedFirstStartTag()
		case GrammarTypeSchemaInformedStartTagContent:
			t.sifst = NewSchemaInformedStartTag()
		default:
			return nil, fmt.Errorf("unexpected grammar type %d for typeEmpty", t.GetGrammarType())
		}
		t.sifst.SetElementContentGrammar(sistElementContent2Empty)

		for i := 0; i < t.GetNumberOfEvents(); i++ {
			prod := t.GetProduction(EventType(i))
			ev := prod.GetEvent()
			ng := prod.GetNextGrammar()

			switch ev.GetEventType() {
			case EventTypeAttribute, EventTypeAttributeNS, EventTypeAttributeGeneric:
				if ng == t {
					t.sifst.AddProduction(ev, t.sifst)
				} else if ng.GetGrammarType() == GrammarTypeSchemaInformedFirstStartTagContent {
					ng2 := ng.(*SchemaInformedFirstStartTag)
					g, err := ng2.GetTypeEmptyInterval()
					if err != nil {
						return nil, err
					}
					t.sifst.AddProduction(ev, g)
				} else if ng.GetGrammarType() == GrammarTypeSchemaInformedStartTagContent {
					ng2 := ng.(*SchemaInformedStartTag)
					g, err := ng2.GetTypeEmptyInterval()
					if err != nil {
						return nil, err
					}
					t.sifst.AddProduction(ev, g)
				} else {
					return nil, fmt.Errorf("unexpected grammar type %d for typeEmpty", ng.GetGrammarType())
				}
			default:
				if !t.sifst.HasEndElement() {
					t.sifst.AddTerminalProduction(endElement)
				}
			}
		}
	}

	return t.sifst, nil
}

/*
	SchemaInformedFirstStartTag implementation
*/

const (
	sifstUseRuntimeEmptyType bool = true
)

type SchemaInformedFirstStartTag struct {
	*SchemaInformedStartTag
	isTypeCastable bool
	isNillable     bool
	typeEmpty      SchemaInformedFirstStartTagGrammar
	typeName       *QName
}

func NewSchemaInformedFirstStartTag() *SchemaInformedFirstStartTag {
	return &SchemaInformedFirstStartTag{
		SchemaInformedStartTag: NewSchemaInformedStartTag(),
		isTypeCastable:         false,
		isNillable:             false,
		typeEmpty:              nil,
		typeName:               nil,
	}
}

func NewSchemaInformedFirstStartTagWithEC2(elementContent2 SchemaInformedGrammar) *SchemaInformedFirstStartTag {
	return &SchemaInformedFirstStartTag{
		SchemaInformedStartTag: NewSchemaInformedStartTagWithEC2(elementContent2),
		isTypeCastable:         false,
		isNillable:             false,
		typeEmpty:              nil,
		typeName:               nil,
	}
}

func NewSchemaInformedFirstStartTagWithStartTag(startTag SchemaInformedFirstStartTagGrammar) *SchemaInformedFirstStartTag {
	sig := startTag.GetElementContentGrammar().(SchemaInformedGrammar)
	st := NewSchemaInformedFirstStartTagWithEC2(sig)

	// clone top level
	for i := 0; i < startTag.GetNumberOfEvents(); i++ {
		ei := startTag.GetProduction(EventType(i))
		// remove self-reference
		next := ei.GetNextGrammar()
		if next == startTag {
			next = st
		}
		st.AddProduction(ei.GetEvent(), next)
	}

	return st
}

func (t *SchemaInformedFirstStartTag) GetGrammarType() GrammarType {
	return GrammarTypeSchemaInformedFirstStartTagContent
}

func (t *SchemaInformedFirstStartTag) SetTypeCastable(isTypeCastable bool) {
	t.isTypeCastable = isTypeCastable
}

func (t *SchemaInformedFirstStartTag) IsTypeCastable() bool {
	return t.isTypeCastable
}

func (t *SchemaInformedFirstStartTag) SetNillable(isNillable bool) {
	t.isNillable = isNillable
}

func (t *SchemaInformedFirstStartTag) IsNillable() bool {
	return t.isNillable
}

func (t *SchemaInformedFirstStartTag) SetTypeEmpty(typeEmpty SchemaInformedFirstStartTagGrammar) {
	t.typeEmpty = typeEmpty
}

func (t *SchemaInformedFirstStartTag) GetTypeEmpty() (SchemaInformedFirstStartTagGrammar, error) {
	if sifstUseRuntimeEmptyType {
		t, err := t.GetTypeEmptyInterval()
		if err != nil {
			return nil, err
		}
		return t.(SchemaInformedFirstStartTagGrammar), nil
	} else {
		return t.typeEmpty, nil
	}
}

/*
	SchemaInformedFragmentContent implementation
*/

type SchemaInformedFragmentContent struct {
	*AbstractSchemaInformedGrammar
}

func NewSchemaInformedFragmentContent() *SchemaInformedFragmentContent {
	return &SchemaInformedFragmentContent{
		AbstractSchemaInformedGrammar: NewAbstractSchemaInformedGrammar(),
	}
}

func NewSchemaInformedFragmentContentWithLabel(label string) *SchemaInformedFragmentContent {
	return &SchemaInformedFragmentContent{
		AbstractSchemaInformedGrammar: NewAbstractSchemaInformedGrammarWithLabel(&label),
	}
}

func (c *SchemaInformedFragmentContent) GetGrammarType() GrammarType {
	return GrammarTypeSchemaInformedFragmentContent
}
