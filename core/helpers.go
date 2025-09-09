package core

import (
	"errors"
)

/*
	DefaultErrorHandler implementation
*/

type DefaultErrorHandler struct {
	ErrorHandler
}

func NewDefaultErrorHandler() *DefaultErrorHandler {
	return &DefaultErrorHandler{}
}

func (h *DefaultErrorHandler) Warning(err error) {}

func (h *DefaultErrorHandler) Error(err error) {}

/*
	DefaultEXIFactory implementation
*/

type DefaultEXIFactory struct {
	EXIFactory
	grammars                              Grammars
	isFragment                            bool
	codingMode                            CodingMode
	fidelityOptions                       *FidelityOptions
	encodingOptions                       *EncodingOptions
	decodingOptions                       *DecodingOptions
	schemaIDResolver                      SchemaIDResolver
	dtrMapTypes                           *[]QName
	dtrMapRepresentations                 *[]QName
	dtrMapRepresentationsDatatype         map[QName]Datatype
	scElements                            []QName
	scHandler                             SelfContainedHandler
	blockSize                             int
	valueMaxLength                        int
	valuePartitionCapacity                int
	localValuePartitions                  bool
	maximumNumberOfBuiltInElementGrammars int
	maximumNumberOfBuiltInProductions     int
	grammarLearningDisabled               bool
	sharedStrings                         []string
	isUsingNonEvolvingGrammrs             bool
	qnameSort                             func(q1, q2 QName) int
}

func NewDefaultEXIFactory() *DefaultEXIFactory {
	return &DefaultEXIFactory{
		grammars:                              NewSchemaLessGrammars(),
		isFragment:                            false,
		codingMode:                            CodingModeBitPacked,
		fidelityOptions:                       NewDefaultFidelityOptions(),
		encodingOptions:                       NewEncodingOptions(),
		decodingOptions:                       NewDecodingOptions(),
		schemaIDResolver:                      nil,
		dtrMapTypes:                           nil,
		dtrMapRepresentations:                 nil,
		dtrMapRepresentationsDatatype:         nil,
		scElements:                            []QName{},
		scHandler:                             nil,
		blockSize:                             DefaultBlockSize,
		valueMaxLength:                        DefaultValueMaxLength,
		valuePartitionCapacity:                DefaultValuePartitionCapacity,
		localValuePartitions:                  true,
		maximumNumberOfBuiltInElementGrammars: -1,
		maximumNumberOfBuiltInProductions:     -1,
		grammarLearningDisabled:               false,
		sharedStrings:                         []string{},
		isUsingNonEvolvingGrammrs:             false,
		qnameSort:                             QNameCompareFunc,
	}
}

func (f *DefaultEXIFactory) SetFidelityOptions(opts *FidelityOptions) {
	f.fidelityOptions = opts
}

func (f *DefaultEXIFactory) GetFidelityOptions() *FidelityOptions {
	return f.fidelityOptions
}

func (f *DefaultEXIFactory) SetEncodingOptions(opts *EncodingOptions) {
	f.encodingOptions = opts
}

func (f *DefaultEXIFactory) GetEncodingOptions() *EncodingOptions {
	return f.encodingOptions
}

func (f *DefaultEXIFactory) SetDecodingOptions(opts *DecodingOptions) {
	f.decodingOptions = opts
}

func (f *DefaultEXIFactory) GetDecodingOptions() *DecodingOptions {
	return f.decodingOptions
}

func (f *DefaultEXIFactory) SetSchemaIDResolver(resolver SchemaIDResolver) {
	f.schemaIDResolver = resolver
}

func (f *DefaultEXIFactory) GetSchemaIDResolver() SchemaIDResolver {
	return f.schemaIDResolver
}

func (f *DefaultEXIFactory) SetFragment(fragment bool) {
	f.isFragment = fragment
}

func (f *DefaultEXIFactory) IsFragment() bool {
	return f.isFragment
}

func (f *DefaultEXIFactory) SetGrammars(grammars Grammars) {
	if grammars == nil {
		panic("nil grammars")
	}
	f.grammars = grammars
}

