package main

import (
	"event-handler/handler"
	eventSns "event-handler/sns"
	"fmt"
	"runtime"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/gin-gonic/gin"
)

type EventHandlerOutput interface {
	SendEvent(event handler.RawEvent) error
	Run() error
}

type EventHandler struct {
	EventHandlerOutput
	AppName   string
	snsConfig *SnsConfig
	handler   *handler.Handler
}

type SnsConfig struct {
	TopicArn        string
	SubscriptionArn string
	Endpoint        string
	Port            int
	Router          *gin.Engine
}

func main() {

	/*sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))*/

	testApp := Create("test-app").
		CreateSnsEndpointWithVerifiy(3000, true).
		AddEvent("ServerCreated", "test-create", func(eventData map[string]interface{}, source string) error {
			fmt.Printf("testing\n")
			return nil
		}).
		Build()

	err := testApp.SendEvent(handler.RawEvent{
		Timestamp: time.Now(),
		Message:   `{"event": "ServerCreated", "data": {"server_id": "glade-dev"}}`,
	})
	if err != nil {
		panic(err)
	}
}

func Create(appName string) *EventHandler {
	return &EventHandler{
		AppName:   appName,
		handler:   handler.Create(),
		snsConfig: &SnsConfig{},
	}
}

func (e *EventHandler) CreateSnsEndpoint(port int) *EventHandler {
	return e.CreateSnsEndpointWithVerifiy(port, false)
}

func (e *EventHandler) CreateSnsEndpointWithVerifiy(port int, verify bool) *EventHandler {
	httpRouter := eventSns.CreateRouter(port, verify, e.handler)
	go httpRouter.Run(fmt.Sprintf(":%d", port))

	e.snsConfig.Port = port

	return e
}

func (e *EventHandler) SnsSubscribe(host string, topicARN string, sess *session.Session) *EventHandler {
	endpoint := fmt.Sprintf("%s:%d/events/sns", host, e.snsConfig.Port)
	svc := sns.New(sess)
	result, err := svc.Subscribe(&sns.SubscribeInput{
		Protocol:              aws.String("HTTPS"),
		ReturnSubscriptionArn: aws.Bool(true),
		TopicArn:              aws.String(topicARN),
		Endpoint:              aws.String(endpoint),
	})
	if err != nil {
		panic(err)
	}

	e.snsConfig.TopicArn = topicARN
	e.snsConfig.SubscriptionArn = *result.SubscriptionArn
	e.snsConfig.Endpoint = endpoint

	return e
}

func (e *EventHandler) AddEvent(eventName string, callbackId string, callback func(eventData map[string]interface{}, source string) error) *EventHandler {
	e.handler.AddEvent(eventName, callbackId, callback)
	return e
}

func (e *EventHandler) Build() EventHandlerOutput {
	return e
}

func (e *EventHandler) SendEvent(event handler.RawEvent) error {
	return e.handler.RouteEvent(event, "local")
}

func (e *EventHandler) Run() error {
	for {
		runtime.Gosched()
	}
}
