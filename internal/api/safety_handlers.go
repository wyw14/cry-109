package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func (s *Server) safetyStatus(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, s.runtime.SafetyStatus(s.runtime.Clock.Now()))
}

func (s *Server) triggerStorm(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		WindMPS         float64 `json:"wind_mps"`
		EvaluateAfterMS int64   `json:"evaluate_after_ms"`
	}
	if err := readJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	value, err := s.runtime.TriggerStorm(input.WindMPS)
	if err != nil {
		writeError(writer, http.StatusConflict, err.Error())
		return
	}
	if input.EvaluateAfterMS > 0 {
		value.Anchor = s.runtime.AdvanceStorm(s.runtime.Clock.Now().Add(time.Duration(input.EvaluateAfterMS) * time.Millisecond))
	}
	writeJSON(writer, http.StatusOK, value)
}

func (s *Server) secureAnchor(writer http.ResponseWriter, _ *http.Request) {
	if err := s.runtime.Storm.Sequence().Secure(); err != nil {
		writeError(writer, http.StatusConflict, err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, s.runtime.SafetyStatus(s.runtime.Clock.Now()))
}

func (s *Server) applyRegen(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		PowerKW       float64 `json:"power_kw"`
		SinkAvailable bool    `json:"sink_available"`
	}
	if err := readJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	value, err := s.runtime.ApplyRegen(input.PowerKW, input.SinkAvailable)
	if err != nil {
		writeError(writer, http.StatusConflict, err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, value)
}

func (s *Server) changeReeving(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Factor         int     `json:"factor"`
		DrumMeters     float64 `json:"drum_meters"`
		RequiredHeight float64 `json:"required_height"`
	}
	if err := readJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	value, err := s.runtime.ChangeReeving(input.Factor, input.DrumMeters, input.RequiredHeight)
	if err != nil {
		writeJSON(writer, http.StatusConflict, map[string]any{"error": err.Error(), "decision": value})
		return
	}
	writeJSON(writer, http.StatusOK, value)
}

func (s *Server) calibrateLaser(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		MeasuredHeight float64 `json:"measured_height"`
		DrumMeters     float64 `json:"drum_meters"`
	}
	if err := readJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.runtime.Height.CalibrateLaser(input.MeasuredHeight, input.DrumMeters); err != nil {
		writeError(writer, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, s.runtime.Clearance.Check(input.DrumMeters, 0))
}

func (s *Server) raiseIncident(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Kind      string `json:"kind"`
		Component string `json:"component"`
		Message   string `json:"message"`
	}
	if err := readJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(writer, http.StatusCreated, s.runtime.RecordIncident(input.Kind, input.Component, input.Message))
}

func (s *Server) resolveIncident(writer http.ResponseWriter, request *http.Request) {
	if !s.runtime.Incidents.Resolve(chi.URLParam(request, "incidentID")) {
		writeError(writer, http.StatusNotFound, "incident not found")
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"resolved": true})
}
