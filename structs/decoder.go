package structs

import (
	"bufio"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/sderkacs/exi-go/core"
)

// StructResolver defines an interface for resolving structures by name
type StructResolver interface {
	// ResolveStruct creates and returns a pointer to a struct instance based on the element name
	ResolveStruct(elementName string) (any, error)
}

// StructDecoder decodes EXI data directly into Go structures using reflection
type StructDecoder struct {
	noOptionsFactory core.EXIFactory
	exiStream        core.EXIStreamDecoder
	exiBodyOnly      bool
	debug            bool
	typeRegistry     map[string]reflect.Type // Maps field paths to concrete types for interfaces
}

// NewStructDecoder creates a new struct decoder
func NewStructDecoder(noOptionsFactory core.EXIFactory) (*StructDecoder, error) {
	exiStream, err := noOptionsFactory.CreateEXIStreamDecoder()
	if err != nil {
		return nil, err
	}

	return &StructDecoder{
		noOptionsFactory: noOptionsFactory,
		exiStream:        exiStream,
		exiBodyOnly:      false,
		debug:            false,
		typeRegistry:     make(map[string]reflect.Type),
	}, nil
}

// SetFeature sets decoder features like body-only mode
func (d *StructDecoder) SetFeature(name string, value bool) error {
	switch name {
	case core.W3C_EXI_FeatureBodyOnly:
		d.exiBodyOnly = value
	case "debug":
		d.debug = value
	default:
		return fmt.Errorf("struct decoder feature not supported: %s", name)
	}
	return nil
}

// RegisterType registers a concrete type for an interface field at the given path
// fieldPath should be in the format "FieldName" or "FieldName.SubFieldName" etc.
func (d *StructDecoder) RegisterType(fieldPath string, concreteType reflect.Type) {
	d.typeRegistry[fieldPath] = concreteType
}

// RegisterTypeFor is a convenience method to register a type using a sample instance
func (d *StructDecoder) RegisterTypeFor(fieldPath string, sample any) {
	d.typeRegistry[fieldPath] = reflect.TypeOf(sample)
}

// DecodeStruct decodes EXI data from source into the provided Go struct
func (d *StructDecoder) DecodeStruct(source *bufio.Reader, target any) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer to struct")
	}

	targetElem := targetValue.Elem()
	if targetElem.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	var decoder core.EXIBodyDecoder
	var err error
	if d.exiBodyOnly {
		decoder, err = d.exiStream.GetBodyOnlyDecoder(source)
		if err != nil {
			return err
		}
	} else {
		decoder, err = d.exiStream.DecodeHeader(source)
		if err != nil {
			return err
		}
	}

	return d.decodeEXIEvents(decoder, targetElem)
}

// Decode decodes EXI data from source using a resolver to construct the appropriate struct
// based on the root element name discovered during parsing
func (d *StructDecoder) Decode(source *bufio.Reader, resolver StructResolver) (any, error) {
	var decoder core.EXIBodyDecoder
	var err error
	if d.exiBodyOnly {
		decoder, err = d.exiStream.GetBodyOnlyDecoder(source)
		if err != nil {
			return nil, err
		}
	} else {
		decoder, err = d.exiStream.DecodeHeader(source)
		if err != nil {
			return nil, err
		}
	}

	return d.decodeEXIEventsWithResolver(decoder, resolver)
}

