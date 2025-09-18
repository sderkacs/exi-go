package core

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/apd/v3"
	Text "github.com/linkdotnet/golang-stringbuilder"
	"github.com/sderkacs/exi-go/utils"
)

type ValueType int

const (
	ValueTypeBinaryBase64 ValueType = iota
	ValueTypeBinaryHex
	ValueTypeBoolean
	ValueTypeDecimal
	ValueTypeFloat
	ValueTypeInteger
	ValueTypeDateTime
	ValueTypeString
	ValueTypeList
	ValueTypeQName
)

type Value interface {
	GetValueType() ValueType
	GetCharacters() ([]rune, error)
	FillCharactersBuffer(buffer []rune, offset int) error
	GetCharactersLength() (int, error)
	ToString() (string, error)
	BufferToString(buffer []rune, offset int) (string, error)
	Equals(o Value) bool
}

/*
	AbstractValue implementation
*/

type AbstractValue struct {
	Value
	sLen      int
	valueType ValueType
}

func NewAbstractValue(valueType ValueType) *AbstractValue {
	return &AbstractValue{
		sLen:      -1,
		valueType: valueType,
	}
}

func (v *AbstractValue) GetValueType() ValueType {
	return v.valueType
}

func (v *AbstractValue) GetCharacters() ([]rune, error) {
	len, err := v.GetCharactersLength()
	if err != nil {
		return []rune{}, err
	}
	dst := make([]rune, len)
	if err := v.FillCharactersBuffer(dst, 0); err != nil {
		return []rune{}, err
	}

	return dst, nil
}

func (v *AbstractValue) ToString() (string, error) {
	c, err := v.GetCharacters()
	if err != nil {
		return "", err
	}
	return string(c), nil
}

func (v *AbstractValue) BufferToString(buffer []rune, offset int) (string, error) {
	if err := v.FillCharactersBuffer(buffer, offset); err != nil {
		return "", err
	}
	len, err := v.GetCharactersLength()
	if err != nil {
		return "", err
	}

	return string(buffer[offset : offset+len]), nil
}

/*
	AbstractBinaryValue implementation
*/

type AbstractBinaryValue struct {
	*AbstractValue
	bytes []byte
}

func NewAbstractBinaryValue(valueType ValueType, bytes []byte) *AbstractBinaryValue {
	return &AbstractBinaryValue{
		AbstractValue: NewAbstractValue(valueType),
		bytes:         bytes,
	}
}

func (v *AbstractBinaryValue) ToBytes() []byte {
	return v.bytes
}

func (v *AbstractBinaryValue) equals(oBytes []byte) bool {
	if len(v.bytes) == len(oBytes) {
		for i := 0; i < len(v.bytes); i++ {
			if v.bytes[i] != oBytes[i] {
				return false
			}
		}
		return true
	}

	return false
}

/*
	BinaryBase64Value implementation
*/

const (
	eightBit           = 8
	twentyFourBitGroup = 24
)

type BinaryBase64Value struct {
	*AbstractBinaryValue
	fewerThan24bits int
	numberTriplets  int
	numberQuartet   int
}

func NewBinaryBase64Value(bytes []byte) *BinaryBase64Value {
	return &BinaryBase64Value{
		AbstractBinaryValue: NewAbstractBinaryValue(ValueTypeBinaryBase64, bytes),
	}
}

func BinaryBase64ValueParse(val string) *BinaryBase64Value {
	bytes, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return nil
	}

	return NewBinaryBase64Value(bytes)
}

func (v *BinaryBase64Value) GetCharactersLength() (int, error) {
	if v.sLen == -1 {
		lenghtDataBits := len(v.bytes) * eightBit
		if lenghtDataBits == 0 {
			v.sLen = 0
		} else {
			v.fewerThan24bits = lenghtDataBits % twentyFourBitGroup
			v.numberTriplets = lenghtDataBits / twentyFourBitGroup
			if v.fewerThan24bits != 0 {
				v.numberQuartet = v.numberTriplets + 1
			} else {
				v.numberQuartet = v.numberTriplets
			}
			v.sLen = v.numberQuartet * 4
		}
	}

	return v.sLen, nil
}

func (v *BinaryBase64Value) FillCharactersBuffer(buffer []rune, offset int) error {
	v.GetCharactersLength()
	b := base64.StdEncoding.EncodeToString(v.bytes)
	copy(buffer[offset:], []rune(b))
	return nil
}

func (v *BinaryBase64Value) Equals(o Value) bool {
	if o == nil {
		return false
	}
	oi, ok := o.(*BinaryBase64Value)
	if ok {
		return v.equals(oi.bytes)
	} else {
		s, err := o.ToString()
		if err != nil {
			return false
		}
		bv := BinaryBase64ValueParse(s)
		if bv != nil {
			return v.equals(bv.bytes)
		} else {
			return false
		}
	}
}

/*
	BinaryHexValue implementation
*/

type BinaryHexValue struct {
	*AbstractBinaryValue
	lengthData int
}

func NewBinaryHexValue(bytes []byte) *BinaryHexValue {
	return &BinaryHexValue{
		AbstractBinaryValue: NewAbstractBinaryValue(ValueTypeBinaryHex, bytes),
		lengthData:          -1,
	}
}

func BinaryHexValueParse(val string) *BinaryHexValue {
	bytes, err := hex.DecodeString(val)
	if err != nil {
		return nil
	} else {
		return NewBinaryHexValue(bytes)
	}
}

func (v *BinaryHexValue) GetCharactersLength() (int, error) {
	if v.sLen == -1 {
		v.lengthData = len(v.bytes)
		v.sLen = v.lengthData * 2
	}

	return v.sLen, nil
}

func (v *BinaryHexValue) FillCharactersBuffer(buffer []rune, offset int) error {
	v.GetCharactersLength()
	h := hex.EncodeToString(v.bytes)
	copy(buffer[offset:], []rune(h))
	return nil
}

func (v *BinaryHexValue) Equals(o Value) bool {
	if o == nil {
		return false
	}
	oi, ok := o.(*BinaryHexValue)
	if ok {
		return v.equals(oi.bytes)
	} else {
		s, err := o.ToString()
		if err != nil {
			return false
		}
		bv := BinaryHexValueParse(s)
		if bv != nil {
			return v.equals(bv.bytes)
		} else {
			return false
		}
	}
}

/*
	BooleanValue implentation
*/

var (
	BooleanValueFalse = newBooleanValue(false)
	BooleanValueTrue  = newBooleanValue(true)
	BooleanValue0     = newBooleanValueForID(0)
	BooleanValue1     = newBooleanValueForID(1)
	BooleanValue2     = newBooleanValueForID(2)
	BooleanValue3     = newBooleanValueForID(3)
)

type BooleanValue struct {
	*AbstractValue
	b          bool
	characters []rune
	sValue     string
}

func newBooleanValue(b bool) *BooleanValue {
	var characters []rune
	var sValue string

	if b {
		characters = DecodedBooleanTrueArray
		sValue = DecodedBooleanTrue
	} else {
		characters = DecodedBooleanFalseArray
		sValue = DecodedBooleanFalse
	}

	return &BooleanValue{
		AbstractValue: NewAbstractValue(ValueTypeBoolean),
		b:             b,
		characters:    characters,
		sValue:        sValue,
	}
}

func newBooleanValueForID(boolID int) *BooleanValue {
	var b bool
	var characters []rune
	var sValue string

	switch boolID {
	case 0:
		characters = XSDBooleanFalseArray
		sValue = XSDBooleanFalse
		b = false
	case 1:
		characters = XSDBoolean0Array
		sValue = XSDBoolean0
		b = false
	case 2:
		characters = XSDBooleanTrueArray
		sValue = XSDBooleanTrue
		b = true
	case 3:
		characters = XSDBoolean1Array
		sValue = XSDBoolean1
		b = true
	default:
		panic(fmt.Errorf("unknown boolID: %d", boolID))
	}

	return &BooleanValue{
		AbstractValue: NewAbstractValue(ValueTypeBoolean),
		b:             b,
		characters:    characters,
		sValue:        sValue,
	}
}

