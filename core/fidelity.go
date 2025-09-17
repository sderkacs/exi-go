package core

import (
	"fmt"
	"maps"
	"slices"

	"github.com/sderkacs/exi-go/utils"
)

const (
	/* Comments, ProcessingInstructions, DTDs and Prefixes are preserved */
	FeatureComment string = "PRESERVE_COMMENTS"
	FeaturePI      string = "PRESERVE_PIS"
	FeatureDTD     string = "PRESERVE_DTDS"
	FeaturePrefix  string = "PRESERVE_PREFIX"

	/*
	 * Lexical form of element and attribute values is preserved in value
	 * content items
	 */
	FeatureLexicalValue string = "PRESERVE_LEXICAL_VALUES"

	/* Enable the use of self contained elements in the EXI stream */
	FeatureSC string = "SELF_CONTAINED"

	/* Strict interpretation of schemas is used to achieve better compactness */
	FeatureStrict string = "STRICT"
)

type FidelityOptions struct {
	options        map[string]struct{}
	isStrict       bool
	isComment      bool
	isPI           bool
	isDTD          bool
	isPrefix       bool
	isLexicalValue bool
	isSC           bool
}

func NewDefaultFidelityOptions() *FidelityOptions {
	return &FidelityOptions{
		options:        map[string]struct{}{},
		isStrict:       false,
		isComment:      false,
		isPI:           false,
		isDTD:          false,
		isPrefix:       false,
		isLexicalValue: false,
		isSC:           false,
	}
}

func NewStrictFidelityOptions() *FidelityOptions {
	fo := NewDefaultFidelityOptions()

	fo.options[FeatureStrict] = struct{}{}
	fo.isStrict = true

	return fo
}

func NewAllFidelityOptions() *FidelityOptions {
	fo := NewDefaultFidelityOptions()

	fo.options[FeatureComment] = struct{}{}
	fo.isComment = true
	fo.options[FeaturePI] = struct{}{}
	fo.isPI = true
	fo.options[FeatureDTD] = struct{}{}
	fo.isDTD = true
	fo.options[FeaturePrefix] = struct{}{}
	fo.isPrefix = true
	fo.options[FeatureLexicalValue] = struct{}{}
	fo.isLexicalValue = true

	return fo
}

func (fo *FidelityOptions) SetFidelity(key string, decision bool) error {
	switch key {
	case FeatureStrict:
		if decision {
			_, prevContainedLexVal := fo.options[FeatureLexicalValue]

			fo.options = map[string]struct{}{}
			fo.isComment = false
			fo.isPI = false
			fo.isDTD = false
			fo.isPrefix = false
			fo.isLexicalValue = false
			fo.isSC = false

			if prevContainedLexVal {
				fo.options[FeatureLexicalValue] = struct{}{}
				fo.isLexicalValue = true
			}

			fo.options[FeatureStrict] = struct{}{}
			fo.isStrict = true
		} else {
			delete(fo.options, key)
			fo.isStrict = false
		}
	case FeatureLexicalValue:
		if decision {
			fo.options[key] = struct{}{}
			fo.isLexicalValue = true
		} else {
			delete(fo.options, key)
			fo.isLexicalValue = false
		}
	case FeatureComment, FeaturePI, FeatureDTD, FeaturePrefix, FeatureSC:
		if decision {
			if fo.isStrict {
				delete(fo.options, FeatureStrict)
				fo.isStrict = false
			}

			fo.options[key] = struct{}{}
			if key == FeatureComment {
				fo.isComment = true
			}
			if key == FeaturePI {
				fo.isPI = true
			}
			if key == FeatureDTD {
				fo.isDTD = true
			}
			if key == FeaturePrefix {
				fo.isPrefix = true
			}
			if key == FeatureSC {
				fo.isSC = true
			}
		} else {
			delete(fo.options, key)
			if key == FeatureComment {
				fo.isComment = false
			}
			if key == FeaturePI {
				fo.isPI = false
			}
			if key == FeatureDTD {
				fo.isDTD = false
			}
			if key == FeaturePrefix {
				fo.isPrefix = false
			}
			if key == FeatureSC {
				fo.isSC = false
			}
		}
	default:
		return fmt.Errorf("FidelityOption '%s' is unknown", key)
	}

	return nil
}

func (fo *FidelityOptions) IsFidelityEnabled(key string) bool {
	_, exists := fo.options[key]
	return exists
}

func (fo *FidelityOptions) IsStrict() bool {
	return fo.isStrict
}

