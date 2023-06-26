package java

import (
	"fmt"

	intFile "github.com/nextlinux/gosbom/internal/file"
	"github.com/nextlinux/gosbom/gosbom/artifact"
	"github.com/nextlinux/gosbom/gosbom/file"
	"github.com/nextlinux/gosbom/gosbom/pkg"
	"github.com/nextlinux/gosbom/gosbom/pkg/cataloger/generic"
)

var genericZipGlobs = []string{
	"**/*.zip",
}

// TODO: when the generic archive cataloger is implemented, this should be removed (https://github.com/nextlinux/gosbom/issues/246)

// parseZipWrappedJavaArchive is a parser function for java archive contents contained within arbitrary zip files.
func parseZipWrappedJavaArchive(_ file.Resolver, _ *generic.Environment, reader file.LocationReadCloser) ([]pkg.Package, []artifact.Relationship, error) {
	contentPath, archivePath, cleanupFn, err := saveArchiveToTmp(reader.AccessPath(), reader)
	// note: even on error, we should always run cleanup functions
	defer cleanupFn()
	if err != nil {
		return nil, nil, err
	}

	// we use our zip helper functions instead of that from the archiver package or the standard lib. Why? These helper
	// functions support zips with shell scripts prepended to the file. Specifically, the helpers use the central
	// header at the end of the file to determine where the beginning of the zip payload is (unlike the standard lib
	// or archiver).
	fileManifest, err := intFile.NewZipFileManifest(archivePath)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read files from java archive: %w", err)
	}

	// look for java archives within the zip archive
	return discoverPkgsFromZip(reader.Location, archivePath, contentPath, fileManifest, nil)
}