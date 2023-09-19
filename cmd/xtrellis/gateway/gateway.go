package gateway

import (
	"container/list"
	"encoding/binary"
	"errors"

	coord "github.com/simonlangowski/lightning1/coordinator/messages"
)

// A gateway:
// - messages to transmit through the mix-net are sent here
// - clients retrieve messages to transmit from here

// A message queue for each client is used to store messages in real-time as they are received
// and made available for each client round
var messageQueue = make(map[int]*list.List)

func enqueue(index int, val []byte) {
	// Initialize the queue if it doesn't exist
	if _, exists := messageQueue[index]; !exists {
		messageQueue[index] = list.New()
	}
	// Add to the queue
	messageQueue[index].PushBack(val)
}

func dequeue(index int) ([]byte, error) {
	// pop a message from the queue for a specific client
	if queue, exists := messageQueue[index]; exists {
		if element := queue.Front(); element != nil {
			// Pop the front element from the queue
			message := element.Value.([]byte)
			queue.Remove(element)
			return message, nil
		} else {
			return nil, errors.New("queue is empty for the index")
		}
	} else {
		return nil, errors.New("index does not have a queue")
	}
}

func PutMessageForClient(clientId int64, message []byte) error {
	enqueue(int(clientId), message)

	return nil
}

func GetMessageForClient(i *coord.RoundInfo, clientId int64) ([]byte, error) {
	// get next message queued for this client
	message, err := dequeue(int(clientId))

	// if no message found in queue
	if err != nil {
		// use clientId as default message
		message = make([]byte, i.MessageSize)
		binary.LittleEndian.PutUint64(message, uint64(clientId))
	}

	return message, err
}
