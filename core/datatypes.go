package core

import (
	"fmt"

	"github.com/sderkacs/go-exi/utils"
)

type DatatypeID int
type WhiteSpace int

const (
	DataTypeID_EXI_Base64Binary DatatypeID = iota
	DataTypeID_EXI_HexBinary
	DataTypeID_EXI_Boolean
	DataTypeID_EXI_DateTime
	DataTypeID_EXI_Time
	DataTypeID_EXI_Date
	DataTypeID_EXI_GYearMonth
	DataTypeID_EXI_GYear
	DataTypeID_EXI_GMonthDay
	DataTypeID_EXI_GDay
	DataTypeID_EXI_GMonth
	DataTypeID_EXI_Decimal
	DataTypeID_EXI_Double
	DataTypeID_EXI_Integer
	DataTypeID_EXI_String
	DataTypeID_EXI_EString

	WhiteSpacePreserve WhiteSpace = iota
	WhiteSpaceReplace
	WhiteSpaceCollapse
)

type Datatype interface {
	GetBuiltInType() BuiltInType
	GetSchemaType() *QNameContext
	GetBaseDatatype() Datatype
	SetBaseDatatype(datatype Datatype)
	SetGrammarEnumeration(enum EnumDatatype)
	GetGrammarEnumeration() EnumDatatype
	GetWhiteSpace() WhiteSpace
	GetDatatypeID() DatatypeID
	Equals(o Datatype) bool
}

type EnumDatatype interface {
	Datatype
	GetCodingLength() int
	GetEnumerationSize() int
	GetEnumValue(i int) Value
}

/*
	AbstractDatatype implementation
*/

type AbstractDatatype struct {
	Datatype
	builtInType        BuiltInType
	schemaType         *QNameContext
	baseDatatype       Datatype
	grammarEnumeration EnumDatatype
	whiteSpace         WhiteSpace
}

func NewAbstractDatatype(builtInType BuiltInType, schemaType *QNameContext) *AbstractDatatype {
	return &AbstractDatatype{
		builtInType: builtInType,
		schemaType:  schemaType,
		whiteSpace:  WhiteSpaceCollapse,
	}
}

func NewAbstractDatatypeWithWhiteSpace(builtInType BuiltInType, schemaType *QNameContext, whiteSpace WhiteSpace) *AbstractDatatype {
	return &AbstractDatatype{
		builtInType: builtInType,
		schemaType:  schemaType,
		whiteSpace:  whiteSpace,
	}
}

func (d *AbstractDatatype) GetBuiltInType() BuiltInType {
	return d.builtInType
}

func (d *AbstractDatatype) GetSchemaType() *QNameContext {
	return d.schemaType
}

func (d *AbstractDatatype) GetBaseDatatype() Datatype {
	return d.baseDatatype
}

func (d *AbstractDatatype) SetBaseDatatype(datatype Datatype) {
	d.baseDatatype = datatype
}

func (d *AbstractDatatype) SetGrammarEnumeration(enum EnumDatatype) {
	d.grammarEnumeration = enum
}

func (d *AbstractDatatype) GetGrammarEnumeration() EnumDatatype {
	return d.grammarEnumeration
}

func (d *AbstractDatatype) GetWhiteSpace() WhiteSpace {
	return d.whiteSpace
}

func (d *AbstractDatatype) Equals(o Datatype) bool {
	if d.builtInType == o.GetBuiltInType() {
		if d.schemaType == nil {
			return o.GetSchemaType() == nil
		} else {
			return d.schemaType.Equals(o.GetSchemaType())
		}
	}

	return false
}

/*
	AbstractBinaryDatatype implementation
*/

type AbstractBinaryDatatype struct {
	*AbstractDatatype
}

func NewAbstractBinaryDatatype(binaryType BuiltInType, schemaType *QNameContext) *AbstractBinaryDatatype {
	return &AbstractBinaryDatatype{
		AbstractDatatype: NewAbstractDatatype(binaryType, schemaType),
	}
}

/*
	BinaryBase64Datatype implementation
*/

type BinaryBase64Datatype struct {
	*AbstractBinaryDatatype
}

func NewBinaryBase64Datatype(schemaType *QNameContext) *BinaryBase64Datatype {
	return &BinaryBase64Datatype{
		AbstractBinaryDatatype: NewAbstractBinaryDatatype(BuiltInTypeBinaryBase64, schemaType),
	}
}

func (dt *BinaryBase64Datatype) GetDatatypeID() DatatypeID {
	return DataTypeID_EXI_Base64Binary
}

/*
	BinaryHexDatatype implementation
*/

type BinaryHexDatatype struct {
	*AbstractBinaryDatatype
}

