# EXI-GO - Efficient XML Interchange (EXI) Format 1.0 impementation in Go

EXI is a very compact representation for the Extensible Markup Language (XML) Information Set that is intended to simultaneously optimize performance and the utilization of computational resources. The EXI format uses a hybrid approach drawn from the information and formal language theories, plus practical techniques verified by measurements, for entropy encoding XML information. Using a relatively simple algorithm, which is amenable to fast and compact implementation, and a small set of datatype representations, it reliably produces efficient encodings of XML event streams.

EXI-GO is an attempt to rewrite [EXIficient](https://github.com/EXIficient) (the most advanced open-source EXI codec) in the Go language. At the moment the project closely follows the original Java code.

## Installation

```
go get github.com/sderkacs/exi-go
```

## Features and Limitations

The project is in its early stages, and several features have not yet been implemented. Since runtime grammar generation has not yet been implemented, the project relies on a modified version of [exificient-grammars](https://github.com/sderkacs/exificient-grammars) with support for Go code generation. EXI compression is also not yet implemented.

## Goals

This project's ultimate goal is to provide the Go community with a fully functional toolkit for working with EXI documents.

## TODO

- [x] EXI marshaling/unmarshaling with Go structs
- [ ] General code cleanup and better comments/documentation
- [ ] Better error handling
- [ ] Runtime grammars loading from XSD
- [ ] Grammars serialization to Go code
- [ ] Provide tools for parsing XSD and generating Go structures that take into account EXI specifics
- [ ] Compressed EXI messages

## Usage

```
grammars, err := foo.NewFooGrammars()
if err != nil {
    panic(err)
}

// Create and configure EXI factory with grammars
fidelityOptions := core.NewDefaultFidelityOptions()

exiFactory := core.NewDefaultEXIFactory()
exiFactory.SetGrammars(grammars)
exiFactory.SetFragment(false) // <- Enable or disable fragment mode
exiFactory.SetFidelityOptions(fidelityOptions)
exiFactory.SetCodingMode(core.CodingModeBitPacked)

// Decode EXI to XML

exiFile := "foo.exi"
exi, err := os.ReadFile(exiFile)
if err != nil {
    panic(err)
}

exiSource := bufio.NewReader(bytes.NewReader(exi))
resultBuffer := bytes.Buffer{}
xmlEncoder := xml.NewEncoder(&resultBuffer)

exiDecoder, err := sax.NewSAXDecoder(exiFactory)
if err != nil {
    panic(err)
}
rootName, err := exiDecoder.Parse(exiSource, xmlEncoder)
if err != nil {
    panic(err)
}

// Knowing the name of the root element can be useful in some situations.
fmt.Printf("Root element name: %s\n", rootName) 

if err := os.WriteFile("foo.xml", resultBuffer.Bytes(), 0644); err != nil {
    panic(err)
}

// Encode XML to EXI

xmlFile := "foo.xml"
xmlBytes, err := os.ReadFile(xmlFile)
if err != nil {
    panic(err)
}

xmlSource := bufio.NewReader(bytes.NewReader(xmlBytes))
exiEncoder, err := sax.NewSAXEncoder(exiFactory)
if err != nil {
    panic(err)
}
resultBuffer := bytes.Buffer{}
if err := exiEncoder.SetWriter(bufio.NewWriter(&resultBuffer)); err != nil {
    panic(err)
}
if err := exiEncoder.Encode(xmlSource); err != nil {
    panic(err)
}
if err := os.WriteFile("foo.exi", resultBuffer.Bytes(), 0644); err != nil {
    panic(err)
}
```

## Contributing

Contributions are welcome! Whether you're fixing bugs, adding features, or reporting issues, your help is invaluable in improving this project.

## Acknowledgments

This project is a derivative work, adapted into Go from the original [EXIficient](https://github.com/EXIficient), created by Siemens AG under the MIT License.
