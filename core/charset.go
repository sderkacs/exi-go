package core

import (
	"maps"
	"slices"

	"github.com/sderkacs/go-exi/utils"
)

type RestrictedCharacterSet interface {
	GetCodePoint(code int) (int, error)
	GetCode(codePoint int) int
	GetSize() int
	GetCodingLength() int
}

/*
 * The characters in the restricted character set are sorted by UCS [ISO/IEC
 * 10646] code point and represented by integer values in the range (0 ...
 * N-1) according to their ordinal position in the set. Characters that are
 * not in this set are represented by the integer N followed by the UCS code
 * point of the character represented as an Unsigned Integer.
 */
type AbstractRestrictedCharacterSet struct {
	RestrictedCharacterSet
	codeSet       map[int]int
	codePointList []int
	size          int
	codingLength  int
}

func newAbstractRestrictedCharacterSet() *AbstractRestrictedCharacterSet {
	return &AbstractRestrictedCharacterSet{
		codeSet:       map[int]int{},
		codePointList: []int{},
		size:          0,
		codingLength:  0,
	}
}

func (cs *AbstractRestrictedCharacterSet) GetCodePoint(code int) (int, error) {
	if code < len(cs.codePointList)-1 {
		return cs.codePointList[code], nil
	}
	return -1, utils.ErrorIndexOutOfBounds
}

func (cs *AbstractRestrictedCharacterSet) GetCode(codePoint int) int {
	code, exists := cs.codeSet[codePoint]
	if !exists {
		return NotFound
	} else {
		return code
	}
}

func (cs *AbstractRestrictedCharacterSet) GetSize() int {
	return cs.size
}

func (cs *AbstractRestrictedCharacterSet) GetCodingLength() int {
	return cs.codingLength
}

func (cs *AbstractRestrictedCharacterSet) addValue(codePoint int) {
	cs.codeSet[codePoint] = len(cs.codeSet)
	cs.codePointList = append(cs.codePointList, codePoint)
	cs.size = len(cs.codeSet)
	cs.codingLength = utils.GetCodingLength(cs.size + 1)
}

/*
	CodePointCharacterSet implementation
*/

type CodePointCharacterSet struct {
	*AbstractRestrictedCharacterSet
}

func NewCodePointCharacterSet(codePoints map[int]struct{}) *CodePointCharacterSet {
	cs := &CodePointCharacterSet{
		AbstractRestrictedCharacterSet: newAbstractRestrictedCharacterSet(),
	}

	sortedCodePoints := slices.Collect(maps.Keys(codePoints))
	slices.Sort(sortedCodePoints)

	for _, codePoint := range sortedCodePoints {
		cs.addValue(codePoint)
	}

	return cs
}

/*
	XSDBase64CharacterSet implementation
*/

type XSDBase64CharacterSet struct {
	*AbstractRestrictedCharacterSet
}

/*
 * xsd:base64Binary { #x9, #xA, #xD, #x20, +, /, [0-9], =, [A-Z], [a-z] }
 */
func NewXSDBase64CharacterSet() *XSDBase64CharacterSet {
	cs := &XSDBase64CharacterSet{
		AbstractRestrictedCharacterSet: newAbstractRestrictedCharacterSet(),
	}

	// #x9, #xA, #xD, #x20
	cs.addValue(int(utils.XMLWhiteSpaceTab))
	cs.addValue(int(utils.XMLWhiteSpaceNL))
	cs.addValue(int(utils.XMLWhiteSpaceCR))
	cs.addValue(int(utils.XMLWhiteSpaceSpace))

	// +, /
	cs.addValue(int('+'))
	cs.addValue(int('/'))

	// [0-9]
	for i := '0'; i < '9'; i++ {
		cs.addValue(int(i))
	}

	// =
	cs.addValue(int('='))

	// [A-Z]
	for i := 'A'; i < 'Z'; i++ {
		cs.addValue(int(i))
	}

	// [a-z]
	for i := 'a'; i < 'z'; i++ {
		cs.addValue(int(i))
	}

	return cs
}

