package gateway

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"strconv"
	"strings"

	gatewayv1 "github.com/31333337/repo/pb/gen/proto/go/gateway/v1"
	"github.com/simonlangowski/lightning1/cmd/xtrellis/utils"
	coord "github.com/simonlangowski/lightning1/coordinator/messages"
	"google.golang.org/protobuf/proto"
)

// A gateway:
// - receives messages to transmit through the mix-net for a particular client
// - enables mix-net clients to retrieve messages to transmit
// - receives final messages from the mix-net

// Enable gateway? Used externally to conditionally use the gateway
var Enable bool = false

// mix-net message size; initialization required
// Total message size including 8*2 bytes for serialized meta
var messageSize int64 = 0

// packet protocol size; calculated and cached by GetMaxProtocolSize
var protocolSize int64 = 0

// filesystem directory to save final messages sent through the mix-net
var messageDirectory string

// proxy server address and port, listens for messages
var proxyAddress string

func Init(s int64, enable bool, addr string, dir string) {
	messageSize = s
	Enable = enable
	proxyAddress = addr
	messageDirectory = dir

	if !enable {
		return
	}

	// create the message directory if not exists
	err := os.MkdirAll(messageDirectory, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("[Gateway] Error creating message directory: %v\n", err))
	}

	go proxyStart()
}

// Max gateway packet protocol size in bytes, not counting packet data
// This is a good estimate as protobuf serialized message size are inherently variant
// The calculated value is cached
func GetMaxProtocolSize() int64 {
	if protocolSize > 0 {
		// return cached result
		return protocolSize
	}

	// Create a packet with max data sizes, marshal it, then measure the length minus data size.
	// For message wire type sizes, refer to https://protobuf.dev/programming-guides/encoding/#bools-and-enums

	packet := &gatewayv1.Packet{
		Type:     gatewayv1.PacketType_PACKET_TYPE_DUMMY,
		StreamId: math.MaxUint64,
		Sequence: math.MaxUint64,
		Length:   math.MaxUint32,
		Data:     []byte("----"), // not nil
	}

	packed, err := proto.Marshal(packet)
	if err != nil {
		panic(err)
	}

	protocolSize = int64(len(packed) - len(packet.Data))
	return protocolSize
}

// The max size of message packet data accounting for protocol size
func GetMaxDataSize() int64 {
	return messageSize - GetMaxProtocolSize()
}

////////////////////////////////////////////////////////////////////////
// Message Queue
// - each client has two message queues for In and Out
////////////////////////////////////////////////////////////////////////

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

////////////////////////////////////////////////////////////////////////
// Packets & Message Serialization
// - a protocol to transmit data streams as packetized mix-net messages
////////////////////////////////////////////////////////////////////////

// Get space in bytes for the packed packet to meet message size
func getPacketSpace(packet *gatewayv1.Packet) (int64, error) {
	packed, err := proto.Marshal(packet)
	if err != nil {
		return 0, err
	}

	// bytes of space to fill
	space := messageSize - int64(len(packed))

	// if packet data is empty
	if space > 0 && bytes.Equal(packet.Data, []byte("")) {
		// account for 2 protocol bytes when data is not empty
		space -= 2
	}

	return space, nil
}

// Pack a packet by filling with space to meet the target message size
func packetPack(packet *gatewayv1.Packet) ([]byte, error) {
	space, err := getPacketSpace(packet)
	if err != nil {
		return nil, err
	}

	if space < 0 {
		return nil, errors.New("packet too large for message size")
	}

	if space > 0 {
		// fill the space used by the data,
		// the packet length enables decoding
		buffer := new(bytes.Buffer)
		buffer.Write(packet.Data)
		buffer.Write(make([]byte, space))
		packet.Data = buffer.Bytes()
	}

	packed, err := proto.Marshal(packet)
	if err != nil {
		return nil, err
	}

	return packed, nil
}

func packetUnpack(packedPacket []byte) (*gatewayv1.Packet, error) {
	packet := &gatewayv1.Packet{}
	err := proto.Unmarshal(packedPacket, packet)
	if err != nil {
		return nil, err
	}

	// deserialize padded packet data based on packet length
	buffer := bytes.NewReader(packet.Data)
	data := make([]byte, packet.Length)
	buffer.Read(data)
	packet.Data = data

	return packet, nil
}

////////////////////////////////////////////////////////////////////////
// Trellis mix-net hooks
// - called within mix-net simulator
// - round start: clients get messages from their In queue
// - round end: final messages are placed in the clients Out queue
////////////////////////////////////////////////////////////////////////

