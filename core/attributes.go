package core

import (
	"strings"

	"github.com/sderkacs/exi-go/utils"
)

type AttributeList interface {
	Clear()

	// NS
	AddNamespaceDeclaration(uri string, prefix *string)
	GetNumberOfNamespaceDeclarations() int
	GetNamespaceDeclaration(index int) *NamespaceDeclarationContainer

	// Any attribute other that NS
	AddAttribute(uri *string, localName string, prefix *string, value string)
	AddAttributeByQName(at QName, value string)

	// XSI-Type
	HasXsiType() bool
	GetXsiTypeRaw() *string
	GetXsiTypePrefix() *string

	// XSI-Nil
	HasXsiNil() bool
	GetXsiNil() *string
	GetXsiNilPrefix() *string

	// Attributes
	GetNumberOfAttributes() int
	GetAttributeURI(index int) *string
	GetAttributeLocalName(index int) *string
	GetAttributeValue(index int) *string
	GetAttributePrefix(index int) *string
}

/*
	AttributeListImpl implementation
*/

const (
	XMLNS_PrefixStart = len(XML_NS_Attribute) + 1
)

type AttributeListImpl struct {
	AttributeList

	// options
	isSchemaInformed       bool
	isCanonical            bool
	preserveSchemaLocation bool
	preservePrefixes       bool

	// xsi:type
	hasXsiType    bool
	xsiTypeRaw    *string
	xsiTypePrefix *string

	//xsi:nil
	hasXsiNil    bool
	xsiNil       *string
	xsiNilPrefix *string

	// attributes
	attributeURI       []string
	attributeLocalName []string
	attributeValue     []string
	attributePrefix    []string

	// NS, prefix mappings
	nsDecls []NamespaceDeclarationContainer
}

func NewAttributeListImpl(exiFactory EXIFactory) *AttributeListImpl {
	return &AttributeListImpl{
		isSchemaInformed:       exiFactory.GetGrammars().IsSchemaInformed(),
		isCanonical:            exiFactory.GetEncodingOptions().IsOptionEnabled(OptionCanonicalExi),
		preserveSchemaLocation: exiFactory.GetEncodingOptions().IsOptionEnabled(OptionIncludeXsiSchemaLocation),
		preservePrefixes:       exiFactory.GetFidelityOptions().IsFidelityEnabled(FeaturePrefix),
		hasXsiType:             false,
		xsiTypeRaw:             nil,
		xsiTypePrefix:          nil,
		hasXsiNil:              false,
		xsiNil:                 nil,
		xsiNilPrefix:           nil,
		attributeURI:           []string{},
		attributeLocalName:     []string{},
		attributeValue:         []string{},
		attributePrefix:        []string{},
		nsDecls:                []NamespaceDeclarationContainer{},
	}
}

func (list *AttributeListImpl) Clear() {
	list.hasXsiType = false
	list.hasXsiNil = false

	list.attributeURI = []string{}
	list.attributeLocalName = []string{}
	list.attributeValue = []string{}
	list.attributePrefix = []string{}

	list.xsiTypeRaw = nil

	list.nsDecls = []NamespaceDeclarationContainer{}
}

/*
	XSI-Type
*/

func (list *AttributeListImpl) HasXsiType() bool {
	return list.hasXsiType
}

func (list *AttributeListImpl) GetXsiTypeRaw() *string {
	return list.xsiTypeRaw
}

func (list *AttributeListImpl) GetXsiTypePrefix() *string {
	return list.xsiTypePrefix
}

/*
	XSI-Nil
*/

func (list *AttributeListImpl) HasXsiNil() bool {
	return list.hasXsiNil
}

func (list *AttributeListImpl) GetXsiNil() *string {
	return list.xsiNil
}

func (list *AttributeListImpl) GetXsiNilPrefix() *string {
	return list.xsiNilPrefix
}

/*
	Attributes
*/

func (list *AttributeListImpl) GetNumberOfAttributes() int {
	return len(list.attributeURI)
}

func (list *AttributeListImpl) GetAttributeURI(index int) *string {
	return &list.attributeURI[index]
}

func (list *AttributeListImpl) GetAttributeLocalName(index int) *string {
	return &list.attributeLocalName[index]
}

func (list *AttributeListImpl) GetAttributeValue(index int) *string {
	return &list.attributeValue[index]
}

func (list *AttributeListImpl) GetAttributePrefix(index int) *string {
	return &list.attributeValue[index]
}

