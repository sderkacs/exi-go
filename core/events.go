package core

import "github.com/sderkacs/go-exi/utils"

type EventType int

const (
	// Start document
	EventTypeStartDocument EventType = iota
	// Special XSI attributes
	EventTypeAttributeXsiType
	EventTypeAttributeXsiNil
	// Attributes
	EventTypeAttribute
	EventTypeAttributeNS
	EventTypeAttributeGeneric
	EventTypeAttributeInvalidValue
	EventTypeAttributeAnyInvalidValue
	EventTypeAttributeGenericUndeclared
	// StartElements
	EventTypeStartElement
	EventTypeStartElementNS
	EventTypeStartElementGeneric
	EventTypeStartElementGenericUndeclared
	// EndElements
	EventTypeEndElement
	EventTypeEndElementUndeclared
	// Characters
	EventTypeCharacters
	EventTypeCharactersGeneric
	EventTypeCharactersGenericUndeclared
	// EndDocument
	EventTypeEndDocument
	// Features & fidelity options
	EventTypeDocType
	EventTypeNamespaceDeclaration
	EventTypeSelfContained
	EventTypeEntityReference
	EventTypeComment
	EventTypeProcessingInstruction
)

type Event interface {
	GetEventType() EventType
	IsEventType(eventType EventType) bool
	Equals(other Event) bool
}

type DatatypeEvent interface {
	Event
	GetDatatype() Datatype
}

type AbstractEvent struct {
	Event
	eventType EventType
}

func (e *AbstractEvent) GetEventType() EventType {
	return e.eventType
}

func (e *AbstractEvent) IsEventType(eventType EventType) bool {
	return e.eventType == eventType
}

func (e *AbstractEvent) Equals(other Event) bool {
	if other == nil {
		return false
	}
	return e.eventType == other.GetEventType()
}

type AbstractDatatypeEvent struct {
	*AbstractEvent
	datatype Datatype
}

func (e *AbstractDatatypeEvent) GetDatatype() Datatype {
	return e.datatype
}

// Back-compat with earlier naming used in ported code
func (e *AbstractDatatypeEvent) GetDataType() Datatype {
	return e.datatype
}

type Attribute struct {
	*AbstractDatatypeEvent
	qname        utils.QName
	qnameContext *QNameContext
}

func NewAttribute(qnc *QNameContext) *Attribute {
	return NewAttributeWithDatatype(qnc, BuiltInGetDefaultDatatype())
}

func NewAttributeWithDatatype(qnc *QNameContext, datatype Datatype) *Attribute {
	as := &AbstractEvent{
		eventType: EventTypeAttribute,
	}
	a := &Attribute{
		AbstractDatatypeEvent: &AbstractDatatypeEvent{
			AbstractEvent: as,
			datatype:      datatype,
		},
		qnameContext: qnc,
		qname:        qnc.GetQName(),
	}
	as.Event = a
	return a
}

func (a *Attribute) GetQNameContext() *QNameContext {
	return a.qnameContext
}

func (a *Attribute) GetQName() utils.QName {
	return a.qname
}

func (a *Attribute) Equals(other Event) bool {
	otherA, ok := other.(*Attribute)
	if ok {
		if a.qname.Local == otherA.qname.Local && a.qname.Space == otherA.qname.Space {
			return true
		}
	}
	return false
}

type AttributeGeneric struct {
	*AbstractEvent
}

func NewAttributeGeneric() *AttributeGeneric {
	as := &AbstractEvent{
		eventType: EventTypeAttributeGeneric,
	}
	a := &AttributeGeneric{
		AbstractEvent: as,
	}
	as.Event = a

	return a
}

type AttributeNS struct {
	*AbstractEvent
	namespaceUri   string
	namespaceUriID int
}

func NewAttributeNS(namespaceUriID int, namespaceUri string) *AttributeNS {
	as := &AbstractEvent{
		eventType: EventTypeAttributeNS,
	}
	a := &AttributeNS{
		AbstractEvent:  as,
		namespaceUriID: namespaceUriID,
		namespaceUri:   namespaceUri,
	}
	as.Event = a
	return a
}

func (a *AttributeNS) GetNamespaceUri() string {
	return a.namespaceUri
}

func (a *AttributeNS) GetNamespaceUriID() int {
	return a.namespaceUriID
}

type Characters struct {
	*AbstractDatatypeEvent
}

func NewCharacters(datatype Datatype) *Characters {
	as := &AbstractEvent{
		eventType: EventTypeCharacters,
	}
	a := &Characters{
		AbstractDatatypeEvent: &AbstractDatatypeEvent{
			AbstractEvent: as,
			datatype:      datatype,
		},
	}
	as.Event = a

	return a
}

type CharactersGeneric struct {
	*AbstractDatatypeEvent
}

func NewCharactersGeneric() *CharactersGeneric {
	as := &AbstractEvent{
		eventType: EventTypeCharactersGeneric,
	}
	a := &CharactersGeneric{
		AbstractDatatypeEvent: &AbstractDatatypeEvent{
			AbstractEvent: as,
			datatype:      BuiltInGetDefaultDatatype(),
		},
	}
	as.Event = a
	return a
}

