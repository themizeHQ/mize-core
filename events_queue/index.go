package eventsqueue

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"go.uber.org/zap"
	"mize.app/logger"
)

var ProducerClient *azeventhubs.ProducerClient

func ConnectToEventStream() {
	var err error
	// create an Event Hubs producer client using a connection string to the namespace and the event hub
	ProducerClient, err = azeventhubs.NewProducerClientFromConnectionString(os.Getenv("AZURE_EVENT_HUB_CONNECTION_STRING"), "", nil)
	if err != nil {
		panic(err)
	}
}

func CloseProducerClient() {
	ProducerClient.Close(context.TODO())
}

func createEvent(topic EventTopic, payload map[string]interface{}) (*azeventhubs.EventData, error) {
	data, err := json.Marshal(
		Event{
			Topic:   topic,
			Payload: payload,
		})
	if err != nil {
		logger.Error(errors.New("events_queue - could not transform event payload to json string"), zap.Error(err))
		return nil, err
	}
	event := azeventhubs.EventData{
		Body: data,
	}
	return &event, nil
}

func emitEvent(event *azeventhubs.EventData) error {
	// create a batch object and add sample events to the batch
	newBatchOptions := &azeventhubs.EventDataBatchOptions{}

	batch, err := ProducerClient.NewEventDataBatch(context.TODO(), newBatchOptions)
	if err != nil {
		logger.Error(errors.New("eventsqueue - failed to create event databatch"), zap.Error(err), zap.Any("batch", batch), zap.Any("event", event))
		return err
	}

	err = batch.AddEventData(event, nil)
	if err != nil {
		logger.Error(errors.New("eventsqueue - failed to add event to batch"), zap.Error(err), zap.Any("batch", batch), zap.Any("event", event))
		return err
	}

	// send the batch of events to the event hub
	err = ProducerClient.SendEventDataBatch(context.TODO(), batch, nil)
	if err != nil {
		logger.Error(errors.New("eventsqueue - failed to send event data batch"), zap.Error(err), zap.Any("batch", batch))
		return err
	}
	return nil
}

func CreateAndEmitEvent(topic EventTopic, payload map[string]interface{}) bool {
	event, err := createEvent(topic, payload)
	if err != nil {
		return false
	}
	err = emitEvent(event)
	return err == nil
}
