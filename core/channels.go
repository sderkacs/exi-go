package core

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/sderkacs/exi-go/utils"
)

const (
	/* long == 64 bits, 9 * 7bits = 63 bits */
	MaxOctetsForLong int = 9
)

type DecoderChannel interface {
	// Decodes a single byte
	Decode() (int, error)

	// Align to next byte-aligned boundary in the stream if it is not
	// already at such a boundary
	Align() error

	// Skips over and discards <code>n</code> bytes of data from this channel.
	Skip(n int64) error

	// Decodes and returns an n-bit unsigned integer.
	DecodeNBitUnsignedInteger(n int) (int, error)
	DecodeNBitUnsignedIntegerValue(n int) (*IntegerValue, error)

	// Decode a single boolean value. The value false is represented by the bit
	// (byte) 0, and the value true is represented by the bit (byte) 1.
	DecodeBoolean() (bool, error)
	DecodeBooleanValue() (*BooleanValue, error)

	// Decode a binary value as a length-prefixed sequence of octets.
	DecodeBinary() ([]byte, error)

	// Decode a string as a length-prefixed sequence of UCS codepoints, each of
	// which is encoded as an integer.
	DecodeString() ([]rune, error)

	// Decode the characters of a string whose length has already been read.
	DecodeStringOnly(length int) ([]rune, error)

	// Decode an arbitrary precision non negative integer using a sequence of
	// octets. The most significant bit of the last octet is set to zero to
	// indicate sequence termination. Only seven bits per octet are used to
	// store the integer's value.
	DecodeUnsignedInteger() (int, error)
	DecodeUnsignedIntegerValue() (*IntegerValue, error)

	// Decode an arbitrary precision integer using a sign bit followed by a
	// sequence of octets. The most significant bit of the last octet is set to
	// zero to indicate sequence termination. Only seven bits per octet are used
	// to store the integer's value.
	DecodeIntegerValue() (*IntegerValue, error)

	// Decode a decimal represented as a Boolean sign followed by two Unsigned
	// Integers. A sign value of zero (0) is used to represent positive Decimal
	// values and a sign value of one (1) is used to represent negative Decimal
	// values The first Integer represents the integral portion of the Decimal
	// value. The second positive integer represents the fractional portion of
	// the decimal with the digits in reverse order to preserve leading zeros.
	DecodeDecimalValue() (*DecimalValue, error)

	// Decode a Float represented as two consecutive Integers. The first Integer
	// represents the mantissa of the floating point number and the second
	// Integer represents the 10-based exponent of the floating point number.
	DecodeFloatValue() (*FloatValue, error)

	// Decode Date-Time as sequence of values representing the individual
	// components of the Date-Time.
	DecodeDateTimeValue(kind DateTimeType) (*DateTimeValue, error)
}

type EncoderChannel interface {
	//TODO: GetOutputStream() ?
	Flush() error

	// Returns the number of bytes written.
	GetLength() int

	// Align to next byte-aligned boundary in the stream if it is not
	// already at such a boundary
	Align() error
	EncodeInt(b int) error
	EncodeBytes(b []byte, offset, length int) error
	EncodeNBitUnsignedInteger(b, n int) error

	// Encode a single boolean value. A false value is encoded as bit (byte) 0
	// and true value is encode as bit (byte) 1.
	EncodeBoolean(b bool) error

	// Encode a binary value as a length-prefixed sequence of octets.
	EncodeBinary(b []byte) error

	// Encode a string as a length-prefixed sequence of UCS codepoints, each of
	// which is encoded as an integer.
	EncodeString(s string) error

	// Encode a string as a sequence of UCS codepoints, each of which is encoded
	// as an integer.
	EncodeStringOnly(s string) error

	// Encode an arbitrary precision non negative integer using a sequence of
	// octets. The most significant bit of the last octet is set to zero to
	// indicate sequence termination. Only seven bits per octet are used to
	// store the integer's value.
	EncodeUnsignedInteger(n int) error
	EncodeUnsignedIntegerValue(iv *IntegerValue) error

	// Encode an arbitrary precision integer using a sign bit followed by a
	// sequence of octets. The most significant bit of the last octet is set to
	// zero to indicate sequence termination. Only seven bits per octet are used
	// to store the integer's value.
	EncodeInteger(n int) error
	EncodeIntegerValue(iv *IntegerValue) error

	// Encode a decimal represented as a Boolean sign followed by two Unsigned
	// Integers. A sign value of zero (0) is used to represent positive Decimal
	// values and a sign value of one (1) is used to represent negative Decimal
	// values The first Integer represents the integral portion of the Decimal
	// value. The second positive integer represents the fractional portion of
	// the decimal with the digits in reverse order to preserve leading zeros.
	EncodeDecimal(negative bool, integral, reverseFraction *IntegerValue) error

	// Encode a Float represented as two consecutive Integers. The first Integer
	// represents the mantissa of the floating point number and the second
	// Integer represents the 10-based exponent of the floating point number
	EncodeFloat(fv *FloatValue) error

	// The Date-Time datatype representation is a sequence of values
	// representing the individual components of the Date-Time
	EncodeDateTime(cal *DateTimeValue) error
}