func GetBooleanValue(b bool) *BooleanValue {
	if b {
		return BooleanValueTrue
	} else {
		return BooleanValueFalse
	}
}

func GetBooleanValueForID(boolID int) (*BooleanValue, error) {
	switch boolID {
	case 0:
		return BooleanValue0, nil
	case 1:
		return BooleanValue1, nil
	case 2:
		return BooleanValue2, nil
	case 3:
		return BooleanValue3, nil
	default:
		return nil, fmt.Errorf("error while creating boolean pattern facet with boolID==%d", boolID)
	}
}

func BooleanValueParse(value string) *BooleanValue {
	value = strings.TrimSpace(value)
	switch value {
	case XSDBoolean0, XSDBooleanFalse:
		return BooleanValueFalse
	case XSDBoolean1, XSDBooleanTrue:
		return BooleanValueTrue
	default:
		return nil
	}
}

func (v *BooleanValue) ToBoolean() bool {
	return v.b
}

func (v *BooleanValue) GetCharacters() ([]rune, error) {
	return v.characters, nil
}

func (v *BooleanValue) FillCharactersBuffer(buffer []rune, offset int) error {
	if offset+len(v.characters) >= len(buffer) {
		return utils.ErrorIndexOutOfBounds
	}
	copy(buffer[offset:], v.characters)
	return nil
}

func (v *BooleanValue) GetCharactersLength() (int, error) {
	return len(v.characters), nil
}

func (v *BooleanValue) ToString() (string, error) {
	return v.sValue, nil
}

func (v *BooleanValue) BufferToString(buffer []rune, offset int) (string, error) {
	return v.sValue, nil
}

func (v *BooleanValue) Equals(o Value) bool {
	if o == nil {
		return false
	}
	oi, ok := o.(*BooleanValue)
	if ok {
		return v.b == oi.b
	} else {
		return false
	}
}

/*
	DateTimeValue implementation
*/

const (
	DateTimeValue_NumberBitsMonthDay      int = 9
	DateTimeValue_NumberBitsTime          int = 17
	DateTimeValue_NumberBitsTimeZone      int = 11
	DateTimeValue_YearOffset              int = 2000
	DateTimeValue_TimeZoneOffsetInMinutes int = 896
	DateTimeValue_SecondsInMinute         int = 64
	DateTimeValue_SecondsInHour           int = 64 * 64
	DateTimeValue_MonthMultiplicator      int = 32
)

// TODO: Check if day should be 0-30
type DateTimeValue struct {
	*AbstractValue
	kind                    DateTimeType
	year                    int
	monthDay                int
	time                    int
	presenceFractionalSecs  bool
	fractionalSecs          int
	presenceTimezone        bool
	timezone                int
	normalized              bool
	normalizedDateTimeValue *DateTimeValue
	sizeFractionalSecs      int
}

func NewDateTimeValue(kind DateTimeType, year, monthDay, time, fractionalSecs int, presenceTimezone bool, timezone int) *DateTimeValue {
	return NewDateTimeValueWithNormalized(kind, year, monthDay, time, fractionalSecs, presenceTimezone, timezone, false)
}

func NewDateTimeValueWithNormalized(kind DateTimeType, year, monthDay, time, fractionalSecs int, presenceTimezone bool, timezone int, normalized bool) *DateTimeValue {
	// Time: ((Hour * 64) + Minutes) * 64 + seconds
	// Canonical EXI: The Hour value MUST NOT be 24
	{
		hour := time / DateTimeValue_SecondsInHour
		if hour == 24 {
			time -= hour * DateTimeValue_SecondsInHour
			minute := time / DateTimeValue_SecondsInMinute
			time -= minute * DateTimeValue_SecondsInMinute // second

			// add one day / set hour to zero
			monthDay++
			hour = 0
			// adapt time
			time = ((hour*DateTimeValue_SecondsInMinute)+minute)*DateTimeValue_SecondsInMinute + time

			// month & day
			// e.g., 1999-12-31T24:00:00Z --> 2000-01-01T00:00:00Z
			month := monthDay / DateTimeValue_MonthMultiplicator
			//TODO: Check this
			//day := monthDay - (month * DateTimeValue_MonthMultiplicator)

			if month == 13 {
				year++
				month = 1
				day := 1
				monthDay = month*DateTimeValue_MonthMultiplicator + day
			}
		}
	}

	presence := false
	if fractionalSecs != 0 {
		presence = true
	}

	return &DateTimeValue{
		AbstractValue:           NewAbstractValue(ValueTypeDateTime),
		kind:                    kind,
		time:                    time,
		year:                    year,
		monthDay:                monthDay,
		fractionalSecs:          fractionalSecs,
		presenceFractionalSecs:  presence,
		presenceTimezone:        presenceTimezone,
		timezone:                timezone,
		normalized:              normalized,
		normalizedDateTimeValue: nil,
		sizeFractionalSecs:      -1,
	}
}

func (v *DateTimeValue) ToTime() (*time.Time, error) {
	t := time.Time{}

	switch v.kind {
	case DateTimeGYear:
		t = time.Date(v.year, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), nil)
	case DateTimeGYearMonth, DateTimeDate:
		t = time.Date(v.year, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), nil)
		t = dateTimeSetMonthDay(v.monthDay, t)
	case DateTimeDateTime:
		t = time.Date(v.year, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), nil)
		t = dateTimeSetMonthDay(v.monthDay, t)
		t = dateTimeSetTime(v.time, t)
		t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), v.fractionalSecs/1_000_000, nil)
	case DateTimeGMonth, DateTimeGMonthDay, DateTimeGDay:
		t = dateTimeSetMonthDay(v.monthDay, t)
	case DateTimeTime:
		t = dateTimeSetTime(v.time, t)
		t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), v.fractionalSecs/1_000_000, nil)
	default:
		return nil, fmt.Errorf("unsupported date time type: %d", v.kind)
	}
	t = dateTimeSetTimezone(v.timezone, t)

	return &t, nil
}

func (v *DateTimeValue) GetCharactersLength() (int, error) {
	if v.sLen == -1 {
		switch v.kind {
		case DateTimeGYear: // Year, [Time-Zone]
			if v.year < 0 {
				v.sLen = 5
			} else {
				v.sLen = 4
			}
		case DateTimeGYearMonth: // Year, MonthDay, [TimeZone]
			if v.year < 0 {
				v.sLen = 5
			} else {
				v.sLen = 4
			}
			v.sLen += 3
		case DateTimeDate: // Year, MonthDay, [TimeZone]
			if v.year < 0 {
				v.sLen = 5
			} else {
				v.sLen = 4
			}
			v.sLen += 6
		case DateTimeDateTime: // Year, MonthDay, Time, [FractionalSecs], [TimeZone]
			// e.g. "0001-01-01T00:00:00.111+00:33";
			if v.fractionalSecs == 0 {
				v.sizeFractionalSecs = 0
			} else {
				v.sizeFractionalSecs = utils.GetStringSize32(v.fractionalSecs) + 1
			}
			if v.year < 0 {
				v.sLen = 5
			} else {
				v.sLen = 4
			}
			v.sLen += 6 + 9 + v.sizeFractionalSecs
		case DateTimeGMonth: // MonthDay, [TimeZone]
			// e.g. "--12"
			v.sLen = 1 + 3
		case DateTimeGMonthDay: // MonthDay, [TimeZone]
			// e.g. "--01-28"
			v.sLen = 1 + 6
		case DateTimeGDay: // MonthDay, [TimeZone]
			// "---16";
			v.sLen = 3 + 2
		case DateTimeTime: // Time, [FractionalSecs], [TimeZone]
			if v.fractionalSecs == 0 {
				v.sizeFractionalSecs = 0
			} else {
				v.sizeFractionalSecs = utils.GetStringSize32(v.fractionalSecs) + 1
			}
			v.sLen = 8 + v.sizeFractionalSecs
		default:
			return -1, fmt.Errorf("unsupported date time kind: %d", v.kind)
		}

		// [TimeZone]
		if v.presenceTimezone {
			if v.timezone == 0 {
				v.sLen += 1
			} else {
				v.timezone += 6
			}
		}
	}

	return v.sLen, nil
}

