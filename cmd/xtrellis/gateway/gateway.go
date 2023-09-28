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
	"sort"
	"strconv"
	"strings"
	"sync"

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
// Message Queues
// - FIFO queues are maintained for messages In and Out of the Gateway
// - incoming gateway data streams are packetized and packed as messages
// - mix-net clients retrieve the next message for each lightning round
////////////////////////////////////////////////////////////////////////

// A message queue for each client
type MessageQueue struct {
	items *list.List
	mutex sync.Mutex
}

func NewMessageQueue() *MessageQueue {
	return &MessageQueue{
		items: list.New(),
	}
}

// Incoming message queue
// Stores packetized messages as they are received by the gateway proxy
// then made available to mix-clients for transmission each round
var msgQueueIn = NewMessageQueue()

// Outgoing message queue
// Stores packetized messages as they are received from the mix-net
var msgQueueOut = NewMessageQueue()

// Enqueue data into the message queue
func (q *MessageQueue) Enqueue(m []byte) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.items.PushBack(m)
}

// Dequeue data from the queue at the given index
func (q *MessageQueue) Dequeue() ([]byte, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if element := q.items.Front(); element != nil {
		// Pop the front element from the queue
		message := element.Value.([]byte)
		q.items.Remove(element)
		return message, nil
	}

	return nil, errors.New("queue is empty")
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
// - round start: clients get messages from the In queue
// - round end: final messages are checked and placed in the Out queue
////////////////////////////////////////////////////////////////////////

// Get the next message for a client to send through the mix-net
// If message queue is empty, use a dummy message
func GetMessageForClient(i *coord.RoundInfo, clientId int64) ([]byte, error) {
	message, err := msgQueueIn.Dequeue()

	if err != nil {
		packet := &gatewayv1.Packet{
			Type:     gatewayv1.PacketType_PACKET_TYPE_DUMMY,
			StreamId: uint64(clientId), // use client id for uniqueness within round
			Sequence: 0,
			Length:   0,
			Data:     nil,
		}
		message, err = packetPack(packet)
	} else {
		utils.DebugLog("⭐ data IN")
	}

	return message, err
}

// Final messages from a mix-net round are sent here.
// This replaces `coordinator.Check` testing for message ids.
// Note: There are duplicates (why?), so sort unique.
func CheckFinalMessages(messages [][]byte, numExpected int) bool {
	// count unique messages and extract unique non-dummy packets
	uniqueMessages := make(map[uint64]uint64)
	var uniquePackets []*gatewayv1.Packet
	for _, m := range messages {
		packet, err := packetUnpack(m)
		if err != nil {
			panic(err)
		}

		uid := packet.StreamId + packet.Sequence

		// if message is unique
		if uniqueMessages[uid] == 0 {
			// mark message as seen
			uniqueMessages[uid] = uid

			// if not a dummy, collect its packet
			if packet.Type != gatewayv1.PacketType_PACKET_TYPE_DUMMY {
				uniquePackets = append(uniquePackets, packet)
			}
		}
	}

	sortPackets(uniquePackets)

	// enqueue unique sorted non-dummy packets as gateway output messages
	for _, p := range uniquePackets {
		message, err := packetPack(p) // repack... not awesome
		if err != nil {
			panic(err)
		}
		msgQueueOut.Enqueue(message)
		utils.DebugLog("⭐ data OUT [%d][%d] += '%s'", p.StreamId, p.Sequence, p.Data)
	}

	// compare unique messages with number expected
	return len(uniqueMessages) == numExpected
}

// Sort packets by StreamId, Sequence
func sortPackets(packets []*gatewayv1.Packet) {
	sort.Slice(packets, func(i, j int) bool {
		pi, pj := packets[i], packets[j]
		switch {
		case pi.StreamId != pj.StreamId:
			return pi.StreamId < pj.StreamId
		default:
			return pi.Sequence < pj.Sequence
		}
	})
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

	utils.DebugLog("[Gateway] Send packet!\n🔶 %v", packet)

	// stage message for a mix-net client to pick it up
	msgQueueIn.Enqueue(message)
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

	// start a new transmission
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
			// end transmission
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

		// send packetized data
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