/*
	AbstractDecoderChannel implementation
*/

type AbstractDecoderChannel struct {
	DecoderChannel
	/* buffer for reading arbitrary large integer values */
	maskedOctets []int
}

func NewAbstractDecoderChannel() *AbstractDecoderChannel {
	return &AbstractDecoderChannel{
		maskedOctets: make([]int, MaxOctetsForLong),
	}
}

func (c *AbstractDecoderChannel) DecodeBooleanValue() (*BooleanValue, error) {
	b, err := c.DecodeBoolean()
	if err != nil {
		return nil, err
	}
	if b {
		return BooleanValueTrue, nil
	} else {
		return BooleanValueFalse, nil
	}
}

func (c *AbstractDecoderChannel) DecodeString() ([]rune, error) {
	len, err := c.DecodeUnsignedInteger()
	if err != nil {
		return []rune{}, err
	}
	return c.DecodeStringOnly(len)
}

func (c *AbstractDecoderChannel) DecodeStringOnly(length int) ([]rune, error) {
	ca := make([]rune, length)

	for i := 0; i < length; i++ {
		codePoint, err := c.DecodeUnsignedInteger()
		if err != nil {
			return []rune{}, err
		}

		// Validate code point to ensure it's a valid Unicode code point
		//TODO: Check disabled!
		// if codePoint < 0 || codePoint > 0x10FFFF {
		// 	return nil, fmt.Errorf("invalid Unicode code point U+%X at index %d", codePoint, i)
		// }
		ca[i] = rune(codePoint)
	}

	return ca, nil
}

/**
 * Decode an arbitrary precision non negative integer using a sequence of
 * octets. The most significant bit of the last octet is set to zero to
 * indicate sequence termination. Only seven bits per octet are used to
 * store the integer's value.
 */
func (c *AbstractDecoderChannel) DecodeUnsignedInteger() (int, error) {
	// 0XXXXXXX ... 1XXXXXXX 1XXXXXXX
	result, err := c.Decode()
	if err != nil {
		return -1, err
	}

	// < 128: just one byte, optimal case
	// ELSE: multiple bytes...
	if result >= 128 {
		result &= 127
		mShift := 7
		var b int

		for {
			// 1. Read the next octet
			b, err = c.Decode()
			if err != nil {
				return -1, err
			}

			// 2. Multiply the value of the unsigned number represented by the 7 least significant bits
			// of the octet by the current multiplier and add the result to the current value.
			result += (b & 127) << mShift

			// 3. Multiply the multiplier by 128
			mShift += 7

			// 4. If the most significant bit of the octet was 1, go back to step 1
			if !(b >= 128) {
				break
			}
		}
	}

	return result, nil
}

func (c *AbstractDecoderChannel) decodeUnsignedLong() (int64, error) {
	result := int64(0)
	mShift := int64(0)
	var b int
	var err error

	for {
		b, err = c.Decode()
		if err != nil {
			return -1, err
		}

		result += int64(b&127) << mShift
		mShift += 7

		if !((uint32(b) >> 7) == 1) {
			break
		}
	}

	return result, nil
}

/**
 * Decode an arbitrary precision integer using a sign bit followed by a
 * sequence of octets. The most significant bit of the last octet is set to
 * zero to indicate sequence termination. Only seven bits per octet are used
 * to store the integer's value.
 */
func (c *AbstractDecoderChannel) decodeInteger() (int, error) {
	b, err := c.DecodeBoolean()
	if err != nil {
		return -1, err
	}
	i, err := c.DecodeUnsignedInteger()
	if err != nil {
		return -1, err
	}
	if b {
		// For negative values, the Unsigned Integer holds the
		// magnitude of the value minus 1.
		return -(i + 1), nil
	} else {
		// positive
		return i, nil
	}
}

func (c *AbstractDecoderChannel) decodeLong() (int64, error) {
	b, err := c.DecodeBoolean()
	if err != nil {
		return -1, err
	}
	i, err := c.decodeUnsignedLong()
	if err != nil {
		return -1, err
	}
	if b {
		// For negative values, the Unsigned Integer holds the
		// magnitude of the value minus 1.
		return -(i + 1), nil
	} else {
		// positive
		return i, nil
	}
}