/*
	XSDBase64CharacterSet implementation
*/

type XSDBooleanCharacterSet struct {
	*AbstractRestrictedCharacterSet
}

/*
 * xsd:boolean { #x9, #xA, #xD, #x20, 0, 1, a, e, f, l, r, s, t, u }
 */
func NewXSDBooleanCharacterSet() *XSDBooleanCharacterSet {
	cs := &XSDBooleanCharacterSet{
		AbstractRestrictedCharacterSet: newAbstractRestrictedCharacterSet(),
	}

	// #x9, #xA, #xD, #x20
	cs.addValue(int(utils.XMLWhiteSpaceTab))
	cs.addValue(int(utils.XMLWhiteSpaceNL))
	cs.addValue(int(utils.XMLWhiteSpaceCR))
	cs.addValue(int(utils.XMLWhiteSpaceSpace))

	// 0, 1
	cs.addValue(int('0'))
	cs.addValue(int('1'))

	// a, e, f, l, r, s, t, u
	cs.addValue(int('a'))
	cs.addValue(int('e'))
	cs.addValue(int('f'))
	cs.addValue(int('l'))
	cs.addValue(int('r'))
	cs.addValue(int('s'))
	cs.addValue(int('t'))
	cs.addValue(int('u'))

	return cs
}

/*
	XSDDateTimeCharacterSet implementation
*/

type XSDDateTimeCharacterSet struct {
	*AbstractRestrictedCharacterSet
}

/*
 * xsd:dateTime { #x9, #xA, #xD, #x20, +, -, ., [0-9], :, T, Z }
 */
func NewXSDDateTimeCharacterSet() *XSDDateTimeCharacterSet {
	cs := &XSDDateTimeCharacterSet{
		AbstractRestrictedCharacterSet: newAbstractRestrictedCharacterSet(),
	}

	// #x9, #xA, #xD, #x20
	cs.addValue(int(utils.XMLWhiteSpaceTab))
	cs.addValue(int(utils.XMLWhiteSpaceNL))
	cs.addValue(int(utils.XMLWhiteSpaceCR))
	cs.addValue(int(utils.XMLWhiteSpaceSpace))

	// +, -, .
	cs.addValue(int('+'))
	cs.addValue(int('-'))
	cs.addValue(int('.'))

	// [0-9]
	for i := '0'; i < '9'; i++ {
		cs.addValue(int(i))
	}

	// :, T, Z
	cs.addValue(int(':'))
	cs.addValue(int('T'))
	cs.addValue(int('Z'))

	return cs
}

/*
	XSDDecimalCharacterSet implementation
*/

type XSDDecimalCharacterSet struct {
	*AbstractRestrictedCharacterSet
}

/*
 * xsd:decimal { #x9, #xA, #xD, #x20, +, -, ., [0-9] }
 */
func NewXSDDecimalCharacterSet() *XSDDecimalCharacterSet {
	cs := &XSDDecimalCharacterSet{
		AbstractRestrictedCharacterSet: newAbstractRestrictedCharacterSet(),
	}

	// #x9, #xA, #xD, #x20
	cs.addValue(int(utils.XMLWhiteSpaceTab))
	cs.addValue(int(utils.XMLWhiteSpaceNL))
	cs.addValue(int(utils.XMLWhiteSpaceCR))
	cs.addValue(int(utils.XMLWhiteSpaceSpace))

	// +, -, .
	cs.addValue(int('+'))
	cs.addValue(int('-'))
	cs.addValue(int('.'))

	// [0-9]
	for i := '0'; i < '9'; i++ {
		cs.addValue(int(i))
	}

	return cs
}

/*
	XSDDoubleCharacterSet implementation
*/