// decodeEXIEventsWithResolver processes EXI events using a resolver to determine the root structure
func (d *StructDecoder) decodeEXIEventsWithResolver(decoder core.EXIBodyDecoder, resolver StructResolver) (any, error) {
	elementStack := newElementStack()
	var rootStruct any

	eventType, exists, err := decoder.Next()
	if err != nil {
		return nil, err
	}

	for exists {
		if d.debug {
			fmt.Printf("[DEBUG] Processing event: %d\n", eventType)
		}

		switch eventType {
		case core.EventTypeStartDocument:
			if err := decoder.DecodeStartDocument(); err != nil {
				return nil, err
			}

		case core.EventTypeEndDocument:
			if err := decoder.DecodeEndDocument(); err != nil {
				return nil, err
			}

		case core.EventTypeAttributeXsiNil:
			qnc, err := decoder.DecodeAttributeXsiNil()
			if err != nil {
				return nil, err
			}

			if d.debug {
				fmt.Printf("[DEBUG] XSI Nil attribute: %s\n", qnc.GetLocalName())
			}
			if err := d.handleAttribute(elementStack, "nil", "true"); err != nil {
				if d.debug {
					fmt.Printf("[DEBUG] Failed to handle xsi:nil: %v\n", err)
				}
			}

		case core.EventTypeAttributeXsiType:
			qnc, err := decoder.DecodeAttributeXsiType()
			if err != nil {
				return nil, err
			}

			if d.debug {
				fmt.Printf("[DEBUG] XSI Type attribute: %s\n", qnc.GetLocalName())
			}
			if err := d.handleAttribute(elementStack, "type", qnc.GetLocalName()); err != nil {
				if d.debug {
					fmt.Printf("[DEBUG] Failed to handle xsi:type: %v\n", err)
				}
			}

		case core.EventTypeStartElement,
			core.EventTypeStartElementNS,
			core.EventTypeStartElementGeneric,
			core.EventTypeStartElementGenericUndeclared:

			qnc, err := decoder.DecodeStartElement()
			if err != nil {
				return nil, err
			}

			elementName := qnc.GetLocalName()
			if d.debug {
				fmt.Printf("[DEBUG] Start element: %s\n", elementName)
			}

			// Handle start element - for root element, use resolver
			if elementStack.isEmpty() {
				// Root element - use resolver to create the appropriate struct
				target, err := resolver.ResolveStruct(elementName)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve struct for root element '%s': %w", elementName, err)
				}

				// Validate that resolver returned a pointer to struct
				targetValue := reflect.ValueOf(target)
				if targetValue.Kind() != reflect.Ptr {
					return nil, fmt.Errorf("resolver must return a pointer to struct, got %T", target)
				}

				targetElem := targetValue.Elem()
				if targetElem.Kind() != reflect.Struct {
					return nil, fmt.Errorf("resolver must return a pointer to struct, got pointer to %s", targetElem.Kind())
				}

				rootStruct = target
				if d.debug {
					fmt.Printf("[DEBUG] Root element %s resolved to type: %s\n", elementName, targetElem.Type().Name())
				}

				// Push the root struct onto the stack
				elementStack.push(&stackFrame{
					value:       targetElem,
					elementName: elementName,
					fieldPath:   "",
				})
			} else {
				// Non-root element - use existing logic
				if err := d.handleStartElement(elementStack, elementStack.peek().value, elementName); err != nil {
					return nil, err
				}
			}

		case core.EventTypeEndElement, core.EventTypeEndElementUndeclared:
			qnc, err := decoder.DecodeEndElement()
			if err != nil {
				return nil, err
			}

			elementName := qnc.GetLocalName()
			if d.debug {
				fmt.Printf("[DEBUG] End element: %s\n", elementName)
			}

			if err := d.handleEndElement(elementStack); err != nil {
				return nil, err
			}

		case core.EventTypeCharacters, core.EventTypeCharactersGeneric, core.EventTypeCharactersGenericUndeclared:
			val, err := decoder.DecodeCharacters()
			if err != nil {
				return nil, err
			}

			text, err := d.extractTextValue(val)
			if err != nil {
				return nil, err
			}

			if d.debug {
				fmt.Printf("[DEBUG] Characters: %s\n", text)
			}

			if err := d.handleCharacters(elementStack, text); err != nil {
				return nil, err
			}

		case core.EventTypeAttribute,
			core.EventTypeAttributeNS,
			core.EventTypeAttributeGeneric,
			core.EventTypeAttributeGenericUndeclared,
			core.EventTypeAttributeInvalidValue,
			core.EventTypeAttributeAnyInvalidValue:

			qnc, err := decoder.DecodeAttribute()
			if err != nil {
				return nil, err
			}

			attrName := qnc.GetLocalName()
			attrVal := decoder.GetAttributeValue()

			text, err := d.extractTextValue(attrVal)
			if err != nil {
				return nil, err
			}

			if d.debug {
				fmt.Printf("[DEBUG] Attribute: %s = %s\n", attrName, text)
			}

			if err := d.handleAttribute(elementStack, attrName, text); err != nil {
				if d.debug {
					fmt.Printf("[DEBUG] Failed to handle attribute %s: %v\n", attrName, err)
				}
			}

		default:
			if d.debug {
				fmt.Printf("[DEBUG] Skipping event type: %d\n", eventType)
			}
		}

		eventType, exists, err = decoder.Next()
		if err != nil {
			return nil, err
		}
	}

	return rootStruct, nil
}