func (c *AbstractDecoderChannel) DecodeIntegerValue() (*IntegerValue, error) {
	b, err := c.DecodeBoolean()
	if err != nil {
		return nil, err
	}
	return c.decodeUnsignedIntegerValue(b)
}

func (c *AbstractDecoderChannel) DecodeUnsignedIntegerValue() (*IntegerValue, error) {
	return c.decodeUnsignedIntegerValue(false)
}

func (c *AbstractDecoderChannel) decodeUnsignedIntegerValue(negative bool) (*IntegerValue, error) {
	var b int
	var err error

	for i := 0; i < MaxOctetsForLong; i++ {
		// Read the next octet
		b, err = c.Decode()
		if err != nil {
			return nil, err
		}

		// If the most significant bit of the octet was 1,
		// another octet is going to come
		if b < 128 {
			/* no more octets */
			switch i {
			case 0:
				/* one octet only */
				if negative {
					return IntegerValueOf32(-(b + 1)), nil
				} else {
					return IntegerValueOf32(b), nil
				}
			case 1, 2, 3:
				/* integer value */
				c.maskedOctets[i] = b

				/* int == 32 bits, 4 * 7bits = 28 bits */
				iResult := 0
				for k := i; k >= 0; k-- {
					iResult = (iResult << 7) | c.maskedOctets[k]
				}
				// For negative values, the Unsigned Integer holds the
				// magnitude of the value minus 1
				if negative {
					return IntegerValueOf32(-(iResult + 1)), nil
				} else {
					return IntegerValueOf32(iResult), nil
				}
			default:
				/* long value */
				c.maskedOctets[i] = b

				/* long == 64 bits, 9 * 7bits = 63 bits */
				lResult := int64(0)
				for k := i; k >= 0; k-- {
					lResult = (lResult << 7) | int64(c.maskedOctets[k])
				}
				// For negative values, the Unsigned Integer holds the
				// magnitude of the value minus 1
				if negative {
					return IntegerValueOf64(-(lResult + 1)), nil
				} else {
					return IntegerValueOf64(lResult), nil
				}
			}
		} else {
			c.maskedOctets[i] = (b & 127)
		}
	}

	// Grrr, we got a BigInteger value to deal with (note: original comment)
	bResult := big.NewInt(0)
	multiplier := big.NewInt(1)

	// already read bytes
	for i := 0; i < MaxOctetsForLong; i++ {
		bResult = bResult.Add(bResult, new(big.Int).Mul(multiplier, big.NewInt(int64(c.maskedOctets[i]))))
		multiplier = multiplier.Lsh(multiplier, 7)
	}

	// read new bytes
	for {
		// 1. Read the next octet
		b, err = c.Decode()
		if err != nil {
			return nil, err
		}

		// 2. The 7 least significant bits hold the value
		bResult = bResult.Add(bResult, new(big.Int).Mul(multiplier, big.NewInt(int64(b&127))))

		// 3. Multiply the multiplier by 128
		multiplier = multiplier.Lsh(multiplier, 7)

		// If the most significant bit of the octet was 1, another is going to come.
		if !(b > 127) {
			break
		}
	}

	// For negative values, the Unsigned Integer holds the
	// magnitude of the value minus 1
	if negative {
		bResult = bResult.Add(bResult, big.NewInt(1)).Neg(bResult)
	}

	return IntegerValueOfBig(*bResult), nil
}

// Decodes and returns an n-bit unsigned integer as string.
func (c *AbstractDecoderChannel) DecodeNBitUnsignedIntegerValue(n int) (*IntegerValue, error) {
	i, err := c.DecodeNBitUnsignedInteger(n)
	if err != nil {
		return nil, err
	}
	return IntegerValueOf32(i), nil
}

/**
 * Decode a decimal represented as a Boolean sign followed by two Unsigned
 * Integers. A sign value of zero (0) is used to represent positive Decimal
 * values and a sign value of one (1) is used to represent negative Decimal
 * values The first Integer represents the integral portion of the Decimal
 * value. The second positive integer represents the fractional portion of
 * the decimal with the digits in reverse order to preserve leading zeros.
 */
func (c *AbstractDecoderChannel) DecodeDecimalValue() (*DecimalValue, error) {
	negative, err := c.DecodeBoolean()
	if err != nil {
		return nil, err
	}

	integral, err := c.decodeUnsignedIntegerValue(false)
	if err != nil {
		return nil, err
	}
	revFractional, err := c.decodeUnsignedIntegerValue(false)
	if err != nil {
		return nil, err
	}

	return NewDecimalValue(negative, integral, revFractional), nil
}

