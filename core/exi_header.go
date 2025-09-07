package core

import (
	"errors"
	"fmt"
)

/*
	AbstractEXIHeader implementation
*/

const (
	EXIHeader_Header                    string = "header"
	EXIHeader_LessCommon                string = "lesscommon"
	EXIHeader_Uncommon                  string = "uncommon"
	EXIHeader_Alignment                 string = "alignment"
	EXIHeader_Byte                      string = "byte"
	EXIHeader_PreCompress               string = "pre-compress"
	EXIHeader_SelfContained             string = "selfContained"
	EXIHeader_ValueMaxLength            string = "valueMaxLength"
	EXIHeader_ValuePartitionCapacity    string = "valuePartitionCapacity"
	EXIHeader_DatatypeRepresentationMap string = "datatypeRepresentationMap"
	EXIHeader_Preserve                  string = "preserve"
	EXIHeader_Dtd                       string = "dtd"
	EXIHeader_Prefixes                  string = "prefixes"
	EXIHeader_LexicalValues             string = "lexicalValues"
	EXIHeader_Comments                  string = "comments"
	EXIHeader_Pis                       string = "pis"
	EXIHeader_BlockSize                 string = "blockSize"
	EXIHeader_Common                    string = "common"
	EXIHeader_Compression               string = "compression"
	EXIHeader_Fragment                  string = "fragment"
	EXIHeader_SchemaID                  string = "schemaId"
	EXIHeader_Strict                    string = "strict"
	EXIHeader_Profile                   string = "p"

	EXIHeader_NumberOfDistinguishingBits int = 2
	EXIHeader_DistinguishingBitsValue    int = 2
	EXIHeader_NumberOfFormatVersionBits  int = 4
	EXIHeader_FormatVersionContinueValue int = 15
)

type AbstractEXIHeader struct {
	headerFactory EXIFactory
}

func (h *AbstractEXIHeader) GetHeaderFactory() (EXIFactory, error) {
	if h.headerFactory == nil {
		h.headerFactory = NewDefaultEXIFactory()
		grammar, err := NewEXIOptionsHeaderGrammars()
		if err != nil {
			return nil, err
		}
		h.headerFactory.SetGrammars(grammar)
		h.headerFactory.SetFidelityOptions(NewStrictFidelityOptions())
	}

	return h.headerFactory, nil
}

type EXIOptionsHeaderGrammars struct {
	schemaID       *string
	grammarContext *GrammarContext
	document       *Document
	fragment       *Fragment
	sief           SchemaInformedGrammar
}