// decodeEXIEvents processes EXI events and populates the target struct
func (d *StructDecoder) decodeEXIEvents(decoder core.EXIBodyDecoder, target reflect.Value) error {
	elementStack := newElementStack()

	eventType, exists, err := decoder.Next()
	if err != nil {
		return err
	}

	for exists {
		if d.debug {
			fmt.Printf("[DEBUG] Processing event: %d\n", eventType)
		}

		switch eventType {
		case core.EventTypeStartDocument:
			if err := decoder.DecodeStartDocument(); err != nil {
				return err
			}

		case core.EventTypeEndDocument:
			if err := decoder.DecodeEndDocument(); err != nil {
				return err
			}

		case core.EventTypeAttributeXsiNil:
			qnc, err := decoder.DecodeAttributeXsiNil()
			if err != nil {
				return err
			}

			if d.debug {
				fmt.Printf("[DEBUG] XSI Nil attribute: %s\n", qnc.GetLocalName())
			}
			// Handle xsi:nil attribute - typically indicates the element should be nil
			if err := d.handleAttribute(elementStack, "nil", "true"); err != nil {
				if d.debug {
					fmt.Printf("[DEBUG] Failed to handle xsi:nil: %v\n", err)
				}
			}

		case core.EventTypeAttributeXsiType:
			qnc, err := decoder.DecodeAttributeXsiType()
			if err != nil {
				return err
			}

			if d.debug {
				fmt.Printf("[DEBUG] XSI Type attribute: %s\n", qnc.GetLocalName())
			}
			// Handle xsi:type attribute - indicates the runtime type
			if err := d.handleAttribute(elementStack, "type", qnc.GetLocalName()); err != nil {
				if d.debug {
					fmt.Printf("[DEBUG] Failed to handle xsi:type: %v\n", err)
				}
			}

		case core.EventTypeStartElement,
			core.EventTypeStartElementNS,
			core.EventTypeStartElementGeneric,
			core.EventTypeStartElementGenericUndeclared:

			qnc, err := decoder.DecodeStartElement()
			if err != nil {
				return err
			}

			elementName := qnc.GetLocalName()
			if d.debug {
				fmt.Printf("[DEBUG] Start element: %s\n", elementName)
			}

			// Handle start element
			if err := d.handleStartElement(elementStack, target, elementName); err != nil {
				return err
			}

		case core.EventTypeEndElement, core.EventTypeEndElementUndeclared:
			qnc, err := decoder.DecodeEndElement()
			if err != nil {
				return err
			}

			elementName := qnc.GetLocalName()
			if d.debug {
				fmt.Printf("[DEBUG] End element: %s\n", elementName)
			}

			// Handle end element
			if err := d.handleEndElement(elementStack); err != nil {
				return err
			}

		case core.EventTypeCharacters, core.EventTypeCharactersGeneric, core.EventTypeCharactersGenericUndeclared:
			val, err := decoder.DecodeCharacters()
			if err != nil {
				return err
			}

			text, err := d.extractTextValue(val)
			if err != nil {
				return err
			}

			if d.debug {
				fmt.Printf("[DEBUG] Characters: %s\n", text)
			}

			// Set the text value in the current field
			if err := d.handleCharacters(elementStack, text); err != nil {
				return err
			}

		case core.EventTypeAttribute,
			core.EventTypeAttributeNS,
			core.EventTypeAttributeGeneric,
			core.EventTypeAttributeGenericUndeclared,
			core.EventTypeAttributeInvalidValue,
			core.EventTypeAttributeAnyInvalidValue:

			qnc, err := decoder.DecodeAttribute()
			if err != nil {
				return err
			}

			attrName := qnc.GetLocalName()
			attrVal := decoder.GetAttributeValue()

			text, err := d.extractTextValue(attrVal)
			if err != nil {
				return err
			}

			if d.debug {
				fmt.Printf("[DEBUG] Attribute: %s = %s\n", attrName, text)
			}

			// Handle attribute mapping to struct fields
			if err := d.handleAttribute(elementStack, attrName, text); err != nil {
				if d.debug {
					fmt.Printf("[DEBUG] Failed to handle attribute %s: %v\n", attrName, err)
				}
				// Continue processing even if attribute handling fails
			}

		default:
			// Skip other event types for now
			if d.debug {
				fmt.Printf("[DEBUG] Skipping event type: %d\n", eventType)
			}
		}

		eventType, exists, err = decoder.Next()
		if err != nil {
			return err
		}
	}

	return nil
}