/**
 * Decode a Float represented as two consecutive Integers. The first Integer
 * represents the mantissa of the floating point number and the second
 * Integer represents the 10-based exponent of the floating point number
 */
func (c *AbstractDecoderChannel) DecodeFloatValue() (*FloatValue, error) {
	mantissa, err := c.DecodeIntegerValue()
	if err != nil {
		return nil, err
	}
	exponent, err := c.DecodeIntegerValue()
	if err != nil {
		return nil, err
	}

	return NewFloatValue(mantissa, exponent), nil
}

/**
 * Decode Date-Time as sequence of values representing the individual
 * components of the Date-Time.
 */
func (c *AbstractDecoderChannel) DecodeDateTimeValue(kind DateTimeType) (*DateTimeValue, error) {
	year, monthDay, time, fractionalSecs := 0, 0, 0, 0
	var err error

	switch kind {
	case DateTimeGYear: // Year, [Time-Zone]
		year, err = c.decodeInteger()
		if err != nil {
			return nil, err
		}
		year += DateTimeValue_YearOffset
	case DateTimeGYearMonth, DateTimeDate: // Year, MonthDay, [TimeZone]
		year, err = c.decodeInteger()
		if err != nil {
			return nil, err
		}
		year += DateTimeValue_YearOffset

		monthDay, err = c.DecodeNBitUnsignedInteger(DateTimeValue_NumberBitsMonthDay)
		if err != nil {
			return nil, err
		}
	case DateTimeDateTime: // Year, MonthDay, Time, [FractionalSecs], [TimeZone]
		// e.g. "0001-01-01T00:00:00.111+00:33";
		year, err = c.decodeInteger()
		if err != nil {
			return nil, err
		}
		year += DateTimeValue_YearOffset

		monthDay, err = c.DecodeNBitUnsignedInteger(DateTimeValue_NumberBitsMonthDay)
		if err != nil {
			return nil, err
		}

		fallthrough // ! no break
	case DateTimeTime: // Time, [FractionalSecs], [TimeZone]
		// e.g. "12:34:56.135"
		time, err = c.DecodeNBitUnsignedInteger(DateTimeValue_NumberBitsTime)
		if err != nil {
			return nil, err
		}
		presenceFractionalSecs, err := c.DecodeBoolean()
		if err != nil {
			return nil, err
		}

		if presenceFractionalSecs {
			fractionalSecs, err = c.DecodeUnsignedInteger()
			if err != nil {
				return nil, err
			}
		} else {
			fractionalSecs = 0
		}
	case DateTimeGMonth, DateTimeGMonthDay, DateTimeGDay:
		// e.g. "--12"
		// e.g. "--01-28"
		// "---16";
		monthDay, err = c.DecodeNBitUnsignedInteger(DateTimeValue_NumberBitsMonthDay)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported date time type: %d", kind)
	}

	presenceTimezone, err := c.DecodeBoolean()
	if err != nil {
		return nil, err
	}

	var timezone int
	if presenceTimezone {
		timezone, err = c.DecodeNBitUnsignedInteger(DateTimeValue_NumberBitsTimeZone)
		if err != nil {
			return nil, err
		}
		timezone -= DateTimeValue_TimeZoneOffsetInMinutes
	} else {
		timezone = 0
	}

	return NewDateTimeValue(kind, year, monthDay, time, fractionalSecs, presenceTimezone, timezone), nil
}

/*
	AbstractEncoderChannel implementation
*/

type AbstractEncoderChannel struct {
	EncoderChannel
}

func NewAbstractEncoderChannel() *AbstractEncoderChannel {
	return &AbstractEncoderChannel{}
}

/**
 * Encode a binary value as a length-prefixed sequence of octets.
 */
func (c *AbstractEncoderChannel) EncodeBinary(b []byte) error {
	if err := c.EncodeUnsignedInteger(len(b)); err != nil {
		return err
	}
	return c.EncodeBytes(b, 0, len(b))
}

/**
 * Encode a string as a length-prefixed sequence of UCS codepoints, each of
 * which is encoded as an integer.
 */
func (c *AbstractEncoderChannel) EncodeString(s string) error {
	ch := []rune(s)
	if err := c.EncodeUnsignedInteger(len(ch)); err != nil {
		return err
	}
	return c.EncodeStringOnly(s)
}

