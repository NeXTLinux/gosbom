package binary

import (
	"reflect"

	"github.com/nextlinux/gosbom/gosbom/cpe"
	"github.com/nextlinux/gosbom/gosbom/file"
	"github.com/nextlinux/gosbom/gosbom/pkg"
)

func newPackage(classifier classifier, location file.Location, matchMetadata map[string]string) *pkg.Package {
	version, ok := matchMetadata["version"]
	if !ok {
		return nil
	}

	update := matchMetadata["update"]

	var cpes []cpe.CPE
	for _, c := range classifier.CPEs {
		c.Version = version
		c.Update = update
		cpes = append(cpes, c)
	}

	p := pkg.Package{
		Name:    classifier.Package,
		Version: version,
		Locations: file.NewLocationSet(
			location.WithAnnotation(pkg.EvidenceAnnotationKey, pkg.PrimaryEvidenceAnnotation),
		),
		Type:         pkg.BinaryPkg,
		CPEs:         cpes,
		FoundBy:      catalogerName,
		MetadataType: pkg.BinaryMetadataType,
		Metadata: pkg.BinaryMetadata{
			Matches: []pkg.ClassifierMatch{
				{
					Classifier: classifier.Class,
					Location:   location,
				},
			},
		},
	}

	if classifier.Type != "" {
		p.Type = classifier.Type
	}

	if !reflect.DeepEqual(classifier.PURL, emptyPURL) {
		purl := classifier.PURL
		purl.Version = version
		p.PURL = purl.ToString()
	}

	if classifier.Language != "" {
		p.Language = classifier.Language
	}

	p.SetID()

	return &p
}