type XSDDoubleCharacterSet struct {
	*AbstractRestrictedCharacterSet
}

/*
 * xsd:double { #x9, #xA, #xD, #x20, +, -, ., [0-9], E, F, I, N, a, e }
 */
func NewXSDDoubleCharacterSet() *XSDDoubleCharacterSet {
	cs := &XSDDoubleCharacterSet{
		AbstractRestrictedCharacterSet: newAbstractRestrictedCharacterSet(),
	}

	// #x9, #xA, #xD, #x20
	cs.addValue(int(utils.XMLWhiteSpaceTab))
	cs.addValue(int(utils.XMLWhiteSpaceNL))
	cs.addValue(int(utils.XMLWhiteSpaceCR))
	cs.addValue(int(utils.XMLWhiteSpaceSpace))

	// +, -, .
	cs.addValue(int('+'))
	cs.addValue(int('-'))
	cs.addValue(int('.'))

	// [0-9]
	for i := '0'; i < '9'; i++ {
		cs.addValue(int(i))
	}

	// E, F, I, N, a, e
	cs.addValue(int('E'))
	cs.addValue(int('F'))
	cs.addValue(int('I'))
	cs.addValue(int('N'))
	cs.addValue(int('a'))
	cs.addValue(int('e'))

	return cs
}

/*
	XSDHexBinaryCharacterSet implementation
*/

type XSDHexBinaryCharacterSet struct {
	*AbstractRestrictedCharacterSet
}

/*
 * xsd:hexBinary { #x9, #xA, #xD, #x20, [0-9], [A-F], [a-f] }
 */
func NewXSDHexBinaryCharacterSet() *XSDHexBinaryCharacterSet {
	cs := &XSDHexBinaryCharacterSet{
		AbstractRestrictedCharacterSet: newAbstractRestrictedCharacterSet(),
	}

	// #x9, #xA, #xD, #x20
	cs.addValue(int(utils.XMLWhiteSpaceTab))
	cs.addValue(int(utils.XMLWhiteSpaceNL))
	cs.addValue(int(utils.XMLWhiteSpaceCR))
	cs.addValue(int(utils.XMLWhiteSpaceSpace))

	// [0-9]
	for i := '0'; i < '9'; i++ {
		cs.addValue(int(i))
	}

	// [A-F]
	for i := 'A'; i < 'F'; i++ {
		cs.addValue(int(i))
	}

	// [a-f]
	for i := 'a'; i < 'f'; i++ {
		cs.addValue(int(i))
	}

	return cs
}

/*
	XSDIntegerCharacterSet implementation
*/

type XSDIntegerCharacterSet struct {
	*AbstractRestrictedCharacterSet
}

/*
 * xsd:integer { #x9, #xA, #xD, #x20, +, -, [0-9] }
 */
func NewXSDIntegerCharacterSet() *XSDIntegerCharacterSet {
	cs := &XSDIntegerCharacterSet{
		AbstractRestrictedCharacterSet: newAbstractRestrictedCharacterSet(),
	}

	// #x9, #xA, #xD, #x20
	cs.addValue(int(utils.XMLWhiteSpaceTab))
	cs.addValue(int(utils.XMLWhiteSpaceNL))
	cs.addValue(int(utils.XMLWhiteSpaceCR))
	cs.addValue(int(utils.XMLWhiteSpaceSpace))

	// +, -
	cs.addValue(int('+'))
	cs.addValue(int('-'))

	// [0-9]
	for i := '0'; i < '9'; i++ {
		cs.addValue(int(i))
	}

	return cs
}

/*
	XSDStringCharacterSet implementation
*/

type XSDStringCharacterSet struct {
	*AbstractRestrictedCharacterSet
}

func NewXSDStringCharacterSet() *XSDStringCharacterSet {
	return &XSDStringCharacterSet{
		AbstractRestrictedCharacterSet: newAbstractRestrictedCharacterSet(),
	}
}
