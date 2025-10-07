package structs

import (
	"bufio"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/sderkacs/go-exi/core"
)

// StructEncoder encodes Go structures directly into EXI data using reflection
type StructEncoder struct {
	noOptionsFactory core.EXIFactory
	exiStream        core.EXIStreamEncoder
	exiBodyOnly      bool
	debug            bool
	exiAttributes    core.AttributeList
	elementStack     *encoderElementStack
	namespaceDecls   map[string]string // Maps prefix to namespace URI
	skipAnyType      bool              // Whether to skip xsi:type="anyType" attributes
}

// NewStructEncoder creates a new struct encoder
func NewStructEncoder(noOptionsFactory core.EXIFactory) (*StructEncoder, error) {
	exiStream, err := noOptionsFactory.CreateEXIStreamEncoder()
	if err != nil {
		return nil, err
	}

	return &StructEncoder{
		noOptionsFactory: noOptionsFactory,
		exiStream:        exiStream,
		exiBodyOnly:      false,
		debug:            false,
		exiAttributes:    core.NewAttributeListImpl(noOptionsFactory),
		elementStack:     newEncoderElementStack(),
		namespaceDecls:   make(map[string]string),
		skipAnyType:      true, // Default to skipping anyType attributes
	}, nil
}

// SetFeature sets encoder features like body-only mode
func (e *StructEncoder) SetFeature(name string, value bool) error {
	switch name {
	case core.W3C_EXI_FeatureBodyOnly:
		e.exiBodyOnly = value
	case "debug":
		e.debug = value
	case "skipAnyType":
		e.skipAnyType = value
	default:
		return fmt.Errorf("struct encoder feature not supported: %s", name)
	}
	return nil
}

// DeclareNamespace registers a namespace declaration that will be added to the root element
func (e *StructEncoder) DeclareNamespace(prefix, namespaceURI string) {
	e.namespaceDecls[prefix] = namespaceURI
}

// DeclareDefaultNamespace registers the default namespace (empty prefix)
func (e *StructEncoder) DeclareDefaultNamespace(namespaceURI string) {
	e.namespaceDecls[""] = namespaceURI
}

// ClearNamespaces clears all registered namespace declarations
func (e *StructEncoder) ClearNamespaces() {
	e.namespaceDecls = make(map[string]string)
}

// AutoDeclareNamespace automatically declares a namespace if not already declared
func (e *StructEncoder) AutoDeclareNamespace(prefix, namespaceURI string) {
	if _, exists := e.namespaceDecls[prefix]; !exists {
		e.namespaceDecls[prefix] = namespaceURI
		if e.debug {
			fmt.Printf("[DEBUG] Auto-declared namespace: %s -> %s\n", prefix, namespaceURI)
		}
	}
}

// parseAttributeNamespace extracts namespace/prefix and local name from an attribute name
func (e *StructEncoder) parseAttributeNamespace(attrName string) (namespace, localName string) {
	// Handle prefixed attribute names like "ns1:attr" or "xml:lang"
	if colonIndex := strings.Index(attrName, ":"); colonIndex != -1 {
		namespace = attrName[:colonIndex]
		localName = attrName[colonIndex+1:]
		return namespace, localName
	}

	// No prefix - attribute is in no namespace (or default namespace for elements)
	return "", attrName
}

// EncodeStruct encodes a Go struct into EXI data and writes it to the provided writer
func (e *StructEncoder) EncodeStruct(writer *bufio.Writer, source any, rootElementName string, ns string) error {
	sourceValue := reflect.ValueOf(source)

	// Handle pointer to struct
	if sourceValue.Kind() == reflect.Ptr {
		if sourceValue.IsNil() {
			return fmt.Errorf("source cannot be nil")
		}
		sourceValue = sourceValue.Elem()
	}

	if sourceValue.Kind() != reflect.Struct {
		return fmt.Errorf("source must be a struct or pointer to struct")
	}

	// Initialize the EXI encoder
	var encoder core.EXIBodyEncoder
	var err error
	if e.exiBodyOnly {
		// For body-only mode, we'd need to get body-only encoder
		// This would require additional factory methods
		return fmt.Errorf("body-only encoding not yet implemented")
	} else {
		encoder, err = e.exiStream.EncodeHeader(writer)
		if err != nil {
			return err
		}
	}

	// Start document
	if err := encoder.EncodeStartDocument(); err != nil {
		return err
	}

	// Encode the struct as the root element
	//rootElementName := e.getStructElementName(sourceValue.Type())
	if err := e.encodeStruct(encoder, sourceValue, rootElementName, ns); err != nil {
		return err
	}

	// End document and flush
	if err := encoder.EncodeEndDocument(); err != nil {
		return err
	}

	return encoder.Flush()
}