func (c *AbstractEncoderChannel) EncodeStringOnly(s string) error {
	ch := []rune(s)
	length := len(ch)

	for i := range length {
		if err := c.EncodeUnsignedInteger(int(ch[i])); err != nil {
			return err
		}
	}

	return nil
}

/**
 * Encode an arbitrary precision integer using a sign bit followed by a
 * sequence of octets. The most significant bit of the last octet is set to
 * zero to indicate sequence termination. Only seven bits per octet are used
 * to store the integer's value.
 */
func (c *AbstractEncoderChannel) EncodeInteger(n int) error {
	// signalize sign
	if n < 0 {
		if err := c.EncodeBoolean(true); err != nil {
			return err
		}
		// For negative values, the Unsigned Integer holds the
		// magnitude of the value minus 1
		return c.EncodeUnsignedInteger((-n) - 1)
	} else {
		if err := c.EncodeBoolean(false); err != nil {
			return err
		}
		return c.EncodeUnsignedInteger(n)
	}
}

func (c *AbstractEncoderChannel) encodeLong(l int64) error {
	// signalize sign
	if l < 0 {
		if err := c.EncodeBoolean(true); err != nil {
			return err
		}
		return c.encodeUnsignedLong((-l) - 1)
	} else {
		if err := c.EncodeBoolean(false); err != nil {
			return err
		}
		return c.encodeUnsignedLong(l)
	}
}

func (c *AbstractEncoderChannel) encodeBigInteger(bi *big.Int) error {
	if bi.Sign() < 0 {
		if err := c.EncodeBoolean(true); err != nil {
			return err
		}
		return c.encodeUnsignedBigInteger(new(big.Int).Neg(bi).Sub(bi, big.NewInt(1)))
	} else {
		if err := c.EncodeBoolean(false); err != nil {
			return err
		}
		return c.encodeUnsignedBigInteger(bi)
	}
}

func (c *AbstractEncoderChannel) EncodeIntegerValue(iv *IntegerValue) error {
	switch iv.GetIntegerValueType() {
	case IntegerValue32:
		return c.EncodeInteger(iv.Value32())
	case IntegerValue64:
		return c.encodeLong(iv.Value64())
	case IntegerValueBig:
		return c.encodeBigInteger(iv.ValueBig())
	default:
		return fmt.Errorf("unexpected EXI interger value type: %d", iv.GetIntegerValueType())
	}
}

/**
 * Encode an arbitrary precision non negative integer using a sequence of
 * octets. The most significant bit of the last octet is set to zero to
 * indicate sequence termination. Only seven bits per octet are used to
 * store the integer's value.
 */
func (c *AbstractEncoderChannel) EncodeUnsignedInteger(n int) error {
	if n < 0 {
		return fmt.Errorf("integer value must have positive value")
	}

	if n < 128 {
		// write byte as is
		return c.EncodeInt(n)
	} else {
		n7BitBlocks := utils.NumberOf7BitBlocksToRepresent32(uint(n))

		switch n7BitBlocks {
		case 5:
			if err := c.EncodeInt(128 | n); err != nil {
				return err
			}
			n = int(uint32(n) >> 7)
			fallthrough
		case 4:
			if err := c.EncodeInt(128 | n); err != nil {
				return err
			}
			n = int(uint32(n) >> 7)
			fallthrough
		case 3:
			if err := c.EncodeInt(128 | n); err != nil {
				return err
			}
			n = int(uint32(n) >> 7)
			fallthrough
		case 2:
			if err := c.EncodeInt(128 | n); err != nil {
				return err
			}
			n = int(uint32(n) >> 7)
			fallthrough
		case 1:
			return c.EncodeInt(0 | n)
		}
	}

	return nil
}

func (c *AbstractEncoderChannel) encodeUnsignedLong(l int64) error {
	if l < 0 {
		return fmt.Errorf("int64 value must have positive value")
	}

	lastEncode := int(l)
	l = int64(uint64(l) >> 7)

	for l != 0 {
		if err := c.EncodeInt(lastEncode | 128); err != nil {
			return err
		}
		lastEncode = int(l)
		l = int64(uint64(l) >> 7)
	}

	return c.EncodeInt(lastEncode)
}

func (c *AbstractEncoderChannel) encodeUnsignedBigInteger(bi *big.Int) error {
	if bi.Sign() < 0 {
		return fmt.Errorf("big.Int value must have positive value")
	}

	// does not fit into long (64 bits)
	// approach: write byte per byte
	nbytes := bi.BitLen() / 7
	if bi.BitLen()%7 > 0 {
		nbytes++
	}

	biCopy := new(big.Int).Set(bi)

	for nbytes > 1 {
		nbytes--

		// 1XXXXXXX ... 1XXXXXXX
		value := int(biCopy.Int64()&0x7F) | 128
		if err := c.EncodeInt(value); err != nil {
			return err
		}

		biCopy.Rsh(biCopy, 7)
	}

	// 0XXXXXXX
	return c.EncodeInt(int(biCopy.Int64() & 0x7F))
}

