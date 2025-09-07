package core

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/sderkacs/exi-go/utils"
)

const (
	DefaultInitialQNameLists int = 60
)

var (
	EmptyStringValue = NewStringValueFromString(EmptyString)
)

type StringCoder interface {
	GetNumberOfStringValues(qnc *QNameContext) int
	Clear()
	SetSharedStrings(sharedStrings []string) error
	IsLocalValuePartitions() bool
}

type StringDecoder interface {
	StringCoder
	AddValue(qnc *QNameContext, value *StringValue) error
	ReadValue(qnc *QNameContext, channel DecoderChannel) (*StringValue, error)
	ReadValueLocalHit(qnc *QNameContext, channel DecoderChannel) (*StringValue, error)
	ReadValueGlobalHit(channel DecoderChannel) (*StringValue, error)
}

type StringEncoder interface {
	StringCoder
	AddValue(qnc *QNameContext, value string) error
	WriteValue(qnc *QNameContext, channel EncoderChannel, value string) error
	IsStringHit(value string) (bool, error)
	GetValueContainer(value string) *ValueContainer
	GetValueContainerSize() int
}

/*
	ValueContainer implementation
*/

type ValueContainer struct {
	Value         string
	Context       *QNameContext
	LocalValueID  int
	GlobalValueID int
}

func NewValueContainer(value string, qnc *QNameContext, localValueID, globalValueID int) ValueContainer {
	return ValueContainer{
		Value:         value,
		Context:       qnc,
		LocalValueID:  localValueID,
		GlobalValueID: globalValueID,
	}
}

/*
	LocalIDMap implementation
*/

type LocalIDMap struct {
	LocalID int
	Context *QNameContext
}

func NewLocalIDMap(localID int, qnc *QNameContext) LocalIDMap {
	return LocalIDMap{
		LocalID: localID,
		Context: qnc,
	}
}

/*
	AbstractStringCoder implementation
*/

type AbstractStringCoder struct {
	StringCoder
	localValuePartitions bool
	localValues          map[QNameContextMapKey][]*StringValue
}

func NewAbstractStringCoder(localValuePartitions bool, initialQNameLists int) *AbstractStringCoder {
	return &AbstractStringCoder{
		localValuePartitions: localValuePartitions,
		localValues:          make(map[QNameContextMapKey][]*StringValue, initialQNameLists),
	}
}

func (c *AbstractStringCoder) GetNumberOfStringValues(qnc *QNameContext) int {
	n := 0
	lvs, exists := c.localValues[qnc.GetMapKey()]
	if exists {
		n = len(lvs)
	}
	return n
}

func (c *AbstractStringCoder) Clear() {
	// local context
	if c.localValuePartitions {
		// free strings only, not destroy lists itself
		for key := range maps.Keys(c.localValues) {
			c.localValues[key] = []*StringValue{}
		}
	}
}

func (c *AbstractStringCoder) IsLocalValuePartitions() bool {
	return c.localValuePartitions
}

func (c *AbstractStringCoder) addLocalValue(qnc *QNameContext, value *StringValue) {
	if c.localValuePartitions {
		lvs, exists := c.localValues[qnc.GetMapKey()]
		if !exists {
			lvs = []*StringValue{}
		}
		lvs = append(lvs, value)
		c.localValues[qnc.GetMapKey()] = lvs
	}
}

/*
	StringDecoderImpl implementation
*/

type StringDecoderImpl struct {
	*AbstractStringCoder
	globalValues []*StringValue
}

func NewStringDecoderImpl(localValuePartitions bool) *StringDecoderImpl {
	return NewStringDecoderImplWithInitialQNameLists(localValuePartitions, DefaultInitialQNameLists)
}

func NewStringDecoderImplWithInitialQNameLists(localValuePartitions bool, initialQNameLists int) *StringDecoderImpl {
	return &StringDecoderImpl{
		AbstractStringCoder: NewAbstractStringCoder(localValuePartitions, initialQNameLists),
		globalValues:        []*StringValue{},
	}
}

func (sd *StringDecoderImpl) AddValue(qnc *QNameContext, value *StringValue) error {
	return nil
}