// encodeStruct encodes a struct value as an XML element
func (e *StructEncoder) encodeStruct(encoder core.EXIBodyEncoder, structValue reflect.Value, elementName, namespace string) error {
	if e.debug {
		fmt.Printf("[DEBUG] Encoding struct element: %s (namespace: %s)\n", elementName, namespace)
	}

	// Check if this is the root element (empty stack)
	isRootElement := e.elementStack.isEmpty()

	// Start element with resolved namespace URI and prefix
	namespaceURI := e.resolveNamespace(namespace)
	prefix := e.getPrefixForNamespace(namespace)

	if e.debug {
		prefixStr := "nil"
		if prefix != nil {
			prefixStr = *prefix
		}
		fmt.Printf("[DEBUG] EncodeStartElement (struct): uri=%s, localName=%s, prefix=%s\n", namespaceURI, elementName, prefixStr)
	}

	//TODO:!!!
	prefix = nil
	if err := encoder.EncodeStartElement(namespaceURI, elementName, prefix); err != nil {
		return err
	}

	// Push to stack for context
	e.elementStack.push(&encoderStackFrame{
		elementName: elementName,
		namespace:   namespace,
	})

	// Clear attributes for this element
	e.exiAttributes.Clear()

	// Add namespace declarations to root element only if explicitly declared
	if isRootElement && len(e.namespaceDecls) > 0 {
		e.addNamespaceDeclarations()
	}

	// First pass: collect all attributes
	structType := structValue.Type()
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Parse XML tag
		xmlTag := field.Tag.Get("xml")
		if xmlTag == "-" {
			continue // Skip fields marked with xml:"-"
		}

		fieldName, isAttribute, _, namespace := e.parseXMLTag(xmlTag, field.Name)

		if isAttribute {
			// Handle as attribute - for attributes, we need to resolve namespace to prefix
			var fullAttrName string
			if namespace != "" {
				// Check if namespace is a full URI that we need to resolve to a prefix
				if strings.Contains(namespace, "://") || strings.HasPrefix(namespace, "urn:") {
					// This is a full namespace URI - find the corresponding prefix
					prefix := e.findPrefixForNamespace(namespace)
					if prefix != "" {
						fullAttrName = prefix + ":" + fieldName
					} else {
						// No prefix found, use just local name
						fullAttrName = fieldName
					}
				} else {
					// This is already a prefix, use it directly
					fullAttrName = namespace + ":" + fieldName
				}
			} else {
				fullAttrName = fieldName
			}

			if err := e.encodeAttribute(fullAttrName, namespace, fieldValue); err != nil {
				if e.debug {
					fmt.Printf("[DEBUG] Failed to encode attribute %s: %v\n", fullAttrName, err)
				}
				// Continue with other fields even if one attribute fails
			}
		}
	}

	if e.debug {
		fmt.Printf("[DEBUG] Attribute list: %+v\n", e.exiAttributes)
	}

	// Encode all collected attributes BEFORE any content
	if err := encoder.EncodeAttributeList(e.exiAttributes); err != nil {
		return err
	}

	// Second pass: encode character data fields AFTER attributes and BEFORE child elements
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Parse XML tag
		xmlTag := field.Tag.Get("xml")
		if xmlTag == "-" {
			continue // Skip fields marked with xml:"-"
		}

		_, _, isCharData, _ := e.parseXMLTag(xmlTag, field.Name)

		if isCharData {
			// Handle as character data
			if err := e.encodeCharacterData(encoder, fieldValue); err != nil {
				return err
			}
		}
	}

	// Third pass: encode all element children
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Parse XML tag
		xmlTag := field.Tag.Get("xml")
		if xmlTag == "-" {
			continue // Skip fields marked with xml:"-"
		}

		fieldName, isAttribute, isCharData, namespace := e.parseXMLTag(xmlTag, field.Name)

		if !isAttribute && !isCharData {
			// Handle as element
			if err := e.encodeField(encoder, fieldValue, fieldName, namespace); err != nil {
				return err
			}
		}
	}

	if e.debug {
		fmt.Printf("[DEBUG] End element: %s\n", elementName)
	}

	// End element
	if err := encoder.EncodeEndElement(); err != nil {
		return err
	}

	// Pop from stack
	e.elementStack.pop()

	return nil
}

