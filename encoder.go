package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// XMLEncoder is a half-baked replacement for xml.Encoder from the standard lib
// with a goal of producing the XML that DSpace needs, since the DSpace METS
// (maybe all METS files?) has so many crazy namespaced elements and
// attributes.  The output parity is achieved primarily by not being
// intelligent and instead just reflecting the input structure, even if it's
// not ideal or even technically wrong.
//
// For simplicity, this replacement doesn't actually handle all situations a
// real Encoder needs to handle, it doesn't handle errors, and it can produce
// really busted XML.
type XMLEncoder struct {
	w              io.Writer
	depth          int
	prefix, indent string
	printIndent    bool
	emptyElement   bool
}

// NewXMLEncoder returns an XML encoder which writes to w
func NewXMLEncoder(w io.Writer) *XMLEncoder {
	return &XMLEncoder{w: w}
}

// Indent sets up the encoder to put each tag on a new line prefixed with
// prefix and indented with repeated indent strings per depth level
func (e *XMLEncoder) Indent(prefix, indent string) {
	e.prefix = prefix
	e.indent = indent
}

// EncodeToken writes out the given token to the writer
func (e *XMLEncoder) EncodeToken(t xml.Token) error {
	var out string

	switch element := t.(type) {
	case xml.StartElement:
		e.writeIndent(1)
		out += "<" + identifier(element.Name)
		var attrs []string
		for _, a := range element.Attr {
			attrs = append(attrs, fmt.Sprintf(`%s="%s"`, identifier(a.Name), a.Value))
		}
		if len(attrs) > 0 {
			out += " " + strings.Join(attrs, " ")
		}
		out += ">"
		e.w.Write([]byte(out))
		e.emptyElement = true

	case xml.EndElement:
		if e.emptyElement {
			e.depth--
		} else {
			e.writeIndent(-1)
		}
		e.w.Write([]byte("</" + identifier(element.Name) + ">"))
		e.emptyElement = false

	case xml.CharData:
		if len(element) > 0 && strings.TrimSpace(string(element)) != "" {
			e.emptyElement = false
			e.w.Write(element)
			e.printIndent = false
		}

	case xml.Comment:
		logger.Fatalf("%#v", element)

	case xml.ProcInst:
		e.w.Write([]byte(fmt.Sprintf("<?%s %s?>", element.Target, element.Inst)))

	case xml.Directive:
		logger.Fatalf("%#v", element)
	}

	return nil
}

func (e *XMLEncoder) writeIndent(depthDelta int) {
	if e.prefix == "" && e.indent == "" {
		return
	}

	if depthDelta < 0 && e.depth > 0 {
		e.depth--
	}

	if e.printIndent {
		e.w.Write([]byte("\n"))
		if e.prefix != "" {
			e.w.Write([]byte(e.prefix))
		}
		if e.indent != "" {
			e.w.Write([]byte(strings.Repeat(e.indent, e.depth)))
		}
	} else {
		e.printIndent = true
	}

	if depthDelta > 0 {
		e.depth++
	}
}

// identifier has to convert the namespace URIs Go gives us into meaningful prefixes
func identifier(n xml.Name) string {
	if n.Space == "" || n.Space == "http://www.loc.gov/METS/" {
		return n.Local
	}

	var namespace = ""
	switch n.Space {
	case "xmlns":
		namespace = "xmlns"
	case "http://www.w3.org/2001/XMLSchema-instance":
		namespace = "xsi"
	case "http://www.loc.gov/mods/v3":
		namespace = "mods"
	case "http://www.dspace.org/xmlns/dspace/dim":
		namespace = "dim"
	case "http://www.w3.org/1999/xlink":
		namespace = "xlink"
	default:
		logger.Fatalf("You need to handle namespace %q, sir", n.Space)
	}

	return namespace + ":" + n.Local
}
