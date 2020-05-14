package mongodb

import (
	"context"

	"github.com/jukylin/esim/log"
	"go.mongodb.org/mongo-driver/event"
)

type spyEvent struct {
	nextEvent MgoEvent

	StartWasCalled bool

	SucceededEventWasCalled bool

	FailedEventWasCalled bool

	logger log.Logger
}

func newSpyEvent(logger log.Logger) MgoEvent {

	spyEvent := &spyEvent{}
	spyEvent.logger = logger

	return spyEvent
}

func (se *spyEvent) NextEvent(event MgoEvent) {
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
