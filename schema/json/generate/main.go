package main

import (
	"fmt"
	"os"

	"github.com/dave/jennifer/jen"

	"github.com/nextlinux/gosbom/schema/json/internal"
)

// This program generates internal/generated.go.

const (
	pkgImport = "github.com/nextlinux/gosbom/gosbom/pkg"
	path      = "internal/generated.go"
)

func main() {
	typeNames, err := internal.AllGosbomMetadataTypeNames()
	if err != nil {
		panic(fmt.Errorf("unable to get all metadata type names: %w", err))
	}

	fmt.Printf("updating metadata container object with %+v types\n", len(typeNames))

	f := jen.NewFile("internal")
	f.HeaderComment("DO NOT EDIT: generated by schema/json/generate/main.go")
	f.ImportName(pkgImport, "pkg")
	f.Comment("ArtifactMetadataContainer is a struct that contains all the metadata types for a package, as represented in the pkg.Package.Metadata field.")
	f.Type().Id("ArtifactMetadataContainer").StructFunc(func(g *jen.Group) {
		for _, typeName := range typeNames {
			g.Id(typeName).Qual(pkgImport, typeName)
		}
	})

	rendered := fmt.Sprintf("%#v", f)

	fh, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(fmt.Errorf("unable to open file: %w", err))
	}
	_, err = fh.WriteString(rendered)
	if err != nil {
		panic(fmt.Errorf("unable to write file: %w", err))
	}
	if err := fh.Close(); err != nil {
		panic(fmt.Errorf("unable to close file: %w", err))
	}
}