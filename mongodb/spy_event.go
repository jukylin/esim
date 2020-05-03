package mongodb

import (
	"context"

	"github.com/jukylin/esim/log"
	"go.mongodb.org/mongo-driver/event"
)

type spyEvent struct {
	nextEvent MonitorEvent

	StartWasCalled bool

	SucceededEventWasCalled bool

	FailedEventWasCalled bool

	logger log.Logger
}

func NewSpyEvent(logger log.Logger) *spyEvent {

	spyEvent := &spyEvent{}
	spyEvent.logger = logger

	return spyEvent
}

func (se *spyEvent) NextEvent(event MonitorEvent) {
	se.nextEvent = event
}

func (se *spyEvent) EventName() string {
	return "spy_proxy"
}

func (se *spyEvent) Start(ctx context.Context, starEv *event.CommandStartedEvent) {
	se.StartWasCalled = true
	se.logger.Infof("StartWasCalled")
	if se.nextEvent != nil {
		se.nextEvent.Start(ctx, starEv)
	}
}

func (se *spyEvent) SucceededEvent(ctx context.Context,
	succEvent *event.CommandSucceededEvent) {
	se.logger.Infof("SucceededEvent")
	se.SucceededEventWasCalled = true
	if se.nextEvent != nil {
		se.nextEvent.SucceededEvent(ctx, succEvent)
	}

}

func (se *spyEvent) FailedEvent(ctx context.Context, failedEvent *event.CommandFailedEvent) {
	se.logger.Infof("FailedEvent")
	se.FailedEventWasCalled = true
	if se.nextEvent != nil {
		se.nextEvent.FailedEvent(ctx, failedEvent)
	}
}
