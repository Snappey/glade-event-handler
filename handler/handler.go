package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Handler struct {
	Callbacks map[string][]Callback
}

type Callback struct {
	Id   string
	Func func(eventData map[string]interface{}, source string) error
}

type RawEvent struct {
	Message   string
	Timestamp time.Time
}

func Create() *Handler {
	handler := Handler{
		Callbacks: map[string][]Callback{},
	}

	return &handler
}

func (h *Handler) AddEvent(eventName string, callbackId string, callback func(eventData map[string]interface{}, source string) error) *Handler {
	event, eventExists := h.Callbacks[eventName]
	if !eventExists {
		event = make([]Callback, 0)
	}

	h.Callbacks[eventName] = append(event, Callback{
		Id:   callbackId,
		Func: callback,
	})

	return h
}

func (h *Handler) RemoveEvent(eventName string, callbackId string) error {
	_, eventExists := h.Callbacks[eventName]
	if !eventExists {
		return errors.New("event does not exist")
	}

	var idxs []int
	for i, v := range h.Callbacks[eventName] {
		if v.Id == callbackId {
			idxs = append(idxs, i)
		}
	}

	for _, v := range idxs {
		h.Callbacks[eventName] = append(h.Callbacks[eventName][:v], h.Callbacks[eventName][v+1:]...)
	}

	return nil
}

func (h *Handler) RouteEvent(event RawEvent, source string) error {
	var jsonMessage map[string]interface{}
	err := json.Unmarshal([]byte(event.Message), &jsonMessage)
	if err != nil {
		return err
	}

	eventKey, hasEvent := jsonMessage["event"]
	if !hasEvent {
		return errors.New("event field not defined")
	}

	// Route event to the callback function in the user defined callbacks
	for _, c := range h.Callbacks[fmt.Sprintf("%s", eventKey)] {
		err := c.Func(jsonMessage, source)
		if err != nil {
			fmt.Printf("error executing callback '%s': %s\n", c.Id, err.Error())
		}
	}

	return nil
}