// handleStartElement processes a start element event
func (d *StructDecoder) handleStartElement(stack *elementStack, target reflect.Value, elementName string) error {
	if stack.isEmpty() {
		// Root element - find matching field or use target directly if names match
		targetType := target.Type()
		if d.debug {
			fmt.Printf("[DEBUG] Root element %s, target type: %s\n", elementName, targetType.Name())
		}

		// For root element, just push the target onto the stack
		stack.push(&stackFrame{
			value:       target,
			elementName: elementName,
			fieldPath:   "",
		})
		return nil
	}

	// Get current context
	current := stack.peek()
	if current == nil {
		return fmt.Errorf("no current context on stack")
	}

	// Find field in current struct that matches element name
	field, fieldName, err := d.findFieldWithName(current.value, elementName)
	if err != nil {
		if d.debug {
			fmt.Printf("[DEBUG] Field not found for element %s: %v\n", elementName, err)
		}
		// Push nil to maintain stack balance
		stack.push(&stackFrame{
			value:       reflect.Value{},
			elementName: elementName,
			fieldPath:   d.buildFieldPath(current.fieldPath, elementName),
			isSliceItem: false,
		})
		return nil
	}

	// Build the field path
	fieldPath := d.buildFieldPath(current.fieldPath, fieldName)

	if d.debug {
		fmt.Printf("[DEBUG] Found field for element %s, type: %s, path: %s\n", elementName, field.Type(), fieldPath)
	}

	// Handle slice types - but only for struct/interface/pointer slices
	// Simple slices ([]byte, []string, etc.) are handled by setFieldValue via character data
	if field.Kind() == reflect.Slice {
		elemType := field.Type().Elem()

		// Handle complex slice elements (structs, pointers to structs, interfaces)
		// Also handle custom types that are themselves slices
		// Simple built-in types like []byte, []string are handled via character data in setFieldValue
		if elemType.Kind() == reflect.Struct ||
			elemType.Kind() == reflect.Interface ||
			elemType.Kind() == reflect.Slice || // Custom types like CertificateType ([]byte)
			(elemType.Kind() == reflect.Ptr && elemType.Elem().Kind() == reflect.Struct) {

			// Create a new element for the slice
			workingValue, appendValue, err := d.createSliceItem(field, fieldPath)
			if err != nil {
				if d.debug {
					fmt.Printf("[DEBUG] Failed to create slice item for path %s: %v\n", fieldPath, err)
				}
				// Push nil to maintain stack balance
				stack.push(&stackFrame{
					value:       reflect.Value{},
					elementName: elementName,
					fieldPath:   fieldPath,
					isSliceItem: false,
				})
				return nil
			}

			// Push the slice item onto the stack
			if d.debug {
				fmt.Printf("[DEBUG] Creating slice item for element %s, current slice length: %d\n",
					elementName, field.Len())
			}
			stack.push(&stackFrame{
				value:        workingValue,
				elementName:  elementName,
				fieldPath:    fieldPath,
				isSliceItem:  true,
				sliceField:   field,
				sliceItemPtr: appendValue,
			})
			return nil
		}
		// For simple slice types, fall through to normal field handling
		// They will be populated via character data
	}

	// Handle interface types
	if field.Kind() == reflect.Interface {
		field, err = d.createInterfaceInstance(field, fieldPath)
		if err != nil {
			if d.debug {
				fmt.Printf("[DEBUG] Failed to create interface instance for path %s: %v\n", fieldPath, err)
			}
			// Push nil to maintain stack balance
			stack.push(&stackFrame{
				value:       reflect.Value{},
				elementName: elementName,
				fieldPath:   fieldPath,
				isSliceItem: false,
			})
			return nil
		}
	}

	// Handle pointer to struct types (optional fields)
	if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct {
		// Create new instance if nil
		if field.IsNil() {
			newInstance := reflect.New(field.Type().Elem())
			field.Set(newInstance)
			if d.debug {
				fmt.Printf("[DEBUG] Created new instance for pointer field %s\n", fieldName)
			}
		}
		// Dereference the pointer to work with the struct
		field = field.Elem()
	}

	// Push the field onto the stack
	stack.push(&stackFrame{
		value:       field,
		elementName: elementName,
		fieldPath:   fieldPath,
		isSliceItem: false,
	})

	return nil
}