func (v *DateTimeValue) FillCharactersBuffer(buffer []rune, offset int) error {
	switch v.kind {
	case DateTimeGYear: // Year, [Time-Zone]
		dateTimeAppendYear(buffer, &offset, v.year)
	case DateTimeGYearMonth: // Year, MonthDay, [TimeZone]
		dateTimeAppendYear(buffer, &offset, v.year)
		dateTimeAppendMonth(buffer, &offset, v.monthDay)
	case DateTimeDate: // Year, MonthDay, [TimeZone]
		dateTimeAppendYear(buffer, &offset, v.year)
		dateTimeAppendMonth(buffer, &offset, v.monthDay)
	case DateTimeDateTime: // Year, MonthDay, Time, [FractionalSecs], [TimeZone]
		// e.g. "0001-01-01T00:00:00.111+00:33";
		dateTimeAppendYear(buffer, &offset, v.year)
		dateTimeAppendMonth(buffer, &offset, v.monthDay)
		buffer[offset] = 'T'
		offset++
		dateTimeAppendTime(buffer, &offset, v.time)
		dateTimeAppendFractionalSeconds(buffer, &offset, v.fractionalSecs, v.sizeFractionalSecs-1)
	case DateTimeGMonth: // MonthDay, [TimeZone]
		// e.g. "--12"
		buffer[offset] = '-'
		offset++
		dateTimeAppendMonth(buffer, &offset, v.monthDay)
	case DateTimeGMonthDay: // MonthDay, [TimeZone]
		// e.g. "--01-28"
		buffer[offset] = '-'
		offset++
		dateTimeAppendMonthDay(buffer, &offset, v.monthDay)
	case DateTimeGDay: // MonthDay, [TimeZone]
		// "---16";
		buffer[offset] = '-'
		offset++
		buffer[offset] = '-'
		offset++
		buffer[offset] = '-'
		offset++
		dateTimeAppendDay(buffer, &offset, v.monthDay)
	case DateTimeTime: // Time, [FractionalSecs], [TimeZone]
		// e.g. "12:34:56.135"
		dateTimeAppendTime(buffer, &offset, v.time)
		dateTimeAppendFractionalSeconds(buffer, &offset, v.fractionalSecs, v.sizeFractionalSecs-1)
	default:
		return fmt.Errorf("unsupported date time type: %d", v.kind)
	}

	// [TimeZone]
	if v.presenceTimezone {
		dateTimeAppendTimezone(buffer, &offset, v.timezone)
	}

	return nil
}

func (v *DateTimeValue) Normalize() *DateTimeValue {
	if v.normalized {
		return v
	}
	if v.normalizedDateTimeValue == nil {
		v.normalizedDateTimeValue = v.doNormalize()
	}

	return v.normalizedDateTimeValue
}

func (v *DateTimeValue) doNormalize() *DateTimeValue {
	// year & month & day
	year := v.year
	month := v.monthDay / DateTimeValue_MonthMultiplicator
	day := v.monthDay - (month * DateTimeValue_MonthMultiplicator)
	// time
	hour := v.time / DateTimeValue_SecondsInHour
	time := v.time
	time -= hour * DateTimeValue_SecondsInHour
	minutes := time / DateTimeValue_SecondsInMinute
	seconds := time - minutes*DateTimeValue_SecondsInMinute

	// start Algorithm with not touching seconds to support leap-seconds
	// https://www.w3.org/TR/2004/REC-xmlschema-2-20041028/#adding-durations-to-dateTimes
	// if(seconds > 59) {
	// seconds -= 60; // remove one minute
	// minutes++; // adds one minute
	// }
	// if(minutes > 59) {
	// minutes -= 60; // remove an hour
	// hour++; // add one hour
	// }

	// timezone, per default 'Z'
	tzMinutes := 0
	tzHours := 0
	if v.presenceTimezone && v.timezone != 0 {
		tz := v.timezone // +/-
		// hours
		tzHours = tz / 64
		// minutes
		tzMinutes = tz - (tzHours * 64)
	}

	negate := -1

	// Minutes temp := S[minute] + D[minute] + carry E[minute] :=
	// modulo(temp, 60) carry := fQuotient(temp, 60)
	temp := minutes + negate*tzMinutes
	minutes = utils.Modulo(temp, 60)
	carry := utils.FQuotient(temp, 60)

	// Hours temp := S[hour] + D[hour] + carry E[hour] := modulo(temp, 24)
	// carry := fQuotient(temp, 24)
	temp = hour + negate*tzHours + carry
	hour = utils.Modulo(temp, 24)
	carry = utils.FQuotient(temp, 24)

	// Days
	var tempDays int

	if day > utils.MaximumDayInMonth(year, month) { // if S[day] > maximumDayInMonthFor(E[year], E[month])
		tempDays = utils.MaximumDayInMonth(year, month)
	} else if day < 1 { // else if S[day] < 1
		tempDays = 1
	} else {
		tempDays = day
	}

	// E[day] := tempDays + D[day] + carry
	day = tempDays + carry

	for {
		if day < 1 {
			day = day + utils.MaximumDayInMonth(year, month-1)
			carry = -1
		} else if day > utils.MaximumDayInMonth(year, month) {
			day = day - utils.MaximumDayInMonth(year, month)
			carry = 1
		} else {
			break
		}
		temp = month + carry
		month = utils.ModuloLoHi(temp, 1, 13)
		year = year + utils.FQuotientLoHi(temp, 1, 13)
	}

	// create new DateTimeValue
	monthDay := month*32 + day              // Month * 32 + Day
	time = ((hour*64)+minutes)*64 + seconds // ((Hour * 64) + Minutes) * 64 + seconds

	// TODO add timezone ONLY if timezone was present before
	presenceTimezone := v.presenceTimezone
	timezone := 0

	return NewDateTimeValueWithNormalized(v.kind, year, monthDay, time, v.fractionalSecs, presenceTimezone, timezone, true)
}

func (v *DateTimeValue) equals(o *DateTimeValue) bool {
	if o == nil {
		return false
	}
	ret := true
	if v.kind == o.kind && v.year == o.year && v.monthDay == o.monthDay && v.time == o.time {
		if v.presenceFractionalSecs == o.presenceFractionalSecs {
			if v.fractionalSecs != o.fractionalSecs {
				ret = false
			}
		}
		if ret && v.presenceTimezone == o.presenceTimezone {
			if v.timezone != o.timezone {
				ret = false
			}
		}
	} else {
		ret = false
	}

	if ret {
		// easy match
		return ret
	} else {
		// normalize both (if not already)
		if v.normalized && o.normalized {
			// not equal
		} else {
			tn := v.Normalize()
			on := o.Normalize()

			ret = tn.equals(on)
		}
	}

	return ret
}

func (v *DateTimeValue) Equals(o Value) bool {
	if o == nil {
		return false
	}
	oi, ok := o.(*DateTimeValue)
	if ok {
		return v.equals(oi)
	} else {
		s, err := o.ToString()
		if err != nil {
			return false
		}
		bv, err := DateTimeParse(s, v.kind)
		if err != nil {
			return false
		}
		if bv != nil {
			return v.equals(bv)
		} else {
			return false
		}
	}
}

