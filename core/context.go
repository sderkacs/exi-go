package core

import (
	"slices"
	"strconv"

	"github.com/sderkacs/go-exi/utils"
)

/*
	QNameContextMapKey implementation
*/

type QNameContextMapKey struct {
	NamespaceUriID int
	LocalNameID    int
}

func NewQNameContextMapKey(qnc *QNameContext) QNameContextMapKey {
	return QNameContextMapKey{
		NamespaceUriID: qnc.GetNamespaceUriID(),
		LocalNameID:    qnc.GetLocalNameID(),
	}
}

/*
	QNameContext implementation
*/

type QNameContext struct {
	namespaceUriId         int
	localNameId            int
	qName                  utils.QName
	defaultQNameAsString   string
	defaultPrefix          string
	grammarGlobalElement   *StartElement
	grammarGlobalAttribute *Attribute
	typeGrammar            SchemaInformedFirstStartTagGrammar
	mapKey                 QNameContextMapKey
}

func NewQNameContext(namespaceUriId int, localNameId int, qName utils.QName) *QNameContext {
	var defaultPrefix string
	var defaultQNameAsString string

	switch namespaceUriId {
	case 0:
		defaultPrefix = ""
		defaultQNameAsString = qName.Local
	case 1:
		defaultPrefix = "xml"
		defaultQNameAsString = "xml:" + qName.Local
	case 2:
		defaultPrefix = "xsi"
		defaultQNameAsString = "xsi:" + qName.Local
	default:
		defaultPrefix = "ns" + strconv.FormatInt(int64(namespaceUriId), 10)
		defaultQNameAsString = defaultPrefix + ":" + qName.Local
	}

	return &QNameContext{
		namespaceUriId:       namespaceUriId,
		localNameId:          localNameId,
		qName:                qName,
		defaultPrefix:        defaultPrefix,
		defaultQNameAsString: defaultQNameAsString,
		mapKey: QNameContextMapKey{
			NamespaceUriID: namespaceUriId,
			LocalNameID:    localNameId,
		},
	}
}

func (q *QNameContext) GetMapKey() QNameContextMapKey {
	return q.mapKey
}

func (q *QNameContext) GetQName() utils.QName {
	return q.qName
}

func (q *QNameContext) GetDefaultQNameAsString() string {
	return q.defaultQNameAsString
}

func (q *QNameContext) GetDefaultPrefix() string {
	return q.defaultPrefix
}

func (q *QNameContext) GetLocalNameID() int {
	return q.localNameId
}

func (q *QNameContext) GetLocalName() string {
	return q.qName.Local
}

func (q *QNameContext) SetGlobalStartElement(grammarGlobalElement *StartElement) {
	q.grammarGlobalElement = grammarGlobalElement
}

func (q *QNameContext) GetGlobalStartElement() *StartElement {
	return q.grammarGlobalElement
}

func (q *QNameContext) SetGlobalAttribute(grammarGlobalAttribute *Attribute) {
	q.grammarGlobalAttribute = grammarGlobalAttribute
}

func (q *QNameContext) GetGlobalAttribute() *Attribute {
	return q.grammarGlobalAttribute
}

func (q *QNameContext) SetTypeGrammar(typeGrammar SchemaInformedFirstStartTagGrammar) {
	q.typeGrammar = typeGrammar
}

func (q *QNameContext) GetTypeGrammar() SchemaInformedFirstStartTagGrammar {
	return q.typeGrammar
}

func (q *QNameContext) GetNamespaceUriID() int {
	return q.namespaceUriId
}

func (q *QNameContext) GetNamespaceUri() string {
	return q.qName.Space
}

func (q *QNameContext) Equals(other *QNameContext) bool {
	if other == nil {
		return false
	}
	return q.localNameId == other.localNameId && q.namespaceUriId == other.namespaceUriId
}

/*
	UriContext implementation
*/

type UriContext interface {
	GetNamespaceUriID() int
	GetNamespaceUri() *string
	GetNumberOfQNames() int
	GetNumberOfPrefixes() int
	GetPrefix(prefixId int) string
	GetPrefixID(prefix string) int
	GetQNameContextById(localNameId int) *QNameContext
	GetQNameContextByName(localName string) *QNameContext
}

