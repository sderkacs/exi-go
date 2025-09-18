package sax

import (
	"bufio"
	"encoding/xml"
	"io"
	"strings"

	"github.com/sderkacs/exi-go/core"
)

type SAXEncoder struct {
	factory       core.EXIFactory
	exiStream     core.EXIStreamEncoder
	encoder       core.EXIBodyEncoder
	exiAttributes core.AttributeList
}

func NewSAXEncoder(factory core.EXIFactory) (*SAXEncoder, error) {
	exiStream, err := factory.CreateEXIStreamEncoder()
	if err != nil {
		return nil, err
	}

	return &SAXEncoder{
		factory:       factory,
		exiStream:     exiStream,
		encoder:       nil,
		exiAttributes: core.NewAttributeListImpl(factory),
	}, nil
}

func (s *SAXEncoder) SetWriter(writer *bufio.Writer) error {
	enc, err := s.exiStream.EncodeHeader(writer)
	if err != nil {
		return err
	}
	s.encoder = enc
	return nil
}

func (s *SAXEncoder) StartPrefixMapping(prefix *string, uri string) error {
	s.exiAttributes.AddNamespaceDeclaration(uri, prefix)
	return nil
}

func (s *SAXEncoder) StartElement(uri, local string, raw *string, attributes []xml.Attr) error {
	return s.startElementPfx(uri, local, nil, attributes)
}

func (s *SAXEncoder) startElementPfx(uri, local string, prefix *string, attributes []xml.Attr) error {
	if err := s.encoder.EncodeStartElement(uri, local, prefix); err != nil {
		return err
	}

	for _, attr := range attributes {
		prefix := s.getPrefixOf(&attr)

		// Skip namespace declarations
		if attr.Name.Space == core.XML_NS_Attribute {
			continue
		}

		s.exiAttributes.AddAttribute(&attr.Name.Space, attr.Name.Local, &prefix, attr.Value)
	}

	if err := s.encoder.EncodeAttributeList(s.exiAttributes); err != nil {
		return err
	}
	s.exiAttributes.Clear()

	return nil
}

func (s *SAXEncoder) getPrefixOf(attr *xml.Attr) string {
	idx := strings.Index(attr.Name.Local, ":")
	if idx == -1 {
		return core.XMLDefaultNSPrefix
	} else {
		return attr.Name.Local[:idx]
	}
}

func (s *SAXEncoder) StartDocument() error {
	return s.encoder.EncodeStartDocument()
}

func (s *SAXEncoder) EndDocument() error {
	if err := s.encoder.EncodeEndDocument(); err != nil {
		return err
	}
	return s.encoder.Flush()
}

func (s *SAXEncoder) EndElement(uri, local string, raw *string) error {
	return s.encoder.EncodeEndElement()
}

func (s *SAXEncoder) Characters(ch []rune, start, length int) error {
	return s.encoder.EncodeCharacters(core.NewStringValueFromSlice(ch[start : start+length]))
}

func (s *SAXEncoder) Encode(reader *bufio.Reader, reference []byte) error {
	dec := xml.NewDecoder(reader)

	start := true

	for {
		token, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				if err := s.EndDocument(); err != nil {
					return err
				}

				return nil
			}
			return err
		}

		if start {
			if err := s.StartDocument(); err != nil {
				return err
			}
			start = false
		}

		switch tok := token.(type) {
		case xml.StartElement:
			if err := s.StartElement(tok.Name.Space, tok.Name.Local, nil, tok.Attr); err != nil {
				return err
			}
		case xml.EndElement:
			if err := s.EndElement(tok.Name.Space, tok.Name.Local, nil); err != nil {
				return err
			}
		case xml.CharData:
			str := string(tok)

			if err := s.Characters([]rune(str), 0, len(str)); err != nil {
				return err
			}
		default:
			// Skip for now
		}
	}
}
