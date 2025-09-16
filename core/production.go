package core

type Production interface {
	GetEvent() Event
	GetNextGrammar() Grammar
	GetEventCode() int
	Duplicate() Production
}

// type AbstractProductionType interface {
// 	Production
// 	GetEventCode() int
// }

type AbstractProduction struct {
	Production
	next      Grammar
	eventCode int
	event     Event
}

func NewAbstractProduction(next Grammar, event Event, eventCode int) *AbstractProduction {
	return &AbstractProduction{
		next:      next,
		eventCode: eventCode,
		event:     event,
	}
}

func (p *AbstractProduction) GetEvent() Event {
	return p.event
}

func (p *AbstractProduction) GetNextGrammar() Grammar {
	return p.next
}

func (p *AbstractProduction) Duplicate() Production {
	return &AbstractProduction{
		next:      p.next,
		eventCode: p.eventCode,
		event:     p.event,
	}
}

type SchemaInformedProduction struct {
	*AbstractProduction
}

func NewSchemaInformedProduction(next Grammar, event Event, eventCode int) *SchemaInformedProduction {
	ap := NewAbstractProduction(next, event, eventCode)
	sip := &SchemaInformedProduction{
		AbstractProduction: ap,
	}
	ap.Production = sip

	return sip
}

func (p *SchemaInformedProduction) GetEventCode() int {
	return p.eventCode
}

type SchemaLessProduction struct {
	*AbstractProduction
	parent Grammar
}

func NewSchemaLessProduction(parent, next Grammar, event Event, eventCode int) *SchemaLessProduction {
	ap := NewAbstractProduction(next, event, eventCode)
	slp := &SchemaLessProduction{
		AbstractProduction: ap,
		parent:             parent,
	}
	ap.Production = slp

	return slp
}

func (p *SchemaLessProduction) GetEventCode() int {
	return p.parent.GetNumberOfEvents() - 1 - p.eventCode
}