// handleCharacters processes character data
func (d *StructDecoder) handleCharacters(stack *elementStack, text string) error {
	current := stack.peek()
	if current == nil || !current.value.IsValid() {
		return nil // No current context or invalid field
	}

	// If current value is a struct, look for a field tagged with xml:",chardata"
	if current.value.Kind() == reflect.Struct {
		field, fieldName, err := d.findCharDataField(current.value)
		if err != nil {
			if d.debug {
				fmt.Printf("[DEBUG] No chardata field found in struct: %v\n", err)
			}
			// If no chardata field found, this might be a struct that doesn't accept character data
			return nil
		}

		if d.debug {
			fmt.Printf("[DEBUG] Found chardata field %s for character data\n", fieldName)
		}

		// Set the character data on the specific field
		return d.setFieldValue(field, text)
	}

	// For non-struct values, set the value directly
	return d.setFieldValue(current.value, text)
}

// handleEndElement processes an end element event and handles slice item completion
func (d *StructDecoder) handleEndElement(stack *elementStack) error {
	current := stack.pop()
	if current == nil {
		return nil
	}

	// If this was a slice item, append it to the slice
	if current.isSliceItem && current.sliceField.IsValid() {
		if d.debug {
			fmt.Printf("[DEBUG] Completing slice item for element %s, slice length before: %d\n",
				current.elementName, current.sliceField.Len())
		}

		// Use the appropriate value to append (handles pointer vs value types)
		valueToAppend := current.sliceItemPtr
		if !valueToAppend.IsValid() {
			valueToAppend = current.value
		}

		// Append the completed item to the slice
		newSlice := reflect.Append(current.sliceField, valueToAppend)
		current.sliceField.Set(newSlice)

		if d.debug {
			fmt.Printf("[DEBUG] Slice now has %d items\n", newSlice.Len())
		}
	}

	return nil
}

// handleAttribute processes an attribute and maps it to a struct field
func (d *StructDecoder) handleAttribute(stack *elementStack, attrName, attrValue string) error {
	current := stack.peek()
	if current == nil || !current.value.IsValid() {
		return nil // No current context or invalid field
	}

	// Only handle attributes for struct types
	if current.value.Kind() != reflect.Struct {
		return nil
	}

	// Find field that matches this attribute
	field, fieldName, err := d.findAttributeField(current.value, attrName)
	if err != nil {
		// Attribute field not found - this is not an error, just skip
		return nil
	}

	if d.debug {
		fmt.Printf("[DEBUG] Found attribute field %s for attribute %s\n", fieldName, attrName)
	}

	// Set the attribute value
	return d.setFieldValue(field, attrValue)
}

