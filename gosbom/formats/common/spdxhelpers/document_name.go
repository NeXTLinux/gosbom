package spdxhelpers

import (
	"github.com/nextlinux/gosbom/gosbom/source"
)

func DocumentName(srcMetadata source.Metadata) string {
	if srcMetadata.Name != "" {
		return srcMetadata.Name
	}

	switch srcMetadata.Scheme {
	case source.ImageScheme:
		return srcMetadata.ImageMetadata.UserInput
	case source.DirectoryScheme, source.FileScheme:
		return srcMetadata.Path
	default:
		return "unknown"
	}
}