func (fo *FidelityOptions) Get1stLevelEventCodeLength(grammar Grammar) int {
	var cl1 int

	switch grammar.GetGrammarType() {
	case GrammarTypeDocument, GrammarTypeFragment:
		cl1 = 0
	case GrammarTypeDocEnd:
		if fo.isComment || fo.isPI {
			cl1 = 1
		} else {
			cl1 = 0
		}
	case GrammarTypeSchemaInformedDocContent, GrammarTypeBuiltInDocContent:
		inc := 0
		if fo.isDTD || fo.isComment || fo.isPI {
			inc = 1
		}
		cl1 = utils.GetCodingLength(grammar.GetNumberOfEvents() + inc)
	case GrammarTypeSchemaInformedFragmentContent, GrammarTypeBuiltInFragmentContent:
		inc := 0
		if fo.isComment || fo.isPI {
			inc = 1
		}
		cl1 = utils.GetCodingLength(grammar.GetNumberOfEvents() + inc)
	case GrammarTypeSchemaInformedFirstStartTagContent, GrammarTypeSchemaInformedStartTagContent, GrammarTypeSchemaInformedElementContent:
		inc := 0
		if fo.Get2ndLevelCharacteristics(grammar) > 0 {
			inc = 1
		}
		cl1 = utils.GetCodingLength(grammar.GetNumberOfEvents() + inc)
	case GrammarTypeBuiltInStartTagContent, GrammarTypeBuiltInElementContent:
		cl1 = utils.GetCodingLength(grammar.GetNumberOfEvents() + 1)
	default:
		cl1 = -1
	}

	return cl1
}

func (fo *FidelityOptions) Get2ndLevelEventType(ec2 int, grammar Grammar) EventType {
	var eventType EventType = EventType(NotFound)

	switch grammar.GetGrammarType() {
	case GrammarTypeDocument, GrammarTypeFragment, GrammarTypeDocEnd, GrammarTypeSchemaInformedFragmentContent, GrammarTypeBuiltInFragmentContent:
		// Root grammars
	case GrammarTypeSchemaInformedDocContent, GrammarTypeBuiltInDocContent:
		if fo.isDTD && ec2 == 0 {
			eventType = EventTypeDocType
		}
	case GrammarTypeSchemaInformedFirstStartTagContent:
		sifst := grammar.(SchemaInformedFirstStartTagGrammar)
		if fo.isStrict {
			if sifst.IsTypeCastable() {
				switch ec2 {
				case 0:
					eventType = EventTypeAttributeXsiType
				case 1:
					eventType = EventTypeAttributeXsiNil
				}
			} else if sifst.IsNillable() && ec2 == 0 {
				eventType = EventTypeAttributeXsiNil
			}
		} else {
			// {0,EE?, 1,xsi:type, 2,xsi:nil, 3,AT*, 4,AT-untyped, 5,NS,
			// 6,SC, 7,SE*, 8,CH, 9,ER, {CM, PI}}
			dec := 0
			if sifst.HasEndElement() {
				dec++
			}
			if ec2 == 0-dec {
				eventType = EventTypeEndElementUndeclared
			} else {
				switch ec2 {
				case 1 - dec:
					eventType = EventTypeAttributeXsiType
				case 2 - dec:
					eventType = EventTypeAttributeXsiNil
				case 3 - dec:
					eventType = EventTypeAttributeGenericUndeclared
				case 4 - dec:
					eventType = EventTypeAttributeInvalidValue
				default:
					if !fo.isPrefix {
						dec++
					}
					if ec2 == 5-dec {
						eventType = EventTypeNamespaceDeclaration
					} else {
						if !fo.isSC {
							dec++
						}
						if ec2 == 6-dec {
							eventType = EventTypeSelfContained
						} else {
							switch ec2 {
							case 7 - dec:
								eventType = EventTypeStartElementGenericUndeclared
							case 8 - dec:
								eventType = EventTypeCharactersGenericUndeclared
							default:
								if !fo.isDTD {
									dec++
								}
								if ec2 == 9-dec {
									eventType = EventTypeEntityReference
								}
							}
						}
					}
				}
			}
		}
	case GrammarTypeSchemaInformedStartTagContent:
		sist := grammar.(SchemaInformedStartTagGrammar)
		if fo.isStrict {
			// no events
		} else {
			// {0,EE?, 1,AT*, 2,AT-untyped, 3,SE*, 4,CH, 5,ER, {CM, PI}}
			dec := 0
			if sist.HasEndElement() {
				dec++
			}
			if ec2 == 0-dec {
				eventType = EventTypeEndElementUndeclared
			} else {
				switch ec2 {
				case 1 - dec:
					eventType = EventTypeAttributeGenericUndeclared
				case 2 - dec:
					eventType = EventTypeAttributeInvalidValue
				case 3 - dec:
					eventType = EventTypeStartElementGenericUndeclared
				case 4 - dec:
					eventType = EventTypeCharactersGenericUndeclared
				default:
					if !fo.isDTD {
						dec++
					}
					if ec2 == 5-dec {
						eventType = EventTypeEntityReference
					}
				}
			}
		}
	case GrammarTypeSchemaInformedElementContent:
		sig := grammar.(SchemaInformedGrammar)
		if fo.isStrict {
			// no events
		} else {
			// {0,EE?, 1,SE*, 2,CH*, 3,ER?, {CM, PI}}
			dec := 0
			if sig.HasEndElement() {
				dec++
			}
			switch ec2 {
			case 0 - dec:
				eventType = EventTypeEndElementUndeclared
			case 1 - dec:
				eventType = EventTypeStartElementGenericUndeclared
			case 2 - dec:
				eventType = EventTypeCharactersGenericUndeclared
			default:
				if !fo.isDTD {
					dec++
				}
				if ec2 == 3-dec {
					eventType = EventTypeEntityReference
				}
			}
		}
	case GrammarTypeBuiltInStartTagContent:
		// {0,EE, 1,AT*, 2,NS, 3,SC, 4,SE*, 5,CH, 6,ER, {CM, PI}}
		switch ec2 {
		case 0:
			eventType = EventTypeEndElementUndeclared
		case 1:
			eventType = EventTypeAttributeGenericUndeclared
		default:
			dec := 0
			if !fo.isPrefix {
				dec++
			}
			if ec2 == 2-dec {
				eventType = EventTypeNamespaceDeclaration
			} else {
				if !fo.isSC {
					dec++
				}
				switch ec2 {
				case 3 - dec:
					eventType = EventTypeSelfContained
				case 4 - dec:
					eventType = EventTypeStartElementGenericUndeclared
				case 5 - dec:
					eventType = EventTypeCharactersGenericUndeclared
				default:
					if !fo.isDTD {
						dec++
					}
					if ec2 == 6-dec {
						eventType = EventTypeEntityReference
					}
				}
			}
		}
	case GrammarTypeBuiltInElementContent:
		// {0,SE*, 1,CH, 2,ER, {CM, PI}}
		switch ec2 {
		case 0:
			eventType = EventTypeStartElementGenericUndeclared
		case 1:
			eventType = EventTypeCharactersGenericUndeclared
		default:
			if fo.isDTD && ec2 == 2 {
				eventType = EventTypeEntityReference
			}
		}
	}

	return eventType
}

