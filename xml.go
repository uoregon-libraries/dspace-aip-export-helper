package main

import (
	"bytes"
	"encoding/xml"
	"os"
)

// decodeMets takes the full mets.xml data as a slice of bytes rather than a
// reader so we can close everything prior to extracting data from the mets
// blobs.  The return is a list of handles found as dependencies.
func decodeMets(data []byte) []string {
	var buf = bytes.NewBuffer(data)
	var d = xml.NewDecoder(buf)
	var handles []string

	var capture bool
	for {
		var t, err = d.Token()
		if t == nil {
			break
		}
		if err != nil {
			logger.Printf("Error reading mets.xml: %s", err)
			os.Exit(1)
		}

		switch element := t.(type) {
		case xml.StartElement:
			capture = isRelationElement(element)
		case xml.CharData:
			if capture {
				handles = append(handles, string(element))
			}
			capture = false
		}
	}

	return handles
}

// isRelationElement returns whether the xml element is defining the "is part
// of" relationship in METS
func isRelationElement(element xml.StartElement) bool {
	if element.Name.Local != "field" {
		return false
	}

	var hasRelationAttr, hasQualifierPartOfAttr bool
	for _, a := range element.Attr {
		if a.Name.Local == "element" && a.Value == "relation" {
			hasRelationAttr = true
		}
		if a.Name.Local == "qualifier" && a.Value == "isPartOf" {
			hasQualifierPartOfAttr = true
		}
	}
	return hasRelationAttr && hasQualifierPartOfAttr
}