func (list *AttributeListImpl) setXsiType(rawType *string, xsiPrefix *string) {
	list.hasXsiType = true
	list.xsiTypeRaw = rawType
	list.xsiTypePrefix = xsiPrefix
}

func (list *AttributeListImpl) setXsiNil(rawNil, xsiPrefix *string) {
	list.hasXsiNil = true
	list.xsiNil = rawNil
	list.xsiNilPrefix = xsiPrefix
}

// NS

func (list *AttributeListImpl) AddNamespaceDeclaration(uri string, prefix *string) {
	// Canonical EXI defines that namespace declarations MUST be sorted
	// lexicographically according to the NS prefix
	if len(list.nsDecls) == 0 || !list.isCanonical {
		list.nsDecls = append(list.nsDecls, NewNamespaceDeclarationContainer(uri, prefix))
	} else {
		i := len(list.nsDecls)

		for i > 0 && list.isGreaterNS(i-1, prefix) {
			// move right
			i--
		}

		// update position i
		list.nsDecls = utils.SliceAddAtIndex(list.nsDecls, i, NewNamespaceDeclarationContainer(uri, prefix))
	}
}

func (list *AttributeListImpl) GetNumberOfNamespaceDeclarations() int {
	return len(list.nsDecls)
}

func (list *AttributeListImpl) GetNamespaceDeclaration(index int) NamespaceDeclarationContainer {
	if index < 0 || index >= len(list.nsDecls) {
		panic("index out of bounds")
	}
	return list.nsDecls[index]
}

// Any attribute other that NS

func (list *AttributeListImpl) AddAttribute(uri *string, localName string, prefix *string, value string) {
	if uri == nil {
		uri = utils.AsPtr(XMLNullNS_URI)
	}
	if prefix == nil {
		prefix = utils.AsPtr(XMLDefaultNSPrefix)
	}

	// xsi:*
	if *uri == XMLSchemaInstanceNS_URI {
		// xsi:type
		if localName == XSIType {
			// Note: prefix to uri mapping is done later on
			list.setXsiType(&value, prefix)
		} else if localName == XSINil { // xsi:nil
			list.setXsiNil(&value, prefix)
		} else if (localName == XSISchemaLocation || localName == XSINoNamespaceSchemaLocation) && !list.preserveSchemaLocation {
			// prune xsi:schemaLocation
		} else {
			list.insertAttribute(uri, localName, prefix, value)
		}
	} else {
		// == Attribute (AT)
		// When schemas are used, attribute events occur in lexical order
		// sorted first by qname localName then by qname uri.
		list.insertAttribute(uri, localName, prefix, value)
	}
}

func (list *AttributeListImpl) AddAttributeByQName(at QName, value string) {
	list.AddAttribute(&at.Space, at.Local, at.Prefix, value)
}

func (list *AttributeListImpl) insertAttribute(uri *string, localName string, prefix *string, value string) {
	if list.isSchemaInformed || list.isCanonical {
		// sorted attributes
		i := len(list.attributeURI)

		for i > 0 && list.IsGreaterAttribute(i-1, *uri, localName) {
			// move right
			i--
		}

		list.attributeURI = utils.SliceAddAtIndex(list.attributeURI, i, *uri)
		list.attributeLocalName = utils.SliceAddAtIndex(list.attributeLocalName, i, localName)
		list.attributePrefix = utils.SliceAddAtIndex(list.attributePrefix, i, *prefix)
		list.attributeValue = utils.SliceAddAtIndex(list.attributeValue, i, value)
	} else {
		// attribute order does not matter
		list.attributeURI = append(list.attributeURI, *uri)
		list.attributeLocalName = append(list.attributeLocalName, localName)
		list.attributePrefix = append(list.attributePrefix, *prefix)
		list.attributeValue = append(list.attributeValue, value)
	}
}

func (list *AttributeListImpl) isGreaterNS(nsIndex int, prefix *string) bool {
	compPrefix := strings.Compare(*list.GetNamespaceDeclaration(nsIndex).prefix, *prefix)

	return compPrefix > 0
}

func (list *AttributeListImpl) IsGreaterAttribute(attributeIndex int, uri string, localName string) bool {
	compLocalName := strings.Compare(*list.GetAttributeLocalName(attributeIndex), localName)

	if compLocalName > 0 {
		return true
	} else if compLocalName < 0 {
		return false
	} else {
		return strings.Compare(*list.GetAttributeURI(attributeIndex), uri) > 0
	}
}
