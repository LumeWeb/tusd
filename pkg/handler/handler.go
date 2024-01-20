package handler

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// Handler is a ready to use handler with routing (using pat)
type Handler struct {
	*UnroutedHandler
	http.Handler
}

// NewHandler creates a routed tus protocol handler. This is the simplest
// way to use tusd but may not be as configurable as you require. If you are
// integrating this into an existing app you may like to use tusd.NewUnroutedHandler
// instead. Using tusd.NewUnroutedHandler allows the tus handlers to be combined into
// your existing router (aka mux) directly. It also allows the GET and DELETE
// endpoints to be customized. These are not part of the protocol so can be
// changed depending on your needs.
func NewHandler(config Config) (*Handler, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}

	handler, err := NewUnroutedHandler(config)
	if err != nil {
		return nil, err
	}

	routedHandler := &Handler{
		UnroutedHandler: handler,
	}

	mux := httprouter.New()
	mux.RedirectTrailingSlash = false

	routedHandler.Handler = handler.Middleware(mux)

	mux.POST("/", handler.PostFile)
	mux.HEAD("/:id", handler.HeadFile)
	mux.PATCH("/:id", handler.PatchFile)
	if !config.DisableDownload {
		mux.GET("/:id", handler.GetFile)
	}

	// Only attach the DELETE handler if the Terminate() method is provided
	if config.StoreComposer.UsesTerminater && !config.DisableTermination {
		mux.DELETE("/:id", handler.DelFile)
	}

	return routedHandler, nil
}