func (sd *StringDecoderImpl) ReadValue(qnc *QNameContext, channel DecoderChannel) (*StringValue, error) {
	var value *StringValue = nil
	var err error

	i, err := channel.DecodeUnsignedInteger()
	if err != nil {
		return nil, err
	}

	switch i {
	case 0:
		// local value partition
		if sd.localValuePartitions {
			value, err = sd.ReadValueLocalHit(qnc, channel)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("EXI stream contains local-value hit even though profile options indicate otherwise")
		}
	case 1:
		// found in global value partition
		value, err = sd.ReadValueGlobalHit(channel)
		if err != nil {
			return nil, err
		}
	default:
		// not found in global value (and local value) partition
		// ==> string literal is encoded as a String with the length
		// incremented by two.
		len := i - 2

		/*
		 * If length L is greater than zero the string S is added
		 */
		if len > 0 {
			runes, err := channel.DecodeStringOnly(len)
			if err != nil {
				return nil, err
			}
			value = NewStringValueFromSlice(runes)
			// After encoding the string value, it is added to both the
			// associated "local" value string table partition and the
			// global value string table partition.
			if err := sd.AddValue(qnc, value); err != nil {
				return nil, err
			}
		} else {
			value = EmptyStringValue
		}
	}

	if value == nil {
		return nil, errors.New("nil value")
	}
	return value, nil
}

func (sd *StringDecoderImpl) ReadValueLocalHit(qnc *QNameContext, channel DecoderChannel) (*StringValue, error) {
	if !sd.localValuePartitions {
		return nil, errors.New("local value partitions are not used")
	}

	n := utils.GetCodingLength(sd.GetNumberOfStringValues(qnc))
	localID, err := channel.DecodeNBitUnsignedInteger(n)
	if err != nil {
		return nil, err
	}
	lvs, ok := sd.localValues[qnc.GetMapKey()]
	if !ok {
		return nil, fmt.Errorf("local values does not found for QNameContext: %+v", qnc.GetMapKey())
	}
	if localID >= len(lvs) {
		return nil, errors.New("out of bounds")
	}

	return lvs[localID], nil
}

func (sd *StringDecoderImpl) ReadValueGlobalHit(channel DecoderChannel) (*StringValue, error) {
	numberBitsGlobal := utils.GetCodingLength(len(sd.globalValues))
	globalID, err := channel.DecodeNBitUnsignedInteger(numberBitsGlobal)
	if err != nil {
		return nil, err
	}
	return sd.globalValues[globalID], nil
}

func (sd *StringDecoderImpl) Clear() {
	sd.AbstractStringCoder.Clear()
	sd.globalValues = []*StringValue{}
}

func (sd *StringDecoderImpl) SetSharedStrings(sharedStrings []string) error {
	for _, s := range sharedStrings {
		if err := sd.AddValue(nil, NewStringValueFromString(s)); err != nil {
			return err
		}
	}
	return nil
}

/*
	StringEncoderImpl implementation
*/

type StringEncoderImpl struct {
	*AbstractStringCoder
	stringValues map[string]ValueContainer
}

func NewStringEncoderImpl(localValuePartitions bool) *StringEncoderImpl {
	return NewStringEncoderImplWithInitialQNameLists(localValuePartitions, DefaultInitialQNameLists)
}

func NewStringEncoderImplWithInitialQNameLists(localValuePartitions bool, initialQNameLists int) *StringEncoderImpl {
	return &StringEncoderImpl{
		AbstractStringCoder: NewAbstractStringCoder(localValuePartitions, initialQNameLists),
		stringValues:        map[string]ValueContainer{},
	}
}

func (se *StringEncoderImpl) AddValue(qnc *QNameContext, value string) error {
	if utils.ContainsKey(se.stringValues, value) {
		panic("attempt to add dupplicate global string value")
	}

	// global context
	se.stringValues[value] = NewValueContainer(value, qnc, se.GetNumberOfStringValues(qnc), len(se.stringValues))
	// local context
	se.addLocalValue(qnc, NewStringValueFromString(value))

	return nil
}

