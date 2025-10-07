package sax

import (
	"bufio"
	"encoding/xml"
	"fmt"

	"github.com/sderkacs/go-exi/core"
)

const (
	SAX_DefaultCharBufferSize int = 4096
)

type SAXDecoder struct {
	noOptionsFactory  core.EXIFactory
	exiStream         core.EXIStreamDecoder
	namespaces        bool
	namespacePrefixes bool
	exiBodyOnly       bool
	cbuffer           []rune
	debug             bool
	attributeList     []xml.Attr
	namespaceList     []core.NamespaceDeclarationContainer
	isFirstElement    bool
}

func NewSAXDecoder(noOptionsFactory core.EXIFactory) (*SAXDecoder, error) {
	return NewSAXDecoderWithBuffer(noOptionsFactory, make([]rune, SAX_DefaultCharBufferSize))
}

func NewSAXDecoderWithBuffer(noOptionsFactory core.EXIFactory, cbuffer []rune) (*SAXDecoder, error) {
	exiStream, err := noOptionsFactory.CreateEXIStreamDecoder()
	if err != nil {
		return nil, err
	}

	return &SAXDecoder{
		noOptionsFactory:  noOptionsFactory,
		exiStream:         exiStream,
		namespaces:        true,
		namespacePrefixes: noOptionsFactory.GetFidelityOptions().IsFidelityEnabled(core.FeaturePrefix),
		exiBodyOnly:       false,
		cbuffer:           cbuffer,
		debug:             false,
		attributeList:     []xml.Attr{},
		namespaceList:     []core.NamespaceDeclarationContainer{},
		isFirstElement:    true,
	}, nil
}

func (d *SAXDecoder) GetFeature(name string) (bool, error) {
	switch name {
	case "http://xml.org/sax/features/namespaces":
		return d.namespaces, nil
	case "http://xml.org/sax/features/namespace-prefixes":
		return d.namespacePrefixes, nil
	default:
		return false, nil
	}
}

func (d *SAXDecoder) SetFeature(name string, value bool) error {
	switch name {
	case "http://xml.org/sax/features/namespaces":
		d.namespaces = value
	case "http://xml.org/sax/features/namespace-prefixes":
		d.namespacePrefixes = value
	case core.W3C_EXI_FeatureBodyOnly:
		d.exiBodyOnly = value
	default:
		return fmt.Errorf("SAX feature not supported: %s", name)
	}

	return nil
}

func (d *SAXDecoder) reset() {
	d.attributeList = []xml.Attr{}
	d.namespaceList = []core.NamespaceDeclarationContainer{}
	d.isFirstElement = true
}

// Parse decodes an EXI-encoded message from the provided bufio.Reader source,
// writes the corresponding XML tokens to the given xml.Encoder writer, and returns
// the local name of the document's root element. If an error occurs during decoding
// or writing, it returns an empty string and the error.
func (d *SAXDecoder) Parse(source *bufio.Reader, writer *xml.Encoder) (string, error) {
	d.reset()

	var decoder core.EXIBodyDecoder
	var err error
	if d.exiBodyOnly {
		decoder, err = d.exiStream.GetBodyOnlyDecoder(source)
		if err != nil {
			return "", err
		}
	} else {
		decoder, err = d.exiStream.DecodeHeader(source)
		if err != nil {
			return "", err
		}
	}

	rootName, err := d.parseEXIEvents(decoder, writer)
	if err != nil {
		return "", err
	}

	return rootName, nil
}

