package utils

import (
	"fmt"
	"math"
	"math/big"

	"github.com/cockroachdb/apd/v3"
	Text "github.com/linkdotnet/golang-stringbuilder"
)

var (
	log2div          = math.Log(2.0)
	sizeTable []int  = []int{9, 99, 999, 9999, 99999, 999999, 9999999, 99999999, 999999999, math.MaxInt32}
	digitOnes []rune = []rune{'0', '1', '2', '3', '4', '5', '6', '7',
		'8', '9', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0',
		'1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3',
		'4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5', '6',
		'7', '8', '9', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2',
		'3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
		'6', '7', '8', '9', '0', '1', '2', '3', '4', '5', '6', '7', '8',
		'9'}
	digitTens []rune = []rune{'0', '0', '0', '0', '0', '0', '0', '0',
		'0', '0', '1', '1', '1', '1', '1', '1', '1', '1', '1', '1', '2',
		'2', '2', '2', '2', '2', '2', '2', '2', '2', '3', '3', '3', '3',
		'3', '3', '3', '3', '3', '3', '4', '4', '4', '4', '4', '4', '4',
		'4', '4', '4', '5', '5', '5', '5', '5', '5', '5', '5', '5', '5',
		'6', '6', '6', '6', '6', '6', '6', '6', '6', '6', '7', '7', '7',
		'7', '7', '7', '7', '7', '7', '7', '8', '8', '8', '8', '8', '8',
		'8', '8', '8', '8', '9', '9', '9', '9', '9', '9', '9', '9', '9',
		'9'}
	digits []rune = []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8',
		'9', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l',
		'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y',
		'z'}
	IntegerMinValueCharArray []rune = []rune("-2147483648")
	LongMinValueCharArray    []rune = []rune("-9223372036854775808")
)

func GetCodingLength(characteristics int) int {
	if characteristics == 0 || characteristics == 1 {
		return 0
	}
	if characteristics == 2 {
		return 1
	}
	if characteristics == 3 || characteristics == 4 {
		return 2
	}
	if 5 <= characteristics && characteristics <= 8 {
		return 3
	}
	if 9 <= characteristics && characteristics <= 16 {
		return 4
	}
	if 17 <= characteristics && characteristics <= 32 {
		return 5
	}
	if 33 <= characteristics && characteristics <= 64 {
		return 6
	}

	if characteristics < 129 {
		// 65 .. 128
		return 7
	} else if characteristics < 257 {
		// 129 .. 256
		return 8
	} else if characteristics < 513 {
		// 257 .. 512
		return 9
	} else if characteristics < 1025 {
		// 513 .. 1024
		return 10
	} else if characteristics < 2049 {
		// 1025 .. 2048
		return 11
	} else if characteristics < 4097 {
		// 2049 .. 4096
		return 12
	} else if characteristics < 8193 {
		// 4097 .. 8192
		return 13
	} else if characteristics < 16385 {
		// 8193 .. 16384
		return 14
	} else if characteristics < 32769 {
		// 16385 .. 32768
		return 15
	} else {
		return int(math.Ceil(math.Log(float64(characteristics))) / log2div)
	}
}

func stringSize32(x int) int {
	for i := 0; ; i++ {
		if x <= sizeTable[i] {
			return i + 1
		}
	}
}

func stringSize64(x int64) int {
	p := int64(10)
	for i := 1; i < 19; i++ {
		if x < p {
			return i
		}
		p = 10 * p
	}
	return 19
}

func GetStringSize32(i int) int {
	if i < 0 {
		return stringSize32(-i) + 1
	} else {
		return stringSize32(i)
	}
}

func GetStringSize64(l int64) int {
	if l == math.MinInt64 {
		return 20
	} else {
		if l < 0 {
			return stringSize64(-l) + 1
		} else {
			return stringSize64(l)
		}
	}
}

/*
 * Places characters representing the integer i into the character array
 * buf. The characters are placed into the buffer backwards starting with
 * the least significant digit at the specified index (exclusive), and
 * working backwards from there.
 *
 * Will fail if i == math.MinInt32
 */