func (f *DefaultEXIFactory) GetGrammars() Grammars {
	return f.grammars
}

func (f *DefaultEXIFactory) SetCodingMode(mode CodingMode) {
	f.codingMode = mode
}

func (f *DefaultEXIFactory) GetCodingMode() CodingMode {
	return f.codingMode
}

func (f *DefaultEXIFactory) SetBlockSize(size int) {
	if size < 0 {
		panic("negative block size")
	}
	f.blockSize = size
}

func (f *DefaultEXIFactory) GetBlockSize() int {
	return f.blockSize
}

func (f *DefaultEXIFactory) SetValueMaxLength(maxLength int) {
	f.valueMaxLength = maxLength
}

func (f *DefaultEXIFactory) GetValueMaxLength() int {
	return f.valueMaxLength
}

func (f *DefaultEXIFactory) SetValuePartitionCapacity(capacity int) {
	f.valuePartitionCapacity = capacity
}

func (f *DefaultEXIFactory) GetValuePartitionCapacity() int {
	return f.valuePartitionCapacity
}

func (f *DefaultEXIFactory) SetDatatypeRepresentationMap(dtpMapTypes *[]QName, dtrMapRepresentations *[]QName) {
	if dtpMapTypes == nil || dtrMapRepresentations == nil || len(*dtpMapTypes) != len(*dtrMapRepresentations) || len(*dtpMapTypes) == 0 {
		// un-set dtrMap
		f.dtrMapTypes = nil
		f.dtrMapRepresentations = nil
	} else {
		f.dtrMapTypes = dtpMapTypes
		f.dtrMapRepresentations = dtrMapRepresentations
	}
}

func (f *DefaultEXIFactory) RegisterDatatypeRepresentationMapDatatype(dtrMapRepresentation QName, datatype Datatype) Datatype {
	f.dtrMapRepresentationsDatatype = map[QName]Datatype{}
	prev := f.dtrMapRepresentationsDatatype[dtrMapRepresentation]
	f.dtrMapRepresentationsDatatype[dtrMapRepresentation] = datatype
	return prev
}

func (f *DefaultEXIFactory) GetDatatypeRepresentationMapTypes() *[]QName {
	return f.dtrMapTypes
}

func (f *DefaultEXIFactory) GetDatatypeRepresentationMapRepresentations() *[]QName {
	return f.dtrMapRepresentations
}

func (f *DefaultEXIFactory) SetSelfContainedElements(elements []QName) {
	f.SetSelfContainedElementsWithHandler(elements, nil)
}

func (f *DefaultEXIFactory) SetSelfContainedElementsWithHandler(elements []QName, handler SelfContainedHandler) {
	f.scElements = elements
	f.scHandler = handler
}

func (f *DefaultEXIFactory) IsSelfContainedElement(element QName) bool {
	for _, e := range f.scElements {
		if e == element {
			return true
		}
	}

	return false
}

func (f *DefaultEXIFactory) GetSelfContainedHandler() SelfContainedHandler {
	return f.scHandler
}

func (f *DefaultEXIFactory) SetLocalValuePartitions(lvp bool) {
	f.localValuePartitions = lvp
}

func (f *DefaultEXIFactory) IsLocalValuePartitions() bool {
	return f.localValuePartitions
}

func (f *DefaultEXIFactory) SetMaximumNumberOfBuiltInElementGrammars(num int) {
	if num >= 0 {
		f.maximumNumberOfBuiltInElementGrammars = num
	} else {
		f.maximumNumberOfBuiltInElementGrammars = -1
	}
}

func (f *DefaultEXIFactory) GetMaximumNumberOfBuiltInElementGrammars() int {
	return f.maximumNumberOfBuiltInElementGrammars
}

func (f *DefaultEXIFactory) SetMaximumNumberOfBuiltInProductions(num int) {
	if num >= 0 {
		f.maximumNumberOfBuiltInProductions = num
	} else {
		f.maximumNumberOfBuiltInProductions = -1
	}
}

func (f *DefaultEXIFactory) GetMaximumNumberOfBuiltInProductions() int {
	return f.maximumNumberOfBuiltInProductions
}