// encodeField encodes a struct field as an XML element
func (e *StructEncoder) encodeField(encoder core.EXIBodyEncoder, fieldValue reflect.Value, fieldName, namespace string) error {
	// Handle nil pointers
	if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
		// Skip nil pointers (optional elements)
		return nil
	}

	// Dereference pointers
	if fieldValue.Kind() == reflect.Ptr {
		fieldValue = fieldValue.Elem()
	}

	switch fieldValue.Kind() {
	case reflect.Struct:
		return e.encodeStruct(encoder, fieldValue, fieldName, namespace)

	case reflect.Slice:
		return e.encodeSlice(encoder, fieldValue, fieldName, namespace)

	case reflect.Interface:
		// Handle interface types by encoding the concrete value
		if fieldValue.IsNil() {
			return nil // Skip nil interfaces
		}
		// Get the concrete value and encode it
		concreteValue := fieldValue.Elem()
		return e.encodeField(encoder, concreteValue, fieldName, namespace)

	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Bool:
		return e.encodeSimpleField(encoder, fieldValue, fieldName, namespace)

	default:
		if e.debug {
			fmt.Printf("[DEBUG] Skipping unsupported field type: %s\n", fieldValue.Kind())
		}
		return nil
	}
}

// encodeSlice encodes a slice field
func (e *StructEncoder) encodeSlice(encoder core.EXIBodyEncoder, sliceValue reflect.Value, fieldName, namespace string) error {
	if sliceValue.IsNil() {
		return nil // Skip nil slices
	}

	sliceType := sliceValue.Type()
	elemType := sliceType.Elem()

	// Handle different slice element types
	switch elemType.Kind() {
	case reflect.Struct:
		// Handle slices of structs - each element becomes a separate XML element
		for i := 0; i < sliceValue.Len(); i++ {
			item := sliceValue.Index(i)
			if err := e.encodeField(encoder, item, fieldName, namespace); err != nil {
				return err
			}
		}

	case reflect.Ptr:
		// Handle slices of pointers (usually to structs)
		for i := 0; i < sliceValue.Len(); i++ {
			item := sliceValue.Index(i)
			if err := e.encodeField(encoder, item, fieldName, namespace); err != nil {
				return err
			}
		}

	case reflect.Interface:
		// Handle slices of interfaces - each element becomes a separate XML element
		for i := 0; i < sliceValue.Len(); i++ {
			item := sliceValue.Index(i)
			if err := e.encodeField(encoder, item, fieldName, namespace); err != nil {
				return err
			}
		}

	case reflect.Uint8:
		// Handle []byte as character data
		if elemType == reflect.TypeOf(byte(0)) {
			bytes := sliceValue.Bytes()
			return e.encodeSimpleValue(encoder, string(bytes), fieldName, namespace)
		}
		fallthrough

	default:
		// For primitive types, encode as space-separated values in a single element
		// For complex types that can't be stringified, encode each as separate elements
		if e.isSimpleSliceType(elemType) {
			// Handle simple slice types as space-separated values
			var parts []string
			for i := 0; i < sliceValue.Len(); i++ {
				item := sliceValue.Index(i)
				str, err := e.valueToString(item)
				if err != nil {
					if e.debug {
						fmt.Printf("[DEBUG] Failed to convert slice item to string, skipping: %v\n", err)
					}
					continue // Skip items that can't be converted to string
				}
				parts = append(parts, str)
			}
			if len(parts) > 0 {
				return e.encodeSimpleValue(encoder, strings.Join(parts, " "), fieldName, namespace)
			}
		} else {
			// For complex types, encode each item as a separate element
			for i := 0; i < sliceValue.Len(); i++ {
				item := sliceValue.Index(i)
				if err := e.encodeField(encoder, item, fieldName, namespace); err != nil {
					return err
				}
			}
		}
		return nil
	}

	return nil
}

// encodeSimpleField encodes a simple field (string, number, bool) as an element with character data
func (e *StructEncoder) encodeSimpleField(encoder core.EXIBodyEncoder, fieldValue reflect.Value, fieldName, namespace string) error {
	str, err := e.valueToString(fieldValue)
	if err != nil {
		return err
	}

	return e.encodeSimpleValue(encoder, str, fieldName, namespace)
}