func Itos32(i int, index *int, buf []rune) {
	var q, r int
	var sign rune = 0

	if i < 0 {
		sign = '-'
		i = -i
	}

	// Generate two digits per iteration
	for i >= 65536 {
		q = i / 100
		// really: r = i - (q * 100);
		r = i - ((q << 6) + (q << 5) + (q << 2))
		i = q

		*index--
		buf[*index] = digitOnes[r]
		*index--
		buf[*index] = digitTens[r]
	}

	// Fall thru to fast mode for smaller numbers
	// assert(i <= 65536, i);
	for {
		q = int(uint(i*52429) >> uint(16+3))
		r = i - ((q << 3) + (q << 1))
		*index--
		buf[*index] = digits[r]
		i = q
		if i == 0 {
			break
		}
	}

	if sign != 0 {
		*index--
		buf[*index] = sign
	}
}

/*
*
  - Places characters representing the integer i into the character array
  - buf. The characters are placed into the buffer backwards starting with
  - the least significant digit at the specified index (exclusive), and
  - working backwards from there.
*/
func Itos64(l int64, index *int, buf []rune) {
	if l == math.MinInt64 {
		*index--
		buf[*index] = '8'
		*index--
		buf[*index] = '0'
		*index--
		buf[*index] = '8'
		*index--
		buf[*index] = '5'
		*index--
		buf[*index] = '7'
		*index--
		buf[*index] = '7'
		*index--
		buf[*index] = '4'
		*index--
		buf[*index] = '5'
		*index--
		buf[*index] = '8'
		*index--
		buf[*index] = '6'
		*index--
		buf[*index] = '3'
		*index--
		buf[*index] = '0'
		*index--
		buf[*index] = '2'
		*index--
		buf[*index] = '7'
		*index--
		buf[*index] = '3'
		*index--
		buf[*index] = '3'
		*index--
		buf[*index] = '2'
		*index--
		buf[*index] = '2'
		*index--
		buf[*index] = '9'
		*index--
		buf[*index] = '-'

		return
	}

	if l == math.MinInt64 {
		panic("l == math.MinInt64")
	}

	var q int64
	var r int
	var sign rune = 0

	if l < 0 {
		sign = '-'
		l = -l
	}

	// Get 2 digits/iteration using longs until quotient fits into an int
	for l > math.MaxInt32 {
		q = l / 100
		// really: r = i - (q * 100);
		r = int(l - ((q << 6) + (q << 5) + (q << 2)))
		l = q
		*index--
		buf[*index] = digitOnes[r]
		*index--
		buf[*index] = digitTens[r]
	}

	// Get 2 digits/iteration using ints
	var q2 int
	i2 := int(l)
	for i2 >= 65536 {
		q2 = i2 / 100
		// really: r = i2 - (q * 100);
		r = i2 - ((q2 << 6) + (q2 << 5) + (q2 << 2))
		i2 = q2
		*index--
		buf[*index] = digitOnes[r]
		*index--
		buf[*index] = digitTens[r]
	}

	// Fall thru to fast mode for smaller numbers
	// assert(i2 <= 65536, i2);
	for {
		q2 = int(uint32(i2*52429) >> (16 + 3))
		r = i2 - ((q2 << 3) + (q2 << 1))
		*index--
		buf[*index] = digits[r]
		i2 = q2
		if i2 == 0 {
			break
		}
	}
	if sign != 0 {
		*index--
		buf[*index] = sign
	}
}

/*
 * Places characters representing the integer i into the character array buf
 * in reverse order.
 *
 * Will fail if i < 0 (zero)
 */
func ItosReverse32(i int, index *int, buf []rune) int {
	var q, r int
	posChar := *index

	// Generate two digits per iteration
	for i >= 65536 {
		q = i / 100
		// really: r = i - (q * 100);
		r = i - ((q << 6) + (q << 5) + (q << 2))
		i = q

		buf[posChar] = digitOnes[r]
		posChar++
		buf[posChar] = digitTens[r]
		posChar++
	}

	// Fall thru to fast mode for smaller numbers
	// assert(i <= 65536, i);
	for {
		q = int(uint(i*52429) >> (16 + 3))
		r = i - ((q << 3) + (q << 1))
		buf[posChar] = digits[r]
		posChar++
		i = q
		if i == 0 {
			break
		}
	}

	return posChar - *index // number of written chars
}

