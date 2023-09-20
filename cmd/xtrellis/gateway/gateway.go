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
// - messages to transmit through the mix-net are sent here
// - clients retrieve messages to transmit from here

// Enable gateway? Used externally to conditionally use the gateway
var Enable bool = false

// Total message size including 8*2 bytes for serialized meta; initialization required
var messageSize int64 = 0

// A message queue for each client is used to store messages in real-time as they are received
// and made available for each client round
var messageQueue = make(map[int]*list.List)

func Init(s int64, enable bool) {
	messageSize = s
	Enable = enable
}

// enqueue data into the queue at the given index
func enqueue(index int, val []byte) {
	// Initialize the queue if it doesn't exist
	if _, exists := messageQueue[index]; !exists {
		messageQueue[index] = list.New()
	}
	// Add to the queue
	messageQueue[index].PushBack(val)
}

// dequeue data from the queue at the given index
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
	maxDataSize := messageSize - 8*2 // 8 bytes for each uint64
	if int64(len(data)) > maxDataSize {
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
	enqueue(int(clientId), message)

	return nil
}

// Output a message from the gateway for a given client
// if no messages in the client's message queue, then use a default
func GetMessageForClient(i *coord.RoundInfo, clientId int64) ([]byte, error) {
	// get next message data queued for this client
	data, err := dequeue(int(clientId))

	// if no message found in queue, use default data
	if err != nil {
		data = []byte("---")
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
	utils.DebugLog("CheckFinalMessages... numExpected=%d", numExpected)
	seen := make(map[uint64]bool)

	// test messages are consecutive integers up to numExpected
	for _, m := range messages {
		c, _, err := messageUnserialize(m)
		if err != nil {
			panic(err)
		}
		utils.DebugLog("... c=%x m=%x", c, m)
		if c < uint64(numExpected) {
			seen[c] = true
		} else {
			return false
		}
	}
	return len(seen) == numExpected
}
