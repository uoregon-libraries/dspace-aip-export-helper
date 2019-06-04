package main

import (
	"archive/zip"
	"io"
	"os"
)

// Zippie wraps logging and error handling around our mets-specific zip operations
type Zippie struct {
	Filename string
}

// NewZippie returns a wrapper for working with zip files
func NewZippie(filename string) *Zippie {
	return &Zippie{filename}
}

// extractMets reads zipfile and returns the "mets.xml" blob's bytes
func (z *Zippie) extractMets() []byte {
	var zr = getZipReadCloser(z.Filename)
	defer zr.Close()

	var f = z.getMetsFile(zr)
	var rc = z.openZipFile(f)
	defer rc.Close()

	var buf = make([]byte, f.UncompressedSize)
	var _, err = io.ReadFull(rc, buf)
	if err != nil {
		logger.Printf("Error reading mets.xml from %q: %s", z.Filename, err)
		os.Exit(1)
	}

	return buf
}

func getZipReadCloser(zipfile string) *zip.ReadCloser {
	var rc, err = zip.OpenReader(zipfile)
	if err != nil {
		logger.Printf("Error opening zip file %q: %s", zipfile, err)
		os.Exit(1)
	}

	return rc
}

func (z *Zippie) getMetsFile(zr *zip.ReadCloser) *zip.File {
	for _, f := range zr.File {
		if f.Name == "mets.xml" {
			return f
		}
	}

	logger.Printf("Zipfile %q doesn't contain a mets.xml", z.Filename)
	os.Exit(1)
	return nil
}

func (z *Zippie) openZipFile(f *zip.File) io.ReadCloser {
	var rc, err = f.Open()
	if err != nil {
		logger.Printf("Error opening %s from zipfile %q: %s", f.Name, z.Filename, err)
		os.Exit(1)
	}

	return rc
}
