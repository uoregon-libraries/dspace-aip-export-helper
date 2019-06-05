// This application tells DSpace to export a given item and all its children,
// then reverses its way back up all containers (collections and communities)
// to export those (without their child objects).  Exports are put in temporary
// directories starting at "/tmp/aip-exports/9999/" and decrementing the final
// directory name by one on each parent traversed; e.g.,
// "/tmp/aip-exports/9998/", "/tmp/aip-exports/9997/", etc.  This allows the
// import to pull communities and collections sequentially in order to get the
// parents of the export before trying to get the object(s) in question so as
// to avoid DSpace errors.
//
// This is a one-off "script", so things like the dspace binary are very
// hard-coded.  If (when) we find a need to maintain this further, a refactor
// would be lovely.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const csi = "\x1b["
const warn = csi + "31;1m"
const clear = csi + "0m"

var logger = StderrLogger()

func usage(code int) {
	fmt.Fprintln(os.Stderr, "Usage: export <handle>")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Tells DSpace to export a given item and all its children,")
	fmt.Fprintln(os.Stderr, "exporting all dependencies (parent communities and collections)")
	fmt.Fprintln(os.Stderr, "along the way.")
	os.Exit(code)
}

func usageError(message string) {
	fmt.Fprintln(os.Stderr, warn+message+clear)
	fmt.Fprintln(os.Stderr)
	usage(1)
}

func main() {
	if len(os.Args) < 2 {
		usageError("You must specify an object handle")
	}

	var handle = os.Args[1]
	if len(handle) < 5 || handle[0:5] != "1794/" {
		handle = "1794/" + handle
	}

	logger.Printf("Exporting %q and all objects it owns...\n", handle)
	exportRecursive([]string{handle}, 0)

	logger.Println("Exporting top-level site AIP and fixing any empty groups")
	exportSite()
}

// exportRecursive exports one or more items at the given depth, then checks
// all zipped METS files for parents and exports them at depth + 1.  As the
// name suggests, this recurses until we see no new handles.
func exportRecursive(handles []string, depth int) {
	var e *Exporter
	for _, h := range handles {
		// We only want one site export
		if h == "1794/0" {
			continue
		}
		logger.Printf("Exporting handle %q", h)
		e = exportOne(h, depth)
	}

	// This happens when there are no handles other than the site root
	if e == nil {
		return
	}

	var parentHandles = extractHandlesFromZipfiles(e.Path())
	var nextHandles []string
	for _, h := range parentHandles {
		nextHandles = append(nextHandles, strings.Replace(h, "hdl:", "", 1))
	}

	if len(nextHandles) == 0 {
		return
	}

	exportRecursive(nextHandles, depth+1)
}

// exportOne runs an export and returns the generated Exporter object
func exportOne(handle string, depth int) *Exporter {
	var e = NewExporter("jechols@uoregon.edu", handle)
	e.Depth = depth

	var err = e.Run(logger)
	if err != nil {
		logger.Println(err)
		os.Exit(1)
	}

	return e
}

// extractHandlesFromZipfiles extracts the METS files from all zipped items in
// the given path and returns all parent handles found
func extractHandlesFromZipfiles(path string) []string {
	var globPattern = filepath.Join(path, "*.zip")

	var zipFiles, err = filepath.Glob(globPattern)
	if err != nil {
		logger.Println("Error reading zip files:", err)
		os.Exit(1)
	}

	var handles []string
	for _, file := range zipFiles {
		var z = NewZippie(file)
		handles = append(handles, decodeMETSHandles(z.extractMETS())...)
	}

	return handles
}

func exportSite() {
	// Export the site, obviously
	var e = exportOne("1794/0", maxDepth)
	e.Path()
	var z = NewZippie(e.Filename())
	var oldMETS = z.extractMETS()
	var newMETS = fixEmptyGroups(oldMETS)
	fmt.Println(string(newMETS))
	z.rewriteMETS(newMETS)
}