func NewEXIOptionsHeaderGrammars() (*EXIOptionsHeaderGrammars, error) {
	/* BEGIN GrammarContext ----- */
	ns0 := ""
	grammarQNames0 := []*QNameContext{}
	grammarPrefixes0 := []string{""}
	guc0 := NewGrammarUriContext(0, ns0, grammarQNames0, grammarPrefixes0)

	ns1 := "http://www.w3.org/XML/1998/namespace"
	qnc0 := NewQNameContext(1, 0, QName{Space: ns1, Local: "base"})
	qnc1 := NewQNameContext(1, 1, QName{Space: ns1, Local: "id"})
	qnc2 := NewQNameContext(1, 2, QName{Space: ns1, Local: "lang"})
	qnc3 := NewQNameContext(1, 3, QName{Space: ns1, Local: "space"})
	grammarQNames1 := []*QNameContext{qnc0, qnc1, qnc2, qnc3}
	grammarPrefixes1 := []string{"xml"}
	guc1 := NewGrammarUriContext(1, ns1, grammarQNames1, grammarPrefixes1)

	ns2 := "http://www.w3.org/2001/XMLSchema-instance"
	qnc4 := NewQNameContext(2, 0, QName{Space: ns2, Local: "nil"})
	qnc5 := NewQNameContext(2, 1, QName{Space: ns2, Local: "type"})
	grammarQNames2 := []*QNameContext{qnc4, qnc5}
	grammarPrefixes2 := []string{"xsi"}
	guc2 := NewGrammarUriContext(2, ns2, grammarQNames2, grammarPrefixes2)

	ns3 := "http://www.w3.org/2001/XMLSchema"
	qnc6 := NewQNameContext(3, 0, QName{Space: ns3, Local: "ENTITIES"})
	qnc7 := NewQNameContext(3, 1, QName{Space: ns3, Local: "ENTITY"})
	qnc8 := NewQNameContext(3, 2, QName{Space: ns3, Local: "ID"})
	qnc9 := NewQNameContext(3, 3, QName{Space: ns3, Local: "IDREF"})
	qnc10 := NewQNameContext(3, 4, QName{Space: ns3, Local: "IDREFS"})
	qnc11 := NewQNameContext(3, 5, QName{Space: ns3, Local: "NCName"})
	qnc12 := NewQNameContext(3, 6, QName{Space: ns3, Local: "NMTOKEN"})
	qnc13 := NewQNameContext(3, 7, QName{Space: ns3, Local: "NMTOKENS"})
	qnc14 := NewQNameContext(3, 8, QName{Space: ns3, Local: "NOTATION"})
	qnc15 := NewQNameContext(3, 9, QName{Space: ns3, Local: "Name"})
	qnc16 := NewQNameContext(3, 10, QName{Space: ns3, Local: "QName"})
	qnc17 := NewQNameContext(3, 11, QName{Space: ns3, Local: "anySimpleType"})
	qnc18 := NewQNameContext(3, 12, QName{Space: ns3, Local: "anyType"})
	qnc19 := NewQNameContext(3, 13, QName{Space: ns3, Local: "anyURI"})
	qnc20 := NewQNameContext(3, 14, QName{Space: ns3, Local: "base64Binary"})
	qnc21 := NewQNameContext(3, 15, QName{Space: ns3, Local: "boolean"})
	qnc22 := NewQNameContext(3, 16, QName{Space: ns3, Local: "byte"})
	qnc23 := NewQNameContext(3, 17, QName{Space: ns3, Local: "date"})
	qnc24 := NewQNameContext(3, 18, QName{Space: ns3, Local: "dateTime"})
	qnc25 := NewQNameContext(3, 19, QName{Space: ns3, Local: "decimal"})
	qnc26 := NewQNameContext(3, 20, QName{Space: ns3, Local: "double"})
	qnc27 := NewQNameContext(3, 21, QName{Space: ns3, Local: "duration"})
	qnc28 := NewQNameContext(3, 22, QName{Space: ns3, Local: "float"})
	qnc29 := NewQNameContext(3, 23, QName{Space: ns3, Local: "gDay"})
	qnc30 := NewQNameContext(3, 24, QName{Space: ns3, Local: "gMonth"})
	qnc31 := NewQNameContext(3, 25, QName{Space: ns3, Local: "gMonthDay"})
	qnc32 := NewQNameContext(3, 26, QName{Space: ns3, Local: "gYear"})
	qnc33 := NewQNameContext(3, 27, QName{Space: ns3, Local: "gYearMonth"})
	qnc34 := NewQNameContext(3, 28, QName{Space: ns3, Local: "hexBinary"})
	qnc35 := NewQNameContext(3, 29, QName{Space: ns3, Local: "int"})
	qnc36 := NewQNameContext(3, 30, QName{Space: ns3, Local: "integer"})
	qnc37 := NewQNameContext(3, 31, QName{Space: ns3, Local: "language"})
	qnc38 := NewQNameContext(3, 32, QName{Space: ns3, Local: "long"})
	qnc39 := NewQNameContext(3, 33, QName{Space: ns3, Local: "negativeInteger"})
	qnc40 := NewQNameContext(3, 34, QName{Space: ns3, Local: "nonNegativeInteger"})
	qnc41 := NewQNameContext(3, 35, QName{Space: ns3, Local: "nonPositiveInteger"})
	qnc42 := NewQNameContext(3, 36, QName{Space: ns3, Local: "normalizedString"})
	qnc43 := NewQNameContext(3, 37, QName{Space: ns3, Local: "positiveInteger"})
	qnc44 := NewQNameContext(3, 38, QName{Space: ns3, Local: "short"})
	qnc45 := NewQNameContext(3, 39, QName{Space: ns3, Local: "string"})
	qnc46 := NewQNameContext(3, 40, QName{Space: ns3, Local: "time"})
	qnc47 := NewQNameContext(3, 41, QName{Space: ns3, Local: "token"})
	qnc48 := NewQNameContext(3, 42, QName{Space: ns3, Local: "unsignedByte"})
	qnc49 := NewQNameContext(3, 43, QName{Space: ns3, Local: "unsignedInt"})
	qnc50 := NewQNameContext(3, 44, QName{Space: ns3, Local: "unsignedLong"})
	qnc51 := NewQNameContext(3, 45, QName{Space: ns3, Local: "unsignedShort"})
	grammarQNames3 := []*QNameContext{qnc6, qnc7, qnc8, qnc9, qnc10, qnc11, qnc12, qnc13, qnc14, qnc15, qnc16, qnc17, qnc18, qnc19, qnc20, qnc21, qnc22, qnc23, qnc24, qnc25, qnc26, qnc27, qnc28, qnc29, qnc30, qnc31, qnc32, qnc33, qnc34, qnc35, qnc36, qnc37, qnc38, qnc39, qnc40, qnc41, qnc42, qnc43, qnc44, qnc45, qnc46, qnc47, qnc48, qnc49, qnc50, qnc51}
	grammarPrefixes3 := []string{}
	guc3 := NewGrammarUriContext(3, ns3, grammarQNames3, grammarPrefixes3)

	ns4 := "http://www.w3.org/2009/exi"
	qnc52 := NewQNameContext(4, 0, QName{Space: ns4, Local: "alignment"})
	qnc53 := NewQNameContext(4, 1, QName{Space: ns4, Local: "base64Binary"})
	qnc54 := NewQNameContext(4, 2, QName{Space: ns4, Local: "blockSize"})
	qnc55 := NewQNameContext(4, 3, QName{Space: ns4, Local: "boolean"})
	qnc56 := NewQNameContext(4, 4, QName{Space: ns4, Local: "byte"})
	qnc57 := NewQNameContext(4, 5, QName{Space: ns4, Local: "comments"})
	qnc58 := NewQNameContext(4, 6, QName{Space: ns4, Local: "common"})
	qnc59 := NewQNameContext(4, 7, QName{Space: ns4, Local: "compression"})
	qnc60 := NewQNameContext(4, 8, QName{Space: ns4, Local: "datatypeRepresentationMap"})
	qnc61 := NewQNameContext(4, 9, QName{Space: ns4, Local: "date"})
	qnc62 := NewQNameContext(4, 10, QName{Space: ns4, Local: "dateTime"})
	qnc63 := NewQNameContext(4, 11, QName{Space: ns4, Local: "decimal"})
	qnc64 := NewQNameContext(4, 12, QName{Space: ns4, Local: "double"})
	qnc65 := NewQNameContext(4, 13, QName{Space: ns4, Local: "dtd"})
	qnc66 := NewQNameContext(4, 14, QName{Space: ns4, Local: "fragment"})
	qnc67 := NewQNameContext(4, 15, QName{Space: ns4, Local: "gDay"})
	qnc68 := NewQNameContext(4, 16, QName{Space: ns4, Local: "gMonth"})
	qnc69 := NewQNameContext(4, 17, QName{Space: ns4, Local: "gMonthDay"})
	qnc70 := NewQNameContext(4, 18, QName{Space: ns4, Local: "gYear"})
	qnc71 := NewQNameContext(4, 19, QName{Space: ns4, Local: "gYearMonth"})
	qnc72 := NewQNameContext(4, 20, QName{Space: ns4, Local: "header"})
	qnc73 := NewQNameContext(4, 21, QName{Space: ns4, Local: "hexBinary"})
	qnc74 := NewQNameContext(4, 22, QName{Space: ns4, Local: "ieeeBinary32"})
	qnc75 := NewQNameContext(4, 23, QName{Space: ns4, Local: "ieeeBinary64"})
	qnc76 := NewQNameContext(4, 24, QName{Space: ns4, Local: "integer"})
	qnc77 := NewQNameContext(4, 25, QName{Space: ns4, Local: "lesscommon"})
	qnc78 := NewQNameContext(4, 26, QName{Space: ns4, Local: "lexicalValues"})
	qnc79 := NewQNameContext(4, 27, QName{Space: ns4, Local: "pis"})
	qnc80 := NewQNameContext(4, 28, QName{Space: ns4, Local: "pre-compress"})
	qnc81 := NewQNameContext(4, 29, QName{Space: ns4, Local: "prefixes"})
	qnc82 := NewQNameContext(4, 30, QName{Space: ns4, Local: "preserve"})
	qnc83 := NewQNameContext(4, 31, QName{Space: ns4, Local: "schemaId"})
	qnc84 := NewQNameContext(4, 32, QName{Space: ns4, Local: "selfContained"})
	qnc85 := NewQNameContext(4, 33, QName{Space: ns4, Local: "strict"})
	qnc86 := NewQNameContext(4, 34, QName{Space: ns4, Local: "string"})
	qnc87 := NewQNameContext(4, 35, QName{Space: ns4, Local: "time"})
	qnc88 := NewQNameContext(4, 36, QName{Space: ns4, Local: "uncommon"})
	qnc89 := NewQNameContext(4, 37, QName{Space: ns4, Local: "valueMaxLength"})
	qnc90 := NewQNameContext(4, 38, QName{Space: ns4, Local: "valuePartitionCapacity"})
	grammarQNames4 := []*QNameContext{qnc52, qnc53, qnc54, qnc55, qnc56, qnc57, qnc58, qnc59, qnc60, qnc61, qnc62, qnc63, qnc64, qnc65, qnc66, qnc67, qnc68, qnc69, qnc70, qnc71, qnc72, qnc73, qnc74, qnc75, qnc76, qnc77, qnc78, qnc79, qnc80, qnc81, qnc82, qnc83, qnc84, qnc85, qnc86, qnc87, qnc88, qnc89, qnc90}
	grammarPrefixes4 := []string{}
	guc4 := NewGrammarUriContext(4, ns4, grammarQNames4, grammarPrefixes4)

	grammarUriContexts := []*GrammarUriContext{guc0, guc1, guc2, guc3, guc4}
	gc := NewGrammarContext(grammarUriContexts, 91)
	/* END GrammarContext ----- */

	/* BEGIN Grammars ----- */
	g0 := NewDocument()
	g1 := NewSchemaInformedDocContent()
	g2 := NewDocEnd()
	g3 := NewFragment()
	g4 := NewSchemaInformedFragmentContent()
	g35 := NewSchemaInformedElement()
	g36 := NewSchemaInformedElement()
	g37 := NewSchemaInformedElement()
	g38 := NewSchemaInformedElement()
	g39 := NewSchemaInformedElement()
	g40 := NewSchemaInformedElement()
	g41 := NewSchemaInformedElement()
	g42 := NewSchemaInformedElement()
	g43 := NewSchemaInformedElement()
	g44 := NewSchemaInformedElement()
	g45 := NewSchemaInformedElement()
	g46 := NewSchemaInformedElement()
	g47 := NewSchemaInformedElement()
	g48 := NewSchemaInformedElement()
	g49 := NewSchemaInformedElement()
	g50 := NewSchemaInformedElement()
	g51 := NewSchemaInformedElement()
	g52 := NewSchemaInformedElement()
	g53 := NewSchemaInformedElement()
	g54 := NewSchemaInformedElement()
	g55 := NewSchemaInformedElement()
	g56 := NewSchemaInformedElement()
	g57 := NewSchemaInformedElement()
	g58 := NewSchemaInformedElement()
	g59 := NewSchemaInformedElement()
	g60 := NewSchemaInformedElement()
	g61 := NewSchemaInformedElement()
	g62 := NewSchemaInformedElement()
	g63 := NewSchemaInformedElement()
	g64 := NewSchemaInformedElement()
	g65 := NewSchemaInformedElement()
	g66 := NewSchemaInformedElement()
	g67 := NewSchemaInformedElement()
	g68 := NewSchemaInformedElement()
	g69 := NewSchemaInformedElement()
	g70 := NewSchemaInformedElement()
	g71 := NewSchemaInformedElement()
	g72 := NewSchemaInformedElement()
	g73 := NewSchemaInformedElement()
	g74 := NewSchemaInformedElement()
	g75 := NewSchemaInformedElement()
	g76 := NewSchemaInformedElement()
	g77 := NewSchemaInformedElement()
	g78 := NewSchemaInformedElement()
	g79 := NewSchemaInformedElement()
	/* END Grammars ----- */

	/* BEGIN Grammars with element content ----- */
	g5 := NewSchemaInformedFirstStartTagWithEC2(g60)
	g6 := NewSchemaInformedFirstStartTagWithEC2(g53)
	g7 := NewSchemaInformedFirstStartTagWithEC2(g45)
	g8 := NewSchemaInformedFirstStartTagWithEC2(g37)
	g9 := NewSchemaInformedFirstStartTagWithEC2(g36)
	g10 := NewSchemaInformedFirstStartTagWithEC2(g40)
	g11 := NewSchemaInformedFirstStartTagWithEC2(g44)
	g12 := NewSchemaInformedFirstStartTagWithEC2(g51)
	g13 := NewSchemaInformedFirstStartTagWithEC2(g58)
	g14 := NewSchemaInformedFirstStartTagWithEC2(g57)
	g15 := NewSchemaInformedFirstStartTagWithEC2(g61)
	g16 := NewSchemaInformedFirstStartTagWithEC2(g62)
	g17 := NewSchemaInformedFirstStartTagWithEC2(g57)
	g18 := NewSchemaInformedFirstStartTagWithEC2(g63)
	g19 := NewSchemaInformedFirstStartTagWithEC2(g64)
	g20 := NewSchemaInformedFirstStartTagWithEC2(g65)
	g21 := NewSchemaInformedFirstStartTagWithEC2(g66)
	g22 := NewSchemaInformedFirstStartTagWithEC2(g67)
	g23 := NewSchemaInformedFirstStartTagWithEC2(g68)
	g24 := NewSchemaInformedFirstStartTagWithEC2(g69)
	g25 := NewSchemaInformedFirstStartTagWithEC2(g70)
	g26 := NewSchemaInformedFirstStartTagWithEC2(g71)
	g27 := NewSchemaInformedFirstStartTagWithEC2(g72)
	g28 := NewSchemaInformedFirstStartTagWithEC2(g73)
	g29 := NewSchemaInformedFirstStartTagWithEC2(g74)
	g30 := NewSchemaInformedFirstStartTagWithEC2(g75)
	g31 := NewSchemaInformedFirstStartTagWithEC2(g76)
	g32 := NewSchemaInformedFirstStartTagWithEC2(g77)
	g33 := NewSchemaInformedFirstStartTagWithEC2(g78)
	g34 := NewSchemaInformedFirstStartTagWithEC2(g79)

	globalSE72 := NewStartElementWithGrammar(qnc72, g5)

	/* BEGIN GlobalElements ----- */
	qnc72.SetGlobalStartElement(globalSE72)
	/* END GlobalElements ----- */

	/* BEGIN GlobalAttributes ----- */
	/* END GlobalAttributes ----- */

	/* BEGIN TypeGrammar ----- */
	qnc6.SetTypeGrammar(g16)
	qnc7.SetTypeGrammar(g17)
	qnc8.SetTypeGrammar(g17)
	qnc9.SetTypeGrammar(g17)
	qnc10.SetTypeGrammar(g16)
	qnc11.SetTypeGrammar(g17)
	qnc12.SetTypeGrammar(g17)
	qnc13.SetTypeGrammar(g16)
	qnc14.SetTypeGrammar(g17)
	qnc15.SetTypeGrammar(g17)
	qnc16.SetTypeGrammar(g17)
	qnc17.SetTypeGrammar(g17)
	qnc18.SetTypeGrammar(g18)
	qnc19.SetTypeGrammar(g17)
	qnc20.SetTypeGrammar(g19)
	qnc21.SetTypeGrammar(g20)
	qnc22.SetTypeGrammar(g21)
	qnc23.SetTypeGrammar(g22)
	qnc24.SetTypeGrammar(g23)
	qnc25.SetTypeGrammar(g24)
	qnc26.SetTypeGrammar(g25)
	qnc27.SetTypeGrammar(g17)
	qnc28.SetTypeGrammar(g25)
	qnc29.SetTypeGrammar(g26)
	qnc30.SetTypeGrammar(g27)
	qnc31.SetTypeGrammar(g28)
	qnc32.SetTypeGrammar(g29)
	qnc33.SetTypeGrammar(g30)
	qnc34.SetTypeGrammar(g31)
	qnc35.SetTypeGrammar(g32)
	qnc36.SetTypeGrammar(g32)
	qnc37.SetTypeGrammar(g17)
	qnc38.SetTypeGrammar(g32)
	qnc39.SetTypeGrammar(g32)
	qnc40.SetTypeGrammar(g10)
	qnc41.SetTypeGrammar(g32)
	qnc42.SetTypeGrammar(g17)
	qnc43.SetTypeGrammar(g10)
	qnc44.SetTypeGrammar(g32)
	qnc45.SetTypeGrammar(g17)
	qnc46.SetTypeGrammar(g33)
	qnc47.SetTypeGrammar(g17)
	qnc48.SetTypeGrammar(g34)
	qnc49.SetTypeGrammar(g10)
	qnc50.SetTypeGrammar(g10)
	qnc51.SetTypeGrammar(g10)
	qnc53.SetTypeGrammar(g19)
	qnc55.SetTypeGrammar(g20)
	qnc61.SetTypeGrammar(g22)
	qnc62.SetTypeGrammar(g23)
	qnc63.SetTypeGrammar(g24)
	qnc64.SetTypeGrammar(g25)
	qnc67.SetTypeGrammar(g26)
	qnc68.SetTypeGrammar(g27)
	qnc69.SetTypeGrammar(g28)
	qnc70.SetTypeGrammar(g29)
	qnc71.SetTypeGrammar(g30)
	qnc73.SetTypeGrammar(g31)
	qnc74.SetTypeGrammar(g25)
	qnc75.SetTypeGrammar(g25)
	qnc76.SetTypeGrammar(g32)
	qnc86.SetTypeGrammar(g17)
	qnc87.SetTypeGrammar(g33)
	/* END TypeGrammar ----- */

	/* BEGIN Grammar Events ----- */
	g0.AddProduction(NewStartDocument(), g1)
	g1.AddProduction(globalSE72, g2)
	g1.AddProduction(NewStartElementGeneric(), g2)
	g2.AddProduction(NewEndDocument(), g35)
	g3.AddProduction(NewStartDocument(), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc52, g8), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc54, g10), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc56, g9), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc57, g9), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc58, g13), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc59, g9), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc60, g11), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc65, g9), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc66, g9), g4)
	g4.AddProduction(globalSE72, g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc77, g6), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc78, g9), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc79, g9), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc80, g9), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc81, g9), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc82, g12), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc83, g14), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc84, g9), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc85, g9), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc88, g7), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc89, g10), g4)
	g4.AddProduction(NewStartElementWithGrammar(qnc90, g10), g4)
	g4.AddProduction(NewStartElementGeneric(), g4)
	g4.AddProduction(NewEndDocument(), g35)
	g5.AddProduction(NewStartElementWithGrammar(qnc77, g6), g54)
	g5.AddProduction(NewStartElementWithGrammar(qnc58, g13), g59)
	g5.AddProduction(NewStartElementWithGrammar(qnc85, g9), g36)
	g5.AddProduction(NewEndElement(), g35)
	g6.AddProduction(NewStartElementWithGrammar(qnc88, g7), g46)
	g6.AddProduction(NewStartElementWithGrammar(qnc82, g12), g52)
	g6.AddProduction(NewStartElementWithGrammar(qnc54, g10), g36)
	g6.AddProduction(NewEndElement(), g35)
	g7.AddProduction(NewStartElementWithGrammar(qnc52, g8), g38)
	g7.AddProduction(NewStartElementWithGrammar(qnc84, g9), g39)
	g7.AddProduction(NewStartElementWithGrammar(qnc89, g10), g41)
	g7.AddProduction(NewStartElementWithGrammar(qnc90, g10), g42)
	g7.AddProduction(NewStartElementWithGrammar(qnc60, g11), g42)
	g7.AddProduction(NewStartElementGeneric(), g45)
	g7.AddProduction(NewEndElement(), g35)
	g8.AddProduction(NewStartElementWithGrammar(qnc56, g9), g36)
	g8.AddProduction(NewStartElementWithGrammar(qnc80, g9), g36)
	g9.AddProduction(NewEndElement(), g35)
	g10.AddProduction(NewCharacters(NewUnsignedIntegerDatatype(qnc49)), g36)
	g11.AddProduction(NewStartElementGeneric(), g43)
	g12.AddProduction(NewStartElementWithGrammar(qnc65, g9), g47)
	g12.AddProduction(NewStartElementWithGrammar(qnc81, g9), g48)
	g12.AddProduction(NewStartElementWithGrammar(qnc78, g9), g49)
	g12.AddProduction(NewStartElementWithGrammar(qnc57, g9), g50)
	g12.AddProduction(NewStartElementWithGrammar(qnc79, g9), g36)
	g12.AddProduction(NewEndElement(), g35)
	g13.AddProduction(NewStartElementWithGrammar(qnc59, g9), g55)
	g13.AddProduction(NewStartElementWithGrammar(qnc66, g9), g56)
	g13.AddProduction(NewStartElementWithGrammar(qnc83, g14), g36)
	g13.AddProduction(NewEndElement(), g35)
	g14.AddProduction(NewCharacters(NewStringDatatype(qnc45)), g36)
	g15.AddProduction(NewAttributeGeneric(), g15)
	g15.AddProduction(NewStartElementWithGrammar(qnc52, g8), g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc54, g10), g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc56, g9), g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc57, g9), g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc58, g13), g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc59, g9), g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc60, g11), g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc65, g9), g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc66, g9), g61)
	g15.AddProduction(globalSE72, g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc77, g6), g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc78, g9), g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc79, g9), g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc80, g9), g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc81, g9), g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc82, g12), g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc83, g14), g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc84, g9), g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc85, g9), g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc88, g7), g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc89, g10), g61)
	g15.AddProduction(NewStartElementWithGrammar(qnc90, g10), g61)
	g15.AddProduction(NewStartElementGeneric(), g61)
	g15.AddProduction(NewEndElement(), g35)
	g15.AddProduction(NewCharactersGeneric(), g61)
	g16.AddProduction(NewCharacters(NewListDatatype(NewStringDatatype(qnc7), qnc6)), g36)
	g17.AddProduction(NewCharacters(NewStringDatatype(qnc7)), g36)
	g18.AddProduction(NewAttributeGeneric(), g18)
	g18.AddProduction(NewStartElementGeneric(), g63)
	g18.AddProduction(NewEndElement(), g35)
	g18.AddProduction(NewCharactersGeneric(), g63)
	g19.AddProduction(NewCharacters(NewBinaryBase64Datatype(qnc20)), g36)
	g20.AddProduction(NewCharacters(NewBooleanDatatype(qnc21)), g36)
	g21.AddProduction(NewCharacters(NewNBitUnsignedIntegerDatatype(NewIntegerValue32(-128), NewIntegerValue32(127), qnc22)), g36)
	g22.AddProduction(NewCharacters(NewDatetimeDatatype(DateTimeDate, qnc23)), g36)
	g23.AddProduction(NewCharacters(NewDatetimeDatatype(DateTimeDateTime, qnc24)), g36)
	g24.AddProduction(NewCharacters(NewDecimalDatatype(qnc25)), g36)
	g25.AddProduction(NewCharacters(NewFloatDatatype(qnc26)), g36)
	g26.AddProduction(NewCharacters(NewDatetimeDatatype(DateTimeGDay, qnc29)), g36)
	g27.AddProduction(NewCharacters(NewDatetimeDatatype(DateTimeGMonth, qnc30)), g36)
	g28.AddProduction(NewCharacters(NewDatetimeDatatype(DateTimeGMonthDay, qnc31)), g36)
	g29.AddProduction(NewCharacters(NewDatetimeDatatype(DateTimeGYear, qnc32)), g36)
	g30.AddProduction(NewCharacters(NewDatetimeDatatype(DateTimeGYearMonth, qnc33)), g36)
	g31.AddProduction(NewCharacters(NewBinaryHexDatatype(qnc34)), g36)
	g32.AddProduction(NewCharacters(NewIntegerDatatype(qnc35)), g36)
	g33.AddProduction(NewCharacters(NewDatetimeDatatype(DateTimeTime, qnc46)), g36)
	g34.AddProduction(NewCharacters(NewNBitUnsignedIntegerDatatype(NewIntegerValue32(0), NewIntegerValue32(255), qnc48)), g36)
	g36.AddProduction(NewEndElement(), g35)
	g37.AddProduction(NewStartElementWithGrammar(qnc56, g9), g36)
	g37.AddProduction(NewStartElementWithGrammar(qnc80, g9), g36)
	g38.AddProduction(NewStartElementWithGrammar(qnc84, g9), g39)
	g38.AddProduction(NewStartElementWithGrammar(qnc89, g10), g41)
	g38.AddProduction(NewStartElementWithGrammar(qnc90, g10), g42)
	g38.AddProduction(NewStartElementWithGrammar(qnc60, g11), g42)
	g38.AddProduction(NewEndElement(), g35)
	g39.AddProduction(NewStartElementWithGrammar(qnc89, g10), g41)
	g39.AddProduction(NewStartElementWithGrammar(qnc90, g10), g42)
	g39.AddProduction(NewStartElementWithGrammar(qnc60, g11), g42)
	g39.AddProduction(NewEndElement(), g35)
	g40.AddProduction(NewCharacters(NewUnsignedIntegerDatatype(qnc49)), g36)
	g41.AddProduction(NewStartElementWithGrammar(qnc90, g10), g42)
	g41.AddProduction(NewStartElementWithGrammar(qnc60, g11), g42)
	g41.AddProduction(NewEndElement(), g35)
	g42.AddProduction(NewStartElementWithGrammar(qnc60, g11), g42)
	g42.AddProduction(NewEndElement(), g35)
	g43.AddProduction(NewStartElementGeneric(), g36)
	g44.AddProduction(NewStartElementGeneric(), g43)
	g45.AddProduction(NewStartElementWithGrammar(qnc52, g8), g38)
	g45.AddProduction(NewStartElementWithGrammar(qnc84, g9), g39)
	g45.AddProduction(NewStartElementWithGrammar(qnc89, g10), g41)
	g45.AddProduction(NewStartElementWithGrammar(qnc90, g10), g42)
	g45.AddProduction(NewStartElementWithGrammar(qnc60, g11), g42)
	g45.AddProduction(NewStartElementGeneric(), g45)
	g45.AddProduction(NewEndElement(), g35)
	g46.AddProduction(NewStartElementWithGrammar(qnc82, g12), g52)
	g46.AddProduction(NewStartElementWithGrammar(qnc54, g10), g36)
	g46.AddProduction(NewEndElement(), g35)
	g47.AddProduction(NewStartElementWithGrammar(qnc81, g9), g48)
	g47.AddProduction(NewStartElementWithGrammar(qnc78, g9), g49)
	g47.AddProduction(NewStartElementWithGrammar(qnc57, g9), g50)
	g47.AddProduction(NewStartElementWithGrammar(qnc79, g9), g36)
	g47.AddProduction(NewEndElement(), g35)
	g48.AddProduction(NewStartElementWithGrammar(qnc78, g9), g49)
	g48.AddProduction(NewStartElementWithGrammar(qnc57, g9), g50)
	g48.AddProduction(NewStartElementWithGrammar(qnc79, g9), g36)
	g48.AddProduction(NewEndElement(), g35)
	g49.AddProduction(NewStartElementWithGrammar(qnc57, g9), g50)
	g49.AddProduction(NewStartElementWithGrammar(qnc79, g9), g36)
	g49.AddProduction(NewEndElement(), g35)
	g50.AddProduction(NewStartElementWithGrammar(qnc79, g9), g36)
	g50.AddProduction(NewEndElement(), g35)
	g51.AddProduction(NewStartElementWithGrammar(qnc65, g9), g47)
	g51.AddProduction(NewStartElementWithGrammar(qnc81, g9), g48)
	g51.AddProduction(NewStartElementWithGrammar(qnc78, g9), g49)
	g51.AddProduction(NewStartElementWithGrammar(qnc57, g9), g50)
	g51.AddProduction(NewStartElementWithGrammar(qnc79, g9), g36)
	g51.AddProduction(NewEndElement(), g35)
	g52.AddProduction(NewStartElementWithGrammar(qnc54, g10), g36)
	g52.AddProduction(NewEndElement(), g35)
	g53.AddProduction(NewStartElementWithGrammar(qnc88, g7), g46)
	g53.AddProduction(NewStartElementWithGrammar(qnc82, g12), g52)
	g53.AddProduction(NewStartElementWithGrammar(qnc54, g10), g36)
	g53.AddProduction(NewEndElement(), g35)
	g54.AddProduction(NewStartElementWithGrammar(qnc58, g13), g59)
	g54.AddProduction(NewStartElementWithGrammar(qnc85, g9), g36)
	g54.AddProduction(NewEndElement(), g35)
	g55.AddProduction(NewStartElementWithGrammar(qnc66, g9), g56)
	g55.AddProduction(NewStartElementWithGrammar(qnc83, g14), g36)
	g55.AddProduction(NewEndElement(), g35)
	g56.AddProduction(NewStartElementWithGrammar(qnc83, g14), g36)
	g56.AddProduction(NewEndElement(), g35)
	g57.AddProduction(NewCharacters(NewStringDatatype(qnc45)), g36)
	g58.AddProduction(NewStartElementWithGrammar(qnc59, g9), g55)
	g58.AddProduction(NewStartElementWithGrammar(qnc66, g9), g56)
	g58.AddProduction(NewStartElementWithGrammar(qnc83, g14), g36)
	g58.AddProduction(NewEndElement(), g35)
	g59.AddProduction(NewStartElementWithGrammar(qnc85, g9), g36)
	g59.AddProduction(NewEndElement(), g35)
	g60.AddProduction(NewStartElementWithGrammar(qnc77, g6), g54)
	g60.AddProduction(NewStartElementWithGrammar(qnc58, g13), g59)
	g60.AddProduction(NewStartElementWithGrammar(qnc85, g9), g36)
	g60.AddProduction(NewEndElement(), g35)
	g61.AddProduction(NewStartElementWithGrammar(qnc52, g8), g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc54, g10), g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc56, g9), g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc57, g9), g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc58, g13), g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc59, g9), g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc60, g11), g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc65, g9), g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc66, g9), g61)
	g61.AddProduction(globalSE72, g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc77, g6), g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc78, g9), g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc79, g9), g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc80, g9), g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc81, g9), g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc82, g12), g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc83, g14), g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc84, g9), g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc85, g9), g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc88, g7), g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc89, g10), g61)
	g61.AddProduction(NewStartElementWithGrammar(qnc90, g10), g61)
	g61.AddProduction(NewStartElementGeneric(), g61)
	g61.AddProduction(NewEndElement(), g35)
	g61.AddProduction(NewCharactersGeneric(), g61)
	g62.AddProduction(NewCharacters(NewListDatatype(NewStringDatatype(qnc7), qnc6)), g36)
	g63.AddProduction(NewStartElementGeneric(), g63)
	g63.AddProduction(NewEndElement(), g35)
	g63.AddProduction(NewCharactersGeneric(), g63)
	g64.AddProduction(NewCharacters(NewBinaryBase64Datatype(qnc20)), g36)
	g65.AddProduction(NewCharacters(NewBooleanDatatype(qnc21)), g36)
	g66.AddProduction(NewCharacters(NewNBitUnsignedIntegerDatatype(NewIntegerValue32(-128), NewIntegerValue32(127), qnc22)), g36)
	g67.AddProduction(NewCharacters(NewDatetimeDatatype(DateTimeDate, qnc23)), g36)
	g68.AddProduction(NewCharacters(NewDatetimeDatatype(DateTimeDateTime, qnc24)), g36)
	g69.AddProduction(NewCharacters(NewDecimalDatatype(qnc25)), g36)
	g70.AddProduction(NewCharacters(NewFloatDatatype(qnc26)), g36)
	g71.AddProduction(NewCharacters(NewDatetimeDatatype(DateTimeGDay, qnc29)), g36)
	g72.AddProduction(NewCharacters(NewDatetimeDatatype(DateTimeGMonth, qnc30)), g36)
	g73.AddProduction(NewCharacters(NewDatetimeDatatype(DateTimeGMonthDay, qnc31)), g36)
	g74.AddProduction(NewCharacters(NewDatetimeDatatype(DateTimeGYear, qnc32)), g36)
	g75.AddProduction(NewCharacters(NewDatetimeDatatype(DateTimeGYearMonth, qnc33)), g36)
	g76.AddProduction(NewCharacters(NewBinaryHexDatatype(qnc34)), g36)
	g77.AddProduction(NewCharacters(NewIntegerDatatype(qnc35)), g36)
	g78.AddProduction(NewCharacters(NewDatetimeDatatype(DateTimeTime, qnc46)), g36)
	g79.AddProduction(NewCharacters(NewNBitUnsignedIntegerDatatype(NewIntegerValue32(0), NewIntegerValue32(255), qnc48)), g36)
	/* END Grammar Events ----- */

	/* BEGIN FirstStartGrammar ----- */
	g5.SetElementContentGrammar(g60)
	g6.SetElementContentGrammar(g53)
	g7.SetElementContentGrammar(g45)
	g8.SetElementContentGrammar(g37)
	g9.SetElementContentGrammar(g36)
	g10.SetElementContentGrammar(g40)
	g11.SetElementContentGrammar(g44)
	g12.SetElementContentGrammar(g51)
	g13.SetElementContentGrammar(g58)
	g14.SetElementContentGrammar(g57)
	g14.SetNillable(true)
	g15.SetElementContentGrammar(g61)
	g15.SetTypeCastable(true)
	g15.SetNillable(true)
	g16.SetElementContentGrammar(g62)
	g17.SetElementContentGrammar(g57)
	g18.SetElementContentGrammar(g63)
	g19.SetElementContentGrammar(g64)
	g20.SetElementContentGrammar(g65)
	g21.SetElementContentGrammar(g66)
	g22.SetElementContentGrammar(g67)
	g23.SetElementContentGrammar(g68)
	g24.SetElementContentGrammar(g69)
	g25.SetElementContentGrammar(g70)
	g26.SetElementContentGrammar(g71)
	g27.SetElementContentGrammar(g72)
	g28.SetElementContentGrammar(g73)
	g29.SetElementContentGrammar(g74)
	g30.SetElementContentGrammar(g75)
	g31.SetElementContentGrammar(g76)
	g32.SetElementContentGrammar(g77)
	g33.SetElementContentGrammar(g78)
	g34.SetElementContentGrammar(g79)
	/* END FirstStartGrammar ----- */

	return &EXIOptionsHeaderGrammars{
		schemaID:       nil,
		grammarContext: gc,
		document:       g0,
		fragment:       g3,
		sief:           g15,
	}, nil
}