func (fo *FidelityOptions) Get2ndLevelEventCode(eventType EventType, grammar Grammar) int {
	ec2 := NotFound

	switch grammar.GetGrammarType() {
	case GrammarTypeDocument, GrammarTypeFragment, GrammarTypeDocEnd, GrammarTypeSchemaInformedFragmentContent, GrammarTypeBuiltInFragmentContent:
		// Root grammars
	case GrammarTypeSchemaInformedDocContent, GrammarTypeBuiltInDocContent:
		/* Schema-informed Document and Fragment Grammars */
		/* Built-in Document and Fragment Grammars */
		if fo.isDTD && eventType == EventTypeDocType {
			ec2 = 0
		}
	case GrammarTypeSchemaInformedFirstStartTagContent:
		sifst := grammar.(SchemaInformedFirstStartTagGrammar)
		if fo.isStrict {
			if sifst.IsTypeCastable() {
				switch eventType {
				case EventTypeAttributeXsiType:
					ec2 = 0
				case EventTypeAttributeXsiNil:
					ec2 = 1
				}
			} else if sifst.IsNillable() && eventType == EventTypeAttributeXsiNil {
				ec2 = 0
			}
		} else {
			// {0,EE?, 1,xsi:type, 2,xsi:nil, 3,AT*, 4,AT-untyped, 5,NS,
			// 6,SC, 7,SE*, 8,CH, 9,ER, {CM, PI}}
			dec := 0
			if sifst.HasEndElement() {
				dec++
			}
			switch eventType {
			case EventTypeEndElementUndeclared:
				ec2 = 0 - dec
			case EventTypeAttributeXsiType:
				ec2 = 1 - dec
			case EventTypeAttributeXsiNil:
				ec2 = 2 - dec
			case EventTypeAttributeGenericUndeclared:
				ec2 = 3 - dec
			case EventTypeAttributeInvalidValue:
				ec2 = 4 - dec
			default:
				if !fo.isPrefix {
					dec++
				}
				if eventType == EventTypeNamespaceDeclaration {
					ec2 = 5 - dec
				} else {
					if !fo.isSC {
						dec++
					}
					switch eventType {
					case EventTypeSelfContained:
						ec2 = 6 - dec
					case EventTypeStartElementGenericUndeclared:
						ec2 = 7 - dec
					case EventTypeCharactersGenericUndeclared:
						ec2 = 8 - dec
					default:
						if !fo.isDTD {
							dec++
						}
						if eventType == EventTypeEntityReference {
							ec2 = 9 - dec
						}
					}
				}
			}
		}
	case GrammarTypeSchemaInformedStartTagContent:
		sist := grammar.(SchemaInformedStartTagGrammar)
		if fo.isStrict {
			// no events
		} else {
			// {0,EE?, 1,AT*, 2,AT-untyped, 3,SE*, 4,CH, 5,ER, {CM, PI}}
			dec := 0
			if sist.HasEndElement() {
				dec++
			}
			switch eventType {
			case EventTypeEndElementUndeclared:
				ec2 = 0 - dec
			case EventTypeAttributeGenericUndeclared:
				ec2 = 1 - dec
			case EventTypeAttributeInvalidValue:
				ec2 = 2 - dec
			case EventTypeStartElementGenericUndeclared:
				ec2 = 3 - dec
			case EventTypeCharactersGenericUndeclared:
				ec2 = 4 - dec
			default:
				if !fo.isDTD {
					dec++
				}
				if eventType == EventTypeEntityReference {
					ec2 = 5 - dec
				}
			}
		}
	case GrammarTypeSchemaInformedElementContent:
		sig := grammar.(SchemaInformedGrammar)
		if fo.isStrict {
			// no events
		} else {
			// {0,EE?, 1,SE*, 2,CH*, 3,ER?, {CM, PI}}
			dec := 0
			if sig.HasEndElement() {
				dec++
			}
			switch eventType {
			case EventTypeEndElementUndeclared:
				ec2 = 0 - dec
			case EventTypeStartElementGenericUndeclared:
				ec2 = 1 - dec
			case EventTypeCharactersGenericUndeclared:
				ec2 = 2 - dec
			default:
				if !fo.isDTD {
					dec++
				}
				if eventType == EventTypeEntityReference {
					ec2 = 3 - dec
				}
			}
		}
	case GrammarTypeBuiltInStartTagContent:
		// {0,EE, 1,AT*, 2,NS, 3,SC, 4,SE*, 5,CH, 6,ER, {CM, PI}}
		switch eventType {
		case EventTypeEndElementUndeclared:
			ec2 = 0
		case EventTypeAttributeGenericUndeclared:
			ec2 = 1
		default:
			dec := 0
			if !fo.isPrefix {
				dec++
			}
			if eventType == EventTypeNamespaceDeclaration {
				ec2 = 2 - dec
			} else {
				if !fo.isSC {
					dec++
				}
				switch eventType {
				case EventTypeSelfContained:
					ec2 = 3 - dec
				case EventTypeStartElementGenericUndeclared:
					ec2 = 4 - dec
				case EventTypeCharactersGenericUndeclared:
					ec2 = 5 - dec
				default:
					if !fo.isDTD {
						dec++
					}
					if eventType == EventTypeEntityReference {
						ec2 = 6 - dec
					}
				}
			}
		}
	case GrammarTypeBuiltInElementContent:
		// {0,SE*, 1,CH, 2,ER, {CM, PI}}
		switch eventType {
		case EventTypeStartElementGenericUndeclared:
			ec2 = 0
		case EventTypeCharactersGenericUndeclared:
			ec2 = 1
		default:
			if fo.isDTD && eventType == EventTypeEntityReference {
				ec2 = 2
			}
		}
	}

	return ec2
}