func (f *DefaultEXIFactory) IsGrammarLearningDisabled() bool {
	return f.grammarLearningDisabled
}

func (f *DefaultEXIFactory) SetSharedStrings(sharedStrings []string) {
	f.sharedStrings = sharedStrings
}

func (f *DefaultEXIFactory) GetSharedStrings() *[]string {
	return &f.sharedStrings
}

func (f *DefaultEXIFactory) SetUsingNonEvolvingGrammars(nonEvolving bool) {
	f.isUsingNonEvolvingGrammrs = nonEvolving
}

func (f *DefaultEXIFactory) IsUsingNonEvolvingGrammars() bool {
	return f.isUsingNonEvolvingGrammrs
}

func (f *DefaultEXIFactory) doSanityCheck() error {
	if f.fidelityOptions.IsFidelityEnabled(FeatureSC) && (f.codingMode == CodingModeCompression || f.codingMode == CodingModePreCompression) {
		return errors.New("(pre-)compression and selfContained elements cannot work together")
	}

	if !f.grammars.IsSchemaInformed() {
		f.maximumNumberOfBuiltInElementGrammars = -1
		f.maximumNumberOfBuiltInProductions = -1
		f.grammarLearningDisabled = false
	}

	if f.GetEncodingOptions().IsOptionEnabled(OptionCanonicalExi) {
		if err := f.updateFactoryAccordingCanonicalEXI(); err != nil {
			return err
		}
	}

	return nil
}

func (f *DefaultEXIFactory) CreateEXIBodyEncoder() (EXIBodyEncoder, error) {
	if err := f.doSanityCheck(); err != nil {
		return nil, err
	}

	if f.codingMode == CodingModeCompression || f.codingMode == CodingModePreCompression {
		return NewEXIBodyEncoderInOrderSC(f)
	} else {
		return NewEXIBodyEncoderInOrder(f)
	}
}

func (f *DefaultEXIFactory) CreateEXIStreamEncoder() (EXIStreamEncoder, error) {
	if err := f.doSanityCheck(); err != nil {
		return nil, err
	}

	return NewEXIStreamEncoderImpl(f)
}

func (f *DefaultEXIFactory) updateFactoryAccordingCanonicalEXI() error {
	// update canonical options according to canonical EXI rules

	// * A Canonical EXI Header MUST NOT begin with the optional EXI Cookie
	f.GetEncodingOptions().UnsetOption(OptionIncludeCookie)
	// * When the alignment option compression is set, pre-compress MUST be
	// used instead of compression.
	if f.GetCodingMode() == CodingModeCompression || f.GetCodingMode() == CodingModePreCompression {
		f.SetCodingMode(CodingModePreCompression)
	}
	// * datatypeRepresentationMap: the tuples are to be sorted
	// lexicographically according to the schema datatype first by {name}
	// then by {namespace}
	if f.dtrMapTypes != nil && len(*f.dtrMapTypes) > 0 {
		f.bubbleSort(f.dtrMapTypes, f.dtrMapRepresentations)
	}

	return nil
}

func (f *DefaultEXIFactory) bubbleSort(dtrMapTypes, dtrMapRepresentations *[]QName) {
	swapped := true
	j := 0
	var tmpType QName
	var tmpRep QName

	for swapped {
		swapped = false
		j++

		for i := 0; i < len(*dtrMapTypes)-j; i++ {
			if f.qnameSort((*dtrMapTypes)[i], (*dtrMapTypes)[i+1]) > 0 {
				tmpType = (*dtrMapTypes)[i]
				(*dtrMapTypes)[i] = (*dtrMapTypes)[i+1]
				(*dtrMapTypes)[i+1] = tmpType
				tmpRep = (*dtrMapRepresentations)[i]
				(*dtrMapRepresentations)[i] = (*dtrMapRepresentations)[i+1]
				(*dtrMapRepresentations)[i+1] = tmpRep
				swapped = true
			}
		}
	}
}