func NewBinaryHexDatatype(schemaType *QNameContext) *BinaryHexDatatype {
	return &BinaryHexDatatype{
		AbstractBinaryDatatype: NewAbstractBinaryDatatype(BuiltInTypeBinaryHex, schemaType),
	}
}

func (dt *BinaryHexDatatype) GetDatatypeID() DatatypeID {
	return DataTypeID_EXI_HexBinary
}

/*
	BooleanDatatype implementation
*/

type BooleanDatatype struct {
	*AbstractDatatype
}

func NewBooleanDatatype(schemaType *QNameContext) *BooleanDatatype {
	return &BooleanDatatype{
		AbstractDatatype: NewAbstractDatatype(BuiltInTypeBoolean, schemaType),
	}
}

func (dt *BooleanDatatype) GetDatatypeID() DatatypeID {
	return DataTypeID_EXI_Boolean
}

/*
	BooleanFacetDatatype implementation
*/

type BooleanFacetDatatype struct {
	*AbstractDatatype
}

func NewBooleanFacetDatatype(schemaType *QNameContext) *BooleanFacetDatatype {
	return &BooleanFacetDatatype{
		AbstractDatatype: NewAbstractDatatype(BuiltInTypeBooleanFacet, schemaType),
	}
}

func (dt *BooleanFacetDatatype) GetDatatypeID() DatatypeID {
	return DataTypeID_EXI_Boolean
}

/*
	DatetimeDatatype implementation
*/

type DatetimeDatatype struct {
	*AbstractDatatype
	dateType          DateTimeType
	lastValidDateTime *DateTimeValue
}

func NewDatetimeDatatype(dateType DateTimeType, schemaType *QNameContext) *DatetimeDatatype {
	return &DatetimeDatatype{
		AbstractDatatype: NewAbstractDatatype(BuiltInTypeDateTime, schemaType),
		dateType:         dateType,
	}
}

func (dt *DatetimeDatatype) GetDatatypeID() DatatypeID {
	switch dt.dateType {
	case DateTimeDateTime:
		return DataTypeID_EXI_DateTime
	case DateTimeTime:
		return DataTypeID_EXI_Time
	case DateTimeDate:
		return DataTypeID_EXI_Date
	case DateTimeGYearMonth:
		return DataTypeID_EXI_GMonthDay
	case DateTimeGYear:
		return DataTypeID_EXI_GYear
	case DateTimeGMonthDay:
		return DataTypeID_EXI_GMonthDay
	case DateTimeGDay:
		return DataTypeID_EXI_GDay
	case DateTimeGMonth:
		return DataTypeID_EXI_GMonth
	default:
		//TODO: Avoid panics
		panic(fmt.Sprintf("unsupported date time type: %d", dt.dateType))
	}
}

func (dt *DatetimeDatatype) GetDatetimeType() DateTimeType {
	return dt.dateType
}

func (dt *DatetimeDatatype) isValidString(value string) bool {
	d, err := DateTimeParse(value, dt.dateType)
	if err != nil {
		return false
	}
	dt.lastValidDateTime = d
	return true
}

func (dt *DatetimeDatatype) IsValid(value Value) (bool, error) {
	dateTime, ok := value.(*DateTimeValue)
	if ok {
		dt.lastValidDateTime = dateTime
		return true, nil
	} else {
		s, err := value.ToString()
		if err != nil {
			return false, err
		}
		return dt.isValidString(s), nil
	}
}

/*
	DecimalDatatype implementation
*/

type DecimalDatatype struct {
	*AbstractDatatype
}

func NewDecimalDatatype(schemaType *QNameContext) *DecimalDatatype {
	return &DecimalDatatype{
		AbstractDatatype: NewAbstractDatatype(BuiltInTypeDecimal, schemaType),
	}
}

func (dt *DecimalDatatype) GetDatatypeID() DatatypeID {
	return DataTypeID_EXI_Decimal
}

/*
	EnumerationDatatype implementation
*/

type EnumerationDatatype struct {
	*AbstractDatatype
	dtEnumValues Datatype
	codingLength int
	enumValues   []Value
}

func NewEnumerationDatatype(enumValues []Value, dtEnumValues Datatype, schemaType *QNameContext) *EnumerationDatatype {
	if dtEnumValues.GetBuiltInType() != BuiltInTypeQName && dtEnumValues.GetBuiltInType() != BuiltInTypeEnumeration {
		return &EnumerationDatatype{
			AbstractDatatype: NewAbstractDatatype(BuiltInTypeEnumeration, schemaType),
			dtEnumValues:     dtEnumValues,
			enumValues:       enumValues,
			codingLength:     utils.GetCodingLength(len(enumValues)),
		}
	} else {
		panic(fmt.Errorf("enumeration type values can't be of type Enumeration or QName"))
	}
}