func (d *SAXDecoder) parseEXIEvents(decoder core.EXIBodyDecoder, writer *xml.Encoder) (string, error) {
	var deferredStartElement *core.QNameContext = nil
	var err error
	isStartElementDeferred := false
	rootName := ""

	eventType, exists, err := decoder.Next()
	if err != nil {
		return "", err
	}
	for exists {
		switch eventType {
		case core.EventTypeStartDocument:
			if err := decoder.DecodeStartDocument(); err != nil {
				return "", err
			}
		case core.EventTypeEndDocument:
			if err := decoder.DecodeEndDocument(); err != nil {
				return "", err
			}
		case core.EventTypeAttributeXsiNil:
			qnc, err := decoder.DecodeAttributeXsiNil()
			if err != nil {
				return "", err
			}

			if err := d.handleAttribute(decoder, qnc); err != nil {
				return "", err
			}
		case core.EventTypeAttributeXsiType:
			qnc, err := decoder.DecodeAttributeXsiType()
			if err != nil {
				return "", err
			}
			if err := d.handleAttribute(decoder, qnc); err != nil {
				return "", err
			}
		case core.EventTypeAttribute,
			core.EventTypeAttributeNS,
			core.EventTypeAttributeGeneric,
			core.EventTypeAttributeGenericUndeclared,
			core.EventTypeAttributeInvalidValue,
			core.EventTypeAttributeAnyInvalidValue:

			qnc, err := decoder.DecodeAttribute()
			if err != nil {
				return "", err
			}

			if err := d.handleAttribute(decoder, qnc); err != nil {
				return "", err
			}
		case core.EventTypeNamespaceDeclaration:
			nsDecl, err := decoder.DecodeNamespaceDeclaration()
			if err != nil {
				return "", err
			}
			if d.debug {
				fmt.Printf("NSDECL: %+v\n", nsDecl)
			}

			d.namespaceList = append(d.namespaceList, *nsDecl)
		case core.EventTypeSelfContained:
			if err := decoder.DecodeStartSelfContainedFragment(); err != nil {
				return "", err
			}
		case core.EventTypeStartElement,
			core.EventTypeStartElementNS,
			core.EventTypeStartElementGeneric,
			core.EventTypeStartElementGenericUndeclared:
			if isStartElementDeferred {
				// handle deferred element if any first
				if err := d.handleDeferredStartElement(decoder, deferredStartElement, writer); err != nil {
					return "", err
				}
				d.isFirstElement = false
			}
			// defer start element and keep on processing
			deferredStartElement, err = decoder.DecodeStartElement()
			if err != nil {
				return "", err
			}
			isStartElementDeferred = true

			if d.isFirstElement {
				rootName = deferredStartElement.GetLocalName()
			}
		case core.EventTypeEndElement, core.EventTypeEndElementUndeclared:
			if isStartElementDeferred {
				// handle deferred element if any first
				if err := d.handleDeferredStartElement(decoder, deferredStartElement, writer); err != nil {
					return "", err
				}
				d.isFirstElement = false
				isStartElementDeferred = false
			}

			// eePrefixes := []core.NamespaceDeclarationContainer{}
			// if d.namespaces {
			// 	eePrefixes = decoder.GetDeclaredPrefixDeclarations()
			// }
			// eeQNameAsString := decoder.GetAttributeQNameAsString()
			eeQName, err := decoder.DecodeEndElement()
			if err != nil {
				return "", err
			}

			// ENCODE
			if err := writer.EncodeToken(xml.EndElement{
				Name: xml.Name{
					Local: eeQName.GetDefaultQNameAsString(),
				},
			}); err != nil {
				return "", err
			}
			if d.debug {
				fmt.Printf("[ENCODE] EndElement{Space: %s, Local: %s}\n", eeQName.GetNamespaceUri(), eeQName.GetLocalName())
			}
		case core.EventTypeCharacters, core.EventTypeCharactersGeneric, core.EventTypeCharactersGenericUndeclared:
			// handle deferred element if any first
			if isStartElementDeferred {
				// handle deferred element if any first
				if err := d.handleDeferredStartElement(decoder, deferredStartElement, writer); err != nil {
					return "", err
				}
				d.isFirstElement = false
				isStartElementDeferred = false
			}

			val, err := decoder.DecodeCharacters()
			if err != nil {
				return "", err
			}

			switch val.GetValueType() {
			case core.ValueTypeBoolean, core.ValueTypeString:
				chars, err := val.GetCharacters()
				if err != nil {
					return "", err
				}

				// ENCODE
				if err := writer.EncodeToken(xml.CharData(string(chars))); err != nil {
					return "", err
				}
				if d.debug {
					fmt.Printf("[ENCODE] CharData: %s\n", string(chars))
				}
			case core.ValueTypeList:
				lv := val.(*core.ListValue)
				values := lv.ToValues()

				if len(values) > 0 {
					vt := values[0].GetValueType()

					for _, val2 := range values {
						switch vt {
						case core.ValueTypeBoolean, core.ValueTypeString:
							chars, err := val2.GetCharacters()
							if err != nil {
								return "", err
							}

							// ENCODE
							if err := writer.EncodeToken(xml.CharData(string(chars))); err != nil {
								return "", err
							}
							if d.debug {
								fmt.Printf("[ENCODE] CharData: %s\n", string(chars))
							}
							// ENCODE
							if err := writer.EncodeToken(xml.CharData(string(core.XSDListDelimCharArray))); err != nil {
								return "", err
							}
							if d.debug {
								fmt.Printf("[ENCODE] CharData: %s\n", string(core.XSDListDelimCharArray))
							}
						default:
							offset := 0
							length, err := val2.GetCharactersLength()
							if err != nil {
								return "", err
							}

							// Weird java code here
							if len(d.cbuffer) < (offset + length + 1) {
								// ENCODE
								if err := writer.EncodeToken(xml.CharData(string(d.cbuffer[:offset]))); err != nil {
									return "", err
								}
								if d.debug {
									fmt.Printf("[ENCODE] CharData: %s\n", string(d.cbuffer[:offset]))
								}
							}

							if err := val2.FillCharactersBuffer(d.cbuffer, offset); err != nil {
								return "", err
							}
							offset += length
							d.cbuffer[offset] = ' '
							offset++

							// ENCODE
							if err := writer.EncodeToken(xml.CharData(string(d.cbuffer[:offset]))); err != nil {
								return "", err
							}
							if d.debug {
								fmt.Printf("[ENCODE] CharData: %s\n", string(d.cbuffer[:offset]))
							}
						}
					}
				}
			default:
				slen, err := val.GetCharactersLength()
				if err != nil {
					return "", err
				}
				if err := d.ensureBufferCapacity(slen); err != nil {
					return "", err
				}

				// fills char array with value
				if err := val.FillCharactersBuffer(d.cbuffer, 0); err != nil {
					return "", err
				}

				// ENCODE
				if err := writer.EncodeToken(xml.CharData(string(d.cbuffer[0:slen]))); err != nil {
					return "", err
				}
				if d.debug {
					fmt.Printf("[ENCODE] CharData: %s\n", string(d.cbuffer[0:slen]))
				}
			}
		case core.EventTypeDocType:
			if isStartElementDeferred {
				// handle deferred element if any first
				if err := d.handleDeferredStartElement(decoder, deferredStartElement, writer); err != nil {
					return "", err
				}
				d.isFirstElement = false
				isStartElementDeferred = false
			}

			docType, err := decoder.DecodeDocType()
			if err != nil {
				return "", err
			}
			if err := d.handleDocType(docType); err != nil {
				return "", err
			}
		case core.EventTypeEntityReference:
			if isStartElementDeferred {
				// handle deferred element if any first
				if err := d.handleDeferredStartElement(decoder, deferredStartElement, writer); err != nil {
					return "", err
				}
				d.isFirstElement = false
				isStartElementDeferred = false
			}

			ref, err := decoder.DecodeEntityReference()
			if err != nil {
				return "", err
			}

			if err := d.handleEntityReference(ref); err != nil {
				return "", err
			}
		case core.EventTypeComment:
			if isStartElementDeferred {
				// handle deferred element if any first
				if err := d.handleDeferredStartElement(decoder, deferredStartElement, writer); err != nil {
					return "", err
				}
				d.isFirstElement = false
				isStartElementDeferred = false
			}

			com, err := decoder.DecodeComment()
			if err != nil {
				return "", err
			}

			if err := d.handleComment(com); err != nil {
				return "", err
			}
		case core.EventTypeProcessingInstruction:
			if isStartElementDeferred {
				// handle deferred element if any first
				if err := d.handleDeferredStartElement(decoder, deferredStartElement, writer); err != nil {
					return "", err
				}
				d.isFirstElement = false
				isStartElementDeferred = false
			}

			pi, err := decoder.DecodeProcessingInstruction()
			if err != nil {
				return "", err
			}

			// ENCODE
			if err := writer.EncodeToken(xml.ProcInst{
				Target: pi.Target,
				Inst:   []byte(pi.Data),
			}); err != nil {
				return "", err
			}
			if d.debug {
				fmt.Printf("[ENCODE] ProcInst{Target = %s, Data = %s}\n", pi.Target, pi.Data)
			}
		default:
			return "", fmt.Errorf("unexpected EXI event: %d", eventType)
		}

		eventType, exists, err = decoder.Next()
		if d.debug {
			fmt.Printf("[NEXT] ET: %d, Exists: %v, Err: %v\n", eventType, exists, err)
		}
		if err != nil {
			return "", err
		}
	}

	return rootName, writer.Flush()
}