type Comment struct {
	*AbstractEvent
}

func NewComment() *Comment {
	as := &AbstractEvent{
		eventType: EventTypeComment,
	}
	a := &Comment{
		AbstractEvent: as,
	}
	as.Event = a
	return a
}

type DocType struct {
	*AbstractEvent
}

func NewDocType() *DocType {
	as := &AbstractEvent{
		eventType: EventTypeDocType,
	}
	a := &DocType{
		AbstractEvent: as,
	}
	as.Event = a
	return a
}

type EndDocument struct {
	*AbstractEvent
}

func NewEndDocument() *EndDocument {
	as := &AbstractEvent{
		eventType: EventTypeEndDocument,
	}
	a := &EndDocument{
		AbstractEvent: as,
	}
	as.Event = a
	return a
}

type EndElement struct {
	*AbstractEvent
}

func NewEndElement() *EndElement {
	as := &AbstractEvent{
		eventType: EventTypeEndElement,
	}
	a := &EndElement{
		AbstractEvent: as,
	}
	as.Event = a
	return a
}

type EntityReference struct {
	*AbstractEvent
}

func NewEntityReference() *EntityReference {
	as := &AbstractEvent{
		eventType: EventTypeEntityReference,
	}
	a := &EntityReference{
		AbstractEvent: as,
	}
	as.Event = a
	return a
}

type NamespaceDeclaration struct {
	*AbstractEvent
}

func NewNamespaceDeclaration() *NamespaceDeclaration {
	as := &AbstractEvent{
		eventType: EventTypeNamespaceDeclaration,
	}
	a := &NamespaceDeclaration{
		AbstractEvent: as,
	}
	as.Event = a
	return a
}

type ProcessingInstruction struct {
	*AbstractEvent
}

func NewProcessingInstruction() *ProcessingInstruction {
	as := &AbstractEvent{
		eventType: EventTypeProcessingInstruction,
	}
	a := &ProcessingInstruction{
		AbstractEvent: as,
	}
	as.Event = a
	return a
}

type SelfContained struct {
	*AbstractEvent
}

func NewSelfContained() *SelfContained {
	as := &AbstractEvent{
		eventType: EventTypeSelfContained,
	}
	a := &SelfContained{
		AbstractEvent: as,
	}
	as.Event = a
	return a
}

type StartDocument struct {
	*AbstractEvent
}

func NewStartDocument() *StartDocument {
	as := &AbstractEvent{
		eventType: EventTypeStartDocument,
	}
	a := &StartDocument{
		AbstractEvent: as,
	}
	as.Event = a
	return a
}

type StartElement struct {
	*AbstractEvent
	qname        utils.QName
	qnameContext *QNameContext
	grammar      Grammar
}

func NewStartElement(qnc *QNameContext) *StartElement {
	as := &AbstractEvent{
		eventType: EventTypeStartElement,
	}
	a := &StartElement{
		AbstractEvent: as,
		qnameContext:  qnc,
		qname:         qnc.qName,
	}
	as.Event = a
	return a
}

func NewStartElementWithGrammar(qnc *QNameContext, grammar Grammar) *StartElement {
	se := NewStartElement(qnc)
	se.grammar = grammar
	return se
}

func (e *StartElement) GetQNameContext() *QNameContext {
	return e.qnameContext
}

func (e *StartElement) GetQName() utils.QName {
	return e.qname
}

func (e *StartElement) SetGrammar(grammar Grammar) {
	e.grammar = grammar
}

func (e *StartElement) GetGrammar() Grammar {
	return e.grammar
}

func (e *StartElement) Equals(other Event) bool {
	if other == nil {
		return false
	}
	otherA, ok := other.(*StartElement)
	if ok {
		if e.qnameContext.Equals(otherA.qnameContext) {
			return true
		}
	}
	return false
}

type StartElementGeneric struct {
	*AbstractEvent
}

func NewStartElementGeneric() *StartElementGeneric {
	as := &AbstractEvent{
		eventType: EventTypeStartElementGeneric,
	}
	a := &StartElementGeneric{
		AbstractEvent: as,
	}
	as.Event = a
	return a
}

type StartElementNS struct {
	*AbstractEvent
	namespaceUri   string
	namespaceUriID int
}

func NewStartElementNS(namespaceUriID int, namespaceUri string) *StartElementNS {
	as := &AbstractEvent{
		eventType: EventTypeStartElementNS,
	}
	a := &StartElementNS{
		AbstractEvent:  as,
		namespaceUriID: namespaceUriID,
		namespaceUri:   namespaceUri,
	}
	as.Event = a
	return a
}

func (e *StartElementNS) GetNamespaceUri() string {
	return e.namespaceUri
}

func (e *StartElementNS) GetNamespaceUriID() int {
	return e.namespaceUriID
}
