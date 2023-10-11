package gateway

import (
	"bytes"
	"reflect"
	"testing"

	gatewayv1 "github.com/31333337/trellis/pb/gen/proto/go/gateway/v1"
)

func TestMessageQueue(t *testing.T) {
	v1 := []byte("test1")
	v2 := []byte("test2")

	q := NewMessageQueue()

	q.Enqueue(v1)
	q.Enqueue(v2)

	m1, err := q.Dequeue()
	if !bytes.Equal(m1, v1) {
		t.Log("bytes not equal")
		t.FailNow()
	}

	m2, err := q.Dequeue()
	if !bytes.Equal(m2, v2) {
		t.Log("bytes not equal")
		t.FailNow()
	}

	_, err = q.Dequeue()
	if err == nil {
		t.Log("expected queue to be empty")
		t.FailNow()
	}
}

func tPacketPack(t *testing.T, success bool, typ gatewayv1.PacketType, id uint64, sequence uint64, data []byte) {
	header := &gatewayv1.Packet{
		Type:     typ,
		StreamId: id,
		Sequence: sequence,
	}

	message, err := packetPack(header, data)

	if !success {
		if err == nil {
			t.Log("should fail")
			t.FailNow()
		}
		return
	}

	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if len(message) != int(messageSize) {
		t.Log(err)
		t.Log("message not equal to message size")
		t.FailNow()
	}
}

func TestPacketPack(t *testing.T) {
	messageSize = 64

	var td = gatewayv1.PacketType_PACKET_TYPE_DATA

	// should pass
	tPacketPack(t, true, td, 0, 0, nil)
	tPacketPack(t, true, td, 1, 1, []byte("1"))
	tPacketPack(t, true, td, 1000000000, 1000000000, []byte("1234567890"))

	// should fail
	messageSize = 20
	tPacketPack(t, false, td, 1000000000, 1000000000, []byte("1234567890"))

	// should pass
	messageSize = 100000
	tPacketPack(t, true, td, 1000000000, 1000000000, []byte("1234567890"))
}

func TestPacketUnpack(t *testing.T) {
	messageSize = 64

	ptype := gatewayv1.PacketType_PACKET_TYPE_START
	streamid := uint64(100)
	sequence := uint64(100)
	data := []byte("1234")

	h1 := &gatewayv1.Packet{
		Type:     ptype,
		StreamId: streamid,
		Sequence: sequence,
	}

	packed, err := packetPack(h1, data)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	h2, data2, err := packetUnpack(packed)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if ptype != h2.Type {
		t.Log("packet Type not equal")
		t.FailNow()
	}

	if streamid != h2.StreamId {
		t.Log("packet StreamId not equal")
		t.FailNow()
	}

	if sequence != h2.Sequence {
		t.Log("packet Sequence not equal")
		t.FailNow()
	}

	if !bytes.Equal(data, data2) {
		t.Log("packet Data not equal")
		t.FailNow()
	}
}

func TestPacketsSort(t *testing.T) {
	var packets = []*gatewayv1.Packet{
		{
			StreamId: 20,
			Sequence: 0,
		},
		{
			StreamId: 10,
			Sequence: 2,
		},
		{
			StreamId: 10,
			Sequence: 1,
		},
		{
			StreamId: 20,
			Sequence: 1,
		},
	}

	var sorted = []*gatewayv1.Packet{
		{
			StreamId: 10,
			Sequence: 1,
		},
		{
			StreamId: 10,
			Sequence: 2,
		},
		{
			StreamId: 20,
			Sequence: 0,
		},
		{
			StreamId: 20,
			Sequence: 1,
		},
	}

	sortPackets(packets)

	if !reflect.DeepEqual(packets, sorted) {
		t.Log("packets not sorted")
		t.FailNow()
	}
}
