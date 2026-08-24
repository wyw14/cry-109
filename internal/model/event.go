package model

import (
	"encoding/json"
	"time"
)

type Event struct {
	ID          string          `json:"id"`
	AggregateID string          `json:"aggregate_id"`
	Kind        string          `json:"kind"`
	OccurredAt  time.Time       `json:"occurred_at"`
	Data        json.RawMessage `json:"data"`
}

func NewEvent(id, aggregateID, kind string, at time.Time, value any) (Event, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return Event{}, err
	}
	return Event{ID: id, AggregateID: aggregateID, Kind: kind, OccurredAt: at, Data: data}, nil
}