/*
	DecimalValue implementation
*/

var (
	BigDecimalOne = new(apd.Decimal).SetInt64(1)
)

type DecimalValue struct {
	*AbstractValue
	negative      bool
	integral      *IntegerValue
	revFractional *IntegerValue
	bval          *apd.Decimal
}

func NewDecimalValue(negative bool, integral, revFractional *IntegerValue) *DecimalValue {
	if negative && ZeroIntegerValue.Equals(integral) && ZeroIntegerValue.Equals(revFractional) {
		negative = false
	}
	// normalize "-0.0" to "0.0"
	return &DecimalValue{
		AbstractValue: NewAbstractValue(ValueTypeDecimal),
		negative:      negative,
		integral:      integral,
		revFractional: revFractional,
	}
}

func DecimalValueParseBig(decimal *apd.Decimal) (*DecimalValue, error) {
	// e.g, -1.30
	negative := decimal.Sign() == -1
	if negative {
		decimal = decimal.Neg(decimal)
	}

	ctx := apd.BaseContext

	// Fractional part (e.g, 0.30)
	fractional := &apd.Decimal{}
	_, err := ctx.Rem(fractional, decimal, BigDecimalOne)
	if err != nil {
		return nil, err
	}
	fracS := fractional.String()
	revFractional, err := IntegerValueParse(utils.ReverseString(fracS[2:]))
	if err != nil {
		return nil, err
	}

	// integral part
	integral := &apd.Decimal{}
	_, err = ctx.Sub(integral, decimal, fractional)
	if err != nil {
		return nil, err
	}

	bint, err := utils.TryBigInt(integral)
	if err != nil {
		return nil, err
	}

	return NewDecimalValue(negative, IntegerValueOfBig(*bint), revFractional), nil
}

func DecimalValueParseString(decimal string) (*DecimalValue, error) {
	sNegative := false
	var sIntegral, sRevFractional *IntegerValue
	var err error
	decimal = strings.TrimSpace(decimal)

	if len(decimal) < 1 {
		return nil, utils.ErrorIndexOutOfBounds
	}
	switch decimal[0] {
	case '-':
		sNegative = true
		decimal = decimal[1:]
	case '+':
		decimal = decimal[1:]
	}

	decPoint := strings.Index(decimal, ".")

	switch decPoint {
	case -1:
		// no decimal point at all
		sIntegral, err = IntegerValueParse(decimal)
		if err != nil {
			return nil, err
		}
		sRevFractional = ZeroIntegerValue
	case 0:
		if decPoint+1 >= len(decimal) {
			return nil, utils.ErrorIndexOutOfBounds
		}
		// e.g. ".234"
		sIntegral = ZeroIntegerValue
		sRevFractional, err = IntegerValueParse(utils.ReverseString(decimal[decPoint+1:]))
		if err != nil {
			return nil, err
		}
	default:
		if decPoint+1 >= len(decimal) {
			return nil, utils.ErrorIndexOutOfBounds
		}
		sIntegral, err = IntegerValueParse(decimal[:decPoint])
		if err != nil {
			return nil, err
		}
		sRevFractional, err = IntegerValueParse(utils.ReverseString(decimal[decPoint+1:]))
		if err != nil {
			return nil, err
		}
	}

	if sIntegral == nil || sRevFractional == nil {
		return nil, nil
	} else {
		return NewDecimalValue(sNegative, sIntegral, sRevFractional), nil
	}
}

func (v *DecimalValue) IsNegative() bool {
	return v.negative
}

func (v *DecimalValue) GetIntegral() *IntegerValue {
	return v.integral
}

func (v *DecimalValue) GetRevFractional() *IntegerValue {
	return v.revFractional
}

func (v *DecimalValue) ToBigDecimal() (*apd.Decimal, error) {
	if v.bval == nil {
		len, err := v.GetCharactersLength()
		if err != nil {
			return nil, err
		}
		characters := make([]rune, len)
		err = v.FillCharactersBuffer(characters, 0)
		if err != nil {
			return nil, err
		}
		v.bval, _, err = apd.NewFromString(string(characters))
		if err != nil {
			return nil, err
		}
	}

	return v.bval, nil
}

func (v *DecimalValue) GetCharactersLength() (int, error) {
	if v.sLen == -1 {
		// +12.34
		iLen, err := v.integral.GetCharactersLength()
		if err != nil {
			return -1, err
		}
		revLen, err := v.revFractional.GetCharactersLength()
		if err != nil {
			return -1, err
		}

		if v.negative {
			v.sLen = 1
		}
		v.sLen = iLen + 1 + revLen
	}

	return v.sLen, nil
}

func (v *DecimalValue) FillCharactersBuffer(buffer []rune, offset int) error {
	// negative
	if v.negative {
		buffer[offset] = '-'
		offset++
	}

	// integral
	err := v.integral.FillCharactersBuffer(buffer, offset)
	if err != nil {
		return err
	}
	len, err := v.integral.GetCharactersLength()
	if err != nil {
		return err
	}
	offset += len

	// dot
	buffer[offset] = '.'
	offset++

	// fractional
	switch v.revFractional.GetIntegerValueType() {
	case IntegerValue32:
		utils.ItosReverse32(v.revFractional.ival, &offset, buffer)
	case IntegerValue64:
		utils.ItosReverse64(v.revFractional.lval, &offset, buffer)
	case IntegerValueBig:
		sb := Text.NewStringBuilderFromString(v.revFractional.bval.String())
		s := sb.Reverse().ToString()

		copy(buffer[offset:], []rune(s))
	default:
		return fmt.Errorf("unknown integer value type: %d", v.revFractional.GetIntegerValueType())
	}

	return nil
}

func (v *DecimalValue) equals(o *DecimalValue) bool {
	if o == nil {
		return false
	}
	return v.negative == o.negative &&
		v.integral.equals(o.integral) &&
		v.revFractional.equals(o.revFractional)
}

func (v *DecimalValue) Equals(o Value) bool {
	if o == nil {
		return false
	}
	dv, ok := o.(*DecimalValue)
	if !ok {
		return false
	}
	return v.equals(dv)
}

/*
	FloatValue implementation
*/

var (
	FloatValueSpecialValues = IntegerValueOf32(FloatSpecialValues)
	FloatNegativeInfinity   = IntegerValueOf32(-1)
	FloatPositiveInfinity   = IntegerValueOf32(1)
	FloatNaN                = ZeroIntegerValue
)

type FloatValue struct {
	*AbstractValue
	mantissa     *IntegerValue
	exponent     *IntegerValue
	slenMantissa int
	f            *float64
}

func NewFloatValue(mantissa, exponent *IntegerValue) *FloatValue {
	// http://www.w3.org/TR/exi-c14n/#dt-float
	if ZeroIntegerValue.Equals(mantissa) {
		// If the mantissa is 0 and the exponent value is not -(2^14)
		// to indicate one of the special values then the exponent MUST be 0.
		if !FloatValueSpecialValues.Equals(exponent) {
			exponent = ZeroIntegerValue
		}
	} else {
		// If the mantissa is not 0, mantissas MUST have no trailing zeros
		// e.g., 12300E0 --> 123E2
		lm := mantissa.Value64()
		le := exponent.Value64()

		modified := false

		for lm%10 == 0 {
			// multiple of 10
			lm /= 10
			le++
			modified = true
		}

		if modified {
			mantissa = IntegerValueOf64(lm)
			exponent = IntegerValueOf64(le)
		}
	}

	// If the exponent value is -(2^14) and the mantissa value is neither 1
	// nor -1, to indicate the special value not-a-number (NaN), the
	// mantissa MUST be 0.
	if FloatValueSpecialValues.Equals(exponent) &&
		!FloatNegativeInfinity.Equals(mantissa) &&
		!FloatPositiveInfinity.Equals(mantissa) {
		mantissa = FloatNaN // 0
	}

	return &FloatValue{
		AbstractValue: NewAbstractValue(ValueTypeFloat),
		mantissa:      mantissa,
		exponent:      exponent,
		slenMantissa:  -1,
	}

}