func (f *DefaultEXIFactory) CreateEXIBodyDecoder() (EXIBodyDecoder, error) {
	if err := f.doSanityCheck(); err != nil {
		return nil, err
	}

	if f.codingMode == CodingModeCompression || f.codingMode == CodingModePreCompression {
		//return NewEXIBodyDecoderReordered(f), nil
		return nil, errors.New("stream compression is not supported yet")
	} else {
		if f.fidelityOptions.IsFidelityEnabled(FeatureSC) {
			return NewEXIBodyDecoderInOrderSC(f)
		} else {
			return NewEXIBodyDecoderInOrder(f)
		}
	}
}

func (f *DefaultEXIFactory) CreateEXIStreamDecoder() (EXIStreamDecoder, error) {
	if err := f.doSanityCheck(); err != nil {
		return nil, err
	}

	return NewEXIStreamDecoderImpl(f)
}

func (f *DefaultEXIFactory) CreateStringEncoder() StringEncoder {
	var encoder StringEncoder
	if f.GetValueMaxLength() != DefaultValueMaxLength || f.GetValuePartitionCapacity() != DefaultValuePartitionCapacity {
		encoder = NewBoundedStringEncoderImpl(f.IsLocalValuePartitions(), f.GetValueMaxLength(), f.GetValuePartitionCapacity())
	} else {
		encoder = NewStringEncoderImpl(f.IsLocalValuePartitions())
	}

	return encoder
}

func (f *DefaultEXIFactory) isSchemaInformed() bool {
	return f.grammars.IsSchemaInformed()
}

func (f *DefaultEXIFactory) checkDtrMap() error {
	if f.dtrMapTypes == nil {
		f.dtrMapRepresentations = nil
	} else {
		if f.dtrMapRepresentations == nil || len(*f.dtrMapTypes) != len(*f.dtrMapRepresentations) {
			return errors.New("number of arguments for DTR map must match")
		}
	}
	return nil
}

func (f *DefaultEXIFactory) CreateTypeEncoder() (TypeEncoder, error) {
	if f.isSchemaInformed() {
		if err := f.checkDtrMap(); err != nil {
			return nil, err
		}

		if f.fidelityOptions.IsFidelityEnabled(FeatureLexicalValue) {
			return NewLexicalTypeEncoder(f.dtrMapTypes, f.dtrMapRepresentations, &f.dtrMapRepresentationsDatatype)
		} else {
			doNormalize := f.GetEncodingOptions().IsOptionEnabled(OptionUtcTime)
			return NewTypedTypeEncoderWithNormalize(f.dtrMapTypes, f.dtrMapRepresentations, &f.dtrMapRepresentationsDatatype, doNormalize)
		}
	} else {
		// use strings only
		return NewStringTypeEncoder()
	}
}

func (f *DefaultEXIFactory) CreateStringDecoder() StringDecoder {
	var decoder StringDecoder
	if f.GetValueMaxLength() != DefaultValueMaxLength || f.GetValuePartitionCapacity() != DefaultValuePartitionCapacity {
		decoder = NewBoundedStringDecoderImpl(f.IsLocalValuePartitions(), f.GetValueMaxLength(), f.GetValuePartitionCapacity())
	} else {
		decoder = NewStringDecoderImpl(f.IsLocalValuePartitions())
	}

	return decoder
}

func (f *DefaultEXIFactory) CreateTypeDecoder() (TypeDecoder, error) {
	if f.isSchemaInformed() {
		if err := f.checkDtrMap(); err != nil {
			return nil, err
		}

		if f.fidelityOptions.IsFidelityEnabled(FeatureLexicalValue) {
			return NewLexicalTypeDecoder(f.dtrMapTypes, f.dtrMapRepresentations, &f.dtrMapRepresentationsDatatype)
		} else {
			return NewTypedTypeDecoder(f.dtrMapTypes, f.dtrMapRepresentations, &f.dtrMapRepresentationsDatatype)
		}
	} else {
		// use strings only
		return NewStringTypeDecoder()
	}
}

func (f *DefaultEXIFactory) Clone() EXIFactory {
	//TODO: Deep copy?
	z := *f
	return &z
}