func NewEnumerationDatatypeChecked(enumValues []Value, dtEnumValues Datatype, schemaType *QNameContext) (*EnumerationDatatype, error) {
	if dtEnumValues.GetBuiltInType() != BuiltInTypeQName && dtEnumValues.GetBuiltInType() != BuiltInTypeEnumeration {
		return &EnumerationDatatype{
			AbstractDatatype: NewAbstractDatatype(BuiltInTypeEnumeration, schemaType),
			dtEnumValues:     dtEnumValues,
			enumValues:       enumValues,
			codingLength:     utils.GetCodingLength(len(enumValues)),
		}, nil
	} else {
		return nil, fmt.Errorf("enumeration type values can't be of type Enumeration or QName")
	}
}

func (dt *EnumerationDatatype) GetEnumValueDatatype() Datatype {
	return dt.dtEnumValues
}

func (dt *EnumerationDatatype) GetDatatypeID() DatatypeID {
	return dt.dtEnumValues.GetDatatypeID()
}

func (dt *EnumerationDatatype) GetEnumerationSize() int {
	return len(dt.enumValues)
}

func (dt *EnumerationDatatype) GetCodingLength() int {
	return dt.codingLength
}

func (dt *EnumerationDatatype) GetEnumValue(idx int) Value {
	if idx < len(dt.enumValues)-1 {
		return dt.enumValues[idx]
	}
	return nil
}

/*
	ExtendedStringDatatype implementation
*/

type ExtendedStringDatatype struct {
	*AbstractDatatype
	lastValue      *string
	sharedStrings  []string
	grammarStrings EnumDatatype
}

func NewExtendedStringDatatype(schemaType *QNameContext) *ExtendedStringDatatype {
	return &ExtendedStringDatatype{
		AbstractDatatype: NewAbstractDatatypeWithWhiteSpace(BuiltInTypeExtendedString, schemaType, WhiteSpacePreserve),
		lastValue:        nil,
		sharedStrings:    []string{},
		grammarStrings:   nil,
	}
}

func NewExtendedStringDatatypeWithWhiteSpace(schemaType *QNameContext, whiteSpace WhiteSpace) *ExtendedStringDatatype {
	return &ExtendedStringDatatype{
		AbstractDatatype: NewAbstractDatatypeWithWhiteSpace(BuiltInTypeExtendedString, schemaType, whiteSpace),
		lastValue:        nil,
		sharedStrings:    []string{},
		grammarStrings:   nil,
	}
}

func (dt *ExtendedStringDatatype) GetDatatypeID() DatatypeID {
	return DataTypeID_EXI_EString
}

func (dt *ExtendedStringDatatype) SetSharedStrings(sharedStrings []string) {
	dt.sharedStrings = sharedStrings
}

func (dt *ExtendedStringDatatype) SetGrammarStrings(grammarStrings EnumDatatype) {
	dt.grammarStrings = grammarStrings
}

func (dt *ExtendedStringDatatype) GetGrammarStrings() EnumDatatype {
	return dt.grammarStrings
}

/*
	FloatDatatype implementation
*/

type FloatDatatype struct {
	*AbstractDatatype
}

func NewFloatDatatype(schemaType *QNameContext) *FloatDatatype {
	return &FloatDatatype{
		AbstractDatatype: NewAbstractDatatype(BuiltInTypeFloat, schemaType),
	}
}

func (dt *FloatDatatype) GetDatatypeID() DatatypeID {
	return DataTypeID_EXI_Double
}

/*
	IntegerDatatype implementation
*/

type IntegerDatatype struct {
	*AbstractDatatype
}

func NewIntegerDatatype(schemaType *QNameContext) *IntegerDatatype {
	return &IntegerDatatype{
		AbstractDatatype: NewAbstractDatatype(BuiltInTypeInteger, schemaType),
	}
}

func (dt *IntegerDatatype) GetDatatypeID() DatatypeID {
	return DataTypeID_EXI_Integer
}

/*
	ListDatatype implementation
*/

type ListDatatype struct {
	*AbstractDatatype
	listDatatype Datatype
}

func NewListDatatype(listDatatype Datatype, schemaType *QNameContext) *ListDatatype {
	if listDatatype.GetBuiltInType() == BuiltInTypeList {
		panic(fmt.Errorf("list type values can't be of type List"))
	}
	return &ListDatatype{
		AbstractDatatype: NewAbstractDatatype(BuiltInTypeList, schemaType),
		listDatatype:     listDatatype,
	}
}

func NewListDatatypeChecked(listDatatype Datatype, schemaType *QNameContext) (*ListDatatype, error) {
	if listDatatype.GetBuiltInType() == BuiltInTypeList {
		return nil, fmt.Errorf("list type values can't be of type List")
	}
	return &ListDatatype{
		AbstractDatatype: NewAbstractDatatype(BuiltInTypeList, schemaType),
		listDatatype:     listDatatype,
	}, nil
}