// encodeSimpleValue encodes a string value as an element with character data
func (e *StructEncoder) encodeSimpleValue(encoder core.EXIBodyEncoder, value, elementName, namespace string) error {
	if e.debug {
		fmt.Printf("[DEBUG] Encoding simple element: %s = %s (namespace: %s)\n", elementName, value, namespace)
	}

	// Start element with resolved namespace URI and prefix
	namespaceURI := e.resolveNamespace(namespace)
	prefix := e.getPrefixForNamespace(namespace)

	if e.debug {
		prefixStr := "nil"
		if prefix != nil {
			prefixStr = *prefix
		}
		fmt.Printf("[DEBUG] EncodeStartElement (simple): uri=%s, localName=%s, prefix=%s\n", namespaceURI, elementName, prefixStr)
	}

	//TODO:!!!
	prefix = nil
	if err := encoder.EncodeStartElement(namespaceURI, elementName, prefix); err != nil {
		return err
	}

	// Encode empty attribute list BEFORE character data (simple elements have no attributes)
	e.exiAttributes.Clear()
	if err := encoder.EncodeAttributeList(e.exiAttributes); err != nil {
		return err
	}

	// Encode character data AFTER attributes
	if value != "" {
		if err := encoder.EncodeCharacters(core.NewStringValueFromSlice([]rune(value))); err != nil {
			return err
		}
	}

	// End element
	return encoder.EncodeEndElement()
}

// encodeAttribute adds an attribute to the current attribute list
func (e *StructEncoder) encodeAttribute(attrName string, namespace string, fieldValue reflect.Value) error {
	// Handle nil pointers
	if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
		return nil // Skip nil pointer attributes
	}

	// Dereference pointers
	if fieldValue.Kind() == reflect.Ptr {
		fieldValue = fieldValue.Elem()
	}

	str, err := e.valueToString(fieldValue)
	if err != nil {
		return err
	}

	// Filter out xsi:type attributes with anyType values
	if e.shouldSkipAttribute(attrName, str) {
		if e.debug {
			fmt.Printf("[DEBUG] Skipping attribute: %s = %s\n", attrName, str)
		}
		return nil
	}

	if e.debug {
		fmt.Printf("[DEBUG] Adding attribute: %s = %s\n", attrName, str)
	}

	// For attributes, use a simpler approach - don't auto-resolve namespaces
	// Only use namespace if explicitly specified in the attribute name
	if strings.Contains(attrName, ":") {
		// Attribute has explicit prefix, split it
		colonIndex := strings.Index(attrName, ":")
		prefix := attrName[:colonIndex]
		localName := attrName[colonIndex+1:]

		// Look up the namespace URI for this prefix
		var namespaceURI string
		if uri, exists := e.namespaceDecls[prefix]; exists {
			namespaceURI = uri
		}

		if e.debug {
			fmt.Printf("[DEBUG] Adding prefixed attribute: uri=%s, localName=%s, prefix=%s, value=%s\n",
				namespaceURI, localName, prefix, str)
		}

		var namespacePtr *string
		// if namespaceURI != "" {
		// 	namespacePtr = &namespaceURI
		// }
		namespacePtr = &namespace
		prefixPtr := &prefix
		e.exiAttributes.AddAttribute(namespacePtr, localName, prefixPtr, str)
	} else {
		// No prefix - attribute is in no namespace
		if e.debug {
			fmt.Printf("[DEBUG] Adding simple attribute: localName=%s, value=%s\n", attrName, str)
		}
		e.exiAttributes.AddAttribute(&namespace, attrName, nil, str)
	}

	return nil
}

// encodeCharacterData encodes a field as character data within the current element
func (e *StructEncoder) encodeCharacterData(encoder core.EXIBodyEncoder, fieldValue reflect.Value) error {
	// Handle nil pointers
	if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
		return nil // Skip nil pointer fields
	}

	// Dereference pointers
	if fieldValue.Kind() == reflect.Ptr {
		fieldValue = fieldValue.Elem()
	}

	str, err := e.valueToString(fieldValue)
	if err != nil {
		return err
	}

	// Only encode if there's actual content
	if str != "" {
		if e.debug {
			fmt.Printf("[DEBUG] Encoding character data: %s\n", str)
		}

		// Encode character data directly
		return encoder.EncodeCharacters(core.NewStringValueFromSlice([]rune(str)))
	}

	return nil
}

