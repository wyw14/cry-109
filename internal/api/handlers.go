package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/wyw14/cry-109/internal/control"
	"github.com/wyw14/cry-109/internal/model"
)

func (s *Server) health(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{"status": "ok", "service": "portcrane"})
}

func (s *Server) ready(writer http.ResponseWriter, _ *http.Request) {
	diagnostics := s.runtime.Diagnostics()
	status := http.StatusOK
	if !diagnostics.Ready {
		status = http.StatusServiceUnavailable
	}
	writeJSON(writer, status, diagnostics)
}

func (s *Server) diagnostics(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, s.runtime.Diagnostics())
}

func (s *Server) snapshot(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, s.runtime.Snapshot())
}

func (s *Server) listLifts(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{"lifts": s.runtime.Lifts.List()})
}

func (s *Server) startLift(writer http.ResponseWriter, request *http.Request) {
	var input control.StartLiftRequest
	if err := readJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	value, err := s.runtime.StartLift(input)
	if err != nil {
		writeError(writer, http.StatusConflict, err.Error())
		return
	}
	writeJSON(writer, http.StatusCreated, value)
}

func (s *Server) planTrolley(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		LiftID        string  `json:"lift_id"`
		DistanceM     float64 `json:"distance_m"`
		MaxSpeedMPS   float64 `json:"max_speed_mps"`
		RopeLengthM   float64 `json:"rope_length_m"`
		PayloadTonnes float64 `json:"payload_tonnes"`
		Replan        bool    `json:"replan"`
	}
	if err := readJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	var value any
	var err error
	if input.Replan {
		value, err = s.runtime.ReplanTrolley(input.LiftID, input.DistanceM, input.MaxSpeedMPS, input.RopeLengthM, input.PayloadTonnes)
	} else {
		value, err = s.runtime.PlanTrolley(input.LiftID, input.DistanceM, input.MaxSpeedMPS, input.RopeLengthM, input.PayloadTonnes)
	}
	if err != nil {
		writeError(writer, http.StatusConflict, err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, value)
}

func (s *Server) observeLanding(writer http.ResponseWriter, request *http.Request) {
	var input control.LandingRequest
	if err := readJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	release := s.runtime.PrimeLanding(input.SessionID)
	result, err := s.runtime.ObserveLanding(release, input)
	if err != nil {
		writeError(writer, http.StatusConflict, err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, result)
}

func (s *Server) spreaderStatus(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{
		"geometry":      s.runtime.Spreader.Telescope().Geometry(),
		"engage_state":  s.runtime.Spreader.EngageController().State(),
		"lock_proof":    s.runtime.Spreader.EngageController().Proof(),
		"cached_models": s.runtime.Models.Size(),
	})
}

func (s *Server) telescope(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		LengthFeet int     `json:"length_feet"`
		MassTonnes float64 `json:"mass_tonnes"`
	}
	if err := readJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	geometry, err := s.runtime.Spreader.Telescope().Extend(input.LengthFeet, input.MassTonnes)
	if err != nil {
		writeError(writer, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, geometry)
}

func (s *Server) motionStatus(writer http.ResponseWriter, request *http.Request) {
	destination, err := queryFloat(request, "destination", 320)
	if err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	status, err := s.runtime.MotionStatus(destination)
	if err != nil {
		writeError(writer, http.StatusConflict, err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, status)
}

func (s *Server) railOrigin(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Origin      float64 `json:"origin"`
		Destination float64 `json:"destination"`
	}
	if err := readJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	status, err := s.runtime.RecalibrateRail(input.Origin, input.Destination)
	if err != nil {
		writeError(writer, http.StatusConflict, err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, status)
}

func (s *Server) vesselDraft(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		DraftMeters float64 `json:"draft_meters"`
		DeckMeters  float64 `json:"deck_meters"`
		HatchMeters float64 `json:"hatch_meters"`
	}
	if err := readJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	profile, err := s.runtime.UpdateVesselDraft(input.DraftMeters, input.DeckMeters, input.HatchMeters)
	if err != nil {
		writeError(writer, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, profile)
}

func (s *Server) balanceTandem(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Cycle   uint64  `json:"cycle"`
		LeftKN  float64 `json:"left_kn"`
		RightKN float64 `json:"right_kn"`
		Disable bool    `json:"disable"`
	}
	if err := readJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	if input.Disable {
		s.runtime.Tandem.Disable()
		writeJSON(writer, http.StatusOK, map[string]any{"enabled": false, "reason": "operator safe degradation"})
		return
	}
	value, err := s.runtime.BalanceTandem(input.Cycle, input.LeftKN, input.RightKN)
	if err != nil {
		writeError(writer, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, value)
}

func (s *Server) applyHeave(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		HeaveCapturedAt time.Time `json:"heave_captured_at"`
		HoistCapturedAt time.Time `json:"hoist_captured_at"`
		HeaveMeters     float64   `json:"heave_meters"`
		HeaveVelocity   float64   `json:"heave_velocity"`
		HookMeters      float64   `json:"hook_meters"`
		HoistVelocity   float64   `json:"hoist_velocity"`
	}
	if err := readJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	if input.HeaveCapturedAt.IsZero() {
		input.HeaveCapturedAt = s.runtime.Clock.Now()
	}
	if input.HoistCapturedAt.IsZero() {
		input.HoistCapturedAt = s.runtime.Clock.Now()
	}
	command := s.runtime.ApplyHeave(
		model.HeaveSample{CapturedAt: input.HeaveCapturedAt, Meters: input.HeaveMeters, Velocity: input.HeaveVelocity},
		model.HoistSample{CapturedAt: input.HoistCapturedAt, HookMeters: input.HookMeters, Velocity: input.HoistVelocity},
	)
	writeJSON(writer, http.StatusOK, command)
}

func (s *Server) stopTrolley(writer http.ResponseWriter, _ *http.Request) {
	s.runtime.Trolley.Stop()
	writeJSON(writer, http.StatusOK, map[string]any{"stopped": true})
}

func queryFloat(request *http.Request, key string, fallback float64) (float64, error) {
	raw := request.URL.Query().Get(key)
	if raw == "" {
		return fallback, nil
	}
	return strconv.ParseFloat(raw, 64)
}
