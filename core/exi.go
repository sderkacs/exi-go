package core

import "github.com/sderkacs/go-exi/utils"

type CodingMode int

const (
	CodingModeBitPacked CodingMode = iota
	CodingModeBytePacked
	CodingModePreCompression
	CodingModeCompression
)

type SchemaIDResolver interface {
	ResolveSchemaID(schemaID string) (Grammars, error)
}

type SelfContainedHandler interface {
	ScElement(uri, localName *string, channel EncoderChannel) error
}

type ErrorHandler interface {
	Warning(err error)
	Error(err error)
}

type EXIFactory interface {
	// Sets the fidelity options used by the EXI factory (e.g. preserving XML
	// comments or DTDs).
	SetFidelityOptions(opts *FidelityOptions)

	// Returns the fidelity options used by the EXI factory (e.g. preserving XML
	// comments or DTDs).
	GetFidelityOptions() *FidelityOptions

	// Sets the header options used by the EXI Encoder(e.g., include EXI Cookie,
	// EXI Options document).
	SetEncodingOptions(opts *EncodingOptions)

	// Returns the header options used by the EXI encoder.
	GetEncodingOptions() *EncodingOptions

	// Sets the options used by the EXI Decoder(e.g., ignore schemaId).
	SetDecodingOptions(opts *DecodingOptions)

	// Returns the options used by the EXI decoder.
	GetDecodingOptions() *DecodingOptions

	// Sets specific schemaID resolver.
	SetSchemaIDResolver(resolver SchemaIDResolver)

	// Returns schemaID resolver for this factory.
	GetSchemaIDResolver() SchemaIDResolver

	// Informs the factory that we are dealing with an XML fragment instead of
	// an XML document.
	SetFragment(fragment bool)

	// Returns whether we deal with a fragment.
	IsFragment() bool

	// Sets the EXI <code>Grammars</code> used for coding.
	SetGrammars(grammars Grammars)

	// Returns the currently used EXI <code>Grammars</code>. By default a
	// <code>SchemaLessGrammars</code> is used.
	GetGrammars() Grammars

	// Re-sets the coding mode used by the factory.
	SetCodingMode(mode CodingMode)

	// Returns the currently used <code>CodingMode</code>. By default BIT_PACKED
	// is used.
	GetCodingMode() CodingMode

	// The default blockSize is intentionally large (1,000,000) but can be
	// reduced for processing large documents on devices with limited memory.
	SetBlockSize(size int)

	// The blockSize option specifies the block size used for EXI compression.
	// When the "blockSize" element is absent in the EXI Options document, the
	// default blocksize of 1,000,000 is used.
	GetBlockSize() int

	// The valueMaxLength option specifies the maximum length of value content
	// items to be considered for addition to the string table. The default
	// value "unbounded" is assumed when the "valueMaxLength" element is absent
	// in the EXI Options document.
	//
	// See http://www.w3.org/TR/exi/#key-valueMaxLengthOption
	SetValueMaxLength(maxLength int)

	// The default value "unbounded" is assumed when the "valueMaxLength"
	// element is absent. Return value OR negative for unbounded.
	GetValueMaxLength() int

	// The valuePartitionCapacity option specifies the maximum number of value
	// content items in the string table at any given time. The default value
	// "unbounded" is assumed when the "valuePartitionCapacity" element is absent.
	//
	// See http://www.w3.org/TR/exi/#key-valuePartitionCapacityOption
	SetValuePartitionCapacity(capacity int)

	// The default value "unbounded" is assumed when the "valuePartitionCapacity"
	// element is absent. Return value OR negative for unbounded.
	GetValuePartitionCapacity() int

	// By default, each typed value in an EXI stream is represented by the
	// associated built-in EXI datatype representation. However, EXI processors
	// MAY provide the capability to specify different built-in EXI datatype
	// representations or user-defined datatype representations for representing
	// specific schema datatypes. This capability is called Datatype
	// Representation Map.
	SetDatatypeRepresentationMap(dtpMapTypes *[]utils.QName, dtrMapRepresentations *[]utils.QName)

	// The DTR map representation may use built-in String datatypes (e.g.,
	// <code>exi:string</code>) or use user-defined type representations. This
	// method allows to register the datatype that should be used.
	RegisterDatatypeRepresentationMapDatatype(dtrMapRepresentation utils.QName, datatype Datatype) Datatype

	// EXI processors MAY provide the capability to specify different built-in
	// EXI datatype representations or user-defined datatype representations for
	// representing specific schema datatypes.
	GetDatatypeRepresentationMapTypes() *[]utils.QName

	// EXI processors MAY provide the capability to specify different built-in
	// EXI datatype representations or user-defined datatype representations for
	// representing specific schema datatypes.
	GetDatatypeRepresentationMapRepresentations() *[]utils.QName

	// Self-contained elements may be read independently from the rest of the
	// EXI body, allowing them to be indexed for random access. The
	// "selfContained" element MUST NOT appear in an EXI options document when
	// one of "compression", "pre-compression" or "strict" elements are present
	// in the same options document.
	SetSelfContainedElements(elements []utils.QName)

	// Self-contained elements may be read independently from the rest of the
	// EXI body, allowing them to be indexed for random access. The
	// "selfContained" element MUST NOT appear in an EXI options document when
	// one of "compression", "pre-compression" or "strict" elements are present
	// in the same options document.
	SetSelfContainedElementsWithHandler(elements []utils.QName, handler SelfContainedHandler)

	// Returns boolean value telling whether a certain element is encoded as
	// selfContained fragment.
	IsSelfContainedElement(element utils.QName) bool

	// Returns selfContained element handler.
	GetSelfContainedHandler() SelfContainedHandler

	// The EXI profile defines a parameter that can disable the use of local
	// value references. Global value indexing may be controlled using the
	// options defined in the EXI 1.0 specification
	//
	// The localValuePartitions option of the EXI profile is a Boolean used to
	// indicate whether local value partitions are used. The value "0"
	// indicates that no local value partition is used while "1" represents the
	// behavior of the EXI 1.0 specification
	SetLocalValuePartitions(lvp bool)

	// The localValuePartitions option of the EXI profile is a Boolean used to
	// indicate whether local value partitions are used. The value "0"
	// indicates that no local value partition is used while "1" represents the
	// behavior of the EXI 1.0 specification
	IsLocalValuePartitions() bool

	// The EXI profile defines a parameter that restricts the maximum number of
	// elements for which evolving built-in element grammars can be
	// instantiated.
	//
	// The value "unbounded" (-1) indicates that no restrictions are used and
	// represents the behavior of the EXI 1.0 specification
	SetMaximumNumberOfBuiltInElementGrammars(num int)

	// The EXI profile defines a parameter that restricts the maximum number of
	// elements for which evolving built-in element grammars can be
	// instantiated.
	GetMaximumNumberOfBuiltInElementGrammars() int

	// The EXI profile defines a parameter that restricts the maximum number of
	// top-level productions that can be dynamically inserted in built-in
	// element grammars.
	//
	// The value "unbounded" (-1) indicates that no restrictions are used and
	// represents the behavior of the EXI 1.0 specification
	SetMaximumNumberOfBuiltInProductions(num int)

	// The EXI profile defines a parameter that restricts the maximum number of
	// top-level productions that can be dynamically inserted in built-in
	// element grammars.
	GetMaximumNumberOfBuiltInProductions() int

	// The EXI profile defines parameters that restrict grammar learning. This
	// is a convenience method to indicate whether grammar restriction is in
	// use.
	IsGrammarLearningDisabled() bool

	// (Experimental) Feature to pre-agree on shared strings.
	SetSharedStrings(sharedStrings []string)

	// (Experimental) Return list of shared strings.
	GetSharedStrings() *[]string

	// (Experimental) Feature which dictates that grammar does not grow in any
	// circumstance.
	SetUsingNonEvolvingGrammars(nonEvolving bool)

	// (Experimental) Returns whether non-evolving grammars are used.
	IsUsingNonEvolvingGrammars() bool

	// Returns an <code>EXIBodyEncoder</code>.
	CreateEXIBodyEncoder() (EXIBodyEncoder, error)

	// Returns an <code>EXIStreamEncoder</code>.
	CreateEXIStreamEncoder() (EXIStreamEncoder, error)

	// Returns an <code>EXIBodyDecoder</code>.
	CreateEXIBodyDecoder() (EXIBodyDecoder, error)

	// Returns an <code>EXIStreamDecoder</code>.
	CreateEXIStreamDecoder() (EXIStreamDecoder, error)

	//Returns an EXI <code>StringEncoder</code> according coding options.
	CreateStringEncoder() StringEncoder

	// Returns an EXI <code>TypeEncoder</code> according coding options such as
	// schema-informed or schema-less grammar and options like
	// PreserveLexicalValues.
	CreateTypeEncoder() (TypeEncoder, error)

	// Returns an EXI {@link StringDecoder} according coding options.
	CreateStringDecoder() StringDecoder

	// Returns an EXI <code>TypeDecoder</code> according coding options such as
	// schema-informed or schema-less grammar and options like
	// PreserveLexicalValues.
	CreateTypeDecoder() (TypeDecoder, error)

	// Returns a shallow copy of this EXI factory.
	Clone() EXIFactory
}