func (g *EXIOptionsHeaderGrammars) IsSchemaInformed() bool {
	return true
}

func (g *EXIOptionsHeaderGrammars) GetSchemaID() *string {
	return g.schemaID
}

func (g *EXIOptionsHeaderGrammars) SetSchemaID(schemaID *string) error {
	g.schemaID = schemaID
	return nil
}

func (g *EXIOptionsHeaderGrammars) IsBuiltInXMLSchemaTypesOnly() bool {
	return false
}

func (g *EXIOptionsHeaderGrammars) GetDocumentGrammar() Grammar {
	return g.document
}

func (g *EXIOptionsHeaderGrammars) GetFragmentGrammar() Grammar {
	return g.fragment
}

func (g *EXIOptionsHeaderGrammars) GetGrammarContext() *GrammarContext {
	return g.grammarContext
}

func (g *EXIOptionsHeaderGrammars) GetSchemaInformedGrammars() (*SchemaInformedGrammars, error) {
	gs := NewSchemaInformedGrammars(g.grammarContext, g.document, g.fragment, g.sief)
	if err := gs.SetSchemaID(g.schemaID); err != nil {
		return nil, err
	}
	return gs, nil
}

/*
	EXIHeaderDecoder implementation
*/

type EXIHeaderDecoder struct {
	*AbstractEXIHeader
	lastSE                *QNameContext
	dtrSection            bool
	dtrMapTypes           []QName
	dtrMapRepresentations []QName
}

