package core

const (
	W3C_EXI_NS_URI          string = "http://www.w3.org/2009/exi"
	W3C_EXI_LN_Base64Binary string = "base64Binary"
	W3C_EXI_LN_HexBinary    string = "hexBinary"
	W3C_EXI_LN_Boolean      string = "boolean"
	W3C_EXI_LN_DateTime     string = "dateTime"
	W3C_EXI_LN_Time         string = "time"
	W3C_EXI_LN_Date         string = "date"
	W3C_EXI_LN_GYearMonth   string = "gYearMonth"
	W3C_EXI_LN_GYear        string = "gYear"
	W3C_EXI_LN_GMonthDay    string = "gMonthDay"
	W3C_EXI_LN_GDay         string = "gDay"
	W3C_EXI_LN_GMonth       string = "gMonth"
	W3C_EXI_LN_Decimal      string = "decimal"
	W3C_EXI_LN_Double       string = "double"
	W3C_EXI_LN_Integer      string = "integer"
	W3C_EXI_LN_String       string = "string"
	W3C_EXI_FeatureBodyOnly string = "http://www.w3.org/exi/features/exi-body-only"

	EmptyString string = ""

	XSISchemaLocation            string = "schemaLocation"
	XSINoNamespaceSchemaLocation string = "noNamespaceSchemaLocation"

	XML_NS_Prefix           string = "xml"
	XMLNullNS_URI           string = ""
	XMLDefaultNSPrefix      string = ""
	XML_NS_AttributeNS_URI  string = "http://www.w3.org/2000/xmlns/"
	XML_NS_Attribute        string = "xmlns"
	XML_NS_URI              string = "http://www.w3.org/XML/1998/namespace"
	XMLSchemaInstanceNS_URI string = "http://www.w3.org/2001/XMLSchema-instance"
	XMLSchemaNS_URI         string = "http://www.w3.org/2001/XMLSchema"

	XSIPrefix string = "xsi"
	XSIType   string = "type"
	XSINil    string = "nil"

	XSDListDelim     string = " "
	XSDListDelimChar rune   = ' '

	XSDAnyType      string = "anyType"
	XSDBooleanTrue  string = "true"
	XSDBoolean1     string = "1"
	XSDBooleanFalse string = "false"
	XSDBoolean0     string = "0"

	DecodedBooleanTrue  string = XSDBooleanTrue
	DecodedBooleanFalse string = XSDBooleanFalse

	NotFound int = -1

	/*
	 * Block & Channel settings Maximal Number of Values (per Block / Channel)
	 */
	MaxNumberOfValues int = 100
	DefaultBlockSize  int = 1000000

	/*
	 * StringTable settings
	 */
	DefaultValueMaxLength         int = -1
	DefaultValuePartitionCapacity int = -1

	/*
	 * Float & Double Values
	 */
	FloatInfinity      string = "INF"
	FloatMinusInfinity string = "-INF"
	FloatNotANumber    string = "NaN"

	/* -(2^14) == -16384 */
	FloatSpecialValues         int = -16384
	FloatMantissaInfinity      int = 1
	FloatMantissaMinusInfinity int = -1
	FloatMantissaNotANumber    int = 0

	/* -(2^14-1) == -16383 */
	FloatExponentMinRange int64 = -16383
	FloatExponentMaxRange int64 = 16383
	FloatMantissaMinRange int64 = -9223372036854775808
	FloatMantissaMaxRange int64 = 9223372036854775807
)

var (
	PrefixesEmpty   = []string{""}
	LocalNamesEmpty = []string{}
	PrefixesXML     = []string{"xml"}
	LocalNamesXML   = []string{"base", "id", "lang", "space"}
	PrefixesXSI     = []string{"xsi"}
	LocalNamesXSI   = []string{"nil", "type"}
	PrefixesXSD     = []string{}
	LocalNamesXSD   = []string{"ENTITIES", "ENTITY", "ID",
		"IDREF", "IDREFS", "NCName", "NMTOKEN", "NMTOKENS", "NOTATION",
		"Name", "QName", "anySimpleType", "anyType", "anyURI",
		"base64Binary", "boolean", "byte", "date", "dateTime", "decimal",
		"double", "duration", "float", "gDay", "gMonth", "gMonthDay",
		"gYear", "gYearMonth", "hexBinary", "int", "integer", "language",
		"long", "negativeInteger", "nonNegativeInteger",
		"nonPositiveInteger", "normalizedString", "positiveInteger",
		"short", "string", "time", "token", "unsignedByte", "unsignedInt",
		"unsignedLong", "unsignedShort"}

	XSDListDelimCharArray = []rune{' '}
	XSDBooleanTrueArray   = []rune(XSDBooleanTrue)
	XSDBoolean1Array      = []rune(XSDBoolean1)
	XSDBooleanFalseArray  = []rune(XSDBooleanFalse)
	XSDBoolean0Array      = []rune(XSDBoolean0)

	DecodedBooleanTrueArray  []rune = XSDBooleanTrueArray
	DecodedBooleanFalseArray []rune = XSDBooleanFalseArray

	FloatInfinityCharArray      []rune = []rune(FloatInfinity)
	FloatMinusInfinityCharArray []rune = []rune(FloatMinusInfinity)
	FloatNotANumberCharArray    []rune = []rune(FloatNotANumber)
)
