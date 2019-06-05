package main

import (
	"bytes"
	"encoding/xml"
	"os"
)

// decodeMETSHandles takes the full mets.xml data as a slice of bytes rather than a
// reader so we can close everything prior to extracting data from the METS
// blobs.  The return is a list of handles found as dependencies.
func decodeMETSHandles(data []byte) []string {
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

// fixEmptyGroups takes the full mets.xml data and looks for groups that have
// no content, replacing them with a copy of the administrator group's data.
// The new METS data is returns as a byte slice.
func fixEmptyGroups(data []byte) []byte {
	var inbuf = bytes.NewBuffer(data)
	var outbuf = new(bytes.Buffer)

	var d = xml.NewDecoder(inbuf)
	var e = NewXMLEncoder(outbuf)
	e.Indent("", "  ")

	var adminMemberTokens = getAdminMemberTokens(data)
	var inGroupElement, inMembersElement, groupHasMembers bool
	for {
		var t, err = d.Token()
		if t == nil {
			break
		}
		if err != nil {
			logger.Printf("Error tokenizing xml data: %s", err)
			os.Exit(1)
		}

		switch element := t.(type) {
		case xml.StartElement:
			switch element.Name.Local {
			case "Group":
				inGroupElement = true
				groupHasMembers = false
			case "Members":
				if inGroupElement {
					inMembersElement = true
				}
			case "Member":
				if inMembersElement {
					groupHasMembers = true
				}
			}
		case xml.EndElement:
			if element.Name.Local == "Group" {
				if !groupHasMembers {
					for _, t := range adminMemberTokens {
						e.EncodeToken(t)
					}
				}
				inGroupElement = false
			}
		}

		// Always encode the new token - if we had to shim anything in, it already happened above
		e.EncodeToken(t)
	}

	return outbuf.Bytes()
}

// isAdministrativeGroupElement returns whether the xml element is defining the
// admin group
func isAdministrativeGroupElement(element xml.StartElement) bool {
	if element.Name.Local != "Group" {
		return false
	}

	for _, a := range element.Attr {
		if a.Name.Local == "Name" && a.Value == "Administrator" {
			return true
		}
	}
	return false
}

// getAdminMemberTokens finds the <Group> tag named "Administrators" and returns all
// tokens necessary to replicate the members
func getAdminMemberTokens(data []byte) []xml.Token {
	var buf = bytes.NewBuffer(data)

	var d = xml.NewDecoder(buf)
	var adminTokens []xml.Token

	var inAdminGroupElement bool
	for {
		var t, err = d.Token()
		if t == nil {
			break
		}
		if err != nil {
			logger.Printf("Error tokenizing xml data: %s", err)
			os.Exit(1)
		}

		switch element := t.(type) {
		case xml.StartElement:
			if !inAdminGroupElement {
				if isAdministrativeGroupElement(element) {
					inAdminGroupElement = true
				}
				continue
			}
			if element.Name.Local == "Members" || element.Name.Local == "Member" {
				adminTokens = append(adminTokens, t)
			}
		case xml.EndElement:
			if !inAdminGroupElement {
				continue
			}

			if element.Name.Local == "Group" {
				inAdminGroupElement = false
			}
			if element.Name.Local == "Members" || element.Name.Local == "Member" {
				adminTokens = append(adminTokens, t)
			}
		}
	}

	return adminTokens
}