func NewEXIHeaderDecoder() *EXIHeaderDecoder {
	return &EXIHeaderDecoder{
		AbstractEXIHeader: &AbstractEXIHeader{
			headerFactory: nil,
		},
		dtrSection:            false,
		dtrMapTypes:           []QName{},
		dtrMapRepresentations: []QName{},
	}
}

func (d *EXIHeaderDecoder) clear() {
	d.lastSE = nil
	d.dtrSection = false
	d.dtrMapTypes = []QName{}
	d.dtrMapRepresentations = []QName{}
}

func (d *EXIHeaderDecoder) Parse(headerChannel BitDecoderChannel, noOptionsFactory EXIFactory) (EXIFactory, error) {
	ch, err := headerChannel.LookAhead()
	if err != nil {
		return nil, err
	}
	if rune(ch) == '$' {
		h0, err := headerChannel.Decode()
		if err != nil {
			return nil, err
		}
		h1, err := headerChannel.Decode()
		if err != nil {
			return nil, err
		}
		h2, err := headerChannel.Decode()
		if err != nil {
			return nil, err
		}
		h3, err := headerChannel.Decode()
		if err != nil {
			return nil, err
		}
		if rune(h0) != '$' || rune(h1) != 'E' || rune(h2) != 'X' || rune(h3) != 'I' {
			return nil, errors.New("no valid EXI Cookie ($EXI)")
		}
	}

	// An EXI header starts with Distinguishing Bits part, which is a
	// two bit field 1 0
	dbits, err := headerChannel.DecodeNBitUnsignedInteger(EXIHeader_NumberOfDistinguishingBits)
	if err != nil {
		return nil, err
	}
	if dbits != EXIHeader_DistinguishingBitsValue {
		return nil, errors.New("no valid EXI document according distinguishing bits")
	}

	// Presence Bit for EXI Options
	presenceOptions, err := headerChannel.DecodeBoolean()
	if err != nil {
		return nil, err
	}

	// EXI Format Version (1 4+)

	// The first bit of the version field indicates whether the version
	// is a preview or final version of the EXI format.
	// A value of 0 indicates this is a final version and a value of 1
	// indicates this is a preview version.
	previewVersion, err := headerChannel.DecodeBoolean()
	if err != nil {
		return nil, err
	}
	if previewVersion {
		return nil, fmt.Errorf("preview version of EXI")
	}

	// one or more 4-bit unsigned integers represent the version number
	// 1. Read next 4 bits as an unsigned integer value.
	// 2. Add the value that was just read to the version number.
	// 3. If the value is 15, go to step 1, otherwise (i.e. the value
	// being in the range of 0-14), use the current value of the version
	// number as the EXI version number.
	value := -1
	version := 0

	for {
		value, err = headerChannel.DecodeNBitUnsignedInteger(EXIHeader_NumberOfFormatVersionBits)
		if err != nil {
			return nil, err
		}
		version += value

		if !(value == EXIHeader_FormatVersionContinueValue) {
			break
		}
	}

	if version != 0 {
		return nil, fmt.Errorf("incorrect EXI version: %d", version)
	}

	// [EXI Options] ?
	var exiFactory EXIFactory
	if presenceOptions {
		// use default options and re-set if needed
		exiFactory, err = d.ReadEXIOptions(headerChannel, noOptionsFactory)
		if err != nil {
			return nil, err
		}
	} else {
		exiFactory = noOptionsFactory
	}

	// other than bit-packed has [Padding Bits]
	codingMode := exiFactory.GetCodingMode()
	if codingMode != CodingModeBitPacked {
		if err := headerChannel.Align(); err != nil {
			return nil, err
		}
	}

	return exiFactory, nil
}

