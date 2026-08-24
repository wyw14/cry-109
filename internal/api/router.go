package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/wyw14/cry-109/internal/control"
)

type Server struct {
	runtime *control.Runtime
	router  chi.Router
}

func NewServer(runtime *control.Runtime) *Server {
	server := &Server{runtime: runtime}
	router := chi.NewRouter()
	router.Use(server.recoverer, server.requestLog)
	router.Get("/healthz", server.health)
	router.Get("/readyz", server.ready)
	router.Get("/lifts", server.page("lifts"))
	router.Get("/spreader", server.page("spreader"))
	router.Get("/motion", server.page("motion"))
	router.Get("/safety", server.page("safety"))
	router.Route("/api", func(api chi.Router) {
		api.Get("/snapshot", server.snapshot)
		api.Get("/diagnostics", server.diagnostics)
		api.Get("/lifts", server.listLifts)
		api.Post("/lifts", server.startLift)
		api.Post("/lifts/trajectory", server.planTrolley)
		api.Post("/lifts/landing", server.observeLanding)
		api.Get("/spreader", server.spreaderStatus)
		api.Post("/spreader/telescope", server.telescope)
		api.Get("/motion", server.motionStatus)
		api.Post("/motion/rail-origin", server.railOrigin)
		api.Post("/motion/vessel-draft", server.vesselDraft)
		api.Post("/motion/tandem", server.balanceTandem)
		api.Post("/motion/heave", server.applyHeave)
		api.Post("/motion/stop", server.stopTrolley)
		api.Get("/safety", server.safetyStatus)
		api.Post("/safety/storm", server.triggerStorm)
		api.Post("/safety/anchor/secure", server.secureAnchor)
		api.Post("/safety/regen", server.applyRegen)
		api.Post("/safety/reeving", server.changeReeving)
		api.Post("/safety/laser", server.calibrateLaser)
		api.Post("/safety/incidents", server.raiseIncident)
		api.Post("/safety/incidents/{incidentID}/resolve", server.resolveIncident)
	})
	server.router = router
	return server
}

func (s *Server) Handler() http.Handler { return s.router }
