package cyclonedxxml

import (
	"io"

	"github.com/CycloneDX/cyclonedx-go"

	"github.com/nextlinux/gosbom/gosbom/formats/common/cyclonedxhelpers"
	"github.com/nextlinux/gosbom/gosbom/sbom"
)

func encoder(output io.Writer, s sbom.SBOM) error {
	bom := cyclonedxhelpers.ToFormatModel(s)
	enc := cyclonedx.NewBOMEncoder(output, cyclonedx.BOMFileFormatXML)
	enc.SetPretty(true)

	err := enc.Encode(bom)
	return err
}