func (d *EXIHeaderDecoder) ReadEXIOptions(headerChannel BitDecoderChannel, noOptionsFactory EXIFactory) (EXIFactory, error) {
	factory, err := d.GetHeaderFactory()
	if err != nil {
		return nil, err
	}
	ebd, err := factory.CreateEXIBodyDecoder()
	if err != nil {
		return nil, err
	}
	decoder := ebd.(*EXIBodyDecoderInOrder)

	// schemaId = null;
	// schemaIdSet = false;

	// // clone factory
	// EXIFactory exiOptionsFactory = noOptionsFactory.clone();
	exiOptionsFactory := NewDefaultEXIFactory()
	// re-use important settings
	exiOptionsFactory.SetSchemaIDResolver(noOptionsFactory.GetSchemaIDResolver())
	exiOptionsFactory.SetDecodingOptions(noOptionsFactory.GetDecodingOptions())
	// re-use schema knowledge
	exiOptionsFactory.SetGrammars(noOptionsFactory.GetGrammars())

	// // STRICT is special, there is no NON STRICT flag --> per default set
	// to
	// // non strict
	// if (exiOptionsFactory.getFidelityOptions().isStrict()) {
	// exiOptionsFactory.getFidelityOptions().setFidelity(
	// FidelityOptions.FEATURE_STRICT, false);
	// }

	d.clear()

	eventType, exists, err := decoder.Next()
	if err != nil {
		return nil, err
	}
	for exists {
		switch eventType {
		case EventTypeStartDocument:
			if err := decoder.DecodeStartDocument(); err != nil {
				return nil, err
			}
		case EventTypeEndDocument:
			if err := decoder.DecodeEndDocument(); err != nil {
				return nil, err
			}
		case EventTypeAttributeXsiNil:
			if _, err := decoder.DecodeAttributeXsiNil(); err != nil {
				return nil, err
			}
			if err := d.handleXsiNil(decoder.GetAttributeValue(), exiOptionsFactory); err != nil {
				return nil, err
			}
		case EventTypeAttributeXsiType:
			if _, err := decoder.DecodeAttributeXsiType(); err != nil {
				return nil, err
			}
		case EventTypeAttribute, EventTypeAttributeNS, EventTypeAttributeGeneric,
			EventTypeAttributeGenericUndeclared, EventTypeAttributeInvalidValue, EventTypeAttributeAnyInvalidValue:
			if _, err := decoder.DecodeAttribute(); err != nil {
				return nil, err
			}
		case EventTypeNamespaceDeclaration:
			if _, err := decoder.DecodeNamespaceDeclaration(); err != nil {
				return nil, err
			}
		case EventTypeStartElement, EventTypeStartElementNS, EventTypeStartElementGeneric, EventTypeStartElementGenericUndeclared:
			se, err := decoder.DecodeStartElement()
			if err != nil {
				return nil, err
			}
			if err := d.handleStartElement(se, exiOptionsFactory); err != nil {
				return nil, err
			}
		case EventTypeEndElement, EventTypeEndElementUndeclared:
			ee, err := decoder.DecodeEndElement()
			if err != nil {
				return nil, err
			}
			if err := d.handleEndElement(ee, exiOptionsFactory); err != nil {
				return nil, err
			}
		case EventTypeCharacters, EventTypeCharactersGeneric, EventTypeCharactersGenericUndeclared:
			ch, err := decoder.DecodeCharacters()
			if err != nil {
				return nil, err
			}
			if err := d.handleCharacters(ch, exiOptionsFactory); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unexpected EXI event in header: %d", eventType)
		}

		eventType, exists, err = decoder.Next()
		if err != nil {
			return nil, err
		}
	}

	if len(d.dtrMapTypes) == len(d.dtrMapRepresentations) && len(d.dtrMapTypes) > 0 {
		dtrMapTypesA := make([]QName, len(d.dtrMapTypes))
		copy(dtrMapTypesA, d.dtrMapTypes)
		dtrMapRepresentationsA := make([]QName, len(d.dtrMapRepresentations))
		copy(dtrMapRepresentationsA, d.dtrMapRepresentations)
		exiOptionsFactory.SetDatatypeRepresentationMap(&dtrMapTypesA, &dtrMapRepresentationsA)
	}

	return exiOptionsFactory, nil
}