// findFieldWithName finds a struct field that matches the given element name and returns the field name
func (d *StructDecoder) findFieldWithName(structValue reflect.Value, elementName string) (reflect.Value, string, error) {
	if !structValue.IsValid() || structValue.Kind() != reflect.Struct {
		return reflect.Value{}, "", fmt.Errorf("not a valid struct")
	}

	structType := structValue.Type()

	// Try exact match first
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Check XML tag first
		if xmlTag := field.Tag.Get("xml"); xmlTag != "" {
			// Parse XML tag which may contain namespace URI and local name
			// Format: "namespace localname" or just "localname"
			tagParts := strings.Fields(xmlTag)
			var localName string

			if len(tagParts) >= 2 {
				// Has namespace: "http://someNamespace SomeValue"
				localName = tagParts[len(tagParts)-1] // Take the last part as local name
			} else if len(tagParts) == 1 {
				// No namespace, split on comma for attributes like "name,attr"
				localName = strings.Split(tagParts[0], ",")[0]
			}

			if localName == elementName {
				return structValue.Field(i), field.Name, nil
			}
		}

		// Check direct field name match
		if field.Name == elementName {
			return structValue.Field(i), field.Name, nil
		}

		// Check case-insensitive match
		if strings.EqualFold(field.Name, elementName) {
			return structValue.Field(i), field.Name, nil
		}
	}

	return reflect.Value{}, "", fmt.Errorf("field not found for element: %s", elementName)
}

// findAttributeField finds a struct field that matches the given attribute name
func (d *StructDecoder) findAttributeField(structValue reflect.Value, attrName string) (reflect.Value, string, error) {
	if !structValue.IsValid() || structValue.Kind() != reflect.Struct {
		return reflect.Value{}, "", fmt.Errorf("not a valid struct")
	}

	structType := structValue.Type()

	// Look for fields with attribute tags
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Check XML tag for attribute marker
		if xmlTag := field.Tag.Get("xml"); xmlTag != "" {
			// Parse XML tag - attributes are marked with ",attr" suffix
			// Examples: "name,attr", "urn:namespace localname,attr"

			tagParts := strings.Split(xmlTag, ",")
			isAttribute := false

			// Check if this is marked as an attribute
			for _, part := range tagParts {
				if strings.TrimSpace(part) == "attr" {
					isAttribute = true
					break
				}
			}

			if isAttribute {
				// Extract the attribute name (everything before the first comma)
				attrTagName := strings.TrimSpace(tagParts[0])

				// Handle namespaced attributes like "urn:namespace localname"
				namespaceParts := strings.Fields(attrTagName)
				var localName string

				if len(namespaceParts) >= 2 {
					// Has namespace: take the last part as local name
					localName = namespaceParts[len(namespaceParts)-1]
				} else {
					// No namespace
					localName = attrTagName
				}

				if localName == attrName {
					return structValue.Field(i), field.Name, nil
				}
			}
		}

		// Also check direct field name match for attributes (fallback)
		if field.Name == attrName {
			return structValue.Field(i), field.Name, nil
		}

		// Check case-insensitive match for attributes (fallback)
		if strings.EqualFold(field.Name, attrName) {
			return structValue.Field(i), field.Name, nil
		}
	}

	return reflect.Value{}, "", fmt.Errorf("attribute field not found for: %s", attrName)
}

// findCharDataField finds a struct field that is tagged with xml:",chardata"
func (d *StructDecoder) findCharDataField(structValue reflect.Value) (reflect.Value, string, error) {
	if !structValue.IsValid() || structValue.Kind() != reflect.Struct {
		return reflect.Value{}, "", fmt.Errorf("not a valid struct")
	}

	structType := structValue.Type()

	// Look for field with xml:",chardata" tag
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Check XML tag for chardata marker
		if xmlTag := field.Tag.Get("xml"); xmlTag != "" {
			// Check for ",chardata" suffix or exact match ",chardata"
			if xmlTag == ",chardata" || strings.HasSuffix(xmlTag, ",chardata") {
				return structValue.Field(i), field.Name, nil
			}
		}
	}

	return reflect.Value{}, "", fmt.Errorf("chardata field not found")
}

// buildFieldPath constructs a dot-separated field path
func (d *StructDecoder) buildFieldPath(parentPath, fieldName string) string {
	if parentPath == "" {
		return fieldName
	}
	return parentPath + "." + fieldName
}

