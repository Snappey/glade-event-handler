package main_test

import (
	eventHandler "event-handler"
	"testing"
)

func TestBasicApp(t *testing.T) {

	testApp := eventHandler.Create("test-app").
		AddEvent("TestEvent", "testevent", func(eventData map[string]interface{}, source string) error {
			t.Log("Event ran")
			return nil
		}).
		Build()

	err := testApp.SendEvent(eventHandler.Event{
		Message: `{"event": "TestEvent", "data": {"server_id": "glade-dev"}}`,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestSnsApp(t *testing.T) {

	testApp := eventHandler.Create("sns-app").
		CreateSnsEndpoint(3000).
		AddEvent("SnsEvent", "testevent2", func(eventData map[string]interface{}, source string) error {
			t.Log("Sns Event ran")
			return nil
		}).
		Build()

	err := testApp.SendEvent(eventHandler.Event{
		Message: `{"event": "SnsEvent", "data": {"server_id": "glade-dev"}}`,
	})
	if err != nil {
		t.Error(err)
	}
}