// valueToString converts a reflect.Value to its string representation
func (e *StructEncoder) valueToString(value reflect.Value) (string, error) {
	switch value.Kind() {
	case reflect.String:
		return value.String(), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(value.Int(), 10), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(value.Uint(), 10), nil

	case reflect.Float32:
		return strconv.FormatFloat(value.Float(), 'f', -1, 32), nil

	case reflect.Float64:
		return strconv.FormatFloat(value.Float(), 'f', -1, 64), nil

	case reflect.Bool:
		return strconv.FormatBool(value.Bool()), nil

	case reflect.Interface:
		// Handle interface types by converting the concrete value
		if value.IsNil() {
			return "", nil
		}
		return e.valueToString(value.Elem())

	case reflect.Slice:
		// Handle slice types
		if value.IsNil() {
			return "", nil
		}

		// Special handling for []byte - convert directly to string (common for base64 data)
		if value.Type().Elem().Kind() == reflect.Uint8 {
			bytes := value.Bytes()
			return string(bytes), nil
		}

		// For other slice types, convert elements to space-separated string
		var parts []string
		for i := 0; i < value.Len(); i++ {
			item := value.Index(i)
			str, err := e.valueToString(item)
			if err != nil {
				// If we can't convert slice elements, return empty string
				return "", nil
			}
			parts = append(parts, str)
		}
		return strings.Join(parts, " "), nil

	case reflect.Array:
		// Handle array types similarly to slices

		// Special handling for byte arrays - convert directly to string
		if value.Type().Elem().Kind() == reflect.Uint8 {
			// Convert array to slice and then to string
			slice := value.Slice(0, value.Len())
			bytes := slice.Bytes()
			return string(bytes), nil
		}

		// For other array types, convert elements to space-separated string
		var parts []string
		for i := 0; i < value.Len(); i++ {
			item := value.Index(i)
			str, err := e.valueToString(item)
			if err != nil {
				// If we can't convert array elements, return empty string
				return "", nil
			}
			parts = append(parts, str)
		}
		return strings.Join(parts, " "), nil

	default:
		return "", fmt.Errorf("unsupported value type for string conversion: %s", value.Kind())
	}
}

// parseXMLTag parses an XML struct tag and returns the field name, whether it's an attribute, whether it's chardata, and namespace
func (e *StructEncoder) parseXMLTag(xmlTag, defaultName string) (string, bool, bool, string) {
	if xmlTag == "" {
		return defaultName, false, false, ""
	}

	// Split by comma to separate name from modifiers
	parts := strings.Split(xmlTag, ",")
	namepart := strings.TrimSpace(parts[0])

	// Check for attribute and chardata modifiers
	isAttribute := false
	isCharData := false
	for i := 1; i < len(parts); i++ {
		modifier := strings.TrimSpace(parts[i])
		if modifier == "attr" {
			isAttribute = true
		} else if modifier == "chardata" {
			isCharData = true
		}
	}

	// Handle namespace in the name part
	var fieldName, namespace string
	if namepart == "" {
		fieldName = defaultName
	} else {
		if isAttribute {
			// For attributes, check for space-separated format first (for full URIs)
			nameFields := strings.Fields(namepart)
			if len(nameFields) >= 2 {
				// Format: "namespace localname" - common for full URIs
				namespace = strings.Join(nameFields[:len(nameFields)-1], " ")
				fieldName = nameFields[len(nameFields)-1]
			} else if colonIndex := strings.Index(namepart, ":"); colonIndex != -1 {
				// Check for simple prefix:localname format (but only if no spaces)
				namespace = namepart[:colonIndex]
				fieldName = namepart[colonIndex+1:]
			} else {
				// Just local name - no namespace for attribute
				fieldName = namepart
			}
		} else {
			// For elements, prefer space-separated format for namespace URI
			nameFields := strings.Fields(namepart)
			if len(nameFields) >= 2 {
				// Format: "namespace localname"
				namespace = strings.Join(nameFields[:len(nameFields)-1], " ")
				fieldName = nameFields[len(nameFields)-1]
			} else {
				// Check for prefix:localname format
				if colonIndex := strings.Index(namepart, ":"); colonIndex != -1 {
					namespace = namepart[:colonIndex]
					fieldName = namepart[colonIndex+1:]
				} else {
					// Just local name
					fieldName = namepart
				}
			}
		}
	}

	return fieldName, isAttribute, isCharData, namespace
}