func (dt *ListDatatype) GetDatatypeID() DatatypeID {
	return dt.listDatatype.GetDatatypeID()
}

func (dt *ListDatatype) GetListDatatype() Datatype {
	return dt.listDatatype
}

/*
	NBitUnsignedIntegerDatatype implementation
*/

type NBitUnsignedIntegerDatatype struct {
	*AbstractDatatype
	lowerBound         *IntegerValue
	upperBound         *IntegerValue
	numberOfBits4Range int
}

func NewNBitUnsignedIntegerDatatype(lowerBound *IntegerValue, upperBound *IntegerValue, schemaType *QNameContext) *NBitUnsignedIntegerDatatype {
	diff := upperBound.Sub(lowerBound)

	return &NBitUnsignedIntegerDatatype{
		AbstractDatatype:   NewAbstractDatatype(BuiltInTypeNBitUnsignedInteger, schemaType),
		lowerBound:         lowerBound,
		upperBound:         upperBound,
		numberOfBits4Range: utils.GetCodingLength(diff.Value32() + 1),
	}
}

func (dt *NBitUnsignedIntegerDatatype) GetDatatypeID() DatatypeID {
	return DataTypeID_EXI_Integer
}

func (dt *NBitUnsignedIntegerDatatype) GetLowerBound() *IntegerValue {
	return dt.lowerBound
}

func (dt *NBitUnsignedIntegerDatatype) GetUpperBound() *IntegerValue {
	return dt.upperBound
}

func (dt *NBitUnsignedIntegerDatatype) GetNumberOfBits() int {
	return dt.numberOfBits4Range
}

/*
	RestrictedCharacterSetDatatype implementation
*/

type RestrictedCharacterSetDatatype struct {
	*AbstractDatatype
	rcs RestrictedCharacterSet
}

func NewRestrictedCharacterSetDatatype(rcs RestrictedCharacterSet, schemaType *QNameContext) *RestrictedCharacterSetDatatype {
	return &RestrictedCharacterSetDatatype{
		AbstractDatatype: NewAbstractDatatypeWithWhiteSpace(BuiltInTypeRcsString, schemaType, WhiteSpacePreserve),
		rcs:              rcs,
	}
}

func NewRestrictedCharacterSetDatatypeWithWhiteSpace(rcs RestrictedCharacterSet, schemaType *QNameContext, whiteSpace WhiteSpace) *RestrictedCharacterSetDatatype {
	return &RestrictedCharacterSetDatatype{
		AbstractDatatype: NewAbstractDatatypeWithWhiteSpace(BuiltInTypeRcsString, schemaType, whiteSpace),
		rcs:              rcs,
	}
}

func (dt *RestrictedCharacterSetDatatype) GetDatatypeID() DatatypeID {
	return DataTypeID_EXI_String
}

func (dt *RestrictedCharacterSetDatatype) GetRestrictedCharacterSet() RestrictedCharacterSet {
	return dt.rcs
}

/*
	StringDatatype implementation
*/

type StringDatatype struct {
	*AbstractDatatype
	isDerivedByUnion bool
}

func NewStringDatatypeWithDerive(schemaType *QNameContext, isDerivedByUnion bool) *StringDatatype {
	return &StringDatatype{
		AbstractDatatype: NewAbstractDatatype(BuiltInTypeString, schemaType),
		isDerivedByUnion: isDerivedByUnion,
	}
}

func NewStringDatatypeWithWhiteSpace(schemaType *QNameContext, whiteSpace WhiteSpace) *StringDatatype {
	dt := NewStringDatatypeWithDerive(schemaType, false)
	dt.whiteSpace = whiteSpace
	return dt
}

func NewStringDatatype(schemaType *QNameContext) *StringDatatype {
	return NewStringDatatypeWithWhiteSpace(schemaType, WhiteSpacePreserve)
}

func (d *StringDatatype) GetDatatypeID() DatatypeID {
	return DataTypeID_EXI_String
}

func (d *StringDatatype) IsDerivedByUnion() bool {
	return d.isDerivedByUnion
}

/*
	UnsignedIntegerDatatype implementation
*/

type UnsignedIntegerDatatype struct {
	*AbstractDatatype
}

func NewUnsignedIntegerDatatype(schemaType *QNameContext) *UnsignedIntegerDatatype {
	return &UnsignedIntegerDatatype{
		AbstractDatatype: NewAbstractDatatype(BuiltInTypeUnsignedInteger, schemaType),
	}
}

func (dt *UnsignedIntegerDatatype) GetDatatypeID() DatatypeID {
	return DataTypeID_EXI_Integer
}