func NewFloatValueFrom64(mantissa, exponent int64) *FloatValue {
	return NewFloatValue(IntegerValueOf64(mantissa), IntegerValueOf64(exponent))
}

func FloatValueParseString(value string) (*FloatValue, error) {
	var sMantissa, sExponent int64
	value = strings.TrimSpace(value)

	if len(value) == 0 {
		return nil, fmt.Errorf("empty string")
	} else if value == FloatInfinity {
		sMantissa = int64(FloatMantissaInfinity)
		sExponent = int64(FloatSpecialValues)
	} else if value == FloatMinusInfinity {
		sMantissa = int64(FloatMantissaMinusInfinity)
		sExponent = int64(FloatSpecialValues)
	} else if value == FloatNotANumber {
		sMantissa = int64(FloatMantissaNotANumber)
		sExponent = int64(FloatSpecialValues)
	} else {
		indexE := strings.Index(value, "E")
		if indexE == -1 {
			indexE = strings.Index(value, "e")
		}

		var c rune
		chars := []rune(value)

		// detecting sign
		c = chars[0]
		negative := (c == '-')

		lenMantissa := indexE
		if indexE == -1 {
			lenMantissa = len(chars)
		}
		startMantissa := 0
		if negative || c == '+' {
			startMantissa = 1
		}

		decPoint := false
		decimalDigits := 0
		sMantissa = 0
		sExponent = 0

		// invalid floats
		if lenMantissa == 0 {
			return nil, fmt.Errorf("mantissa length is zero")
		}
		if lenMantissa >= len(chars) {
			return nil, fmt.Errorf("out of bounds")
		}

		// parsing mantissa
		for i := startMantissa; i < lenMantissa; i++ {
			c = chars[i]

			switch c {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				sMantissa = 10*sMantissa + int64(c-'0')
				if decPoint {
					decimalDigits++
				}
			case '.':
				if decPoint {
					// decimal point twice
					return nil, fmt.Errorf("multiple decimal points")
				}
				decPoint = true
			default:
				return nil, fmt.Errorf("unexpected character in mantissa: %c", c)
			}
		}

		// check for mantissa overflow
		if sMantissa < 0 {
			if negative {
				if sMantissa != math.MinInt64 {
					return nil, fmt.Errorf("mantissa overflow")
				}
			} else {
				return nil, fmt.Errorf("mantissa overflow")
			}
		}

		if negative {
			sMantissa = (-1) * sMantissa
		}

		// parsing exponent
		negativeExp := false
		if indexE != -1 {
			for i := indexE + 1; i < len(chars); i++ {
				c = chars[i]

				switch c {
				case '0':
					sExponent = 10 * sExponent
				case '1', '2', '3', '4', '5', '6', '7', '8', '9':
					sExponent = 10*sExponent + int64(c-'0')
				case '-':
					if negativeExp {
						return nil, fmt.Errorf("multiple exponent sign")
					}
					negativeExp = true
				case '+':
					// skip
				default:
					return nil, fmt.Errorf("unexpected character in exponent: %c", c)
				}
			}
		}

		if negativeExp {
			sExponent = (-1) * sExponent
		}
		sExponent -= int64(decimalDigits)

		// too large ranges
		if sMantissa < FloatMantissaMinRange || sMantissa > FloatMantissaMaxRange ||
			sExponent < FloatExponentMinRange || sExponent > FloatExponentMaxRange {
			return nil, fmt.Errorf("out of range")
		}
	}

	return NewFloatValueFrom64(sMantissa, sExponent), nil
}

func FloatValueParseFloat32(value float32) *FloatValue {
	var sMantissa, sExponent int

	if math.IsInf(float64(value), 0) || math.IsNaN(float64(value)) {
		// exponent value is -(2^14),
		// . the mantissa value 1 represents INF,
		// . the mantissa value -1 represents -INF
		// . any other mantissa value represents NaN
		if math.IsNaN(float64(value)) {
			sMantissa = FloatMantissaNotANumber
		} else if value < 0 {
			sMantissa = FloatMantissaMinusInfinity
		} else {
			sMantissa = FloatMantissaInfinity
		}
		// exponent (special value)
		sExponent = FloatSpecialValues // e == -(2^14)
	} else {
		// floating-point according to the IEEE 754 floating-point
		// "single format" bit layout.
		sExponent = 0

		for value-float32(math.Trunc(float64(value))) != 0.0 {
			value *= 10
			sExponent--
		}
		sMantissa = int(math.Trunc(float64(value)))
	}

	return NewFloatValueFrom64(int64(sMantissa), int64(sExponent))
}

func FloatValueParseFloat64(value float64) *FloatValue {
	var sMantissa, sExponent int64

	if math.IsInf(value, 0) || math.IsNaN(value) {
		// exponent value is -(2^14),
		// . the mantissa value 1 represents INF,
		// . the mantissa value -1 represents -INF
		// . any other mantissa value represents NaN
		if math.IsNaN(value) {
			sMantissa = int64(FloatMantissaNotANumber)
		} else if value < 0 {
			sMantissa = int64(FloatMantissaMinusInfinity)
		} else {
			sMantissa = int64(FloatMantissaInfinity)
		}
		// exponent (special value)
		sExponent = int64(FloatSpecialValues) // e == -(2^14)
	} else {
		// floating-point according to the IEEE 754 floating-point
		// "single format" bit layout.
		sExponent = 0

		for value-math.Trunc(value) != 0.0 {
			value *= 10
			sExponent--
		}
		sMantissa = int64(math.Trunc(value))
	}

	return NewFloatValueFrom64(sMantissa, sExponent)
}

func (v *FloatValue) GetMantissa() *IntegerValue {
	return v.mantissa
}

func (v *FloatValue) GetExponent() *IntegerValue {
	return v.exponent
}

func (v *FloatValue) ToFloat32() float32 {
	if v.f == nil {
		v.ToFloat64()
	}

	return float32(*v.f)
}

func (v *FloatValue) ToFloat64() float64 {
	if v.f == nil {
		if v.exponent.Equals(FloatValueSpecialValues) {
			if v.mantissa.Equals(FloatNegativeInfinity) {
				v.f = utils.AsPtr(math.Inf(-1))
			} else if v.mantissa.Equals(FloatPositiveInfinity) {
				v.f = utils.AsPtr(math.Inf(+1))
			} else {
				v.f = utils.AsPtr(math.NaN())
			}
		} else {
			// f = mantissa * (double) (Math.pow(10, exponent));
			lMantissa := v.mantissa.Value64()
			lExponent := v.exponent.Value64()

			v.f = utils.AsPtr(float64(lMantissa) * math.Pow(10, float64(lExponent)))
		}
	}

	return *v.f
}

func (v *FloatValue) GetCharactersLength() (int, error) {
	if v.sLen == -1 {
		if v.exponent.Equals(FloatValueSpecialValues) {
			if v.mantissa.Equals(FloatNegativeInfinity) {
				v.sLen = len(FloatMinusInfinityCharArray)
			} else if v.mantissa.Equals(FloatPositiveInfinity) {
				v.sLen = len(FloatInfinityCharArray)
			} else {
				if v.mantissa.Equals(FloatNaN) {
					return -1, fmt.Errorf("NaN")
				}
				v.sLen = len(FloatNotANumberCharArray)
			}
		} else {
			// iMantissa + "E" + iExponent
			slenMantissa, err := v.mantissa.GetCharactersLength()
			if err != nil {
				return -1, err
			}
			slenExponent, err := v.exponent.GetCharactersLength()
			if err != nil {
				return -1, err
			}

			v.slenMantissa = slenMantissa
			v.sLen = slenMantissa + 1 + slenExponent
		}
	}

	return v.sLen, nil
}

