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
func (z *Zippie) rewriteMETS(METSData []byte) {
	var zr = getZipReadCloser(z.Filename)
	defer zr.Close()

	var tempout = new(bytes.Buffer)
	var zw = zip.NewWriter(tempout)

	for _, fIn := range zr.File {
		if fIn.Name == "mets.xml" {
			continue
		}

		var fOut, err = zw.Create(fIn.Name)
		if err != nil {
			logger.Fatalf("Error adding %s to in-memory zip file: %s", fIn.Name, err)
		}

		var fInRC = z.openZipFile(fIn)
		_, err = io.CopyN(fOut, fInRC, int64(fIn.UncompressedSize))
		if err != nil {
			logger.Fatalf("Error copying %q from zip file %q: %s", fIn.Name, z.Filename, err)
		}
	}

	var fOut, err = zw.Create("mets.xml")
	if err != nil {
		logger.Fatalf("Error adding mets.xml to in-memory zip file: %s", err)
	}
	_, err = fOut.Write(METSData)
	if err != nil {
		logger.Fatalf("Error writing mets.xml to in-memory zip file: %s", err)
	}

	err = zw.Close()
	if err != nil {
		logger.Fatalf("Error trying to close zip buffer when rewriting zip file %q: %s", z.Filename, err)
	}

	var rewrittenFile *os.File
	rewrittenFile, err = os.Create(z.Filename)
	if err != nil {
		logger.Fatalf("Unable to remove old zip file %q: %s", z.Filename, err)
	}
	defer rewrittenFile.Close()

	_, err = io.CopyN(rewrittenFile, tempout, int64(tempout.Len()))
	if err != nil {
		logger.Fatalf("Unable to write new zip file data to %q: %s", z.Filename, err)
	}
	err = rewrittenFile.Close()
	if err != nil {
		logger.Fatalf("Unable to close new zip file %q: %s", z.Filename, err)
	}
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
