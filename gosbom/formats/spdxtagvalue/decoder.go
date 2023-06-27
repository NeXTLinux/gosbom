package spdxtagvalue

import (
	"fmt"
	"io"

	"github.com/nextlinux/gosbom/gosbom/formats/common/spdxhelpers"
	"github.com/nextlinux/gosbom/gosbom/sbom"
	"github.com/spdx/tools-golang/tagvalue"
)

func decoder(reader io.Reader) (*sbom.SBOM, error) {
	doc, err := tagvalue.Read(reader)
	if err != nil {
		return nil, fmt.Errorf("unable to decode spdx-tag-value: %w", err)
	}

	return spdxhelpers.ToGosbomModel(doc)
}
