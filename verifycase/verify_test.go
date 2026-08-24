package verifycase

import (
	"testing"
	"time"

	"github.com/wyw14/cry-109/internal/lift"
	"github.com/wyw14/cry-109/internal/model"
	"github.com/wyw14/cry-109/internal/spreader"
)

func TestLoadTransferRequiresFourCurrentTwistlocks(t *testing.T) {
	session := "engage-current"
	controller := spreader.NewEngageController(session)
	controller.Begin(session)
	for _, corner := range []model.Corner{model.FrontLeft, model.RearRight} {
		controller.Confirm(model.TwistlockAck{SessionID: session, Corner: corner, Locked: true, At: time.Now()})
	}
	permit := lift.NewLoadPermit(session)
	if permit.Evaluate(controller.Proof()) {
		t.Fatal("load transfer was permitted with only a diagonal pair locked")
	}
	controller.Confirm(model.TwistlockAck{SessionID: "engage-old", Corner: model.FrontRight, Locked: true, At: time.Now()})
	if permit.Evaluate(controller.Proof()) {
		t.Fatal("an old-session acknowledgement completed the current proof")
	}
	for _, corner := range []model.Corner{model.FrontRight, model.RearLeft} {
		controller.Confirm(model.TwistlockAck{SessionID: session, Corner: corner, Locked: true, At: time.Now()})
	}
	if !permit.Evaluate(controller.Proof()) {
		_, reason := permit.Status()
		t.Fatalf("four current-session locks did not permit transfer: %s", reason)
	}
}