func (v *FloatValue) FillCharactersBuffer(buffer []rune, offset int) error {
	if _, err := v.GetCharactersLength(); err != nil {
		return err
	}

	if v.exponent.Equals(FloatValueSpecialValues) {
		var a2copy []rune
		if v.mantissa.Equals(FloatNegativeInfinity) {
			a2copy = FloatMinusInfinityCharArray
		} else if v.mantissa.Equals(FloatPositiveInfinity) {
			a2copy = FloatInfinityCharArray
		} else {
			if v.mantissa.Equals(FloatNaN) {
				return fmt.Errorf("NaN")
			}
			a2copy = FloatNotANumberCharArray
		}
		copy(buffer[offset:], a2copy)
	} else {
		if err := v.mantissa.FillCharactersBuffer(buffer, offset); err != nil {
			return err
		}
		offset += +v.slenMantissa
		buffer[offset] = 'E'
		offset++
		if err := v.exponent.FillCharactersBuffer(buffer, offset); err != nil {
			return err
		}
	}

	return nil
}

func (v *FloatValue) ToString() (string, error) {
	if v.exponent.Equals(FloatValueSpecialValues) {
		if v.mantissa.Equals(FloatNegativeInfinity) {
			return FloatMinusInfinity, nil
		} else if v.mantissa.Equals(FloatPositiveInfinity) {
			return FloatInfinity, nil
		} else {
			if v.mantissa.Equals(FloatNaN) {
				return "", fmt.Errorf("NaN")
			}
			return FloatNotANumber, nil
		}
	} else {
		len, err := v.GetCharactersLength()
		if err != nil {
			return "", err
		}
		buffer := make([]rune, len)
		err = v.FillCharactersBuffer(buffer, 0)
		if err != nil {
			return "", err
		}
		return string(buffer), nil
	}
}

func (v *FloatValue) BufferToString(buffer []rune, offset int) (string, error) {
	if v.exponent.Equals(FloatValueSpecialValues) {
		if v.mantissa.Equals(FloatNegativeInfinity) {
			return FloatMinusInfinity, nil
		} else if v.mantissa.Equals(FloatPositiveInfinity) {
			return FloatInfinity, nil
		} else {
			if v.mantissa.Equals(FloatNaN) {
				return "", fmt.Errorf("NaN")
			}
			return FloatNotANumber, nil
		}
	} else {
		return v.AbstractValue.BufferToString(buffer, offset)
	}
}

func (v *FloatValue) multiply(a, b int64) (int64, error) {
	result := a * b
	if a != 0 && b != result/a {
		return -1, fmt.Errorf("overflow")
	}
	return result, nil
}

func (v *FloatValue) equals(o *FloatValue) bool {
	// e.g. 10E-1 vs. 1000E-3
	if v.mantissa == o.mantissa && v.exponent == o.exponent {
		return true
	} else {
		if v.mantissa.Equals(o.mantissa) && v.exponent.Equals(o.exponent) {
			return true
		} else {
			tExponent := v.exponent.Value64()
			oExponent := o.exponent.Value64()
			tMantissa := v.mantissa.Value64()
			oMantissa := o.exponent.Value64()

			if tExponent > oExponent {
				// e.g. 234E2 vs. 2340E1
				diff := tExponent - oExponent
				for i := int64(0); i < diff; i++ {
					// tMantissa *= 10
					m, err := v.multiply(tMantissa, 10)
					if err != nil {
						return false
					}
					tMantissa = m
				}
			} else {
				// e.g. 30E0 vs. 3E1
				diff := oExponent - tExponent
				for i := int64(0); i < diff; i++ {
					// oMantissa *= 10
					m, err := v.multiply(oMantissa, 10)
					if err != nil {
						return false
					}
					oMantissa = m
				}
			}

			return tMantissa == oMantissa
		}
	}
}

func (v *FloatValue) Equals(o Value) bool {
	if o == nil {
		return false
	}
	val, ok := o.(*FloatValue)
	if !ok {
		return false
	}
	return v.equals(val)
}

/*
	IntegerValue implementation
*/

var (
	ZeroIntegerValue = NewIntegerValue32(0)
	MinValue32       = big.NewInt(int64(math.MinInt32))
	MaxValue32       = big.NewInt(int64(math.MaxInt32))
	MinValue64       = big.NewInt(math.MinInt64)
	MaxValue64       = big.NewInt(math.MaxInt64)
)

type IntegerValueType int

const (
	IntegerValue32 = iota
	IntegerValue64
	IntegerValueBig
)

type IntegerValue struct {
	*AbstractValue
	ival     int
	lval     int64
	iValType IntegerValueType
	bval     *big.Int
}

func NewIntegerValue32(ival int) *IntegerValue {
	av := NewAbstractValue(ValueTypeInteger)
	iv := &IntegerValue{
		AbstractValue: av,
		ival:          ival,
		iValType:      IntegerValue32,
		lval:          0,
		bval:          nil,
	}
	av.Value = iv
	return iv
}

func NewIntegerValue64(lval int64) *IntegerValue {
	av := NewAbstractValue(ValueTypeInteger)
	iv := &IntegerValue{
		AbstractValue: av,
		ival:          0,
		iValType:      IntegerValue64,
		lval:          lval,
		bval:          nil,
	}
	return iv
}

func NewIntegerValueBig(bval big.Int) *IntegerValue {
	return &IntegerValue{
		AbstractValue: NewAbstractValue(ValueTypeInteger),
		ival:          0,
		iValType:      IntegerValueBig,
		lval:          0,
		bval:          &bval,
	}
}

func integerValueGetAdjustedValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 0 && value[0] == '+' {
		value = value[1:]
	}
	return value
}

func IntegerValueParse(value string) (*IntegerValue, error) {
	value = integerValueGetAdjustedValue(value)
	l := len(value)

	if l > 0 {
		if value[0] == '-' {
			if l < 11 {
				i, err := strconv.ParseInt(value, 10, 32)
				if err != nil {
					return nil, err
				}
				return NewIntegerValue32(int(i)), nil
			} else if l < 20 {
				l, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return nil, err
				}
				return NewIntegerValue64(l), nil
			} else {
				b := new(big.Int)
				b, ok := b.SetString(value, 10)
				if !ok {
					return nil, fmt.Errorf("not a number")
				}
				return NewIntegerValueBig(*b), nil
			}
		} else {
			if l < 10 {
				i, err := strconv.ParseInt(value, 10, 32)
				if err != nil {
					return nil, err
				}
				return NewIntegerValue32(int(i)), nil
			} else if l < 19 {
				l, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return nil, err
				}
				return NewIntegerValue64(l), nil
			} else {
				b := new(big.Int)
				b, ok := b.SetString(value, 10)
				if !ok {
					return nil, fmt.Errorf("not a number")
				}
				return NewIntegerValueBig(*b), nil
			}
		}
	} else {
		return nil, fmt.Errorf("not a number")
	}
}

func IntegerValueOf32(ival int) *IntegerValue {
	return NewIntegerValue32(ival)
}

func IntegerValueOf64(lval int64) *IntegerValue {
	if lval < math.MinInt32 || lval > math.MaxInt32 {
		return NewIntegerValue64(lval)
	} else {
		return NewIntegerValue32(int(lval))
	}
}

func IntegerValueOfBig(bval big.Int) *IntegerValue {
	// fits into int?
	if bval.Cmp(MinValue32) == -1 || bval.Cmp(MaxValue32) == 1 {
		// fits into long?
		if bval.Cmp(MinValue64) == -1 || bval.Cmp(MaxValue64) == 1 {
			return NewIntegerValueBig(bval)
		} else {
			return NewIntegerValue64(bval.Int64())
		}
	} else {
		return NewIntegerValue32(int(bval.Int64()))
	}
}