func (se *StringEncoderImpl) WriteValue(qnc *QNameContext, channel EncoderChannel, value string) error {
	vc, ok := se.stringValues[value]

	if ok {
		// hit
		if se.localValuePartitions && qnc.Equals(vc.Context) {
			/*
			 * local value hit ==> is represented as zero (0) encoded as an
			 * Unsigned Integer followed by the compact identifier of the
			 * string value in the "local" value partition
			 */
			if err := channel.EncodeUnsignedInteger(0); err != nil {
				return err
			}
			numberBitsLocal := utils.GetCodingLength(se.GetNumberOfStringValues(qnc))
			return channel.EncodeNBitUnsignedInteger(vc.LocalValueID, numberBitsLocal)
		} else {
			/*
			 * global value hit ==> value is represented as one (1) encoded
			 * as an Unsigned Integer followed by the compact identifier of
			 * the String value in the global value partition.
			 */
			if err := channel.EncodeUnsignedInteger(1); err != nil {
				return err
			}
			numberBitsGlobal := utils.GetCodingLength(len(se.stringValues))
			return channel.EncodeNBitUnsignedInteger(vc.GlobalValueID, numberBitsGlobal)
		}
	} else {
		/*
		 * miss [not found in local nor in global value partition] ==>
		 * string literal is encoded as a String with the length incremented
		 * by two.
		 */
		runes := []rune(value)
		len := len(runes)

		if err := channel.EncodeUnsignedInteger(len + 2); err != nil {
			return err
		}
		/*
		 * If length L is greater than zero the string S is added
		 */
		if len > 0 {
			if err := channel.EncodeStringOnly(value); err != nil {
				return err
			}
			// After encoding the string value, it is added to both the
			// associated "local" value string table partition and the
			// global value string table partition.
			if err := se.AddValue(qnc, value); err != nil {
				return err
			}
		}
	}

	return nil
}

func (se *StringEncoderImpl) IsStringHit(value string) (bool, error) {
	return utils.ContainsKey(se.stringValues, value), nil
}

func (se *StringEncoderImpl) GetValueContainer(value string) *ValueContainer {
	vc, ok := se.stringValues[value]
	if ok {
		return &vc
	} else {
		return nil
	}
}

func (se *StringEncoderImpl) GetValueContainerSize() int {
	return len(se.stringValues)
}

func (se *StringEncoderImpl) Clear() {
	se.AbstractStringCoder.Clear()
	se.stringValues = map[string]ValueContainer{}
}

func (se *StringEncoderImpl) SetSharedStrings(sharedStrings []string) error {
	for _, s := range sharedStrings {
		if err := se.AddValue(nil, s); err != nil {
			return err
		}
	}

	return nil
}

/*
	BoundedStringDecoderImpl implementation
*/

type BoundedStringDecoderImpl struct {
	*StringDecoderImpl
	valueMaxLength         int
	valuePartitionCapacity int
	globalID               int
	localIDMapping         []LocalIDMap
}

func NewBoundedStringDecoderImpl(localValuePartitions bool, valueMaxLength, valuePartitionCapacity int) *BoundedStringDecoderImpl {
	lmapSize := 0
	if valuePartitionCapacity > 0 && localValuePartitions {
		lmapSize = valuePartitionCapacity
	}

	return &BoundedStringDecoderImpl{
		StringDecoderImpl:      NewStringDecoderImpl(localValuePartitions),
		valueMaxLength:         valueMaxLength,
		valuePartitionCapacity: valuePartitionCapacity,
		globalID:               -1,
		localIDMapping:         make([]LocalIDMap, lmapSize),
	}
}

func (sd *BoundedStringDecoderImpl) AddValue(qnc *QNameContext, value *StringValue) error {
	clen, err := value.GetCharactersLength()
	if err != nil {
		return err
	}

	// first: check "valueMaxLength"
	if sd.valueMaxLength < 0 || clen <= sd.valueMaxLength {
		// next: check "valuePartitionCapacity"
		if sd.valuePartitionCapacity < 0 {
			// no "valuePartitionCapacity" restriction
			return sd.StringDecoderImpl.AddValue(qnc, value)
		} else {
			// If valuePartitionCapacity is not zero the string S is added
			if sd.valuePartitionCapacity == 0 {
				// no values per partition
			} else {
				/*
				 * When S is added to the global value partition and there was
				 * already a string V in the global value partition associated
				 * with the compact identifier globalID, the string S replaces
				 * the string V in the global table, and the string V is removed
				 * from its associated local value partition by rendering its
				 * compact identifier permanently unassigned.
				 */
				if slices.Contains(sd.globalValues, value) {
					return errors.New("duplicate global string values")
				}

				/*
				 * When the string value is added to the global value partition,
				 * the value of globalID is incremented by one (1). If the
				 * resulting value of globalID is equal to
				 * valuePartitionCapacity, its value is reset to zero (0)
				 */
				sd.globalID++
				if sd.globalID == sd.valuePartitionCapacity {
					sd.globalID = 0
				}

				if len(sd.globalValues) > sd.globalID {
					sd.globalValues[sd.globalID] = value
					// Java source have a quite weird code here!
				} else {
					// Need to check twice?
					if slices.Contains(sd.globalValues, value) {
						return errors.New("duplicate global string values")
					}
					sd.globalValues = append(sd.globalValues, value)
				}

				if sd.localValuePartitions {
					// update local ID mapping
					sd.localIDMapping[sd.globalID] = NewLocalIDMap(sd.GetNumberOfStringValues(qnc), qnc)
					// local value
					sd.addLocalValue(qnc, value)
				}
			}
		}
	}

	return nil
}