func (d *EXIHeaderDecoder) handleStartElement(se *QNameContext, f EXIFactory) error {
	if d.dtrSection {
		if len(d.dtrMapTypes) == len(d.dtrMapRepresentations) {
			// schema datatype
			d.dtrMapTypes = append(d.dtrMapTypes, se.qName)
		} else {
			// datatype representation
			d.dtrMapRepresentations = append(d.dtrMapRepresentations, se.qName)
		}
	} else if se.GetNamespaceUri() == W3C_EXI_NS_URI {
		localName := se.GetLocalName()

		switch localName {
		case EXIHeader_Byte:
			f.SetCodingMode(CodingModeBytePacked)
		case EXIHeader_PreCompress:
			f.SetCodingMode(CodingModePreCompression)
		case EXIHeader_SelfContained:
			if err := f.GetFidelityOptions().SetFidelity(FeatureSC, true); err != nil {
				return err
			}
		case EXIHeader_DatatypeRepresentationMap:
			d.dtrSection = true
		case EXIHeader_Dtd:
			if err := f.GetFidelityOptions().SetFidelity(FeatureDTD, true); err != nil {
				return err
			}
		case EXIHeader_Prefixes:
			if err := f.GetFidelityOptions().SetFidelity(FeaturePrefix, true); err != nil {
				return err
			}
		case EXIHeader_LexicalValues:
			if err := f.GetFidelityOptions().SetFidelity(FeatureLexicalValue, true); err != nil {
				return err
			}
		case EXIHeader_Comments:
			if err := f.GetFidelityOptions().SetFidelity(FeatureComment, true); err != nil {
				return err
			}
		case EXIHeader_Pis:
			if err := f.GetFidelityOptions().SetFidelity(FeaturePI, true); err != nil {
				return err
			}
		case EXIHeader_Compression:
			f.SetCodingMode(CodingModeCompression)
		case EXIHeader_Fragment:
			f.SetFragment(true)
		case EXIHeader_Strict:
			if err := f.GetFidelityOptions().SetFidelity(FeatureStrict, true); err != nil {
				return err
			}
		case EXIHeader_Profile:
			// profile parameters are not used yet
		}
	}

	d.lastSE = se
	return nil
}

func (d *EXIHeaderDecoder) handleEndElement(ee *QNameContext, _ EXIFactory) error {
	if ee.GetNamespaceUri() == W3C_EXI_NS_URI {
		localName := ee.GetLocalName()

		if localName == EXIHeader_DatatypeRepresentationMap {
			d.dtrSection = false
		}
	}

	return nil
}