func (fo *FidelityOptions) Get2ndLevelCharacteristics(grammar Grammar) int {
	ch2 := 0

	switch grammar.GetGrammarType() {
	case GrammarTypeDocument, GrammarTypeFragment:
		// Root grammars
		// ch2 = 0
	case GrammarTypeDocEnd:
		if fo.Get3rdLevelCharacteristics() > 0 {
			ch2++
		}
	case GrammarTypeSchemaInformedDocContent, GrammarTypeBuiltInDocContent:
		if fo.isDTD {
			ch2++
		}
		if fo.Get3rdLevelCharacteristics() > 0 {
			ch2++
		}
	case GrammarTypeSchemaInformedFragmentContent, GrammarTypeBuiltInFragmentContent:
		if fo.Get3rdLevelCharacteristics() > 0 {
			ch2++
		}
	case GrammarTypeSchemaInformedFirstStartTagContent:
		sifst := grammar.(SchemaInformedFirstStartTagGrammar)
		if fo.isStrict {
			cst := 0
			if sifst.IsTypeCastable() {
				cst = 1
			}
			nlb := 0
			if sifst.IsNillable() {
				nlb = 1
			}
			ch2 = cst + nlb
		} else {
			// {EE?, xsi:type, xsi:nil, AT*, AT-untyped, NS, SC, SE*, CH,
			// ER, {CM, PI}}
			if !sifst.HasEndElement() {
				ch2++
			}
			ch2 += 4 // xsi:type, xsi:nil, AT*, AT-untyped
			if fo.isPrefix {
				ch2++
			}
			if fo.isSC {
				ch2++
			}
			ch2 += 2 // SE*, CH
			if fo.isDTD {
				ch2++
			}
			if fo.Get3rdLevelCharacteristics() > 0 {
				ch2++
			}
		}
	case GrammarTypeSchemaInformedStartTagContent:
		sist := grammar.(SchemaInformedStartTagGrammar)
		if fo.isStrict {
			// no events
		} else {
			// {EE?, AT*, AT-untyped, SE*, CH, ER, {CM, PI}}
			if !sist.HasEndElement() {
				ch2++
			}
			ch2 += 4 // AT*, AT-untyped, SE*, CH
			if fo.isDTD {
				ch2++
			}
			if fo.Get3rdLevelCharacteristics() > 0 {
				ch2++
			}
		}
	case GrammarTypeSchemaInformedElementContent:
		sig := grammar.(SchemaInformedGrammar)
		if fo.isStrict {
			// no events
		} else {
			// {EE?, SE*, CH*, ER?, {CM, PI}}
			if !sig.HasEndElement() {
				ch2++
			}
			ch2 += 2 // SE*, CH
			if fo.isDTD {
				ch2++
			}
			if fo.Get3rdLevelCharacteristics() > 0 {
				ch2++
			}
		}
	case GrammarTypeBuiltInStartTagContent:
		// {EE, AT*, NS, SC, SE*, CH, ER, {CM, PI}}
		ch2 += 2 // EE, AT*
		if fo.isPrefix {
			ch2++
		}
		if fo.isSC {
			ch2++
		}
		ch2 += 2 // SE*, CH
		if fo.isDTD {
			ch2++
		}
		if fo.Get3rdLevelCharacteristics() > 0 {
			ch2++
		}
	case GrammarTypeBuiltInElementContent:
		// {SE*, CH, ER, {CM, PI}}
		ch2 += 2 // SE*, CH
		if fo.isDTD {
			ch2++
		}
		if fo.Get3rdLevelCharacteristics() > 0 {
			ch2++
		}
	}

	return ch2
}