func (d *SAXDecoder) handleDeferredStartElement(decoder core.EXIBodyDecoder, deferredStartElement *core.QNameContext, writer *xml.Encoder) error {
	nsAttrs := []xml.Attr{}

	if d.namespaces && d.isFirstElement {
		prefixes := decoder.GetDeclaredPrefixDeclarations()
		for _, prefix := range prefixes {
			p := ""
			if prefix.Prefix != nil {
				p = *prefix.Prefix
			}
			if d.debug {
				fmt.Printf("NSDECL(DEF): %+v, Prefix: %s\n", prefix, p)
			}
			nsAttrs = append(nsAttrs, xml.Attr{
				Name: xml.Name{
					Local: fmt.Sprintf("xmlns:%s", p),
				},
				Value: prefix.NamespaceURI,
			})
		}
	}

	/*
	 * the qualified name is required when the namespace-prefixes property
	 * is true, and is optional when the namespace-prefixes property is
	 * false (the default).
	 */
	// String seQNameAsString = Constants.EMPTY_STRING;
	// if (namespacePrefixes) {
	// seQNameAsString = decoder.getElementQNameAsString();
	// }
	/*
	 * Note: it looks like widely used APIs (Xerces, Saxon, ..) provide the
	 * textual qname even when
	 * http://xml.org/sax/features/namespace-prefixes is set to false
	 * http:// sourceforge.net/projects/exificient/forums/forum/856596/topic
	 * /5839494
	 */

	//seQNameAsString := decoder.GetElementQNameAsString()

	// start so far deferred start element
	// ENCODE

	attrs := []xml.Attr{}
	if d.isFirstElement {
		attrs = append(attrs, nsAttrs...)
	}
	attrs = append(attrs, d.attributeList...)

	if err := writer.EncodeToken(xml.StartElement{
		Name: xml.Name{
			Local: deferredStartElement.GetDefaultQNameAsString(),
		},
		Attr: attrs,
	}); err != nil {
		return err
	}
	if d.debug {
		fmt.Printf("[ENCODE] StartElement{Space: %s, Local: %s}\n", deferredStartElement.GetNamespaceUri(), deferredStartElement.GetLocalName())
	}

	// clear attributes
	d.attributeList = []xml.Attr{}
	return nil
}

