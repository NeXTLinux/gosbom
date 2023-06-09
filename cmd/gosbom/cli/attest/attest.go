package attest

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nextlinux/gosbom/cmd/gosbom/cli/eventloop"
	"github.com/nextlinux/gosbom/cmd/gosbom/cli/options"
	"github.com/nextlinux/gosbom/cmd/gosbom/cli/packages"
	"github.com/nextlinux/gosbom/gosbom"
	"github.com/nextlinux/gosbom/gosbom/event"
	"github.com/nextlinux/gosbom/gosbom/event/monitor"
	"github.com/nextlinux/gosbom/gosbom/formats/gosbomjson"
	"github.com/nextlinux/gosbom/gosbom/formats/table"
	"github.com/nextlinux/gosbom/gosbom/sbom"
	"github.com/nextlinux/gosbom/gosbom/source"
	"github.com/nextlinux/gosbom/internal/bus"
	"github.com/nextlinux/gosbom/internal/config"
	"github.com/nextlinux/gosbom/internal/log"
	"github.com/nextlinux/gosbom/internal/ui"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"
	"golang.org/x/exp/slices"

	"github.com/anchore/stereoscope"
)

func Run(_ context.Context, app *config.Application, args []string) error {
	err := ValidateOutputOptions(app)
	if err != nil {
		return err
	}

	// could be an image or a directory, with or without a scheme
	// TODO: validate that source is image
	userInput := args[0]
	si, err := source.ParseInputWithNameVersion(userInput, app.Platform, app.SourceName, app.SourceVersion, app.DefaultImagePullSource)
	if err != nil {
		return fmt.Errorf("could not generate source input for packages command: %w", err)
	}

	if si.Scheme != source.ImageScheme {
		return fmt.Errorf("attestations are only supported for oci images at this time")
	}

	eventBus := partybus.NewBus()
	stereoscope.SetBus(eventBus)
	gosbom.SetBus(eventBus)
	subscription := eventBus.Subscribe()

	return eventloop.EventLoop(
		execWorker(app, *si),
		eventloop.SetupSignals(),
		subscription,
		stereoscope.Cleanup,
		ui.Select(options.IsVerbose(app), app.Quiet)...,
	)
}

func buildSBOM(app *config.Application, si source.Input, errs chan error) (*sbom.SBOM, error) {
	src, cleanup, err := source.New(si, app.Registry.ToOptions(), app.Exclusions)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to construct source from user input %q: %w", si.UserInput, err)
	}

	s, err := packages.GenerateSBOM(src, errs, app)
	if err != nil {
		return nil, err
	}

	if s == nil {
		return nil, fmt.Errorf("no SBOM produced for %q", si.UserInput)
	}

	return s, nil
}

//nolint:funlen
func execWorker(app *config.Application, si source.Input) <-chan error {
	errs := make(chan error)
	go func() {
		defer close(errs)
		defer bus.Publish(partybus.Event{Type: event.Exit})

		s, err := buildSBOM(app, si, errs)
		if err != nil {
			errs <- fmt.Errorf("unable to build SBOM: %w", err)
			return
		}

		// note: ValidateOutputOptions ensures that there is no more than one output type
		o := app.Outputs[0]

		f, err := os.CreateTemp("", o)
		if err != nil {
			errs <- fmt.Errorf("unable to create temp file: %w", err)
			return
		}
		defer os.Remove(f.Name())

		writer, err := options.MakeSBOMWriter(app.Outputs, f.Name(), app.OutputTemplatePath)
		if err != nil {
			errs <- fmt.Errorf("unable to create SBOM writer: %w", err)
			return
		}

		if err := writer.Write(*s); err != nil {
			errs <- fmt.Errorf("unable to write SBOM to temp file: %w", err)
			return
		}

		// TODO: what other validation here besides binary name?
		cmd := "cosign"
		if !commandExists(cmd) {
			errs <- fmt.Errorf("unable to find cosign in PATH; make sure you have it installed")
			return
		}

		// Select Cosign predicate type based on defined output type
		// As orientation, check: https://github.com/sigstore/cosign/blob/main/pkg/cosign/attestation/attestation.go
		var predicateType string
		switch strings.ToLower(o) {
		case "cyclonedx-json":
			predicateType = "cyclonedx"
		case "spdx-tag-value", "spdx-tv":
			predicateType = "spdx"
		case "spdx-json", "json":
			predicateType = "spdxjson"
		default:
			predicateType = "custom"
		}

		args := []string{"attest", si.UserInput, "--predicate", f.Name(), "--type", predicateType}
		if app.Attest.Key != "" {
			args = append(args, "--key", app.Attest.Key)
		}

		execCmd := exec.Command(cmd, args...)
		execCmd.Env = os.Environ()
		if app.Attest.Key != "" {
			execCmd.Env = append(execCmd.Env, fmt.Sprintf("COSIGN_PASSWORD=%s", app.Attest.Password))
		} else {
			// no key provided, use cosign's keyless mode
			execCmd.Env = append(execCmd.Env, "COSIGN_EXPERIMENTAL=1")
		}

		log.WithFields("cmd", strings.Join(execCmd.Args, " ")).Trace("creating attestation")

		// bus adapter for ui to hook into stdout via an os pipe
		r, w, err := os.Pipe()
		if err != nil {
			errs <- fmt.Errorf("unable to create os pipe: %w", err)
			return
		}
		defer w.Close()

		mon := progress.NewManual(-1)

		bus.Publish(
			partybus.Event{
				Type: event.AttestationStarted,
				Source: monitor.GenericTask{
					Title: monitor.Title{
						Default:      "Create attestation",
						WhileRunning: "Creating attestation",
						OnSuccess:    "Created attestation",
					},
					Context: "cosign",
				},
				Value: &monitor.ShellProgress{
					Reader: r,
					Manual: mon,
				},
			},
		)

		execCmd.Stdout = w
		execCmd.Stderr = w

		// attest the SBOM
		err = execCmd.Run()
		if err != nil {
			mon.SetError(err)
			errs <- fmt.Errorf("unable to attest SBOM: %w", err)
			return
		}

		mon.SetCompleted()
	}()
	return errs
}

func ValidateOutputOptions(app *config.Application) error {
	err := packages.ValidateOutputOptions(app)
	if err != nil {
		return err
	}

	if len(app.Outputs) > 1 {
		return fmt.Errorf("multiple SBOM format is not supported for attest at this time")
	}

	// cannot use table as default output format when using template output
	if slices.Contains(app.Outputs, table.ID.String()) {
		app.Outputs = []string{gosbomjson.ID.String()}
	}

	return nil
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
