package core

import (
	"fmt"
	"maps"

	"github.com/sderkacs/exi-go/utils"
)

/*
	EncodingOptions implementation
*/

const (
	// Include EXI Cookie
	OptionIncludeCookie string = "INCLUDE_COOKIE"

	// Include EXI Options (or opposite of omitOptionsDocument in Canonical EXI)
	OptionIncludeOptions string = "INCLUDE_OPTIONS"

	//  Include schemaID as part of EXI Options
	OptionIncludeSchemaID string = "INCLUDE_SCHEMA_ID"

	// Encode entity references as ER event instead of trying to resolve them
	OptionRetainEntityReference string = "KEEP_ENTITY_REFERENCES_UNRESOLVED"

	// Include attribute "schemaLocation" and "noNamespaceSchemaLocation"
	OptionIncludeXsiSchemaLocation string = "INCLUDE_XSI_SCHEMALOCATION"

	// Include Insignificant xsi:nil values e.g., xsi:nil="false"
	OptionIncludeInsignificanXsiNil string = "INCLUDE_INSIGNIFICANT_XSI_NIL"

	// To indicate that the EXI profile is in use and advertising each parameter
	// value (exi:p element)
	OptionIncludeProfileValues string = "INCLUDE_PROFILE_VALUES"

	// The utcTime option is used to specify whether Date-Time values must be
	// represented using Coordinated Universal Time (UTC, sometimes called
	// "Greenwich Mean Time").
	OptionUtcTime string = "UTC_TIME"

	// To indicate that the EXI stream should respect the Canonical EXI rules
	// (see http://www.w3.org/TR/exi-c14n)
	OptionCanonicalExi string = "http://www.w3.org/TR/exi-c14n"

	// To set the deflate stream with a specified compression level.
	OptionDeflateCompressionValue string = "DEFLATE_COMPRESSION_VALUE"
)

type EncodingOptions struct {
	options map[string]any
}

func NewEncodingOptions() *EncodingOptions {
	return &EncodingOptions{
		options: map[string]any{},
	}
}

func (o *EncodingOptions) SetOption(key string) error {
	return o.SetOptionKeyValue(key, nil)
}

func (o *EncodingOptions) SetOptionKeyValue(key string, value any) error {
	switch key {
	case OptionIncludeCookie, OptionIncludeSchemaID, OptionRetainEntityReference,
		OptionIncludeXsiSchemaLocation, OptionIncludeInsignificanXsiNil,
		OptionIncludeProfileValues, OptionUtcTime:
		o.options[key] = nil
	case OptionCanonicalExi:
		o.options[key] = nil
		// by default the Canonical EXI Option "omitOptionsDocument" is
		// false
		// --> include options
		o.options[OptionIncludeOptions] = nil
	case OptionDeflateCompressionValue:
		if value != nil {
			_, ok := value.(int)
			if ok {
				o.options[key] = value
				break
			}
		}

		return fmt.Errorf("EncodingOption '%s' requires value of type int", key)
	default:
		return fmt.Errorf("EncodingOption '%s' is unknown", key)
	}

	return nil
}

func (o *EncodingOptions) UnsetOption(key string) bool {
	exists := utils.ContainsKey(o.options, key)
	delete(o.options, key)
	return exists
}

func (o *EncodingOptions) IsOptionEnabled(key string) bool {
	return utils.ContainsKey(o.options, key)
}

func (o *EncodingOptions) GetOptionValue(key string) any {
	return o.options[key]
}

func (o *EncodingOptions) Equals(other *EncodingOptions) bool {
	if other == nil {
		return false
	}

	return maps.Equal(o.options, other.options)
}

/*
	DecodingOptions implementation
*/

const (
	// SchemaId in EXI header is not used
	OptionIgnoreSchemaID string = "IGNORE_SCHEMA_ID"

	// Pushback size for multiple streams in one file
	OptionPushbackBufferSize int = 512
)

type DecodingOptions struct {
	options map[string]any
}

func NewDecodingOptions() *DecodingOptions {
	return &DecodingOptions{
		options: map[string]any{},
	}
}

func (o *DecodingOptions) SetOption(key string) error {
	return o.SetOptionKeyValue(key, nil)
}

func (o *DecodingOptions) SetOptionKeyValue(key string, value any) error {
	switch key {
	case OptionIgnoreSchemaID:
		o.options[key] = nil
	default:
		return fmt.Errorf("DecodingOption '%s' is unknown", key)
	}

	return nil
}

func (o *DecodingOptions) UnsetOption(key string) bool {
	exists := utils.ContainsKey(o.options, key)
	delete(o.options, key)
	return exists
}

func (o *DecodingOptions) IsOptionEnabled(key string) bool {
	return utils.ContainsKey(o.options, key)
}

func (o *DecodingOptions) GetOptionValue(key string) any {
	return o.options[key]
}

func (o *DecodingOptions) Equals(other *DecodingOptions) bool {
	if other == nil {
		return false
	}

	return maps.Equal(o.options, other.options)
}
