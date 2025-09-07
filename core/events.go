package core

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
	if e == other {
		return true
	}

	if e.eventType == other.GetEventType() {
		return true
	}

	return false
}

type AbstractDatatypeEvent struct {
	AbstractEvent
	datatype Datatype
}

func (e *AbstractDatatypeEvent) GetDataType() Datatype {
	return e.datatype
}

type Attribute struct {
	*AbstractDatatypeEvent
	qname        QName
	qnameContext *QNameContext
}

func NewAttribute(qnc *QNameContext) *Attribute {
	return NewAttributeWithDatatype(qnc, BuiltInGetDefaultDatatype())
}

func NewAttributeWithDatatype(qnc *QNameContext, datatype Datatype) *Attribute {
	return &Attribute{
		AbstractDatatypeEvent: &AbstractDatatypeEvent{
			AbstractEvent: AbstractEvent{
				eventType: EventTypeAttribute,
			},
			datatype: datatype,
		},
		qnameContext: qnc,
		qname:        qnc.GetQName(),
	}
}

func (a *Attribute) GetQNameContext() *QNameContext {
	return a.qnameContext
}

func (a *Attribute) GetQName() QName {
	return a.qname
}

func (a *Attribute) Equals(other Event) bool {
	if a == other {
		return true
	}
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
	return &AttributeGeneric{
		AbstractEvent: &AbstractEvent{
			eventType: EventTypeAttributeGeneric,
		},
	}
}

type AttributeNS struct {
	*AbstractEvent
	namespaceUri   string
	namespaceUriID int
}

func NewAttributeNS(namespaceUriID int, namespaceUri string) *AttributeNS {
	return &AttributeNS{
		AbstractEvent: &AbstractEvent{
			eventType: EventTypeAttributeNS,
		},
		namespaceUriID: namespaceUriID,
		namespaceUri:   namespaceUri,
	}
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
	return &Characters{
		AbstractDatatypeEvent: &AbstractDatatypeEvent{
			AbstractEvent: AbstractEvent{
				eventType: EventTypeCharacters,
			},
			datatype: datatype,
		},
	}
}

type CharactersGeneric struct {
	*AbstractDatatypeEvent
}

func NewCharactersGeneric() *CharactersGeneric {
	return &CharactersGeneric{
		AbstractDatatypeEvent: &AbstractDatatypeEvent{
			AbstractEvent: AbstractEvent{
				eventType: EventTypeCharactersGeneric,
			},
			datatype: BuiltInGetDefaultDatatype(),
		},
	}
}

type Comment struct {
	*AbstractEvent
}

func NewComment() *Comment {
	return &Comment{
		AbstractEvent: &AbstractEvent{
			eventType: EventTypeComment,
		},
	}
}

type DocType struct {
	*AbstractEvent
}

func NewDocType() *DocType {
	return &DocType{
		AbstractEvent: &AbstractEvent{
			eventType: EventTypeDocType,
		},
	}
}

type EndDocument struct {
	*AbstractEvent
}

func NewEndDocument() *EndDocument {
	return &EndDocument{
		AbstractEvent: &AbstractEvent{
			eventType: EventTypeEndDocument,
		},
	}
}

type EndElement struct {
	*AbstractEvent
}

func NewEndElement() *EndElement {
	return &EndElement{
		AbstractEvent: &AbstractEvent{
			eventType: EventTypeEndElement,
		},
	}
}

type EntityReference struct {
	*AbstractEvent
}

func NewEntityReference() *EntityReference {
	return &EntityReference{
		AbstractEvent: &AbstractEvent{
			eventType: EventTypeEntityReference,
		},
	}
}

type NamespaceDeclaration struct {
	*AbstractEvent
}

func NewNamespaceDeclaration() *NamespaceDeclaration {
	return &NamespaceDeclaration{
		AbstractEvent: &AbstractEvent{
			eventType: EventTypeNamespaceDeclaration,
		},
	}
}

type ProcessingInstruction struct {
	*AbstractEvent
}

func NewProcessingInstruction() *ProcessingInstruction {
	return &ProcessingInstruction{
		AbstractEvent: &AbstractEvent{
			eventType: EventTypeProcessingInstruction,
		},
	}
}

type SelfContained struct {
	*AbstractEvent
}

func NewSelfContained() *SelfContained {
	return &SelfContained{
		AbstractEvent: &AbstractEvent{
			eventType: EventTypeSelfContained,
		},
	}
}

type StartDocument struct {
	*AbstractEvent
}

func NewStartDocument() *StartDocument {
	return &StartDocument{
		AbstractEvent: &AbstractEvent{
			eventType: EventTypeStartDocument,
		},
	}
}

type StartElement struct {
	*AbstractEvent
	qname        QName
	qnameContext *QNameContext
	grammar      Grammar
}

func NewStartElement(qnc *QNameContext) *StartElement {
	return &StartElement{
		AbstractEvent: &AbstractEvent{
			eventType: EventTypeStartElement,
		},
		qnameContext: qnc,
		qname:        qnc.qName,
	}
}

func NewStartElementWithGrammar(qnc *QNameContext, grammar Grammar) *StartElement {
	se := NewStartElement(qnc)
	se.grammar = grammar
	return se
}

func (e *StartElement) GetQNameContext() *QNameContext {
	return e.qnameContext
}

func (e *StartElement) GetQName() QName {
	return e.qname
}

func (e *StartElement) SetGrammar(grammar Grammar) {
	e.grammar = grammar
}

func (e *StartElement) GetGrammar() Grammar {
	return e.grammar
}

type StartElementGeneric struct {
	*AbstractEvent
}

func NewStartElementGeneric() *StartElementGeneric {
	return &StartElementGeneric{
		AbstractEvent: &AbstractEvent{
			eventType: EventTypeStartElementGeneric,
		},
	}
}

type StartElementNS struct {
	*AbstractEvent
	namespaceUri   string
	namespaceUriID int
}

func NewStartElementNS(namespaceUriID int, namespaceUri string) *StartElementNS {
	return &StartElementNS{
		AbstractEvent: &AbstractEvent{
			eventType: EventTypeStartElementNS,
		},
		namespaceUriID: namespaceUriID,
		namespaceUri:   namespaceUri,
	}
}

func (e *StartElementNS) GetNamespaceUri() string {
	return e.namespaceUri
}

func (e *StartElementNS) GetNamespaceUriID() int {
	return e.namespaceUriID
}