func (fo *FidelityOptions) Get3rdLevelEventType(ec3 int) EventType {
	switch ec3 {
	case 0:
		if fo.isComment {
			return EventTypeComment
		} else if fo.isPI {
			return EventTypeProcessingInstruction
		}
	case 1:
		return EventTypeProcessingInstruction
	}

	return EventType(NotFound)
}

func (fo *FidelityOptions) Get3rdLevelEventCode(eventType EventType) int {
	if !fo.isStrict {
		if fo.isComment {
			switch eventType {
			case EventTypeComment:
				return 0
			case EventTypeProcessingInstruction:
				return 1
			}
		} else if fo.isPI {
			if eventType == EventTypeProcessingInstruction {
				return 0
			}
		}
	}

	return NotFound
}

func (fo *FidelityOptions) Get3rdLevelCharacteristics() int {
	ch := 0

	if fo.isComment {
		ch++
	}
	if fo.isPI {
		ch++
	}

	return ch
}

func (fo *FidelityOptions) Equals(other utils.ComparableType) bool {
	if other == nil {
		return false
	}
	if fo2, ok := other.(*FidelityOptions); ok {
		so1 := slices.Collect(maps.Keys(fo.options))
		so2 := slices.Collect(maps.Keys(fo2.options))

		slices.Sort(so1)
		slices.Sort(so2)

		return slices.Compare(so1, so2) == 0
	}

	return false
}