func (c *AbstractEncoderChannel) EncodeUnsignedIntegerValue(iv *IntegerValue) error {
	switch iv.GetIntegerValueType() {
	case IntegerValue32:
		return c.EncodeUnsignedInteger(iv.Value32())
	case IntegerValue64:
		return c.encodeUnsignedLong(iv.Value64())
	case IntegerValueBig:
		return c.encodeUnsignedBigInteger(iv.ValueBig())
	default:
		return fmt.Errorf("unexpcted EXI integer value type: %d", iv.GetIntegerValueType())
	}
}

/**
 * Encode a decimal represented as a Boolean sign followed by two Unsigned
 * Integers. A sign value of zero (0) is used to represent positive Decimal
 * values and a sign value of one (1) is used to represent negative Decimal
 * values The first Integer represents the integral portion of the Decimal
 * value. The second positive integer represents the fractional portion of
 * the decimal with the digits in reverse order to preserve leading zeros.
 */
func (c *AbstractEncoderChannel) EncodeDecimal(negative bool, integral, reverseFraction *IntegerValue) error {
	// sign, integral, reverse fractional
	if err := c.EncodeBoolean(negative); err != nil {
		return err
	}
	if err := c.EncodeUnsignedIntegerValue(integral); err != nil {
		return err
	}
	return c.EncodeUnsignedIntegerValue(reverseFraction)
}

/**
 * Encode a Float represented as two consecutive Integers. The first Integer
 * represents the mantissa of the floating point number and the second
 * Integer represents the 10-based exponent of the floating point number
 */
func (c *AbstractEncoderChannel) EncodeFloat(fv *FloatValue) error {
	// encode mantissa and exponent
	if err := c.EncodeIntegerValue(fv.GetMantissa()); err != nil {
		return err
	}
	return c.EncodeIntegerValue(fv.GetExponent())
}

func (c *AbstractEncoderChannel) EncodeDateTime(datetime *DateTimeValue) error {
	switch datetime.kind {
	case DateTimeGYear: // Year, [Time-Zone]
		return c.EncodeInteger(datetime.year - DateTimeValue_YearOffset)
	case DateTimeGYearMonth, DateTimeDate: // Year, MonthDay, [TimeZone]
		if err := c.EncodeInteger(datetime.year - DateTimeValue_YearOffset); err != nil {
			return err
		}
		if err := c.EncodeNBitUnsignedInteger(datetime.monthDay, DateTimeValue_NumberBitsMonthDay); err != nil {
			return err
		}
	case DateTimeDateTime: // Year, MonthDay, Time, [FractionalSecs],
		if err := c.EncodeInteger(datetime.year - DateTimeValue_YearOffset); err != nil {
			return err
		}
		if err := c.EncodeNBitUnsignedInteger(datetime.monthDay, DateTimeValue_NumberBitsMonthDay); err != nil {
			return err
		}
		fallthrough // ! no break
	case DateTimeTime: // Time, [FractionalSecs], [TimeZone]
		if err := c.EncodeNBitUnsignedInteger(datetime.time, DateTimeValue_NumberBitsTime); err != nil {
			return err
		}
		if datetime.presenceFractionalSecs {
			if err := c.EncodeBoolean(true); err != nil {
				return err
			}
			if err := c.EncodeUnsignedInteger(datetime.fractionalSecs); err != nil {
				return err
			}
		} else {
			if err := c.EncodeBoolean(true); err != nil {
				return err
			}
		}
	case DateTimeGMonth, DateTimeGMonthDay, DateTimeGDay: // MonthDay, [TimeZone]
		if err := c.EncodeNBitUnsignedInteger(datetime.monthDay, DateTimeValue_NumberBitsMonthDay); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unexpected EXI date time type: %d", datetime.kind)
	}

	// [TimeZone]
	if datetime.presenceTimezone {
		if err := c.EncodeBoolean(true); err != nil {
			return err
		}
		return c.EncodeNBitUnsignedInteger(datetime.timezone+DateTimeValue_TimeZoneOffsetInMinutes, DateTimeValue_NumberBitsTimeZone)
	} else {
		return c.EncodeBoolean(false)
	}
}

/*
	BitDecoderChannel implementation
*/