func (v *IntegerValue) GetIntegerValueType() IntegerValueType {
	return v.iValType
}

func (v *IntegerValue) GetCharactersLength() (int, error) {
	if v.sLen == -1 {
		switch v.iValType {
		case IntegerValue32:
			if v.ival == math.MinInt {
				v.sLen = len(utils.IntegerMinValueCharArray)
			} else {
				v.sLen = utils.GetStringSize32(v.ival)
			}
		case IntegerValue64:
			if v.lval == math.MinInt64 {
				v.sLen = len(utils.LongMinValueCharArray)
			} else {
				v.sLen = utils.GetStringSize64(v.lval)
			}
		case IntegerValueBig:
			v.sLen = len(v.bval.String())
		default:
			v.sLen = -1
		}
	}

	return v.sLen, nil
}

func (v *IntegerValue) FillCharactersBuffer(buffer []rune, offset int) error {
	switch v.iValType {
	case IntegerValue32:
		if v.ival == math.MinInt {
			copy(buffer[offset:], utils.IntegerMinValueCharArray)
		} else {
			length, err := v.GetCharactersLength()
			if err != nil {
				return err
			}
			if len(buffer) < length {
				return errors.New("buffer size is smaller than characters length")
			}
			i := offset + length
			utils.Itos32(v.ival, &i, buffer)
		}
	case IntegerValue64:
		if v.lval == math.MinInt64 {
			copy(buffer[offset:], utils.LongMinValueCharArray)
		} else {
			length, err := v.GetCharactersLength()
			if err != nil {
				return err
			}
			if len(buffer) < length {
				return errors.New("buffer size is smaller than characters length")
			}
			i := offset + length
			utils.Itos64(v.lval, &i, buffer)
		}
	case IntegerValueBig:
		sval := v.bval.String()
		copy(buffer, []rune(sval))
	default:
		// return nil
	}

	return nil
}

func (v *IntegerValue) Value32() int {
	switch v.iValType {
	case IntegerValue32:
		return v.ival
	case IntegerValue64:
		return int(v.lval)
	case IntegerValueBig:
		return int(v.bval.Int64())
	}
	panic(fmt.Errorf("unexpected integer value: %d", v.iValType))
}

func (v *IntegerValue) Value64() int64 {
	switch v.iValType {
	case IntegerValue32:
		return int64(v.ival)
	case IntegerValue64:
		return v.lval
	case IntegerValueBig:
		return v.bval.Int64()
	}
	panic(fmt.Errorf("unexpected integer value: %d", v.iValType))
}

func (v *IntegerValue) ValueBig() *big.Int {
	switch v.iValType {
	case IntegerValue32:
		return big.NewInt(int64(v.ival))
	case IntegerValue64:
		return big.NewInt(v.lval)
	case IntegerValueBig:
		return v.bval
	}
	panic(fmt.Errorf("unexpected integer value: %d", v.iValType))
}

func (v *IntegerValue) IsPositive() bool {
	switch v.iValType {
	case IntegerValue32:
		return v.ival >= 0
	case IntegerValue64:
		return v.lval >= 0
	case IntegerValueBig:
		return v.bval.Sign() >= 0
	}
	panic(fmt.Errorf("unexpected integer value: %d", v.iValType))
}

func (v *IntegerValue) Add(o *IntegerValue) *IntegerValue {
	if o.equals(ZeroIntegerValue) {
		return v
	}

	switch v.iValType {
	case IntegerValue32:
		switch o.iValType {
		case IntegerValue32:
			return NewIntegerValue32(v.ival + o.ival)
		case IntegerValue64:
			return IntegerValueOf64(int64(v.ival) + o.lval)
		case IntegerValueBig:
			val := big.NewInt(int64(v.ival))
			val = val.Add(val, o.bval)
			return IntegerValueOfBig(*val)
		default:
			return nil
		}
	case IntegerValue64:
		switch o.iValType {
		case IntegerValue32:
			return NewIntegerValue64(v.lval + int64(o.ival))
		case IntegerValue64:
			return IntegerValueOf64(v.lval + o.lval)
		case IntegerValueBig:
			val := big.NewInt(v.lval)
			val = val.Add(val, o.bval)
			return IntegerValueOfBig(*val)
		default:
			return nil
		}
	case IntegerValueBig:
		switch o.iValType {
		case IntegerValue32:
			val := big.NewInt(int64(o.ival))
			val = val.Add(v.bval, val)
			return NewIntegerValueBig(*val)
		case IntegerValue64:
			val := big.NewInt(o.lval)
			val = val.Add(v.bval, val)
			return IntegerValueOfBig(*val)
		case IntegerValueBig:
			tmp := *v.bval
			val := tmp.Add(v.bval, o.bval)
			return IntegerValueOfBig(*val)
		default:
			return nil
		}
	default:
		panic(fmt.Errorf("unexpected integer value: %d", v.iValType))
	}
}

func (v *IntegerValue) Sub(o *IntegerValue) *IntegerValue {
	if o.equals(ZeroIntegerValue) {
		return v
	}

	switch v.iValType {
	case IntegerValue32:
		switch o.iValType {
		case IntegerValue32:
			return NewIntegerValue32(v.ival - o.ival)
		case IntegerValue64:
			return IntegerValueOf64(int64(v.ival) - o.lval)
		case IntegerValueBig:
			val := big.NewInt(int64(v.ival))
			val = val.Sub(val, o.bval)
			return IntegerValueOfBig(*val)
		default:
			return nil
		}
	case IntegerValue64:
		switch o.iValType {
		case IntegerValue32:
			return NewIntegerValue64(v.lval - int64(o.ival))
		case IntegerValue64:
			return IntegerValueOf64(v.lval - o.lval)
		case IntegerValueBig:
			val := big.NewInt(v.lval)
			val = val.Sub(val, o.bval)
			return IntegerValueOfBig(*val)
		default:
			return nil
		}
	case IntegerValueBig:
		switch o.iValType {
		case IntegerValue32:
			val := big.NewInt(int64(o.ival))
			val = val.Sub(v.bval, val)
			return NewIntegerValueBig(*val)
		case IntegerValue64:
			val := big.NewInt(o.lval)
			val = val.Sub(v.bval, val)
			return IntegerValueOfBig(*val)
		case IntegerValueBig:
			tmp := *v.bval
			val := tmp.Sub(v.bval, o.bval)
			return IntegerValueOfBig(*val)
		default:
			return nil
		}
	default:
		panic(fmt.Errorf("unexpected integer value: %d", v.iValType))
	}
}

func (v *IntegerValue) Equals(o Value) bool {
	if o == nil {
		return false
	}
	val, ok := o.(*IntegerValue)
	if !ok {
		return false
	}
	iv, err := IntegerValueParse(val.String())
	if err != nil {
		return false
	}
	return v.equals(iv)
}

func (v *IntegerValue) equals(o *IntegerValue) bool {
	if o == nil || v.iValType != o.iValType {
		return false
	}

	switch v.iValType {
	case IntegerValue32:
		return v.ival == o.ival
	case IntegerValue64:
		return v.lval == o.lval
	case IntegerValueBig:
		return v.bval.Cmp(o.bval) == 0
	}

	return false
}

func (v *IntegerValue) String() string {
	switch v.iValType {
	case IntegerValue32:
		return strconv.FormatInt(int64(v.ival), 10)
	case IntegerValue64:
		return strconv.FormatInt(v.lval, 10)
	case IntegerValueBig:
		return v.bval.String()
	}
	panic(fmt.Errorf("unexpected integer value: %d", v.iValType))
}

/*
 * Returns a negative integer, zero, or a positive integer as this
 * object is less than, equal to, or greater than the specified object.
 */
