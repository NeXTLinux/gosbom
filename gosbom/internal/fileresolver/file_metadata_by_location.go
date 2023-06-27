package fileresolver

import (
	"github.com/nextlinux/gosbom/gosbom/file"

	"github.com/anchore/stereoscope/pkg/image"
)

func fileMetadataByLocation(img *image.Image, location file.Location) (file.Metadata, error) {
	entry, err := img.FileCatalog.Get(location.Reference())
	if err != nil {
		return file.Metadata{}, err
	}

	return entry.Metadata, nil
}