type BitDecoderChannel struct {
	*AbstractDecoderChannel
	reader *BitReader
}

func NewBitDecoderChannel(reader *bufio.Reader) *BitDecoderChannel {
	adc := NewAbstractDecoderChannel()
	dc := &BitDecoderChannel{
		AbstractDecoderChannel: adc,
		reader:                 NewBitReader(reader),
	}
	adc.DecoderChannel = dc
	return dc
}

func (c *BitDecoderChannel) Decode() (int, error) {
	return c.reader.Read()
}

func (c *BitDecoderChannel) Align() error {
	return c.reader.Align()
}

func (c *BitDecoderChannel) LookAhead() (int, error) {
	return c.reader.LookAhead()
}

func (c *BitDecoderChannel) Skip(n int64) error {
	return c.reader.Skip(n)
}

/**
 * Decodes and returns an n-bit unsigned integer.
 */
func (c *BitDecoderChannel) DecodeNBitUnsignedInteger(n int) (int, error) {
	if n < 0 {
		return -1, fmt.Errorf("length of NBit unsigned integer must have positive value")
	}
	if n == 0 {
		return 0, nil
	} else {
		return c.reader.ReadBits(n)
	}
}

/**
 * Decode a single boolean value. The value false is represented by the bit
 * 0, and the value true is represented by the bit 1.
 */
func (c *BitDecoderChannel) DecodeBoolean() (bool, error) {
	value, err := c.reader.ReadBit()
	if err != nil {
		return false, err
	}
	return (value == 1), nil
}

/**
 * Decode a binary value as a length-prefixed sequence of octets.
 */
func (c *BitDecoderChannel) DecodeBinary() ([]byte, error) {
	length, err := c.DecodeUnsignedInteger()
	if err != nil {
		return []byte{}, nil
	}
	result := make([]byte, length)

	if err := c.reader.ReadToBuffer(result, 0, length); err != nil {
		return []byte{}, err
	}
	return result, nil
}

/*
	BitEncoderChannel implementation
*/

type BitEncoderChannel struct {
	*AbstractEncoderChannel
	writer *BitWriter
}

func NewBitEncoderChannel(writer bufio.Writer) *BitEncoderChannel {
	aec := NewAbstractEncoderChannel()
	bec := &BitEncoderChannel{
		AbstractEncoderChannel: aec,
		writer:                 NewBitWriter(writer),
	}
	return bec
}

func (c *BitEncoderChannel) GetWriter() *BitWriter {
	return c.writer
}

func (c *BitEncoderChannel) GetLength() int {
	return c.writer.GetLength()
}

/**
 * Flush underlying bit output stream.
 */
func (c *BitEncoderChannel) Flush() error {
	return c.writer.Flush()
}

func (c *BitEncoderChannel) Align() error {
	return c.writer.Align()
}

func (c *BitEncoderChannel) Encode(b int) error {
	return c.writer.WriteBits(b, 8)
}

func (c *BitEncoderChannel) EncodeBytes(b []byte, offset, length int) error {
	for i := offset; i < (offset + length); i++ {
		if err := c.writer.WriteBits(int(b[i]), 8); err != nil {
			return err
		}
	}
	return nil
}

/**
 * Encode n-bit unsigned integer. The n least significant bits of parameter
 * b starting with the most significant, i.e. from left to right.
 */
func (c *BitEncoderChannel) EncodeNBitUnsignedInteger(b, n int) error {
	if b < 0 || n < 0 {
		return errors.New("encode negative value as unsigned integer is invalid")
	}

	return c.writer.WriteBits(b, n)
}

/**
 * Encode a single boolean value. A false value is encoded as bit 0 and true
 * value is encode as bit 1.
 */
func (c *BitEncoderChannel) EncodeBoolean(b bool) error {
	if b {
		return c.writer.WriteBit1()
	} else {
		return c.writer.WriteBit0()
	}
}

/*
	ByteDecoderChannel implementation
*/

type ByteDecoderChannel struct {
	*AbstractDecoderChannel
	reader *bufio.Reader
}

func NewByteDecoderChannel(reader *bufio.Reader) *ByteDecoderChannel {
	adc := NewAbstractDecoderChannel()
	bdc := &ByteDecoderChannel{
		AbstractDecoderChannel: adc,
		reader:                 reader,
	}
	return bdc
}

func (c *ByteDecoderChannel) GetReader() *bufio.Reader {
	return c.reader
}

func (c *ByteDecoderChannel) Decode() (int, error) {
	b, err := c.reader.ReadByte()
	if err == io.EOF {
		return -1, errors.New("premature EOS found while reading data")
	}
	if err != nil {
		return -1, err
	}
	return int(b), nil
}

