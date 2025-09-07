package core

type Production interface {
	GetEvent() Event
	GetNextGrammar() Grammar
	GetEventCode() int
}

type AbstractProductionType interface {
	Production
	GetEventCode() int
}

type AbstractProduction struct {
	AbstractProductionType
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

type SchemaInformedProduction struct {
	*AbstractProduction
}

func NewSchemaInformedProduction(next Grammar, event Event, eventCode int) *SchemaInformedProduction {
	return &SchemaInformedProduction{
		AbstractProduction: NewAbstractProduction(next, event, eventCode),
	}
}

func (p *SchemaInformedProduction) GetEventCode() int {
	return p.eventCode
}

type SchemaLessProduction struct {
	*AbstractProduction
	parent Grammar
}

func NewSchemaLessProduction(parent, next Grammar, event Event, eventCode int) *SchemaLessProduction {
	return &SchemaLessProduction{
		AbstractProduction: NewAbstractProduction(next, event, eventCode),
		parent:             parent,
	}
}

func (p *SchemaLessProduction) GetEventCode() int {
	return p.parent.GetNumberOfEvents() - 1 - p.eventCode
}