func (d *EXIHeaderDecoder) handleCharacters(value Value, f EXIFactory) error {
	localName := d.lastSE.GetLocalName()

	switch localName {
	case EXIHeader_ValueMaxLength:
		val, ok := value.(*IntegerValue)
		if ok {
			f.SetValueMaxLength(val.Value32())
		} else {
			return fmt.Errorf("failure while processing: %s", localName)
		}
	case EXIHeader_ValuePartitionCapacity:
		val, ok := value.(*IntegerValue)
		if ok {
			if val.GetIntegerValueType() == IntegerValue32 {
				f.SetValuePartitionCapacity(val.Value32())
			} else {
				return fmt.Errorf("ValuePartitionCapacity other than int not supported")
			}
		} else {
			return fmt.Errorf("failure while processing: %s", localName)
		}
	case EXIHeader_BlockSize:
		val, ok := value.(*IntegerValue)
		if ok {
			if val.GetIntegerValueType() == IntegerValue32 {
				f.SetBlockSize(val.Value32())
			} else {
				return fmt.Errorf("BlockSize other than int not supported")
			}
		} else {
			return fmt.Errorf("failure while processing: %s", localName)
		}
	case EXIHeader_SchemaID:
		if f.GetDecodingOptions().IsOptionEnabled(OptionIgnoreSchemaID) {
			// ignoring
		} else {
			schemaID, err := value.ToString()
			if err != nil {
				return err
			}

			sir := f.GetSchemaIDResolver()
			if sir != nil {
				grammars, err := sir.ResolveSchemaID(schemaID)
				if err != nil {
					return err
				}
				f.SetGrammars(grammars)
			} else {
				return fmt.Errorf("exi header provides schema ID '%s' but no SchemaIDResolver set", schemaID)
			}
		}
	case EXIHeader_Profile:
		if value.GetValueType() == ValueTypeDecimal {
			val := value.(*DecimalValue)
			f.SetLocalValuePartitions(val.IsNegative())
			if val.GetIntegral().GetIntegerValueType() != IntegerValue32 {
				return fmt.Errorf("decimal's integral part is not int")
			}
			f.SetMaximumNumberOfBuiltInElementGrammars(val.GetIntegral().Value32() - 1)
			if val.GetRevFractional().GetIntegerValueType() != IntegerValue32 {
				return fmt.Errorf("decimal's reverse fractional part is not int")
			}
			f.SetMaximumNumberOfBuiltInProductions(val.GetRevFractional().Value32() - 1)
		}
	}

	return nil
}

func (d *EXIHeaderDecoder) handleXsiNil(value Value, f EXIFactory) error {
	localName := d.lastSE.GetLocalName()

	if localName == EXIHeader_SchemaID {
		val, ok := value.(*BooleanValue)
		if ok {
			if val.ToBoolean() {
				// schema-less, default
				f.SetGrammars(NewSchemaLessGrammars())
			}
		} else {
			return fmt.Errorf("failure while processing: %s", localName)
		}
	}

	return nil
}

/*
	EXIHeaderEncoder implementation
*/

type EXIHeaderEncoder struct {
	*AbstractEXIHeader
}

func NewEXIHeaderEncoder() *EXIHeaderEncoder {
	return &EXIHeaderEncoder{
		AbstractEXIHeader: &AbstractEXIHeader{
			headerFactory: nil,
		},
	}
}

func (e *EXIHeaderEncoder) Write(headerChannel *BitEncoderChannel, f EXIFactory) error {
	headerOptions := f.GetEncodingOptions()
	codingMode := f.GetCodingMode()

	if headerOptions.IsOptionEnabled(OptionIncludeCookie) {
		// four byte field consists of four characters " $ " , " E ",
		// " X " and " I " in that order.
		if err := headerChannel.Encode('$'); err != nil {
			return err
		}
		if err := headerChannel.Encode('E'); err != nil {
			return err
		}
		if err := headerChannel.Encode('X'); err != nil {
			return err
		}
		if err := headerChannel.Encode('I'); err != nil {
			return err
		}
	}

	// Distinguishing Bits 10
	if err := headerChannel.EncodeNBitUnsignedInteger(2, 2); err != nil {
		return err
	}

	// Presence Bit for EXI Options 0
	includeOptions := headerOptions.IsOptionEnabled(OptionIncludeOptions)
	if err := headerChannel.EncodeBoolean(includeOptions); err != nil {
		return err
	}

	// EXI Format Version 0-0000
	if err := headerChannel.EncodeBoolean(false); err != nil { // preview
		return err
	}
	if err := headerChannel.EncodeNBitUnsignedInteger(0, 4); err != nil {
		return err
	}

	// EXI Header options and so forth
	if includeOptions {
		if err := e.WriteEXIOptions(f, headerChannel); err != nil {
			return err
		}
	}

	// other than bit-packed requires [Padding Bits]
	if codingMode != CodingModeBitPacked {
		if err := headerChannel.Align(); err != nil {
			return err
		}
		if err := headerChannel.Flush(); err != nil {
			return err
		}
	}

	return nil
}