type UriContextBase struct {
	namespaceUriId int
	namespaceUri   string
}

func (b *UriContextBase) GetNamespaceUriID() int {
	return b.namespaceUriId
}

func (b *UriContextBase) GetNamespaceUri() string {
	return b.namespaceUri
}

func (b *UriContextBase) Equal(other *UriContextBase) bool {
	return b.namespaceUriId == other.namespaceUriId
}

/*
	GrammarContext implementation
*/

type GrammarContext struct {
	grammarUriContexts    []*GrammarUriContext
	numberOfQNameContexts int
}

func NewGrammarContext(grammarUriContexts []*GrammarUriContext, numberOfQNameContexts int) *GrammarContext {
	return &GrammarContext{
		grammarUriContexts:    grammarUriContexts,
		numberOfQNameContexts: numberOfQNameContexts,
	}
}

func (c *GrammarContext) GetNumberOfGrammarUriContexts() int {
	return len(c.grammarUriContexts)
}

func (c *GrammarContext) GetGrammarUriContextByID(id int) *GrammarUriContext {
	return c.grammarUriContexts[id]
}

func (c *GrammarContext) GetGrammarUriContext(namespaceUri string) *GrammarUriContext {
	for _, ctx := range c.grammarUriContexts {
		if ctx.namespaceUri == namespaceUri {
			return ctx
		}
	}

	return nil
}

func (c *GrammarContext) GetNumberOfGrammarQNameContexts() int {
	return c.numberOfQNameContexts
}

/*
	GrammarUriContext implementation
*/

type GrammarUriContext struct {
	grammarQNames   []*QNameContext
	grammarPrefixes []string
	defaultPrefix   string
	UriContextBase
}

func NewGrammarUriContext(namespaceUriId int, namespaceUri string, grammarQNames []*QNameContext, grammarPrefixes []string) *GrammarUriContext {
	var defaultPrefix string

	switch namespaceUriId {
	case 0:
		defaultPrefix = ""
	case 1:
		defaultPrefix = "xml"
	case 2:
		defaultPrefix = "xsi"
	default:
		defaultPrefix = "ns" + strconv.FormatInt(int64(namespaceUriId), 10)
	}

	return &GrammarUriContext{
		grammarQNames:   grammarQNames,
		grammarPrefixes: grammarPrefixes,
		defaultPrefix:   defaultPrefix,
		UriContextBase: UriContextBase{
			namespaceUriId: namespaceUriId,
			namespaceUri:   namespaceUri,
		},
	}
}

func NewGrammarUriContextWithEmptyPrefixes(namespaceUriId int, namespaceUri string, grammarQNames []*QNameContext) *GrammarUriContext {
	return NewGrammarUriContext(namespaceUriId, namespaceUri, grammarQNames, []string{})
}

func (c *GrammarUriContext) GetDefaultPrefix() string {
	return c.defaultPrefix
}

func (c *GrammarUriContext) GetNumberOfQNames() int {
	return len(c.grammarQNames)
}

func (c *GrammarUriContext) GetQNameContextByLocalNameID(localNameId int) *QNameContext {
	if localNameId < len(c.grammarQNames) {
		return c.grammarQNames[localNameId]
	}
	return nil
}

func (c *GrammarUriContext) GetQNameContextByLocalName(localName string) *QNameContext {
	idx := slices.IndexFunc(c.grammarQNames, func(ctx *QNameContext) bool {
		return ctx.qName.Local == localName
	})
	if idx < 0 {
		return nil
	} else {
		return c.grammarQNames[idx]
	}
}

func (c *GrammarUriContext) GetNumberOfPrefixes() int {
	return len(c.grammarPrefixes)
}

func (c *GrammarUriContext) GetPrefix(prefixId int) *string {
	if prefixId < len(c.grammarPrefixes) {
		return &c.grammarPrefixes[prefixId]
	}
	return nil
}

func (c *GrammarUriContext) GetPrefixID(prefix string) int {
	for idx, p := range c.grammarPrefixes {
		if p == prefix {
			return idx
		}
	}
	return -1 //TODO: Introduce constant (Constants.NOT_FOUND)
}
