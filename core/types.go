package core

import (
	"fmt"
	"strings"

	"github.com/sderkacs/exi-go/utils"
)

type BuiltInType int
type DateTimeType int

const (
	BuiltInTypeBinaryBase64 BuiltInType = iota
	BuiltInTypeBinaryHex
	BuiltInTypeBoolean
	BuiltInTypeBooleanFacet
	BuiltInTypeDecimal
	BuiltInTypeFloat
	BuiltInTypeNBitUnsignedInteger
	BuiltInTypeUnsignedInteger
	BuiltInTypeInteger
	BuiltInTypeDateTime
	BuiltInTypeString
	BuiltInTypeRcsString
	BuiltInTypeExtendedString
	BuiltInTypeEnumeration
	BuiltInTypeList
	BuiltInTypeQName

	DateTimeGYear DateTimeType = iota
	DateTimeGYearMonth
	DateTimeDate
	DateTimeDateTime
	DateTimeGMonth
	DateTimeGMonthDay
	DateTimeGDay
	DateTimeTime
)

var (
	XsdBase64Binary       utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "base64Binary"}
	XsdHexBinary          utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "hexBinary"}
	XsdBoolean            utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "boolean"}
	XsdDateTime           utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "dateTime"}
	XsdTime               utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "time"}
	XsdDate               utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "date"}
	XsdGYearMonth         utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "gYearMonth"}
	XsdGYear              utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "gYear"}
	XsdGMonthDay          utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "gMonthDay"}
	XsdGDay               utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "gDay"}
	XsdGMonth             utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "gMonth"}
	XsdDecimal            utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "decimal"}
	XsdFloat              utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "float"}
	XsdDouble             utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "double"}
	XsdInteger            utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "integer"}
	XsdNonNegativeInteger utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "nonNegativeInteger"}
	XsdString             utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "string"}
	XsdExtendedString     utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "estring"}
	XsdAnySimpleType      utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "anySimpleType"}
	XsdQName              utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "QName"}
	XsdNotation           utils.QName = utils.QName{Space: XMLSchemaNS_URI, Local: "Notation"}
	DefaultValueName      utils.QName = XsdString

	defaultDatatype Datatype = NewStringDatatype(NewQNameContext(-1, -1, utils.QName{}))
)

func BuiltInGetDefaultDatatype() Datatype {
	return defaultDatatype
}

type TypeCoder interface{}

type TypeEncoder interface {
	TypeCoder
	IsValid(datatype Datatype, value Value) (bool, error)
	WriteValue(qnc *QNameContext, channel EncoderChannel, encoder StringEncoder) error
}

type TypeDecoder interface {
	ReadValue(datatype Datatype, qnc *QNameContext, channel DecoderChannel, decoder StringDecoder) (Value, error)
}

/*
	AbstractTypeCoder implementation
*/

type AbstractTypeCoder struct {
	TypeCoder
	dtrMapTypes                  *[]utils.QName
	dtrMapRepresentations        *[]utils.QName
	dtrMapRepresentationDatatype *map[utils.QName]Datatype
	dtrMap                       map[utils.QName]Datatype
	dtrMapInUse                  bool
}