// Output a message from the gateway for a given client.
// If no messages in the client's message queue, then use a default.
func GetMessageForClient(i *coord.RoundInfo, clientId int64) ([]byte, error) {
	// get next message data queued for this client
	message, err := msgQueueIn.Dequeue(int(clientId))

	// if no message found in queue, use a dummy
	if err != nil {
		packet := &gatewayv1.Packet{
			Type:     gatewayv1.PacketType_PACKET_TYPE_DUMMY,
			StreamId: uint64(clientId), // use client id as stream id for round uniqueness
			Sequence: 0,
			Length:   0,
			Data:     nil,
		}
		message, err = packetPack(packet)
	}

	utils.DebugLog("data=%x, message=%x", data, message)

	return message, err
}

// Send final messages from the mix-net here.
// This replaces `coordinator.Check` testing for message ids.
// Note: There are duplicates, so sort unique.
func CheckFinalMessages(messages [][]byte, numExpected int) bool {

	/*
		This is inherently hacky within the state/aspect of trellis simulation.
		For now:
		- extract unique messages
		- enqueue non-dummy packets as gateway output
		- compare number of unique messages with number expected
	*/

	// extract unique messages from duplicates
	uniquePackets := make(map[uint64]*gatewayv1.Packet)
	for _, m := range messages {
		packet, err := packetUnpack(m)
		if err != nil {
			panic(err)
		}

		var uid uint64 = 0

		if packet.Type == gatewayv1.PacketType_PACKET_TYPE_DUMMY {
			uid = packet.StreamId
		} else {
			uid = packet.StreamId + packet.Sequence
		}

		uniquePackets[uid] = packet
	}

	// put unique non-dummy packets in the out queue
	for _, p := range uniquePackets {
		if p.Type != gatewayv1.PacketType_PACKET_TYPE_DUMMY {
			msgQueueOut.Enqueue(int(p.StreamId), p.Data) // TODO: support data reassembly from packets+protocol
		}
	}

	return len(uniquePackets) == numExpected
}

////////////////////////////////////////////////////////////////////////
// Gateway Packet I/O
////////////////////////////////////////////////////////////////////////

// Send a message through the mix-net.
func sendPacket(packet *gatewayv1.Packet) {
	// pack the packet into a static-sized message
	message, err := packetPack(packet)
	if err != nil {
		// TODO: handle error less fatally
		panic(fmt.Sprintf("[Gateway] Error creating message from packet: %v\n", err))
	}

	// stage message for a mix-net client to pick it up
	msgQueueIn.Enqueue(0, message) // TODO: message queue per stream, not client
}

// Start gateway proxy to listen for incoming messages
func proxyStart() {
	// Create a listener for incoming connections
	listener, err := net.Listen("tcp", proxyAddress)
	if err != nil {
		log.Printf("[Gateway] Error listening: %v\n", err)
		return
	}
	defer listener.Close()

	log.Printf("[Gateway] Listening on %s...\n", proxyAddress)

	// Accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[Gateway] Error accepting connection: %v\n", err)
			continue
		}

		go proxyHandleConnection(conn)
	}
}

func proxyHandleConnection(conn net.Conn) {
	defer conn.Close()

	streamId := getStreamId(conn)
	var packetCounter uint64 = 0

	utils.DebugLog("[Gateway] Accepted connection from %s id=%d", conn.RemoteAddr(), streamId)

	// send PACKET_TYPE_START for a new transmission
	packet := &gatewayv1.Packet{
		Type:     gatewayv1.PacketType_PACKET_TYPE_START,
		StreamId: streamId,
		Sequence: 0,
		Length:   0,
		Data:     nil,
	}
	sendPacket(packet)
	packetCounter++

	for {
		// TODO: more data space may be available if protocol uses variant sizes, then measure data size per-packet
		dataSize := GetMaxDataSize()

		bufferData := make([]byte, dataSize)
		n, err := conn.Read(bufferData)

		if err != nil {
			packetFinal := &gatewayv1.Packet{
				StreamId: streamId,
				Sequence: packetCounter,
			}

			if err == io.EOF {
				packetFinal.Type = gatewayv1.PacketType_PACKET_TYPE_END
				utils.DebugLog("[Gateway] Finished receiving data stream")

			} else {
				packetFinal.Type = gatewayv1.PacketType_PACKET_TYPE_ERROR
				utils.DebugLog("[Gateway] Error receiving data stream: %s", err.Error())
			}

			sendPacket(packetFinal)
			return
		}

		packet := &gatewayv1.Packet{
			Type:     gatewayv1.PacketType_PACKET_TYPE_DATA,
			StreamId: streamId,
			Sequence: packetCounter,
			Length:   uint32(n),
			Data:     bufferData[:n],
		}
		sendPacket(packet)
		packetCounter++
	}
}

// cheap hack to produce a unique id from a network connection
// TODO?: use UUID module
func getStreamId(conn net.Conn) uint64 {
	id0 := strings.Trim(fmt.Sprintf("%x", conn), "&{}")

	// parse numeric string as hexadecimal integer (base 16)
	id, err := strconv.ParseInt(id0, 16, 64)
	if err != nil {
		panic("[Gateway] uuid conversion failed")
	}

	return uint64(id)
}