func (sd *BoundedStringDecoderImpl) Clear() {
	sd.StringDecoderImpl.Clear()
	sd.globalID = -1
}

/*
	BoundedStringEncoderImpl implementation
*/

type BoundedStringEncoderImpl struct {
	*StringEncoderImpl
	valueMaxLength         int
	valuePartitionCapacity int
	globalID               int
	globalIDMapping        []ValueContainer
}

func NewBoundedStringEncoderImpl(localValuePartitions bool, valueMaxLength, valuePartitionCapacity int) *BoundedStringEncoderImpl {
	return &BoundedStringEncoderImpl{
		StringEncoderImpl:      NewStringEncoderImpl(localValuePartitions),
		valueMaxLength:         valueMaxLength,
		valuePartitionCapacity: valuePartitionCapacity,
		globalID:               -1,
		globalIDMapping:        make([]ValueContainer, utils.Max(0, valuePartitionCapacity)),
	}
}

func (se *BoundedStringEncoderImpl) AddValue(qnc *QNameContext, value string) error {
	// first: check "valueMaxLength"
	if se.valueMaxLength < 0 || len(value) <= se.valueMaxLength {
		// next: check "valuePartitionCapacity"
		if se.valuePartitionCapacity < 0 {
			// no "valuePartitionCapacity" restriction
			if err := se.StringEncoderImpl.AddValue(qnc, value); err != nil {
				return err
			}
		} else {
			// If valuePartitionCapacity is not zero the string S is added
			if se.valuePartitionCapacity == 0 {
				// no values per partition
			} else {
				/*
				 * When S is added to the global value partition and there was
				 * already a string V in the global value partition associated
				 * with the compact identifier globalID, the string S replaces
				 * the string V in the global table, and the string V is removed
				 * from its associated local value partition by rendering its
				 * compact identifier permanently unassigned.
				 */
				if utils.ContainsKey(se.stringValues, value) {
					return errors.New("duplicate global string value")
				}

				se.globalID++
				if se.globalID == se.valuePartitionCapacity {
					se.globalID = 0
				}

				vc := NewValueContainer(value, qnc, se.GetNumberOfStringValues(qnc), se.globalID)

				if len(se.stringValues) == se.valuePartitionCapacity {
					// full --> remove old value
					vcFree := se.globalIDMapping[se.globalID]

					// free local
					if err := se.freeStringValue(vcFree.Context, vcFree.LocalValueID); err != nil {
						return err
					}

					// remove global
					delete(se.stringValues, vcFree.Value)
				}

				// add global
				se.stringValues[value] = vc

				// add local
				se.addLocalValue(qnc, NewStringValueFromString(value))
				se.globalIDMapping[se.globalID] = vc
			}
		}
	}

	return nil
}

func (se *BoundedStringEncoderImpl) freeStringValue(qnc *QNameContext, localValueID int) error {
	if se.localValuePartitions {
		lvs, ok := se.localValues[qnc.GetMapKey()]
		if !ok {
			return fmt.Errorf("local value missing: %+v", qnc.GetMapKey())
		}
		if localValueID >= len(se.localValues) {
			return errors.New("local value ID is larger that local values map size")
		}
		sv := lvs[localValueID]
		if sv == nil {
			return errors.New("local value is nil")
		}
		lvs[localValueID] = nil
	}

	return nil
}

func (se *BoundedStringEncoderImpl) Clear() {
	se.StringEncoderImpl.Clear()
	se.globalID = -1
}