// createInterfaceInstance creates a concrete instance for an interface field
func (d *StructDecoder) createInterfaceInstance(interfaceField reflect.Value, fieldPath string) (reflect.Value, error) {
	// Look up the registered concrete type
	concreteType, exists := d.typeRegistry[fieldPath]
	if !exists {
		return reflect.Value{}, fmt.Errorf("no concrete type registered for interface field at path: %s", fieldPath)
	}

	// Create new instance of the concrete type
	var newInstance reflect.Value
	if concreteType.Kind() == reflect.Ptr {
		newInstance = reflect.New(concreteType.Elem())
	} else {
		newInstance = reflect.New(concreteType)
	}

	// Set the interface field to point to the new instance
	if concreteType.Kind() == reflect.Ptr {
		interfaceField.Set(newInstance)
		return newInstance.Elem(), nil
	} else {
		interfaceField.Set(newInstance.Elem())
		return newInstance.Elem(), nil
	}
}

// createSliceItem creates a new item for a slice field
// Returns (workingValue, pointerValue, error) where:
// - workingValue is what we populate during parsing
// - pointerValue is what gets appended to the slice (for pointer types)
func (d *StructDecoder) createSliceItem(sliceField reflect.Value, fieldPath string) (reflect.Value, reflect.Value, error) {
	sliceType := sliceField.Type()
	elemType := sliceType.Elem()

	if d.debug {
		fmt.Printf("[DEBUG] Creating slice item of type %s for path %s\n", elemType, fieldPath)
	}

	// Handle different element types
	switch elemType.Kind() {
	case reflect.Struct:
		// Create new struct instance
		newItem := reflect.New(elemType).Elem()
		return newItem, newItem, nil

	case reflect.Ptr:
		if elemType.Elem().Kind() == reflect.Struct {
			// Create new pointer to struct
			newPtr := reflect.New(elemType.Elem())
			// Return the dereferenced struct for population, and the pointer for appending
			return newPtr.Elem(), newPtr, nil
		}
		// For other pointer types, create and return the pointer
		newItem := reflect.New(elemType.Elem())
		return newItem, newItem, nil

	case reflect.Slice:
		// Handle custom slice types
		// Create new slice instance
		newItem := reflect.New(elemType).Elem()
		return newItem, newItem, nil

	case reflect.Interface:
		// For interface slices, try to find a registered concrete type
		itemPath := fieldPath + "[]" // Special notation for slice items
		concreteType, exists := d.typeRegistry[itemPath]
		if !exists {
			return reflect.Value{}, reflect.Value{}, fmt.Errorf("no concrete type registered for slice interface at path: %s", itemPath)
		}

		var newInstance reflect.Value
		if concreteType.Kind() == reflect.Ptr {
			newInstance = reflect.New(concreteType.Elem())
			return newInstance.Elem(), newInstance.Elem(), nil
		} else {
			newInstance = reflect.New(concreteType)
			return newInstance.Elem(), newInstance.Elem(), nil
		}

	default:
		// For primitive types, create zero value
		newItem := reflect.New(elemType).Elem()
		return newItem, newItem, nil
	}
}

// setFieldValue sets a field value from a string representation
func (d *StructDecoder) setFieldValue(field reflect.Value, text string) error {
	if !field.IsValid() || !field.CanSet() {
		return nil
	}

	text = strings.TrimSpace(text)

	switch field.Kind() {
	case reflect.String:
		field.SetString(text)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if text == "" {
			return nil
		}
		val, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse int: %v", err)
		}
		field.SetInt(val)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if text == "" {
			return nil
		}
		val, err := strconv.ParseUint(text, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse uint: %v", err)
		}
		field.SetUint(val)

	case reflect.Float32, reflect.Float64:
		if text == "" {
			return nil
		}
		val, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return fmt.Errorf("cannot parse float: %v", err)
		}
		field.SetFloat(val)

	case reflect.Bool:
		if text == "" {
			return nil
		}
		val, err := strconv.ParseBool(text)
		if err != nil {
			return fmt.Errorf("cannot parse bool: %v", err)
		}
		field.SetBool(val)

	case reflect.Ptr:
		// Handle pointer types by creating new instance
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		// For pointer to struct with text content, this might be invalid XML structure
		// but we'll handle it gracefully by setting the dereferenced value
		return d.setFieldValue(field.Elem(), text)

	case reflect.Slice:
		return d.setSliceValue(field, text)

	default:
		if d.debug {
			fmt.Printf("[DEBUG] Unsupported field type for text content: %s\n", field.Kind())
		}
		return nil
	}

	return nil
}