func (v *IntegerValue) Cmp(o *IntegerValue) int {
	switch v.iValType {
	case IntegerValue32:
		switch o.iValType {
		case IntegerValue32:
			if v.ival == o.ival {
				return 0
			} else if v.ival < o.ival {
				return -1
			} else {
				return 1
			}
		case IntegerValue64, IntegerValueBig:
			return -1
		default:
			return -2
		}
	case IntegerValue64:
		switch o.iValType {
		case IntegerValue32:
			return 1
		case IntegerValue64:
			if v.lval == o.lval {
				return 0
			} else if v.lval < o.lval {
				return -1
			} else {
				return 1
			}
		case IntegerValueBig:
			return -1
		default:
			return -2
		}
	case IntegerValueBig:
		switch o.iValType {
		case IntegerValue32, IntegerValue64:
			return 1
		case IntegerValueBig:
			return v.bval.Cmp(o.bval)
		default:
			return -2
		}
	default:
		return -2
	}
}

/*
	ListValue implementation
*/

type ListValue struct {
	*AbstractValue
	values         []Value
	listDatatype   Datatype
	numberOfValues int
}

func NewListValue(values []Value, listDatatype Datatype) *ListValue {
	return &ListValue{
		AbstractValue:  NewAbstractValue(ValueTypeList),
		values:         values,
		listDatatype:   listDatatype,
		numberOfValues: len(values),
	}
}

func ListValueParse(value string, listDatatype Datatype) (*ListValue, error) {
	tokens := strings.Fields(value)
	values := make([]Value, len(tokens))
	index := 0

	for _, token := range tokens {
		next := NewStringValueFromString(token)
		encoder, err := NewTypedTypeEncoder(nil, nil, nil)
		if err != nil {
			return nil, err
		}

		valid, err := encoder.IsValid(listDatatype, next)
		if err != nil {
			return nil, err
		}

		if valid {
			values[index] = next
			index++
		} else {
			return nil, nil
		}
	}

	return NewListValue(values, listDatatype), nil
}

func (v *ListValue) GetNumberOfValues() int {
	return v.numberOfValues
}

func (v *ListValue) ToValues() []Value {
	return v.values
}

func (v *ListValue) GetListDatatype() Datatype {
	return v.listDatatype
}

func (v *ListValue) GetCharactersLength() (int, error) {
	if v.sLen == -1 {
		v.sLen = 0
		if len(v.values) > 0 {
			v.sLen = len(v.values) - 1
		}

		vlen := len(v.values)
		for i := 0; i < vlen; i++ {
			ilen, err := v.values[i].GetCharactersLength()
			if err != nil {
				return -1, err
			}
			v.sLen += ilen
		}
	}

	return v.sLen, nil
}

func (v *ListValue) FillCharactersBuffer(buffer []rune, offset int) error {
	if len(v.values) > 0 {
		// fill buffer (except last item)

		var iVal Value
		vlenMinus1 := len(v.values) - 1

		for i := 0; i < vlenMinus1; i++ {
			iVal = v.values[i]

			if err := iVal.FillCharactersBuffer(buffer, offset); err != nil {
				return err
			}
			ilen, err := iVal.GetCharactersLength()
			if err != nil {
				return err
			}

			offset += ilen
			buffer[offset] = XSDListDelimChar
			offset++
		}

		// last item (no delimiter)
		iVal = v.values[vlenMinus1]
		if err := iVal.FillCharactersBuffer(buffer, offset); err != nil {
			return err
		}
	}

	return nil
}

func (v *ListValue) equals(o *ListValue) bool {
	if o == nil {
		return false
	}

	// datatype
	if v.listDatatype.GetBuiltInType() != o.listDatatype.GetBuiltInType() {
		return false
	}

	// values
	if len(v.values) == len(o.values) {
		for i := 0; i < len(v.values); i++ {
			if !v.values[i].Equals(o.values[i]) {
				return false
			}
		}

		return true
	} else {
		return false
	}
}

func (v *ListValue) Equals(o Value) bool {
	if o == nil {
		return false
	}
	oi, ok := o.(*ListValue)
	if ok {
		return v.equals(oi)
	} else {
		s, err := o.ToString()
		if err != nil {
			return false
		}
		lv, err := ListValueParse(s, v.listDatatype)
		if err != nil {
			return false
		}
		if lv != nil {
			return v.equals(lv)
		} else {
			return false
		}
	}
}

/*
	QNameValue implementation
*/

type QNameValue struct {
	*AbstractValue
	namespaceURI string
	localName    string
	prefix       *string
	characters   *[]rune
	sValue       string
}

func NewQNameValue(namespaceURI, localName string, prefix *string) *QNameValue {
	var sValue string
	if prefix == nil || len(*prefix) == 0 {
		sValue = localName
	} else {
		sValue = utils.AsValue(prefix) + ":" + localName
	}

	return &QNameValue{
		AbstractValue: NewAbstractValue(ValueTypeQName),
		namespaceURI:  namespaceURI,
		localName:     localName,
		prefix:        prefix,
		characters:    nil,
		sValue:        sValue,
	}
}

func (v *QNameValue) GetNamespaceURI() string {
	return v.namespaceURI
}

func (v *QNameValue) GetLocalName() string {
	return v.localName
}

func (v *QNameValue) GetPrefix() *string {
	return v.prefix
}

func (v *QNameValue) GetCharactersLength() (int, error) {
	return len(v.sValue), nil
}

func (v *QNameValue) FillCharactersBuffer(buffer []rune, offset int) error {
	if v.characters == nil {
		v.characters = utils.AsPtr([]rune(v.sValue))
		copy(buffer[offset:], *v.characters)
	}

	return nil
}

func (v *QNameValue) ToString() (string, error) {
	return v.sValue, nil
}

func (v *QNameValue) BufferToString(buffer []rune, offset int) (string, error) {
	return v.sValue, nil
}

func (v *QNameValue) Equals(o Value) bool {
	if o == nil {
		return false
	}
	oi, ok := o.(*QNameValue)
	if ok {
		return v.namespaceURI == oi.namespaceURI && v.localName == oi.localName
	} else {
		return false
	}
}

/*
	StringValue implementation
*/

type StringValue struct {
	*AbstractValue
	characters *[]rune
	sValue     *string
}

func NewStringValueFromSlice(ch []rune) *StringValue {
	return &StringValue{
		AbstractValue: NewAbstractValue(ValueTypeString),
		characters:    &ch,
		sValue:        nil,
	}
}

func NewStringValueFromString(s string) *StringValue {
	return &StringValue{
		AbstractValue: NewAbstractValue(ValueTypeString),
		characters:    nil,
		sValue:        &s,
	}
}

func (v *StringValue) checkCharacters() {
	if v.characters == nil {
		v.characters = utils.AsPtr([]rune(*v.sValue))
	}
}

func (v *StringValue) GetCharactersLength() (int, error) {
	v.checkCharacters()
	return len(*v.characters), nil
}

func (v *StringValue) GetCharacters() ([]rune, error) {
	v.checkCharacters()
	return *v.characters, nil
}

func (v *StringValue) FillCharactersBuffer(buffer []rune, offset int) error {
	if offset+len(*v.characters) > len(buffer) {
		return utils.ErrorIndexOutOfBounds
	}

	v.checkCharacters()
	copy(buffer[offset:], *v.characters)
	return nil
}

func (v *StringValue) ToString() (string, error) {
	if v.sValue == nil {
		v.sValue = utils.AsPtr(string(*v.characters))
	}
	return *v.sValue, nil
}

func (v *StringValue) BufferToString(buffer []rune, offset int) (string, error) {
	return v.ToString()
}

func (v *StringValue) Equals(o Value) bool {
	if o == nil {
		return false
	}
	if v == o {
		return true
	} else {
		vs, err := v.ToString()
		if err != nil {
			return false
		}

		os, err := o.ToString()
		if err != nil {
			return false
		}

		return vs == os
	}
}
