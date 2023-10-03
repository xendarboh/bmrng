package gateway

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/simonlangowski/lightning1/cmd/xtrellis/utils"
	coord "github.com/simonlangowski/lightning1/coordinator/messages"
)

// A gateway:
// - receives messages to transmit through the mix-net for a particular client
// - clients retrieve messages to transmit
// - receives final messages from the mix-net

// Enable gateway? Used externally to conditionally use the gateway
var Enable bool = false

// Total message size including 8*2 bytes for serialized meta; initialization required
var messageSize int64 = 0

func Init(s int64, enable bool) {
	messageSize = s
	Enable = enable
}

// The max size of the data element of a message accounting for protocol data
func GetMaxDataSize() int64 {
	// 8 bytes for each uint64 of serialization protocol
	return messageSize - 8*2
}

// A message queue for each client
type MessageQueue struct {
	items map[int]*list.List
}

func NewMessageQueue() *MessageQueue {
	return &MessageQueue{
		items: make(map[int]*list.List),
	}
}

// Incoming message queue; stores messages as they are received then made available for mix-net xfer each client round
var msgQueueIn = NewMessageQueue()

// Outgoing message queue; stores final messages as they arrive out of the mix-net
var msgQueueOut = NewMessageQueue()

// Enqueue data into the queue at the given index
func (q *MessageQueue) Enqueue(index int, val []byte) {
	// Initialize the queue if it doesn't exist
	if _, exists := q.items[index]; !exists {
		q.items[index] = list.New()
	}
	// Add to the queue
	q.items[index].PushBack(val)
}

// Dequeue data from the queue at the given index
func (q *MessageQueue) Dequeue(index int) ([]byte, error) {
	// pop a message from the queue for a specific client
	if queue, exists := q.items[index]; exists {
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

func messageSerialize(id uint64, data []byte) ([]byte, error) {
	if !(messageSize > 0) {
		return nil, errors.New("Invalid messageSize")
	}

	buffer := new(bytes.Buffer)

	// serialize uint64 for message id
	if err := binary.Write(buffer, binary.LittleEndian, id); err != nil {
		return nil, errors.New(fmt.Sprintf("Error serializing uint64 (id): %s", err))
	}

	// serialize uint64 for data length
	dataLength := uint64(len(data))
	if err := binary.Write(buffer, binary.LittleEndian, dataLength); err != nil {
		return nil, errors.New(fmt.Sprintf("Error serializing uint64 (length): %s", err))
	}

	// serialize message data
	if int64(len(data)) > GetMaxDataSize() {
		return nil, errors.New("data exceeds max size")
	}
	buffer.Write(data)

	// pad to the message size
	padding := make([]byte, messageSize-int64(buffer.Len()))
	buffer.Write(padding)

	return buffer.Bytes(), nil
}

func messageUnserialize(message []byte) (uint64, []byte, error) {
	if !(messageSize > 0) {
		return 0, nil, errors.New("Invalid messageSize")
	}

	buffer := bytes.NewReader(message)

	// deserialize uint64 for message id
	var id uint64
	if err := binary.Read(buffer, binary.LittleEndian, &id); err != nil {
		return 0, nil, errors.New(fmt.Sprintf("Error deserializing uint64 (id): %s", err))
	}

	// deserialize uint64 for data lengthd
	var dataLength uint64
	if err := binary.Read(buffer, binary.LittleEndian, &dataLength); err != nil {
		return 0, nil, errors.New(fmt.Sprintf("Error deserializing uint64 (length): %s", err))
	}

	// deserialize data
	data := make([]byte, dataLength)
	buffer.Read(data)

	return id, data, nil
}

// Input a message into the gateway for a client
func PutMessageForClient(clientId int64, message []byte) error {
	msgQueueIn.Enqueue(int(clientId), message)

	return nil
}

// Output a message from the gateway for a given client.
// If no messages in the client's message queue, then use a default.
func GetMessageForClient(i *coord.RoundInfo, clientId int64) ([]byte, error) {
	// get next message data queued for this client
	data, err := msgQueueIn.Dequeue(int(clientId))

	// if no message found in queue, use default data
	if err != nil {
		data = []byte("")
	}

	// use clientId as message id
	id := uint64(clientId)

	// serialize message
	message, err := messageSerialize(id, data)
	if err != nil {
		panic(err)
	}

	utils.DebugLog("data=%x, message=%x", data, message)

	return message, err
}

// Send final messages from the mix-net here.
// This replaces coordinator.Check testing for message ids within serialization.
// Note: There are duplicates, message ids are used to sort unique.
func CheckFinalMessages(messages [][]byte, numExpected int) bool {
	messageData := make(map[uint64][]byte)

	// test messages are consecutive integers up to numExpected
	for _, m := range messages {
		id, data, err := messageUnserialize(m)
		if err != nil {
			panic(err)
		}
		if id < uint64(numExpected) {
			messageData[id] = data
		} else {
			return false
		}
	}

	// for each final message
	for i, s := range messageData {
		// if the message data has length, then it was fed to the client
		if len(s) > 0 {
			// add the data to the client's out queue
			msgQueueOut.Enqueue(int(i), s)
		}
		utils.DebugLog("messageData[%d] = '%x'", i, s)
	}

	return len(messageData) == numExpected
}