// shouldSkipAttribute determines if an attribute should be skipped during encoding
func (e *StructEncoder) shouldSkipAttribute(attrName, attrValue string) bool {
	// Skip xsi:type attributes with anyType values (if enabled)
	if e.skipAnyType && strings.Contains(attrName, "type") {
		// Check for various forms of anyType
		if strings.Contains(attrValue, "anyType") ||
			strings.Contains(attrValue, ":anyType") ||
			attrValue == "anyType" {
			return true
		}
	}

	// Skip empty attribute values (optional)
	if strings.TrimSpace(attrValue) == "" {
		return true
	}

	return false
}

// resolveNamespace converts a namespace prefix to its full URI
func (e *StructEncoder) resolveNamespace(namespace string) string {
	if namespace == "" {
		return "" // No namespace
	}

	// Check if it's already a full URI (contains "://" or starts with "urn:")
	if strings.Contains(namespace, "://") || strings.HasPrefix(namespace, "urn:") {
		return namespace // Already a full URI
	}

	// Look up prefix in registered namespaces
	if uri, exists := e.namespaceDecls[namespace]; exists {
		return uri
	}

	// If not found, treat as literal namespace URI
	return namespace
}

// findPrefixForNamespace finds the prefix for a given namespace URI
func (e *StructEncoder) findPrefixForNamespace(namespaceURI string) string {
	for prefix, uri := range e.namespaceDecls {
		if uri == namespaceURI {
			return prefix
		}
	}
	return ""
}

// getPrefixForNamespace returns the prefix to use for a given namespace
func (e *StructEncoder) getPrefixForNamespace(namespace string) *string {
	if namespace == "" {
		return nil // No prefix for default namespace
	}

	// If it's already a URI, find the corresponding prefix
	if strings.Contains(namespace, "://") || strings.HasPrefix(namespace, "urn:") {
		for prefix, uri := range e.namespaceDecls {
			if uri == namespace && prefix != "" {
				return &prefix
			}
		}
		// No registered prefix found for this URI
		return nil
	}

	// If it's a prefix that exists in our declarations, return it
	if _, exists := e.namespaceDecls[namespace]; exists {
		if namespace == "" {
			return nil // Empty prefix means default namespace
		}
		prefixCopy := namespace
		return &prefixCopy
	}

	// If namespace looks like a prefix but not registered, still return it
	// This handles cases where struct tags use prefixes not explicitly declared
	if !strings.Contains(namespace, ":") && len(namespace) < 10 {
		prefixCopy := namespace
		return &prefixCopy
	}

	return nil
}

// addNamespaceDeclarations adds registered namespace declarations as attributes
func (e *StructEncoder) addNamespaceDeclarations() {
	for prefix, namespaceURI := range e.namespaceDecls {
		var attrName string
		if prefix == "" {
			// Default namespace declaration
			attrName = "xmlns"
		} else {
			// Prefixed namespace declaration
			attrName = "xmlns:" + prefix
		}

		if e.debug {
			fmt.Printf("[DEBUG] Adding namespace declaration: %s = %s\n", attrName, namespaceURI)
		}

		// Add namespace declaration as attribute
		e.exiAttributes.AddAttribute(nil, attrName, nil, namespaceURI)
	}
}

// isSimpleSliceType determines if a slice element type can be converted to string
func (e *StructEncoder) isSimpleSliceType(elemType reflect.Type) bool {
	switch elemType.Kind() {
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Bool:
		return true
	default:
		return false
	}
}

// getStructElementName gets the element name for a struct type
func (e *StructEncoder) getStructElementName(structType reflect.Type) string {
	// For now, just use the struct type name
	// In a more sophisticated implementation, you might look for xml tags on the struct itself
	return strings.ToLower(structType.Name())
}

// Stack implementation for tracking element context during encoding
type encoderStackFrame struct {
	elementName string
	namespace   string
}

type encoderElementStack struct {
	frames []*encoderStackFrame
}

func newEncoderElementStack() *encoderElementStack {
	return &encoderElementStack{
		frames: make([]*encoderStackFrame, 0),
	}
}

func (s *encoderElementStack) push(frame *encoderStackFrame) {
	s.frames = append(s.frames, frame)
}

func (s *encoderElementStack) pop() *encoderStackFrame {
	if len(s.frames) == 0 {
		return nil
	}

	frame := s.frames[len(s.frames)-1]
	s.frames = s.frames[:len(s.frames)-1]
	return frame
}

func (s *encoderElementStack) peek() *encoderStackFrame {
	if len(s.frames) == 0 {
		return nil
	}
	return s.frames[len(s.frames)-1]
}

func (s *encoderElementStack) isEmpty() bool {
	return len(s.frames) == 0
}
