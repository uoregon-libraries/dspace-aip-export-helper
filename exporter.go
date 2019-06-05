package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const maxDepth = 9999

// Exporter holds necessary information to generate a command-line call to the DSpace exporter
type Exporter struct {
	// EPerson is the email address of the user doing the export - DSpace won't
	// export without a valid username
	EPerson string

	// Handle to the object being exported
	Handle string

	// Depth tells us how far from the original export we are so we can make sure
	// parents are imported before children by simply importing sequentially.  At
	// a depth of 0, the directory name will be prefixed with 9999.  At a depth
	// of 1, the prefix is 9998, etc.  Additionally, at a depth of 0, we know
	// we're on the "primary" export and can use the "-a" flag on the DSpace CLI
	// to to get all child objects.
	Depth int
}

// NewExporter sets up an exporter for a full collection/community export
// operation
func NewExporter(person string, handle string) *Exporter {
	return &Exporter{EPerson: person, Handle: handle}
}

// Path returns the location where all zip files will be stored, ending with
// the reverse-sequential directory we use for sorting on import
func (e *Exporter) Path() string {
	return filepath.Join(os.TempDir(), "aip-export", fmt.Sprintf("%04d", maxDepth-e.Depth))
}

// Filename calculates the full path to the exported zipfile, including our
// sequential prefix (9999, 9998, etc)
func (e *Exporter) Filename() string {
	var hdl = strings.ReplaceAll(e.Handle, "/", "_")
	return filepath.Join(e.Path(), hdl+".zip")
}

func (e *Exporter) args() []string {
	var cmd = []string{"packager", "-u", "-d", "-t", "AIP", "-e", e.EPerson}

	if e.Depth == 0 {
		cmd = append(cmd, "-a")
	}

	return append(cmd, "-i", e.Handle, e.Filename())
}

// Run generates an export command, executes it, and returns any errors which
// occur.  A logger must be passed in to let us write information out
// somewhere, but the provided NullLogger can be used to suppress output if
// that is desired.
func (e *Exporter) Run(logger Logger) error {
	var _, err = os.Stat(e.Filename())
	if err == nil {
		logger.Printf("%q has already been exported", e.Filename())
		return nil
	}

	var binary = "/usr/local/dspace/bin/dspace"
	var args = e.args()
	logger.Printf("Executing `%s %s`", binary, strings.Join(args, " "))
	var cmd = exec.Command(binary, args...)

	var output []byte
	output, err = cmd.CombinedOutput()
	logger.Println(string(output))

	if err != nil {
		return fmt.Errorf("export command error: %s", err)
	}

	return nil
}