func (c *ByteDecoderChannel) Align() error {
	return nil
}

func (c *ByteDecoderChannel) Skip(n int64) error {
	for n != 0 {
		skipped, err := c.reader.Discard(int(n))
		if err != nil {
			return err
		}
		n -= int64(skipped)
	}
	return nil
}

/**
 * Decodes and returns an n-bit unsigned integer using the minimum number of
 * bytes required for n bits.
 */
func (c *ByteDecoderChannel) DecodeNBitUnsignedInteger(n int) (int, error) {
	if n < 0 {
		return -1, errors.New("size of NBit unsigned integer must be postive value")
	}

	bitsRead := 0
	result := 0

	for bitsRead < n {
		b, err := c.Decode()
		if err != nil {
			return -1, err
		}
		result += b << bitsRead
		bitsRead += 8
	}

	return result, nil
}

/**
 * Decode a single boolean value. The value false is represented by the byte
 * 0, and the value true is represented by the byte 1.
 */
func (c *ByteDecoderChannel) DecodeBoolean() (bool, error) {
	b, err := c.Decode()
	if err != nil {
		return false, err
	}
	return b != 0, nil
}

/**
 * Decode a binary value as a length-prefixed sequence of octets.
 */
func (c *ByteDecoderChannel) DecodeBinary() ([]byte, error) {
	length, err := c.DecodeUnsignedInteger()
	if err != nil {
		return []byte{}, err
	}

	result := make([]byte, length)

	readBytes := 0
	for readBytes < length {
		read, err := c.reader.Read(result[readBytes : readBytes+(length-readBytes)])
		if err == io.EOF {
			return []byte{}, errors.New("premature EOS found while reading data")
		}
		if err != nil {
			return []byte{}, err
		}
		readBytes += read
	}

	return result, nil
}

/*
	ByteEncoderChannel implementation
*/

type ByteEncoderChannel struct {
	*AbstractEncoderChannel
	writer bufio.Writer
	len    int
}

func NewByteEncoderChannel(writer bufio.Writer) *ByteEncoderChannel {
	aec := NewAbstractEncoderChannel()
	bec := &ByteEncoderChannel{
		AbstractEncoderChannel: aec,
		writer:                 writer,
		len:                    0,
	}
	return bec
}

func (c *ByteEncoderChannel) GetWriter() *bufio.Writer {
	return &c.writer
}

func (c *ByteEncoderChannel) GetLength() int {
	return c.len
}

func (c *ByteEncoderChannel) Flush() error {
	return c.writer.Flush()
}

func (c *ByteEncoderChannel) Align() error {
	return nil
}

func (c *ByteEncoderChannel) Encode(b int) error {
	if err := c.writer.WriteByte(byte(b & 0xFF)); err != nil {
		return err
	}
	c.len++
	return nil
}

func (c *ByteEncoderChannel) EncodeBytes(b []byte, offset, length int) error {
	if _, err := c.writer.Write(b[offset : offset+length]); err != nil {
		return err
	}
	c.len += length
	return nil
}

func (c *ByteEncoderChannel) EncodeBoolean(b bool) error {
	return c.Encode(utils.BoolToInt(b))
}

func (c *ByteEncoderChannel) EncodeNBitUnsignedInteger(b, n int) error {
	if b < 0 || n < 0 {
		return errors.New("size of NBit unsigned integer must be postive value")
	}

	if n == 0 {
		// 0 bytes
	} else if n < 9 {
		// 1 byte
		return c.Encode(b & 0xFF)
	} else if n < 17 {
		// 2 bytes
		if err := c.Encode(b & 0x00FF); err != nil {
			return err
		}
		return c.Encode((b & 0xFF00) >> 8)
	} else if n < 25 {
		// 3 bytes
		if err := c.Encode(b & 0x0000FF); err != nil {
			return err
		}
		if err := c.Encode((b & 0x00FF00) >> 8); err != nil {
			return err
		}
		return c.Encode((b & 0xFF0000) >> 16)
	} else if n < 33 {
		// 4 bytes
		if err := c.Encode(b & 0x000000FF); err != nil {
			return err
		}
		if err := c.Encode((b & 0x0000FF00) >> 8); err != nil {
			return err
		}
		if err := c.Encode((b & 0x00FF0000) >> 16); err != nil {
			return err
		}
		return c.Encode((b & 0xFF000000) >> 24)
	} else {
		return errors.New("currently no more than 4 Bytes allowed for NBitUnsignedInteger")
	}

	return nil
}