func (e *EXIHeaderEncoder) WriteEXIOptions(f EXIFactory, encoderChannel EncoderChannel) error {
	factory, err := e.GetHeaderFactory()
	if err != nil {
		return err
	}
	enc, err := factory.CreateEXIBodyEncoder()
	if err != nil {
		return err
	}

	encoder := enc.(*EXIBodyEncoderInOrder)
	if err := encoder.SetOutputChannel(encoderChannel); err != nil {
		return err
	}

	if err := encoder.EncodeStartDocument(); err != nil {
		return err
	}
	if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_Header, nil); err != nil {
		return err
	}

	if e.isLessCommon(f) {
		if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_LessCommon, nil); err != nil {
			return err
		}

		if e.isUncommon(f) {
			if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_Uncommon, nil); err != nil {
				return err
			}

			if e.isUserDefinedMetaData(f) {
				if f.GetEncodingOptions().IsOptionEnabled(OptionIncludeProfileValues) {
					// EXI profile options
					if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_Profile, nil); err != nil {
						return err
					}

					/*
					 * 1. The localValuePartitions parameter is encoded as
					 * the sign of the decimal value: the parameter is equal
					 * to 0 if the decimal value is positive and 1 if the
					 * decimal value is negative.
					 */
					negative := f.IsLocalValuePartitions()
					/*
					 * 2. The maximumNumberOfBuiltInElementGrammars
					 * parameter is represented by the first unsigned
					 * integer corresponding to integral portion of the
					 * decimal value: the
					 * maximumNumberOfBuiltInElementGrammars parameter is
					 * unbounded if the unsigned integer value is 0;
					 * otherwise it is equal to the unsigned integer value -
					 * 1.
					 */
					integral := IntegerValueOf32(f.GetMaximumNumberOfBuiltInElementGrammars() + 1)
					/*
					 * 3. The maximumNumberOfBuiltInProductions parameter is
					 * represented by the second unsigned integer
					 * corresponding to the fractional portion in reverse
					 * order of the decimal value: the
					 * maximumNumberOfBuiltInProductions parameter is
					 * unbounded if the unsigned integer value is 0;
					 * otherwise it is equal to the unsigned integer value -
					 * 1.
					 */
					revFractional := IntegerValueOf32(f.GetMaximumNumberOfBuiltInProductions() + 1)

					qnv := NewQNameValue(XMLSchemaNS_URI, "decimal", nil)
					if err := encoder.EncodeAttributeXsiType(qnv, nil); err != nil {
						return err
					}

					dv := NewDecimalValue(negative, integral, revFractional)
					if err := encoder.encodeCharactersForce(dv); err != nil {
						return err
					}

					if err := encoder.EncodeEndElement(); err != nil {
						return err
					}
				}
			}

			if e.isAlignment(f) {
				if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_Alignment, nil); err != nil {
					return err
				}

				if e.isByte(f) {
					if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_Byte, nil); err != nil {
						return err
					}
					if err := encoder.EncodeEndElement(); err != nil {
						return err
					}
				}

				if e.isPreCompress(f) {
					if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_PreCompress, nil); err != nil {
						return err
					}
					if err := encoder.EncodeEndElement(); err != nil {
						return err
					}
				}
			}

			if e.isSelfContained(f) {
				if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_SelfContained, nil); err != nil {
					return err
				}
				if err := encoder.EncodeEndElement(); err != nil {
					return err
				}
			}

			if e.isValueMaxLength(f) {
				if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_ValueMaxLength, nil); err != nil {
					return err
				}
				if err := encoder.EncodeCharacters(IntegerValueOf32(f.GetValueMaxLength())); err != nil {
					return err
				}
				if err := encoder.EncodeEndElement(); err != nil {
					return err
				}
			}

			if e.isValuePartitionCapacity(f) {
				if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_ValuePartitionCapacity, nil); err != nil {
					return err
				}
				if err := encoder.EncodeCharacters(IntegerValueOf32(f.GetValuePartitionCapacity())); err != nil {
					return err
				}
				if err := encoder.EncodeEndElement(); err != nil {
					return err
				}
			}

			if e.isDatatypeRepresentationMap(f) {
				types := f.GetDatatypeRepresentationMapTypes()
				representations := f.GetDatatypeRepresentationMapRepresentations()

				if len(*types) != len(*representations) {
					return errors.New("datatype representations map types size != map representations size")
				}

				// sequence "schema datatype" + datatype representation
				for i := range len(*types) {
					if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_DatatypeRepresentationMap, nil); err != nil {
						return err
					}

					// schema datatype
					kind := (*types)[i]
					if err := encoder.EncodeStartElement(kind.Space, kind.Local, nil); err != nil {
						return err
					}
					if err := encoder.EncodeEndElement(); err != nil {
						return err
					}

					// datatype representation
					representation := (*representations)[i]
					if err := encoder.EncodeStartElement(representation.Space, representation.Local, nil); err != nil {
						return err
					}
					if err := encoder.EncodeEndElement(); err != nil {
						return err
					}

					// datatypeRepresentationMap
					if err := encoder.EncodeEndElement(); err != nil {
						return err
					}
				}
			}

			// uncommon
			if err := encoder.EncodeEndElement(); err != nil {
				return err
			}
		}

		if e.isPreserve(f) {
			if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_Preserve, nil); err != nil {
				return err
			}

			fo := f.GetFidelityOptions()

			if fo.IsFidelityEnabled(FeatureDTD) {
				if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_Dtd, nil); err != nil {
					return err
				}
				if err := encoder.EncodeEndElement(); err != nil {
					return err
				}
			}
			if fo.IsFidelityEnabled(FeaturePrefix) {
				if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_Prefixes, nil); err != nil {
					return err
				}
				if err := encoder.EncodeEndElement(); err != nil {
					return err
				}
			}
			if fo.IsFidelityEnabled(FeatureLexicalValue) {
				if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_LexicalValues, nil); err != nil {
					return err
				}
				if err := encoder.EncodeEndElement(); err != nil {
					return err
				}
			}
			if fo.IsFidelityEnabled(FeatureComment) {
				if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_Comments, nil); err != nil {
					return err
				}
				if err := encoder.EncodeEndElement(); err != nil {
					return err
				}
			}
			if fo.IsFidelityEnabled(FeaturePI) {
				if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_Pis, nil); err != nil {
					return err
				}
				if err := encoder.EncodeEndElement(); err != nil {
					return err
				}
			}

			// preserve
			if err := encoder.EncodeEndElement(); err != nil {
				return err
			}
		}

		if e.isBlockSize(f) {
			if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_BlockSize, nil); err != nil {
				return err
			}
			if err := encoder.EncodeCharacters(IntegerValueOf32(f.GetBlockSize())); err != nil {
				return err
			}
			if err := encoder.EncodeEndElement(); err != nil {
				return err
			}
		}

		// less common
		if err := encoder.EncodeEndElement(); err != nil {
			return err
		}
	}

	if e.isCommon(f) {
		if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_Common, nil); err != nil {
			return err
		}

		if e.isCompression(f) {
			if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_Compression, nil); err != nil {
				return err
			}
			if err := encoder.EncodeEndElement(); err != nil {
				return err
			}
		}

		if e.isFragment(f) {
			if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_Fragment, nil); err != nil {
				return err
			}
			if err := encoder.EncodeEndElement(); err != nil {
				return err
			}
		}

		if e.isSchemaID(f) {
			if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_SchemaID, nil); err != nil {
				return err
			}

			g := f.GetGrammars()

			// When the value of the "schemaID" element is empty, no user
			// defined schema information is used for processing the EXI
			// body; however, the built-in XML schema types are available
			// for use in the EXI body.
			if g.IsBuiltInXMLSchemaTypesOnly() {
				if g.GetSchemaID() != nil || *g.GetSchemaID() != "" {
					return fmt.Errorf("schemaID is not empty")
				}
				if err := encoder.EncodeCharacters(EmptyStringValue); err != nil {
					return err
				}
			} else {
				if g.IsSchemaInformed() {
					// schema-informed
					// An example schemaID scheme is the use of URI that is
					// apt for globally identifying schema resources on the
					// Web.

					// HeaderOptions ho = f.getHeaderOptions();
					// Object schemaId = ho.getOptionValue(HeaderOptions.INCLUDE_SCHEMA_ID);
					schemaID := g.GetSchemaID()
					if schemaID == nil || *schemaID == "" {
						return fmt.Errorf("schemaID is empty")
					}

					if err := encoder.EncodeCharacters(NewStringValueFromString(*schemaID)); err != nil {
						return err
					}
				} else {
					// schema-less
					// When the "schemaID" element in the EXI options
					// document
					// contains the xsi:nil attribute with its value set to
					// true, no
					// schema information is used for processing the EXI
					// body.
					// TODO typed fashion
					if err := encoder.EncodeAttributeXsiNil(BooleanValueTrue, nil); err != nil {
						return err
					}
				}
			}

			if err := encoder.EncodeEndElement(); err != nil {
				return err
			}
		}

		// common
		if err := encoder.EncodeEndElement(); err != nil {
			return err
		}
	}

	if e.isStrict(f) {
		if err := encoder.EncodeStartElement(W3C_EXI_NS_URI, EXIHeader_Strict, nil); err != nil {
			return err
		}
		if err := encoder.EncodeEndElement(); err != nil {
			return err
		}
	}

	// header
	if err := encoder.EncodeEndElement(); err != nil {
		return err
	}
	if err := encoder.EncodeEndDocument(); err != nil {
		return err
	}

	return nil
}

func (e *EXIHeaderEncoder) isLessCommon(f EXIFactory) bool {
	return e.isUncommon(f) || e.isPreserve(f) || e.isBlockSize(f)
}

func (e *EXIHeaderEncoder) isUncommon(f EXIFactory) bool {
	// user defined meta-data, alignment, selfContained, valueMaxLength,
	// valuePartitionCapacity, datatypeRepresentationMap
	return e.isUserDefinedMetaData(f) || e.isAlignment(f) || e.isSelfContained(f) ||
		e.isValueMaxLength(f) || e.isValuePartitionCapacity(f) || e.isDatatypeRepresentationMap(f)
}

func (e *EXIHeaderEncoder) isUserDefinedMetaData(f EXIFactory) bool {
	return f.IsGrammarLearningDisabled() || !f.IsLocalValuePartitions()
}

func (e *EXIHeaderEncoder) isAlignment(f EXIFactory) bool {
	return e.isByte(f) || e.isPreCompress(f)
}

func (e *EXIHeaderEncoder) isByte(f EXIFactory) bool {
	return f.GetCodingMode() == CodingModeBytePacked
}

func (e *EXIHeaderEncoder) isPreCompress(f EXIFactory) bool {
	return f.GetCodingMode() == CodingModePreCompression
}

func (e *EXIHeaderEncoder) isSelfContained(f EXIFactory) bool {
	return f.GetFidelityOptions().IsFidelityEnabled(FeatureSC)
}

func (e *EXIHeaderEncoder) isValueMaxLength(f EXIFactory) bool {
	return f.GetValueMaxLength() != DefaultValueMaxLength
}

func (e *EXIHeaderEncoder) isValuePartitionCapacity(f EXIFactory) bool {
	return f.GetValuePartitionCapacity() >= 0
}

func (e *EXIHeaderEncoder) isDatatypeRepresentationMap(f EXIFactory) bool {
	// Canonical EXI: When the value of the Preserve.lexicalValues fidelity
	// option is true the element datatypeRepresentationMap MUST be omitted
	return !f.GetFidelityOptions().IsFidelityEnabled(FeatureLexicalValue) &&
		f.GetDatatypeRepresentationMapTypes() != nil &&
		len(*f.GetDatatypeRepresentationMapTypes()) > 0
}

func (e *EXIHeaderEncoder) isPreserve(f EXIFactory) bool {
	fo := f.GetFidelityOptions()
	return fo.IsFidelityEnabled(FeatureDTD) ||
		fo.IsFidelityEnabled(FeaturePrefix) ||
		fo.IsFidelityEnabled(FeatureLexicalValue) ||
		fo.IsFidelityEnabled(FeatureComment) ||
		fo.IsFidelityEnabled(FeaturePI)
}

func (e *EXIHeaderEncoder) isBlockSize(f EXIFactory) bool {
	return f.GetBlockSize() != DefaultBlockSize &&
		(f.GetCodingMode() == CodingModeCompression || f.GetCodingMode() == CodingModePreCompression)
}

func (e *EXIHeaderEncoder) isCommon(f EXIFactory) bool {
	return e.isCompression(f) || e.isFragment(f) || e.isSchemaID(f)
}

func (e *EXIHeaderEncoder) isCompression(f EXIFactory) bool {
	return f.GetCodingMode() == CodingModeCompression
}

func (e *EXIHeaderEncoder) isFragment(f EXIFactory) bool {
	return f.IsFragment()
}

func (e *EXIHeaderEncoder) isSchemaID(f EXIFactory) bool {
	return f.GetDecodingOptions().IsOptionEnabled(OptionIncludeSchemaID)
}

func (e *EXIHeaderEncoder) isStrict(f EXIFactory) bool {
	return f.GetFidelityOptions().IsStrict()
}