func (d *SAXDecoder) ensureBufferCapacity(reqSize int) error {
	if reqSize > len(d.cbuffer) {
		newSize := len(d.cbuffer)

		for {
			newSize = newSize << 2

			if !(newSize < reqSize) {
				break
			}
		}

		d.cbuffer = make([]rune, newSize)
	}

	return nil
}

func (d *SAXDecoder) handleAttribute(decoder core.EXIBodyDecoder, atQName *core.QNameContext) error {
	val := decoder.GetAttributeValue()

	var sVal string
	var err error

	switch val.GetValueType() {
	case core.ValueTypeBoolean, core.ValueTypeString:
		sVal, err = val.ToString()
		if err != nil {
			return err
		}
	case core.ValueTypeList:
		lv := val.(*core.ListValue)

		if lv.GetNumberOfValues() > 0 {
			runes := []rune{}

			values := lv.ToValues()
			vt := values[0].GetValueType()

			for _, v2 := range values {
				switch vt {
				case core.ValueTypeBoolean, core.ValueTypeString:
					r, err := v2.GetCharacters()
					if err != nil {
						return err
					}
					runes = append(runes, r...)
				default:
					slen, err := v2.GetCharactersLength()
					if err != nil {
						return err
					}
					if err := d.ensureBufferCapacity(slen); err != nil {
						return err
					}
					if err := v2.FillCharactersBuffer(d.cbuffer, 0); err != nil {
						return err
					}
					runes = append(runes, d.cbuffer[:slen]...)
				}
			}

			sVal = string(runes)
		} else {
			sVal = core.EmptyString
		}
	default:
		slen, err := val.GetCharactersLength()
		if err != nil {
			return err
		}
		if err := d.ensureBufferCapacity(slen); err != nil {
			return err
		}
		sVal, err = val.BufferToString(d.cbuffer, 0)
		if err != nil {
			return err
		}
	}

	// empty string if no qualified name is necessary
	// String atQNameAsString = Constants.EMPTY_STRING;
	// if (namespacePrefixes) {
	// atQNameAsString = decoder.getAttributeQNameAsString();
	//}
	/*
	 * Note: it looks like widely used APIs (Xerces, Saxon, ..) provide the
	 * textual qname even when
	 * http://xml.org/sax/features/namespace-prefixes is set to false
	 * http://
	 * sourceforge.net/projects/exificient/forums/forum/856596/topic/5839494
	 */

	atQNameAsString := decoder.GetAttributeQNameAsString()
	attr := xml.Attr{
		Name: xml.Name{
			Local: atQNameAsString,
		},
		Value: sVal,
	}
	d.attributeList = append(d.attributeList, attr)

	//TODO: Add attr, cdata
	if d.debug {
		fmt.Printf("ADD ATTR: %s = %s\n", attr, sVal)
	}
	return nil
}

func (d *SAXDecoder) handleDocType(docType *core.DocTypeContainer) error {
	if docType != nil {
		if d.debug {
			fmt.Printf("DOC TYPE: %+v\n", *docType)
		}
	}
	return nil
}

func (d *SAXDecoder) handleEntityReference(erName []rune) error {
	if d.debug {
		fmt.Printf("EREF: %s\n", string(erName))
	}
	return nil
}

func (d *SAXDecoder) handleComment(comment []rune) error {
	if d.debug {
		fmt.Printf("COM: %s\n", string(comment))
	}
	return nil
}
