package sns

import (
	"errors"
	"event-handler/handler"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type SnsEvent struct {
	Type             string    `json:"Type"`
	MessageId        string    `json:"MessageId"`
	Token            string    `json:"Token"`
	TopicArn         string    `json:"TopicArn"`
	Message          string    `json:"Message"`
	SubscribeURL     string    `json:"SubscribeURL"`
	Timestamp        time.Time `json:"Timestamp"`
	SignatureVersion string    `json:"SignatureVersion"`
	Signature        string    `json:"Signature"`
	SigningCertURL   string    `json:"SigningCertURL"`
}

func fetchEventHeader(header http.Header) string {
	for name, value := range header {
		if strings.ToLower(name) == "x-amz-sns-message-type" {
			return value[0]
		}
	}
	return ""
}

func processNotification(event SnsEvent, eventHandler *handler.Handler) error {
	return eventHandler.RouteEvent(handler.RawEvent{
		Message:   event.Message,
		Timestamp: event.Timestamp,
	}, "sns")
}

func confirmSubscription(event SnsEvent) error {
	if event.SubscribeURL != "" {
		return errors.New("invalid subscribe url")
	}

	resp, err := http.Get(event.SubscribeURL)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code from sns: %d expected 200", resp.StatusCode)
	}

	return nil
}

func confirmUnsubscription(event SnsEvent) error {
	fmt.Printf("sns event: %s\n", event.Type) // TODO: Implement
	return nil
}

func handleEvent(eventType string, event SnsEvent, eventHandler *handler.Handler) error {
	switch eventType {
	case "SubscriptionConfirmation":
		return confirmSubscription(event)
	case "Notification":
		return processNotification(event, eventHandler)
	case "UnsubscribeConfirmation":
		return confirmUnsubscription(event)
	default:
		return errors.New("unhandled event type")
	}
}

func event(eventHandler *handler.Handler, verify bool) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var event SnsEvent
		if err := c.ShouldBindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if verify {
			err := verifySignature(event)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "signature mismatch"})
				return
			}
		}

		eventType := fetchEventHeader(c.Request.Header)
		err := handleEvent(eventType, event, eventHandler)
		if err != nil {
			fmt.Printf("error handling sns event: %s %s\n", event.Type, err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to handle event"})
		} else {
			c.JSON(http.StatusOK, gin.H{})
		}
	}

	return gin.HandlerFunc(fn)
}
