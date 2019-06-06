package main

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
)

// Zippie wraps logging and error handling around our METS-specific zip operations
type Zippie struct {
	Filename string
}

// NewZippie returns a wrapper for working with zip files
func NewZippie(filename string) *Zippie {
	return &Zippie{filename}
}

// extractMETS reads zipfile and returns the "mets.xml" blob's bytes
func (z *Zippie) extractMETS() []byte {
	var zr = getZipReadCloser(z.Filename)
	defer zr.Close()

	var f = z.getMETSFile(zr)
	var rc = z.openZipFile(f)
	defer rc.Close()

	var buf = make([]byte, f.UncompressedSize)
	var _, err = io.ReadFull(rc, buf)
	if err != nil {
		logger.Fatalf("Error reading mets.xml from %q: %s", z.Filename, err)
	}

	return buf
}

func getZipReadCloser(zipfile string) *zip.ReadCloser {
	var rc, err = zip.OpenReader(zipfile)
	if err != nil {
		logger.Fatalf("Error opening zip file %q: %s", zipfile, err)
	}

	return rc
}

// rewriteMETS takes the given data and rebuilds the zip file entirely, using
// the new METS data to replace the old
func (z *Zippie) rewriteMETS(METSData []byte) error {
	return nil
}

func (z *Zippie) getMETSFile(zr *zip.ReadCloser) *zip.File {
	for _, f := range zr.File {
		if f.Name == "mets.xml" {
			return f
		}
	}

	logger.Fatalf("Zipfile %q doesn't contain a mets.xml", z.Filename)
	return nil
}

func (z *Zippie) openZipFile(f *zip.File) io.ReadCloser {
	var rc, err = f.Open()
	if err != nil {
		logger.Fatalf("Error opening %s from zipfile %q: %s", f.Name, z.Filename, err)
	}

	return rc
}