// setSliceValue sets a slice field value from a string representation
func (d *StructDecoder) setSliceValue(field reflect.Value, text string) error {
	if !field.IsValid() || !field.CanSet() {
		return nil
	}

	text = strings.TrimSpace(text)
	sliceType := field.Type()
	elemType := sliceType.Elem()

	// Handle []byte specifically for character data
	if elemType.Kind() == reflect.Uint8 {
		field.Set(reflect.ValueOf([]byte(text)))
		return nil
	}

	// Handle []rune for Unicode character data
	if elemType.Kind() == reflect.Int32 {
		field.Set(reflect.ValueOf([]rune(text)))
		return nil
	}

	// Handle slices of pointers to structs - this is unusual for text content
	// but we'll handle it by creating a single element
	if elemType.Kind() == reflect.Ptr && elemType.Elem().Kind() == reflect.Struct {
		if text == "" {
			field.Set(reflect.MakeSlice(sliceType, 0, 0))
			return nil
		}

		// Create a single element slice with a new struct instance
		slice := reflect.MakeSlice(sliceType, 1, 1)
		newInstance := reflect.New(elemType.Elem())
		slice.Index(0).Set(newInstance)
		field.Set(slice)
		return nil
	}

	// For other slice types, split text by whitespace and convert each element
	if text == "" {
		field.Set(reflect.MakeSlice(sliceType, 0, 0))
		return nil
	}

	parts := strings.Fields(text)
	slice := reflect.MakeSlice(sliceType, len(parts), len(parts))

	for i, part := range parts {
		elem := slice.Index(i)
		if err := d.setFieldValue(elem, part); err != nil {
			return fmt.Errorf("failed to set slice element %d: %v", i, err)
		}
	}

	field.Set(slice)
	return nil
}

// extractTextValue extracts text from a Value interface
func (d *StructDecoder) extractTextValue(val core.Value) (string, error) {
	switch val.GetValueType() {
	case core.ValueTypeBoolean, core.ValueTypeString:
		return val.ToString()

	case core.ValueTypeList:
		lv := val.(*core.ListValue)
		if lv.GetNumberOfValues() == 0 {
			return "", nil
		}

		values := lv.ToValues()
		var parts []string

		for _, v := range values {
			text, err := d.extractTextValue(v)
			if err != nil {
				return "", err
			}
			parts = append(parts, text)
		}

		return strings.Join(parts, " "), nil

	default:
		slen, err := val.GetCharactersLength()
		if err != nil {
			return "", err
		}

		buffer := make([]rune, slen)

		if err := val.FillCharactersBuffer(buffer, 0); err != nil {
			return "", err
		}

		return string(buffer[0:slen]), nil
	}
}

// Stack implementation for tracking element context
type stackFrame struct {
	value        reflect.Value
	elementName  string
	fieldPath    string        // Tracks the full field path for interface type resolution
	isSliceItem  bool          // Indicates if this frame represents an item in a slice
	sliceField   reflect.Value // Reference to the parent slice field (if isSliceItem is true)
	sliceItemPtr reflect.Value // For pointer types, stores the pointer to append to slice
}

type elementStack struct {
	frames []*stackFrame
}

func newElementStack() *elementStack {
	return &elementStack{
		frames: make([]*stackFrame, 0),
	}
}

func (s *elementStack) push(frame *stackFrame) {
	s.frames = append(s.frames, frame)
}

func (s *elementStack) pop() *stackFrame {
	if len(s.frames) == 0 {
		return nil
	}

	frame := s.frames[len(s.frames)-1]
	s.frames = s.frames[:len(s.frames)-1]
	return frame
}

func (s *elementStack) peek() *stackFrame {
	if len(s.frames) == 0 {
		return nil
	}
	return s.frames[len(s.frames)-1]
}

func (s *elementStack) isEmpty() bool {
	return len(s.frames) == 0
}
