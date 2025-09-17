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
	next      Grammar
	eventCode int
	event     Event
}

func NewSchemaInformedProduction(next Grammar, event Event, eventCode int) *SchemaInformedProduction {
	sip := &SchemaInformedProduction{
		next:      next,
		eventCode: eventCode,
		event:     event,
	}

	return sip
}
func (p *SchemaInformedProduction) GetEvent() Event {
	return p.event
}

func (p *SchemaInformedProduction) GetNextGrammar() Grammar {
	return p.next
}

func (p *SchemaInformedProduction) Duplicate() Production {
	return &SchemaInformedProduction{
		next:      p.next,
		eventCode: p.eventCode,
		event:     p.event,
	}
}

func (p *SchemaInformedProduction) GetEventCode() int {
	return p.eventCode
}

type SchemaLessProduction struct {
	next      Grammar
	eventCode int
	event     Event
	parent    Grammar
}

func NewSchemaLessProduction(parent, next Grammar, event Event, eventCode int) *SchemaLessProduction {
	slp := &SchemaLessProduction{
		next:      next,
		eventCode: eventCode,
		event:     event,
		parent:    parent,
	}

	return slp
}

func (p *SchemaLessProduction) GetEvent() Event {
	return p.event
}

func (p *SchemaLessProduction) GetNextGrammar() Grammar {
	return p.next
}

func (p *SchemaLessProduction) Duplicate() Production {
	return &SchemaLessProduction{
		next:      p.next,
		eventCode: p.eventCode,
		event:     p.event,
		parent:    p.parent,
	}
}

func (p *SchemaLessProduction) GetEventCode() int {
	return p.parent.GetNumberOfEvents() - 1 - p.eventCode
}