func NewAbstractTypeCoder(
	dtrMapTypes *[]utils.QName,
	dtrMapRepresentations *[]utils.QName,
	dtrMapRepresentationDatatype *map[utils.QName]Datatype,
) (*AbstractTypeCoder, error) {
	c := &AbstractTypeCoder{
		dtrMapTypes:                  dtrMapTypes,
		dtrMapRepresentations:        dtrMapRepresentations,
		dtrMapRepresentationDatatype: dtrMapRepresentationDatatype,
		dtrMap:                       map[utils.QName]Datatype{},
		dtrMapInUse:                  (dtrMapTypes != nil),
	}
	if dtrMapTypes != nil {
		if err := c.initDtrMaps(); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *AbstractTypeCoder) initDtrMaps() error {
	if !c.dtrMapInUse {
		return fmt.Errorf("DTR map is not used")
	}
	var err error

	c.dtrMap[XsdBase64Binary], err = c.getDatatypeRepresentation(W3C_EXI_NS_URI, W3C_EXI_LN_Base64Binary)
	if err != nil {
		return err
	}
	c.dtrMap[XsdHexBinary], err = c.getDatatypeRepresentation(W3C_EXI_NS_URI, W3C_EXI_LN_HexBinary)
	if err != nil {
		return err
	}
	c.dtrMap[XsdBoolean], err = c.getDatatypeRepresentation(W3C_EXI_NS_URI, W3C_EXI_LN_Boolean)
	if err != nil {
		return err
	}
	c.dtrMap[XsdDateTime], err = c.getDatatypeRepresentation(W3C_EXI_NS_URI, W3C_EXI_LN_DateTime)
	if err != nil {
		return err
	}
	c.dtrMap[XsdTime], err = c.getDatatypeRepresentation(W3C_EXI_NS_URI, W3C_EXI_LN_Time)
	if err != nil {
		return err
	}
	c.dtrMap[XsdDate], err = c.getDatatypeRepresentation(W3C_EXI_NS_URI, W3C_EXI_LN_Date)
	if err != nil {
		return err
	}
	c.dtrMap[XsdGYearMonth], err = c.getDatatypeRepresentation(W3C_EXI_NS_URI, W3C_EXI_LN_GYearMonth)
	if err != nil {
		return err
	}
	c.dtrMap[XsdGYear], err = c.getDatatypeRepresentation(W3C_EXI_NS_URI, W3C_EXI_LN_GYear)
	if err != nil {
		return err
	}
	c.dtrMap[XsdGMonthDay], err = c.getDatatypeRepresentation(W3C_EXI_NS_URI, W3C_EXI_LN_GMonthDay)
	if err != nil {
		return err
	}
	c.dtrMap[XsdGDay], err = c.getDatatypeRepresentation(W3C_EXI_NS_URI, W3C_EXI_LN_GDay)
	if err != nil {
		return err
	}
	c.dtrMap[XsdGMonth], err = c.getDatatypeRepresentation(W3C_EXI_NS_URI, W3C_EXI_LN_GMonth)
	if err != nil {
		return err
	}
	c.dtrMap[XsdDecimal], err = c.getDatatypeRepresentation(W3C_EXI_NS_URI, W3C_EXI_LN_Decimal)
	if err != nil {
		return err
	}
	c.dtrMap[XsdFloat], err = c.getDatatypeRepresentation(W3C_EXI_NS_URI, W3C_EXI_LN_Double)
	if err != nil {
		return err
	}
	c.dtrMap[XsdDouble], err = c.getDatatypeRepresentation(W3C_EXI_NS_URI, W3C_EXI_LN_Double)
	if err != nil {
		return err
	}
	c.dtrMap[XsdInteger], err = c.getDatatypeRepresentation(W3C_EXI_NS_URI, W3C_EXI_LN_Integer)
	if err != nil {
		return err
	}
	c.dtrMap[XsdString], err = c.getDatatypeRepresentation(W3C_EXI_NS_URI, W3C_EXI_LN_String)
	if err != nil {
		return err
	}
	c.dtrMap[XsdAnySimpleType], err = c.getDatatypeRepresentation(W3C_EXI_NS_URI, W3C_EXI_LN_String)
	if err != nil {
		return err
	}

	for i := 0; i < len(*c.dtrMapTypes); i++ {
		dtrMapRepr := (*c.dtrMapRepresentations)[i]
		representation, err := c.getDatatypeRepresentation(dtrMapRepr.Space, dtrMapRepr.Local)
		if err != nil {
			return err
		}
		kind := (*c.dtrMapTypes)[i]
		c.dtrMap[kind] = representation
	}

	return nil
}

func (c *AbstractTypeCoder) getDatatypeRepresentation(uri string, localPart string) (Datatype, error) {
	if !c.dtrMapInUse {
		return nil, fmt.Errorf("DTR map is not used")
	}

	var datatype Datatype = nil

	// find datatype for given representation
	if uri == W3C_EXI_NS_URI {
		// EXI built-in datatypes
		// see http://www.w3.org/TR/exi/#builtInEXITypes
		switch localPart {
		case XsdBase64Binary.Local:
			datatype = NewBinaryBase64Datatype(nil)
		case XsdHexBinary.Local:
			datatype = NewBinaryHexDatatype(nil)
		case XsdBoolean.Local:
			datatype = NewBooleanDatatype(nil)
		case XsdDateTime.Local:
			datatype = NewDatetimeDatatype(DateTimeDateTime, nil)
		case XsdTime.Local:
			datatype = NewDatetimeDatatype(DateTimeTime, nil)
		case XsdDate.Local:
			datatype = NewDatetimeDatatype(DateTimeDate, nil)
		case XsdGYearMonth.Local:
			datatype = NewDatetimeDatatype(DateTimeGYearMonth, nil)
		case XsdGYear.Local:
			datatype = NewDatetimeDatatype(DateTimeGYear, nil)
		case XsdGMonthDay.Local:
			datatype = NewDatetimeDatatype(DateTimeGMonthDay, nil)
		case XsdGDay.Local:
			datatype = NewDatetimeDatatype(DateTimeGDay, nil)
		case XsdGMonth.Local:
			datatype = NewDatetimeDatatype(DateTimeGMonth, nil)
		case XsdDecimal.Local:
			datatype = NewDecimalDatatype(nil)
		case XsdDouble.Local:
			datatype = NewFloatDatatype(nil)
		case XsdInteger.Local:
			datatype = NewIntegerDatatype(nil)
		case XsdString.Local:
			datatype = NewStringDatatype(nil)
		case XsdExtendedString.Local:
			datatype = NewExtendedStringDatatype(nil)
		default:
			return nil, fmt.Errorf("[exi] unsupported datatype representation: {%s}%s", uri, localPart)
		}
	} else {
		qn := utils.QName{Space: uri, Local: localPart}
		if c.dtrMapRepresentationDatatype != nil {
			datatype = (*c.dtrMapRepresentationDatatype)[qn]
		}
		if datatype == nil {
			return nil, fmt.Errorf("[exi] no datatype instance")
		}
	}

	return datatype, nil
}

func (c *AbstractTypeCoder) getDtrDatatype(datatype Datatype) (Datatype, error) {
	if !c.dtrMapInUse {
		return nil, fmt.Errorf("DTR map is not used")
	}

	var dtrDatatype Datatype = nil
	var err error
	if datatype.Equals(BuiltInGetDefaultDatatype()) {
		// e.g., untyped values are encoded always as String
		dtrDatatype = datatype
	} else {
		schemaType := datatype.GetSchemaType().GetQName()

		// unions
		if datatype.GetBuiltInType() == BuiltInTypeString && (datatype.(*StringDatatype)).IsDerivedByUnion() {
			if utils.ContainsKey(c.dtrMap, schemaType) {
				// direct DTR mapping
				dtrDatatype = c.dtrMap[schemaType]
			} else {
				baseDatatype := datatype.GetBaseDatatype()
				schemaBaseType := baseDatatype.GetSchemaType().GetQName()

				if baseDatatype.GetBuiltInType() == BuiltInTypeString && (baseDatatype.(*StringDatatype)).IsDerivedByUnion() && utils.ContainsKey(c.dtrMap, schemaBaseType) {
					dtrDatatype = c.dtrMap[schemaBaseType]
				} else {
					dtrDatatype = datatype
				}
			}
		}

		// lists
		if dtrDatatype == nil && datatype.GetBuiltInType() == BuiltInTypeList {
			ldt := datatype.(*ListDatatype)

			if utils.ContainsKey(c.dtrMap, schemaType) {
				// direct DTR mapping
				dtrDatatype = c.dtrMap[schemaType]
			} else if utils.ContainsKey(c.dtrMap, ldt.GetListDatatype().GetSchemaType().GetQName()) {
				dt := c.dtrMap[ldt.GetListDatatype().GetSchemaType().GetQName()]
				dtrDatatype, err = NewListDatatypeChecked(dt, datatype.GetSchemaType())
				if err != nil {
					return nil, err
				}
			} else {
				baseDatatype := datatype.GetBaseDatatype()
				schemaBaseType := baseDatatype.GetSchemaType().GetQName()

				if baseDatatype.GetBuiltInType() == BuiltInTypeList && utils.ContainsKey(c.dtrMap, schemaBaseType) {
					dtrDatatype = c.dtrMap[schemaBaseType]
				} else {
					dtrDatatype = datatype
				}
			}
		}

		// enums
		if dtrDatatype == nil && datatype.GetBuiltInType() == BuiltInTypeEnumeration {
			if utils.ContainsKey(c.dtrMap, schemaType) {
				// direct DTR mapping
				dtrDatatype = c.dtrMap[schemaType]
			} else {
				baseDatatype := datatype.GetBaseDatatype()
				schemaBaseType := baseDatatype.GetSchemaType().GetQName()

				if baseDatatype.GetBuiltInType() == BuiltInTypeEnumeration && utils.ContainsKey(c.dtrMap, schemaBaseType) {
					dtrDatatype = c.dtrMap[schemaBaseType]
				} else {
					dtrDatatype = datatype
				}
			}
		}

		if dtrDatatype == nil {
			dtrDatatype = c.dtrMap[schemaType]

			if dtrDatatype == nil {
				dtrDatatype, err = c.updateDtrDatatype(datatype)
				if err != nil {
					return nil, err
				}
			}
		}

		if dtrDatatype.GetDatatypeID() == DataTypeID_EXI_EString && datatype.GetGrammarEnumeration() != nil {
			// add grammar strings et cetera
			esdt := dtrDatatype.(*ExtendedStringDatatype)
			esdt.SetGrammarStrings(datatype.GetGrammarEnumeration())
		}
	}

	return dtrDatatype, nil
}

func (c *AbstractTypeCoder) updateDtrDatatype(datatype Datatype) (Datatype, error) {
	baseDatatype := datatype.GetBaseDatatype()
	simpleBaseType := baseDatatype.GetSchemaType()
	var err error

	dtrDatatype := c.dtrMap[simpleBaseType.GetQName()]
	if dtrDatatype == nil {
		dtrDatatype, err = c.updateDtrDatatype(baseDatatype)
		if err != nil {
			return nil, err
		}
	}

	// special integer handling
	if (dtrDatatype.GetBuiltInType() == BuiltInTypeInteger || dtrDatatype.GetBuiltInType() == BuiltInTypeUnsignedInteger) &&
		(datatype.GetBuiltInType() == BuiltInTypeNBitUnsignedInteger || datatype.GetBuiltInType() == BuiltInTypeUnsignedInteger) {
		dtrDatatype = datatype
	}

	c.dtrMap[datatype.GetSchemaType().GetQName()] = dtrDatatype

	return dtrDatatype, nil
}

/*
	AbstractTypeEncoder implementation
*/

type AbstractTypeEncoder struct {
	*AbstractTypeCoder
}

func NewAbstractTypeEncoder(dtrMapTypes *[]utils.QName,
	dtrMapRepresentations *[]utils.QName,
	dtrMapRepresentationDatatype *map[utils.QName]Datatype,
) (*AbstractTypeEncoder, error) {
	super, err := NewAbstractTypeCoder(dtrMapTypes, dtrMapRepresentations, dtrMapRepresentationDatatype)
	if err != nil {
		return nil, err
	}

	return &AbstractTypeEncoder{
		AbstractTypeCoder: super,
	}, nil
}

func (e *AbstractTypeEncoder) writeRCSValue(rcsDT *RestrictedCharacterSetDatatype, qnc *QNameContext,
	channel EncoderChannel, encoder StringEncoder, lastValidValue string) error {

	hit, err := encoder.IsStringHit(lastValidValue)
	if err != nil {
		return err
	}

	if hit {
		err = encoder.WriteValue(qnc, channel, lastValidValue)
		if err != nil {
			return err
		}
	} else {
		// NO local or global value hit
		// string-table miss ==> restricted character
		// string literal is encoded as a String with the length
		// incremented by two.
		runes := []rune(lastValidValue)
		runesL := len(runes)

		err = channel.EncodeUnsignedInteger(runesL + 2)
		if err != nil {
			return err
		}

		rcs := rcsDT.GetRestrictedCharacterSet()

		// If length L is greater than zero the string S is added
		if runesL > 0 {
			numberOfBits := rcs.GetCodingLength()

			for i := 0; i < runesL; i++ {
				codePoint := runes[i]
				code := rcs.GetCode(int(codePoint))

				if code == NotFound {
					// indicate deviation
					err = channel.EncodeNBitUnsignedInteger(rcs.GetSize(), numberOfBits)
					if err != nil {
						return err
					}
					err = channel.EncodeUnsignedInteger(int(codePoint))
					if err != nil {
						return err
					}
				} else {
					err = channel.EncodeNBitUnsignedInteger(code, numberOfBits)
					if err != nil {
						return err
					}
				}

				// After encoding the string value, it is added to both the
				// associated "local" value string table partition and the
				// global value string table partition.
				encoder.AddValue(qnc, lastValidValue)
			}
		}
	}

	return nil
}

/*
	AbstractTypeDecoder implementation
*/

type AbstractTypeDecoder struct {
	*AbstractTypeCoder
}

func NewAbstractTypeDecoder(dtrMapTypes *[]utils.QName,
	dtrMapRepresentations *[]utils.QName,
	dtrMapRepresentationDatatype *map[utils.QName]Datatype,
) (*AbstractTypeDecoder, error) {
	super, err := NewAbstractTypeCoder(dtrMapTypes, dtrMapRepresentations, dtrMapRepresentationDatatype)
	if err != nil {
		return nil, err
	}

	return &AbstractTypeDecoder{
		AbstractTypeCoder: super,
	}, nil
}

func (d *AbstractTypeDecoder) readRCSValue(rcsDT *RestrictedCharacterSetDatatype, qnc *QNameContext, channel DecoderChannel, decoder StringDecoder) (Value, error) {
	rcs := rcsDT.GetRestrictedCharacterSet()

	var value *StringValue
	var err error

	i, err := channel.DecodeUnsignedInteger()
	if err != nil {
		return nil, err
	}

	switch i {
	case 0:
		// local value partition
		value, err = decoder.ReadValueLocalHit(qnc, channel)
		if err != nil {
			return nil, err
		}
	case 1:
		// found in global value partition
		value, err = decoder.ReadValueGlobalHit(channel)
		if err != nil {
			return nil, err
		}
	default:
		// not found in global value (and local value) partition
		// ==> restricted character string literal is encoded as a String
		// with the length incremented by two.
		l := i - 2

		// If length L is greater than zero the string S is added
		if l > 0 {
			numberOfBits := rcs.GetCodingLength()
			size := rcs.GetSize()

			cValue := make([]rune, l)
			value = NewStringValueFromSlice(cValue)

			for k := 0; k < l; k++ {
				code, err := channel.DecodeNBitUnsignedInteger(numberOfBits)
				if err != nil {
					return nil, err
				}
				var codePoint int
				if code == size {
					// deviation
					codePoint, err = channel.DecodeUnsignedInteger()
					if err != nil {
						return nil, err
					}
				} else {
					codePoint, err = rcs.GetCodePoint(code)
					if err != nil {
						return nil, err
					}
				}

				cValue[k] = rune(codePoint)
			}

			// After encoding the string value, it is added to both the
			// associated "local" value string table partition and the
			// global value string table partition.
			decoder.AddValue(qnc, value)
		} else {
			value = EmptyStringValue
		}
	}

	return value, nil
}

/*
	TypedTypeEncoder implementation
*/

type TypedTypeEncoder struct {
	*AbstractTypeEncoder
	lastDataType       Datatype
	doNormalize        bool
	lastBytes          *[]byte
	lastBool           *BooleanValue
	lastBooleanID      int
	lastBoolean        bool
	lastDecimal        *DecimalValue
	lastFloat          *FloatValue
	lastNBitInteger    *IntegerValue
	lastUnsignedIntger *IntegerValue
	lastInteger        *IntegerValue
	lastDateTime       *DateTimeValue
	lastString         *string
	lastEnumIndex      int
	lastListValues     *ListValue
}

func NewTypedTypeEncoder(dtrMapTypes *[]utils.QName,
	dtrMapRepresentations *[]utils.QName,
	dtrMapRepresentationDatatype *map[utils.QName]Datatype,
) (*TypedTypeEncoder, error) {
	return NewTypedTypeEncoderWithNormalize(dtrMapTypes, dtrMapRepresentations, dtrMapRepresentationDatatype, false)
}

func NewTypedTypeEncoderWithNormalize(dtrMapTypes *[]utils.QName,
	dtrMapRepresentations *[]utils.QName,
	dtrMapRepresentationDatatype *map[utils.QName]Datatype,
	doNormalize bool,
) (*TypedTypeEncoder, error) {
	super, err := NewAbstractTypeEncoder(dtrMapTypes, dtrMapRepresentations, dtrMapRepresentationDatatype)
	if err != nil {
		return nil, err
	}

	return &TypedTypeEncoder{
		AbstractTypeEncoder: super,
		lastDataType:        nil,
		doNormalize:         doNormalize,
		lastBytes:           nil,
		lastBool:            nil,
		lastBooleanID:       -1,
		lastDecimal:         nil,
		lastFloat:           nil,
		lastNBitInteger:     nil,
		lastUnsignedIntger:  nil,
		lastInteger:         nil,
		lastDateTime:        nil,
		lastString:          nil,
		lastEnumIndex:       -1,
		lastListValues:      nil,
	}, nil
}

func (e *TypedTypeEncoder) IsValid(datatype Datatype, value Value) (bool, error) {
	var err error
	if e.dtrMapInUse && datatype.GetBuiltInType() != BuiltInTypeExtendedString {
		e.lastDataType, err = e.getDtrDatatype(datatype)
		if err != nil {
			return false, err
		}
	} else {
		e.lastDataType = datatype
	}

	switch e.lastDataType.GetBuiltInType() {
	case BuiltInTypeBinaryBase64, BuiltInTypeBinaryHex:
		abv, ok := value.(*AbstractBinaryValue)
		if ok {
			e.lastBytes = utils.AsPtr(abv.ToBytes())
		} else {
			s, err := value.ToString()
			if err != nil {
				return false, err
			}
			return e.isValidString(s)
		}
	case BuiltInTypeBoolean:
		b, ok := value.(*BooleanValue)
		if ok {
			e.lastBool = b
			return true, nil
		} else {
			s, err := value.ToString()
			if err != nil {
				return false, err
			}
			return e.isValidString(s)
		}
	case BuiltInTypeBooleanFacet:
		b, ok := value.(*BooleanValue)
		if ok {
			e.lastBool = b
			e.lastBooleanID = 0
			// TODO not fully correct
			if e.lastBoolean {
				e.lastBooleanID = 2
			}
			return true, nil
		} else {
			s, err := value.ToString()
			if err != nil {
				return false, err
			}
			return e.isValidString(s)
		}
	case BuiltInTypeDecimal:
		d, ok := value.(*DecimalValue)
		if ok {
			e.lastDecimal = d
			return true, nil
		} else {
			s, err := value.ToString()
			if err != nil {
				return false, err
			}
			return e.isValidString(s)
		}
	case BuiltInTypeFloat:
		f, ok := value.(*FloatValue)
		if ok {
			e.lastFloat = f
			return true, nil
		} else {
			s, err := value.ToString()
			if err != nil {
				return false, err
			}
			return e.isValidString(s)
		}
	case BuiltInTypeNBitUnsignedInteger:
		nbitDT := e.lastDataType.(*NBitUnsignedIntegerDatatype)
		nbit, ok := value.(*IntegerValue)
		if ok {
			e.lastNBitInteger = nbit
			return e.lastNBitInteger.Cmp(nbitDT.GetLowerBound()) >= 0 && e.lastNBitInteger.Cmp(nbitDT.GetUpperBound()) <= 0, nil
		} else {
			s, err := value.ToString()
			if err != nil {
				return false, err
			}
			return e.isValidString(s)
		}
	case BuiltInTypeUnsignedInteger:
		i, ok := value.(*IntegerValue)
		if ok {
			e.lastUnsignedIntger = i
			return e.lastUnsignedIntger.IsPositive(), nil
		} else {
			s, err := value.ToString()
			if err != nil {
				return false, err
			}
			return e.isValidString(s)
		}
	case BuiltInTypeInteger:
		i, ok := value.(*IntegerValue)
		if ok {
			e.lastInteger = i
			return true, nil
		} else {
			s, err := value.ToString()
			if err != nil {
				return false, err
			}
			return e.isValidString(s)
		}
	case BuiltInTypeDateTime:
		dt, ok := value.(*DateTimeValue)
		if ok {
			e.lastDateTime = dt
			return true, nil
		} else {
			s, err := value.ToString()
			if err != nil {
				return false, err
			}
			return e.isValidString(s)
		}
	case BuiltInTypeString, BuiltInTypeRcsString, BuiltInTypeExtendedString:
		// Note: no validity check needed for RCS strings since any char-sequence
		// can be encoded due to fallback mechanism.
		s, err := value.ToString()
		if err != nil {
			return false, err
		}
		e.lastString = &s
		return true, nil
	case BuiltInTypeEnumeration:
		enumDT := e.lastDataType.(*EnumerationDatatype)
		index := 0

		for index < enumDT.GetEnumerationSize() {
			if enumDT.GetEnumValue(index).Equals(value) {
				e.lastEnumIndex = index
				return true, nil
			}
			index++
		}

		return false, nil
	case BuiltInTypeList:
		lv, ok := value.(*ListValue)
		if ok {
			listDT := e.lastDataType.(*ListDatatype)
			if listDT.GetListDatatype().GetBuiltInType() == lv.GetListDatatype().GetBuiltInType() {
				e.lastListValues = lv
				return true, nil
			}
		} else {
			s, err := value.ToString()
			if err != nil {
				return false, err
			}
			return e.isValidString(s)
		}
	case BuiltInTypeQName:
		/* not allowed datatype */
		return false, nil
	}

	return false, nil
}

func (e *TypedTypeEncoder) isValidString(value string) (bool, error) {
	var err error

	switch e.lastDataType.GetBuiltInType() {
	case BuiltInTypeBinaryBase64:
		value = strings.TrimSpace(value)
		bvb := BinaryBase64ValueParse(value)
		if bvb == nil {
			return false, nil
		} else {
			e.lastBytes = utils.AsPtr(bvb.ToBytes())
			return true, nil
		}
	case BuiltInTypeBinaryHex:
		value = strings.TrimSpace(value)
		bhv := BinaryHexValueParse(value)
		if bhv == nil {
			return false, nil
		} else {
			e.lastBytes = utils.AsPtr(bhv.ToBytes())
			return true, nil
		}
	case BuiltInTypeBoolean:
		e.lastBool = BooleanValueParse(value)
		return (e.lastBool != nil), nil
	case BuiltInTypeBooleanFacet:
		value = strings.TrimSpace(value)
		retValue := true

		switch value {
		case XSDBooleanFalse:
			e.lastBooleanID = 0
			e.lastBoolean = false
		case XSDBoolean0:
			e.lastBooleanID = 1
			e.lastBoolean = false
		case XSDBooleanTrue:
			e.lastBooleanID = 2
			e.lastBoolean = true
		case XSDBoolean1:
			e.lastBooleanID = 3
			e.lastBoolean = true
		default:
			retValue = false
		}

		return retValue, nil
	case BuiltInTypeDecimal:
		e.lastDecimal, err = DecimalValueParseString(value)
		if err != nil {
			return false, err
		}
		return (e.lastDecimal != nil), nil
	case BuiltInTypeFloat:
		e.lastFloat, err = FloatValueParseString(value)
		if err != nil {
			return false, err
		}
		return (e.lastFloat != nil), nil
	case BuiltInTypeNBitUnsignedInteger:
		e.lastNBitInteger, err = IntegerValueParse(value)
		if err != nil {
			return false, err
		}
		if e.lastNBitInteger == nil {
			return false, nil
		} else {
			nbitDT := e.lastDataType.(*NBitUnsignedIntegerDatatype)
			return e.lastNBitInteger.Cmp(nbitDT.GetLowerBound()) >= 0 && e.lastNBitInteger.Cmp(nbitDT.GetUpperBound()) <= 0, nil
		}
	case BuiltInTypeUnsignedInteger:
		e.lastUnsignedIntger, err = IntegerValueParse(value)
		if err != nil {
			return false, err
		}
		if e.lastUnsignedIntger != nil {
			return e.lastUnsignedIntger.IsPositive(), nil
		} else {
			return false, nil
		}
	case BuiltInTypeInteger:
		e.lastInteger, err = IntegerValueParse(value)
		if err != nil {
			return false, err
		}
		return (e.lastInteger != nil), nil
	case BuiltInTypeDateTime:
		datetimeDT := e.lastDataType.(*DatetimeDatatype)
		e.lastDateTime, err = DateTimeParse(value, datetimeDT.GetDatetimeType())
		if err != nil {
			return false, err
		}
		return (e.lastDateTime != nil), nil
	case BuiltInTypeList:
		listDT := e.lastDataType.(*ListDatatype)
		e.lastListValues, err = ListValueParse(value, listDT.GetListDatatype())
		if err != nil {
			return false, err
		}
		return (e.lastListValues != nil), nil
	default:
		return false, nil
	}
}

func (e *TypedTypeEncoder) normalize(datatype Datatype) error {
	switch datatype.GetBuiltInType() {
	case BuiltInTypeDateTime:
		// See https://www.w3.org/TR/2004/REC-xmlschema-2-20041028/#dateTime-canonical-representation
		if e.lastDateTime != nil {
			e.lastDateTime = e.lastDateTime.Normalize()
		}
	case BuiltInTypeList:
		if e.lastListValues != nil {
			dt := e.lastListValues.GetListDatatype()
			for _, v := range e.lastListValues.ToValues() {
				_, err := e.IsValid(dt, v)
				if err != nil {
					return err
				}
				err = e.normalize(dt)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (e *TypedTypeEncoder) WriteValue(qnc *QNameContext, channel EncoderChannel, encoder StringEncoder) error {
	if e.doNormalize {
		if err := e.normalize(e.lastDataType); err != nil {
			return err
		}
	}

	switch e.lastDataType.GetBuiltInType() {
	case BuiltInTypeBinaryBase64, BuiltInTypeBinaryHex:
		if err := channel.EncodeBinary(*e.lastBytes); err != nil {
			return err
		}
	case BuiltInTypeBoolean:
		if err := channel.EncodeBoolean(e.lastBool.ToBoolean()); err != nil {
			return err
		}
	case BuiltInTypeBooleanFacet:
		if err := channel.EncodeNBitUnsignedInteger(e.lastBooleanID, 2); err != nil {
			return err
		}
	case BuiltInTypeDecimal:
		if err := channel.EncodeDecimal(e.lastDecimal.IsNegative(), e.lastDecimal.GetIntegral(), e.lastDecimal.GetRevFractional()); err != nil {
			return err
		}
	case BuiltInTypeFloat:
		if err := channel.EncodeFloat(e.lastFloat); err != nil {
			return err
		}
	case BuiltInTypeNBitUnsignedInteger:
		nbitDT := e.lastDataType.(*NBitUnsignedIntegerDatatype)
		iv := e.lastNBitInteger.Sub(nbitDT.GetLowerBound())
		if err := channel.EncodeNBitUnsignedInteger(iv.Value32(), nbitDT.GetNumberOfBits()); err != nil {
			return err
		}
	case BuiltInTypeUnsignedInteger:
		if err := channel.EncodeUnsignedIntegerValue(e.lastUnsignedIntger); err != nil {
			return err
		}
	case BuiltInTypeInteger:
		if err := channel.EncodeIntegerValue(e.lastInteger); err != nil {
			return err
		}
	case BuiltInTypeDateTime:
		if err := channel.EncodeDateTime(e.lastDateTime); err != nil {
			return err
		}
	case BuiltInTypeString:
		if err := encoder.WriteValue(qnc, channel, *e.lastString); err != nil {
			return err
		}
	case BuiltInTypeRcsString:
		rcsDT := e.lastDataType.(*RestrictedCharacterSetDatatype)
		if err := e.writeRCSValue(rcsDT, qnc, channel, encoder, *e.lastString); err != nil {
			return err
		}
	case BuiltInTypeExtendedString:
		esDT := e.lastDataType.(*ExtendedStringDatatype)
		if err := e.writeExtendedValue(esDT, qnc, channel, encoder, *e.lastString); err != nil {
			return err
		}
	case BuiltInTypeEnumeration:
		enumDT := e.lastDataType.(*EnumerationDatatype)
		if err := channel.EncodeNBitUnsignedInteger(e.lastEnumIndex, enumDT.GetCodingLength()); err != nil {
			return err
		}
	case BuiltInTypeList:
		listDT := e.lastDataType.(*ListDatatype)
		listDatatype := listDT.GetListDatatype()

		// length prefixed sequence of values
		values := e.lastListValues.ToValues()
		if err := channel.EncodeUnsignedInteger(len(values)); err != nil {
			return err
		}

		// iterate over all tokens
		for i := 0; i < len(values); i++ {
			v := values[i]
			valid, err := e.IsValid(listDatatype, v)
			if err != nil {
				return err
			}
			if !valid {
				return fmt.Errorf("list value is not valid") //TODO: ToString
			}

			if err := e.WriteValue(qnc, channel, encoder); err != nil {
				return err
			}
		}
	case BuiltInTypeQName:
		return fmt.Errorf("QName is not allowed as EXI datatype")
	default:
		return fmt.Errorf("EXI datatype not supported")
	}

	return nil
}

func (e *TypedTypeEncoder) getEnumIndex(grammarStrings EnumDatatype, sv *StringValue) int {
	for i := 0; i < grammarStrings.GetEnumerationSize(); i++ {
		v := grammarStrings.GetEnumValue(i)
		if sv.Equals(v) {
			return i
		}
	}
	return -1
}

func (e *TypedTypeEncoder) writeExtendedValue(esDT *ExtendedStringDatatype, qnc *QNameContext, channel EncoderChannel, encoder StringEncoder, value string) error {
	grammarStrings := esDT.GetGrammarStrings()

	vc := encoder.GetValueContainer(value)
	if vc != nil {
		// hit
		if encoder.IsLocalValuePartitions() && qnc.Equals(vc.Context) {
			/*
			 * local value hit ==> is represented as zero (0) encoded as an
			 * Unsigned Integer followed by the compact identifier of the
			 * string value in the "local" value partition
			 */
			if err := channel.EncodeUnsignedInteger(0); err != nil {
				return err
			}
			numberBitsLocal := utils.GetCodingLength(encoder.GetNumberOfStringValues(qnc))
			if err := channel.EncodeNBitUnsignedInteger(vc.LocalValueID, numberBitsLocal); err != nil {
				return err
			}
		} else {
			/*
			 * global value hit ==> value is represented as one (1) encoded
			 * as an Unsigned Integer followed by the compact identifier of
			 * the String value in the global value partition.
			 */
			if err := channel.EncodeUnsignedInteger(1); err != nil {
				return err
			}

			// global value size
			numberBitsGlobal := utils.GetCodingLength(encoder.GetValueContainerSize())
			if err := channel.EncodeNBitUnsignedInteger(vc.GlobalValueID, numberBitsGlobal); err != nil {
				return err
			}
		}
	} else {
		/*
		 * miss [not found in local nor in global value partition] ==>
		 * string literal is encoded as a String with the length incremented
		 * by 6.
		 */

		// --> check grammar strings
		encoded := false
		if grammarStrings != nil {
			gindex := e.getEnumIndex(grammarStrings, NewStringValueFromString(value))

			if gindex >= 0 {
				if err := channel.EncodeUnsignedInteger(2); err != nil {
					return err
				}
				if err := channel.EncodeNBitUnsignedInteger(gindex, grammarStrings.GetCodingLength()); err != nil {
					return err
				}

				encoded = true
			}
		}

		if !encoded {
			// TODO (3)shared string, (4)split string, (5)undefined

			l, err := utils.CodePointCount(value, 0, len(value))
			if err != nil {
				return err
			}

			if err := channel.EncodeUnsignedInteger(l + 6); err != nil {
				return err
			}

			/*
			 * If length L is greater than zero the string S is added
			 */
			if l > 0 {
				if err := channel.EncodeStringOnly(value); err != nil {
					return err
				}
				// After encoding the string value, it is added to both the
				// associated "local" value string table partition and the
				// global value string table partition.
				encoder.AddValue(qnc, value)
			}
		}
	}

	return nil
}

/*
	TypedTypeDecoder implementation
*/

type TypedTypeDecoder struct {
	*AbstractTypeDecoder
}

func NewTypedTypeDecoder(
	dtrMapTypes *[]utils.QName,
	dtrMapRepresentations *[]utils.QName,
	dtrMapRepresentationDatatype *map[utils.QName]Datatype,
) (*TypedTypeDecoder, error) {
	decoder, err := NewAbstractTypeDecoder(dtrMapTypes, dtrMapRepresentations, dtrMapRepresentationDatatype)
	if err != nil {
		return nil, err
	}
	return &TypedTypeDecoder{
		AbstractTypeDecoder: decoder,
	}, nil
}

func (d *TypedTypeDecoder) ReadValue(datatype Datatype, qnc *QNameContext, channel DecoderChannel, decoder StringDecoder) (Value, error) {
	var err error
	if d.dtrMapInUse {
		datatype, err = d.getDtrDatatype(datatype)
		if err != nil {
			return nil, err
		}
	}

	switch datatype.GetBuiltInType() {
	case BuiltInTypeBinaryBase64:
		data, err := channel.DecodeBinary()
		if err != nil {
			return nil, err
		}
		return NewBinaryBase64Value(data), nil
	case BuiltInTypeBinaryHex:
		data, err := channel.DecodeBinary()
		if err != nil {
			return nil, err
		}
		return NewBinaryHexValue(data), nil
	case BuiltInTypeBoolean:
		value, err := channel.DecodeBooleanValue()
		if err != nil {
			return nil, err
		}
		return value, nil
	case BuiltInTypeBooleanFacet:
		booleanID, err := channel.DecodeNBitUnsignedInteger(2)
		if err != nil {
			return nil, err
		}
		value, err := GetBooleanValueForID(booleanID)
		if err != nil {
			return nil, err
		}
		return value, nil
	case BuiltInTypeDecimal:
		value, err := channel.DecodeDecimalValue()
		if err != nil {
			return nil, err
		}
		return value, nil
	case BuiltInTypeFloat:
		value, err := channel.DecodeFloatValue()
		if err != nil {
			return nil, err
		}
		return value, nil
	case BuiltInTypeNBitUnsignedInteger:
		nbitDT := datatype.(*NBitUnsignedIntegerDatatype)
		value, err := channel.DecodeNBitUnsignedIntegerValue(nbitDT.GetNumberOfBits())
		if err != nil {
			return nil, err
		}
		return value.Add(nbitDT.GetLowerBound()), nil
	case BuiltInTypeUnsignedInteger:
		value, err := channel.DecodeUnsignedIntegerValue()
		if err != nil {
			return nil, err
		}
		return value, nil
	case BuiltInTypeInteger:
		value, err := channel.DecodeIntegerValue()
		if err != nil {
			return nil, err
		}
		return value, nil
	case BuiltInTypeDateTime:
		dtDT := datatype.(*DatetimeDatatype)
		value, err := channel.DecodeDateTimeValue(dtDT.GetDatetimeType())
		if err != nil {
			return nil, err
		}
		return value, nil
	case BuiltInTypeString:
		value, err := decoder.ReadValue(qnc, channel)
		if err != nil {
			return nil, err
		}
		return value, nil
	case BuiltInTypeRcsString:
		rcsDT := datatype.(*RestrictedCharacterSetDatatype)
		return d.readRCSValue(rcsDT, qnc, channel, decoder)
	case BuiltInTypeExtendedString:
		esDT := datatype.(*ExtendedStringDatatype)
		return d.readExtendedString(esDT, qnc, channel, decoder)
	case BuiltInTypeEnumeration:
		enumDT := datatype.(*EnumerationDatatype)
		index, err := channel.DecodeNBitUnsignedInteger(enumDT.GetCodingLength())
		if err != nil {
			return nil, err
		}
		if index < 0 || index >= enumDT.GetEnumerationSize() {
			return nil, fmt.Errorf("index out of bounds")
		}
		return enumDT.GetEnumValue(index), nil
	case BuiltInTypeList:
		listDT := datatype.(*ListDatatype)
		listItemDT := listDT.GetListDatatype()

		len, err := channel.DecodeUnsignedInteger()
		if err != nil {
			return nil, err
		}
		values := make([]Value, len)

		for i := 0; i < len; i++ {
			values[i], err = d.ReadValue(listItemDT, qnc, channel, decoder)
			if err != nil {
				return nil, err
			}
		}
		return NewListValue(values, listItemDT), nil
	case BuiltInTypeQName:
		/* not allowed datatype */
		return nil, fmt.Errorf("QName is not an allowed as EXI datatype")
	}

	return nil, nil
}

func (d *TypedTypeDecoder) readExtendedString(esDT *ExtendedStringDatatype, qnc *QNameContext, channel DecoderChannel, decoder StringDecoder) (*StringValue, error) {
	var value *StringValue = nil

	grammarStrings := esDT.GetGrammarStrings()

	i, err := channel.DecodeUnsignedInteger()
	if err != nil {
		return nil, err
	}

	switch i {
	case 0:
		// local value partition
		if decoder.IsLocalValuePartitions() {
			value, err = decoder.ReadValueLocalHit(qnc, channel)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("EXI stream contains local-value hit even though profile options indicate otherwise")
		}
	case 1:
		// found in global value partition
		value, err = decoder.ReadValueGlobalHit(channel)
		if err != nil {
			return nil, err
		}
	case 2:
		// grammar string (enum)
		index, err := channel.DecodeNBitUnsignedInteger(grammarStrings.GetCodingLength())
		if err != nil {
			return nil, err
		}
		if index < 0 || index >= grammarStrings.GetEnumerationSize() {
			return nil, fmt.Errorf("index out of bounds")
		}
		v := grammarStrings.GetEnumValue(index)

		sv, ok := v.(*StringValue)
		if ok {
			value = sv
		} else {
			s, err := v.ToString()
			if err != nil {
				return nil, err
			}
			value = NewStringValueFromString(s)
		}
	case 3:
		// shared string
		return nil, fmt.Errorf("ExtendedString, no support for <shared string>")
	case 4:
		// split string
		return nil, fmt.Errorf("ExtendedString, no support for <split string>")
	case 5:
		// undefined
		return nil, fmt.Errorf("ExtendedString, no support for <undefined>")
	default:
		// not found in global value (and local value) partition
		// ==> string literal is encoded as a String with the length
		// incremented by 6.
		len := i - 6

		// If length 'len' is greater than zero the string S is added.
		if len > 0 {
			ch, err := channel.DecodeStringOnly(len)
			if err != nil {
				return nil, err
			}
			value = NewStringValueFromSlice(ch)
			// After encoding the string value, it is added to both the
			// associated "local" value string table partition and the
			// global value string table partition.
			// AddValue(context, value)
			decoder.AddValue(qnc, value)
		} else {
			value = EmptyStringValue
		}
	}

	if value == nil {
		return nil, fmt.Errorf("unexpected decoder state: value == nil")
	}
	return value, nil
}

/*
	StringTypeDecoder implementation
*/

type StringTypeDecoder struct {
	*AbstractTypeDecoder
}

func NewStringTypeDecoder() (*StringTypeDecoder, error) {
	decoder, err := NewAbstractTypeDecoder(nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return &StringTypeDecoder{
		AbstractTypeDecoder: decoder,
	}, nil
}

func (d *StringTypeDecoder) ReadValue(datatype Datatype, qnc *QNameContext, channel DecoderChannel, decoder StringDecoder) (Value, error) {
	return decoder.ReadValue(qnc, channel)
}

/*
	StringTypeEncoder implementation
*/

type StringTypeEncoder struct {
	*AbstractTypeEncoder
	lastValidValue *string
}

func NewStringTypeEncoder() (*StringTypeEncoder, error) {
	encoder, err := NewAbstractTypeEncoder(nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return &StringTypeEncoder{
		AbstractTypeEncoder: encoder,
		lastValidValue:      nil,
	}, nil
}

func (e *StringTypeEncoder) IsValid(datatype Datatype, value Value) (bool, error) {
	s, err := value.ToString()
	if err != nil {
		return false, err
	}
	e.lastValidValue = &s
	return true, nil
}

func (e *StringTypeEncoder) WriteValue(qnc *QNameContext, channel EncoderChannel, encoder StringEncoder) error {
	return encoder.WriteValue(qnc, channel, *e.lastValidValue)
}

/*
	LexicalTypeDecoder implementation
*/

type LexicalTypeDecoder struct {
	*AbstractTypeDecoder
	rcsBase64Binary *RestrictedCharacterSetDatatype
	rcsHexBinary    *RestrictedCharacterSetDatatype
	rcsBoolean      *RestrictedCharacterSetDatatype
	rcsDateTime     *RestrictedCharacterSetDatatype
	rcsDecimal      *RestrictedCharacterSetDatatype
	rcsDouble       *RestrictedCharacterSetDatatype
	rcsInteger      *RestrictedCharacterSetDatatype
	rcsString       *RestrictedCharacterSetDatatype
}

func NewLexicalTypeDecoder(
	dtrMapTypes *[]utils.QName,
	dtrMapRepresentations *[]utils.QName,
	dtrMapRepresentationDatatype *map[utils.QName]Datatype,
) (*LexicalTypeDecoder, error) {
	decoder, err := NewAbstractTypeDecoder(dtrMapTypes, dtrMapRepresentations, dtrMapRepresentationDatatype)
	if err != nil {
		return nil, err
	}
	return &LexicalTypeDecoder{
		AbstractTypeDecoder: decoder,
		rcsBase64Binary:     NewRestrictedCharacterSetDatatype(NewXSDBase64CharacterSet(), nil),
		rcsHexBinary:        NewRestrictedCharacterSetDatatype(NewXSDHexBinaryCharacterSet(), nil),
		rcsBoolean:          NewRestrictedCharacterSetDatatype(NewXSDBooleanCharacterSet(), nil),
		rcsDateTime:         NewRestrictedCharacterSetDatatype(NewXSDDateTimeCharacterSet(), nil),
		rcsDecimal:          NewRestrictedCharacterSetDatatype(NewXSDDecimalCharacterSet(), nil),
		rcsDouble:           NewRestrictedCharacterSetDatatype(NewXSDDoubleCharacterSet(), nil),
		rcsInteger:          NewRestrictedCharacterSetDatatype(NewXSDIntegerCharacterSet(), nil),
		rcsString:           NewRestrictedCharacterSetDatatype(NewXSDStringCharacterSet(), nil),
	}, nil
}

func (d *LexicalTypeDecoder) ReadValue(datatype Datatype, qnc *QNameContext, channel DecoderChannel, decoder StringDecoder) (Value, error) {
	var err error
	if d.dtrMapInUse {
		datatype, err = d.getDtrDatatype(datatype)
		if err != nil {
			return nil, err
		}
	}

	switch datatype.GetDatatypeID() {
	case DataTypeID_EXI_Base64Binary:
		return d.readRCSValue(d.rcsBase64Binary, qnc, channel, decoder)
	case DataTypeID_EXI_HexBinary:
		return d.readRCSValue(d.rcsHexBinary, qnc, channel, decoder)
	case DataTypeID_EXI_Boolean:
		return d.readRCSValue(d.rcsBoolean, qnc, channel, decoder)
	case DataTypeID_EXI_DateTime,
		DataTypeID_EXI_Time,
		DataTypeID_EXI_Date,
		DataTypeID_EXI_GYearMonth,
		DataTypeID_EXI_GYear,
		DataTypeID_EXI_GMonthDay,
		DataTypeID_EXI_GDay,
		DataTypeID_EXI_GMonth:
		return d.readRCSValue(d.rcsDateTime, qnc, channel, decoder)
	case DataTypeID_EXI_Decimal:
		return d.readRCSValue(d.rcsDecimal, qnc, channel, decoder)
	case DataTypeID_EXI_Double:
		return d.readRCSValue(d.rcsDouble, qnc, channel, decoder)
	case DataTypeID_EXI_Integer:
		return d.readRCSValue(d.rcsInteger, qnc, channel, decoder)
	case DataTypeID_EXI_String:
		// exi:string no restricted character set
		return decoder.ReadValue(qnc, channel)
	default:
		return nil, fmt.Errorf("unsupported datatype ID: %d", datatype.GetDatatypeID())
	}
}

/*
	LexicalTypeEncoder implementation
*/

type LexicalTypeEncoder struct {
	*AbstractTypeEncoder
	rcsBase64Binary *RestrictedCharacterSetDatatype
	rcsHexBinary    *RestrictedCharacterSetDatatype
	rcsBoolean      *RestrictedCharacterSetDatatype
	rcsDateTime     *RestrictedCharacterSetDatatype
	rcsDecimal      *RestrictedCharacterSetDatatype
	rcsDouble       *RestrictedCharacterSetDatatype
	rcsInteger      *RestrictedCharacterSetDatatype
	rcsString       *RestrictedCharacterSetDatatype
	lastValue       Value
	lastDatatype    Datatype
}

func NewLexicalTypeEncoder(
	dtrMapTypes *[]utils.QName,
	dtrMapRepresentations *[]utils.QName,
	dtrMapRepresentationDatatype *map[utils.QName]Datatype,
) (*LexicalTypeEncoder, error) {
	encoder, err := NewAbstractTypeEncoder(dtrMapTypes, dtrMapRepresentations, dtrMapRepresentationDatatype)
	if err != nil {
		return nil, err
	}
	return &LexicalTypeEncoder{
		AbstractTypeEncoder: encoder,
		rcsBase64Binary:     NewRestrictedCharacterSetDatatype(NewXSDBase64CharacterSet(), nil),
		rcsHexBinary:        NewRestrictedCharacterSetDatatype(NewXSDHexBinaryCharacterSet(), nil),
		rcsBoolean:          NewRestrictedCharacterSetDatatype(NewXSDBooleanCharacterSet(), nil),
		rcsDateTime:         NewRestrictedCharacterSetDatatype(NewXSDDateTimeCharacterSet(), nil),
		rcsDecimal:          NewRestrictedCharacterSetDatatype(NewXSDDecimalCharacterSet(), nil),
		rcsDouble:           NewRestrictedCharacterSetDatatype(NewXSDDoubleCharacterSet(), nil),
		rcsInteger:          NewRestrictedCharacterSetDatatype(NewXSDIntegerCharacterSet(), nil),
		rcsString:           NewRestrictedCharacterSetDatatype(NewXSDStringCharacterSet(), nil),
		lastValue:           nil,
		lastDatatype:        nil,
	}, nil
}

func (e *LexicalTypeEncoder) IsValid(datatype Datatype, value Value) (bool, error) {
	var err error
	if e.dtrMapInUse {
		e.lastDatatype, err = e.getDtrDatatype(datatype)
		if err != nil {
			return false, err
		}
	} else {
		e.lastDatatype = datatype
	}
	e.lastValue = value
	return true, nil
}

func (e *LexicalTypeEncoder) WriteValue(qnc *QNameContext, channel EncoderChannel, encoder StringEncoder) error {
	lvs, err := e.lastValue.ToString()
	if err != nil {
		return err
	}

	var rcs *RestrictedCharacterSetDatatype

	switch e.lastDatatype.GetDatatypeID() {
	case DataTypeID_EXI_Base64Binary:
		rcs = e.rcsBase64Binary
	case DataTypeID_EXI_HexBinary:
		rcs = e.rcsHexBinary
	case DataTypeID_EXI_Boolean:
		rcs = e.rcsBoolean
	case DataTypeID_EXI_DateTime,
		DataTypeID_EXI_Time,
		DataTypeID_EXI_Date,
		DataTypeID_EXI_GYearMonth,
		DataTypeID_EXI_GYear,
		DataTypeID_EXI_GMonthDay,
		DataTypeID_EXI_GDay,
		DataTypeID_EXI_GMonth:
		rcs = e.rcsDateTime
	case DataTypeID_EXI_Decimal:
		rcs = e.rcsDecimal
	case DataTypeID_EXI_Double:
		rcs = e.rcsDouble
	case DataTypeID_EXI_Integer:
		rcs = e.rcsInteger
	case DataTypeID_EXI_String:
		// exi:string no restricted character set
		return encoder.WriteValue(qnc, channel, lvs)
	default:
		return fmt.Errorf("unsupported datatype ID: %d", e.lastDatatype.GetDatatypeID())
	}

	if _, err := e.IsValid(rcs, e.lastValue); err != nil {
		return err
	}
	if err := e.writeRCSValue(rcs, qnc, channel, encoder, lvs); err != nil {
		return err
	}

	return nil
}