// Places characters representing the integer i into the character array buf
// in reverse order.
func ItosReverse64(i int64, index *int, buf []rune) {
	var q int64
	var r int

	// Get 2 digits/iteration using longs until quotient fits into an int
	for i > math.MaxInt32 {
		q = i / 100
		// really: r = i - (q * 100);
		r = int(i - ((q << 6) + (q << 5) + (q << 2)))
		i = q

		buf[*index] = digitOnes[r]
		*index++
		buf[*index] = digitTens[r]
		*index++
	}

	// Get 2 digits/iteration using ints
	var q2 int
	i2 := int(i)
	for i >= 65536 {
		q2 = i2 / 100
		// really: r = i - (q * 100);
		r = i2 - ((q2 << 6) + (q2 << 5) + (q2 << 2))
		i2 = q2

		buf[*index] = digitOnes[r]
		*index++
		buf[*index] = digitTens[r]
		*index++
	}

	// Fall thru to fast mode for smaller numbers
	// assert(i2 <= 65536, i2);
	for {
		q2 = int(uint32(i2*52429) >> (16 + 3))
		r = i2 - ((q2 << 3) + (q2 << 1))

		buf[*index] = digits[r]
		*index++

		i2 = q2
		if i2 == 0 {
			break
		}
	}
}

// help function described in W3C PR Schema [E Adding durations to dateTimes]
func FQuotient(a, b int) int {
	// fQuotient(a, b) = the greatest integer less than or equal to a/b
	return int(math.Floor(float64(a) / float64(b)))
}

// help function described in W3C PR Schema [E Adding durations to dateTimes]
func FQuotientLoHi(temp, low, high int) int {
	// fQuotient(a - low, high - low)
	return FQuotient(temp-low, high-low)
}

func Modulo(a, b int) int {
	// a - fQuotient(a,b)*b
	return a - (FQuotient(a, b) * b)
}

// help function described in W3C PR Schema [E Adding durations to dateTimes]
func ModuloLoHi(a, low, high int) int {
	// modulo(a - low, high - low) + low
	return Modulo(a-low, high-low) + low
}

func MaximumDayInMonth(year, month int) int {
	// 31 M = January, March, May, July, August, October, or December
	// 30 M = April, June, September, or November
	// 29 M = February AND (modulo(Y, 400) = 0 OR (modulo(Y, 100) != 0) AND
	// modulo(Y, 4) = 0)
	// 28 Otherwise

	switch month {
	case 1, 3, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	default:
		if month == 2 && (Modulo(year, 400) == 0 || Modulo(year, 100) != 0 && Modulo(year, 4) == 0) {
			return 29
		} else {
			return 28
		}
	}
}

func ReverseString(s string) string {
	return Text.NewStringBuilderFromString(s).Reverse().ToString()
}

// TryBigInt tries to convert the decimal into a `big.Int`.
// Returns an error if the number is not an integer.
// Reuses code from: https://github.com/fardream/decimal
func TryBigInt(x *apd.Decimal) (*big.Int, error) {
	var integ, frac apd.Decimal
	x.Modf(&integ, &frac)
	if !frac.IsZero() {
		return nil, fmt.Errorf("%s: has fractional part", x.String())
	}
	str := x.Text('f')
	r, ok := big.NewInt(0).SetString(str, 10)
	if !ok {
		return nil, fmt.Errorf("%s is not an integer", r)
	}

	return r, nil
}

func BoolToInt(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

/**
 * Returns the least number of 7 bit-blocks that is needed to represent the
 * parameter n. Returns 1 if parameter n is 0.
 */
func NumberOf7BitBlocksToRepresent32(n uint) int {
	if n < 128 { // 7 bits
		return 1
	} else if n < 16384 { // 17 bits
		return 2
	} else if n < 2097152 { // 21 bits
		return 3
	} else if n < 268435456 { // 28 bits
		return 4
	} else { // 35 bits
		return 5
	}
}

/**
 * Returns the least number of 7 bit-blocks that is needed to represent the
 * parameter l. Returns 1 if parameter l is 0.
 */
func NumberOf7BitBlocksToRepresent64(l uint64) int {
	if l < 0xffffffff {
		return NumberOf7BitBlocksToRepresent32(uint(l))
	} else if l < 0x800000000 { // 35 bits
		return 5
	} else if l < 0x40000000000 { // 42 bits
		return 6
	} else if l < 0x2000000000000 { // 49 bits
		return 7
	} else if l < 0x100000000000000 { // 56 bits
		return 8
	} else if l < 0x8000000000000000 { // 63 bits
		return 9
	} else { // 70 bits
		// long, 64 bits
		return 10
	}
}
