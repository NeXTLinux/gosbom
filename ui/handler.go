/*
Package ui provides all public UI elements intended to be repurposed in other applications. Specifically, a single
Handler object is provided to allow consuming applications (such as grype) to check if there are UI elements the handler
can respond to (given a specific event type) and handle the event in context of the given screen frame object.
*/
package ui

import (
	"context"
	"sync"

	gosbomEvent "github.com/nextlinux/gosbom/gosbom/event"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/jotframe/pkg/frame"

	stereoscopeEvent "github.com/anchore/stereoscope/pkg/event"
)

// Handler is an aggregated event handler for the set of supported events (PullDockerImage, ReadImage, FetchImage, PackageCatalogerStarted)
type Handler struct {
}

// NewHandler returns an empty Handler
func NewHandler() *Handler {
	return &Handler{}
}

// RespondsTo indicates if the handler is capable of handling the given event.
func (r *Handler) RespondsTo(event partybus.Event) bool {
	switch event.Type {
	case stereoscopeEvent.PullDockerImage,
		stereoscopeEvent.ReadImage,
		stereoscopeEvent.FetchImage,
		gosbomEvent.PackageCatalogerStarted,
		gosbomEvent.SecretsCatalogerStarted,
		gosbomEvent.FileDigestsCatalogerStarted,
		gosbomEvent.FileMetadataCatalogerStarted,
		gosbomEvent.FileIndexingStarted,
		gosbomEvent.ImportStarted,
		gosbomEvent.AttestationStarted,
		gosbomEvent.CatalogerTaskStarted:
		return true
	default:
		return false
	}
}

// Handle calls the specific event handler for the given event within the context of the screen frame.
func (r *Handler) Handle(ctx context.Context, fr *frame.Frame, event partybus.Event, wg *sync.WaitGroup) error {
	switch event.Type {
	case stereoscopeEvent.PullDockerImage:
		return PullDockerImageHandler(ctx, fr, event, wg)

	case stereoscopeEvent.ReadImage:
		return ReadImageHandler(ctx, fr, event, wg)

	case stereoscopeEvent.FetchImage:
		return FetchImageHandler(ctx, fr, event, wg)

	case gosbomEvent.PackageCatalogerStarted:
		return PackageCatalogerStartedHandler(ctx, fr, event, wg)

	case gosbomEvent.SecretsCatalogerStarted:
		return SecretsCatalogerStartedHandler(ctx, fr, event, wg)

	case gosbomEvent.FileDigestsCatalogerStarted:
		return FileDigestsCatalogerStartedHandler(ctx, fr, event, wg)

	case gosbomEvent.FileMetadataCatalogerStarted:
		return FileMetadataCatalogerStartedHandler(ctx, fr, event, wg)

	case gosbomEvent.FileIndexingStarted:
		return FileIndexingStartedHandler(ctx, fr, event, wg)

	case gosbomEvent.ImportStarted:
		return ImportStartedHandler(ctx, fr, event, wg)

	case gosbomEvent.AttestationStarted:
		return AttestationStartedHandler(ctx, fr, event, wg)

	case gosbomEvent.CatalogerTaskStarted:
		return CatalogerTaskStartedHandler(ctx, fr, event, wg)
	}
	return nil
}
