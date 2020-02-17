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


func (this *spyEvent) NextEvent(event MonitorEvent) {
	this.nextEvent = event
}

func (this *spyEvent) EventName() string {
	return "spy_proxy"
}

func (this *spyEvent) Start(ctx context.Context, starEv *event.CommandStartedEvent) {
	this.StartWasCalled = true
	this.logger.Infof("StartWasCalled")
	if this.nextEvent != nil {
		this.nextEvent.Start(ctx, starEv)
	}
}


func (this *spyEvent) SucceededEvent(ctx context.Context,
	succEvent *event.CommandSucceededEvent) {
	this.logger.Infof("SucceededEvent")
	this.SucceededEventWasCalled = true
	if this.nextEvent != nil {
		this.nextEvent.SucceededEvent(ctx, succEvent)
	}

}


func (this *spyEvent) FailedEvent(ctx context.Context, failedEvent *event.CommandFailedEvent) {
	this.logger.Infof("FailedEvent")
	this.FailedEventWasCalled = true
	if this.nextEvent != nil {
		this.nextEvent.FailedEvent(ctx, failedEvent)
	}
}
