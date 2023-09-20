package gateway

import (
	"bytes"
	"testing"
)

func TestMessageQueue(t *testing.T) {
	id := 3
	v1 := []byte("test1")
	v2 := []byte("test2")

	q := NewMessageQueue()

	q.Enqueue(id, v1)
	q.Enqueue(id, v2)

	_, err := q.Dequeue(4)
	if err == nil {
		t.Log("expected index to not have a queue")
		t.FailNow()
	}

	m1, err := q.Dequeue(id)
	if !bytes.Equal(m1, v1) {
		t.Log("bytes not equal")
		t.FailNow()
	}

	m2, err := q.Dequeue(id)
	if !bytes.Equal(m2, v2) {
		t.Log("bytes not equal")
		t.FailNow()
	}
}

func TestMessageSerialization(t *testing.T) {
	id := uint64(3)
	data := []byte("test1")
	_, err := messageSerialize(id, data)

	if err == nil {
		t.Log("Expected error: Invalid messageSize")
		t.FailNow()
	}

	messageSize = 32

	b, err1 := messageSerialize(id, data)
	if err1 != nil {
		t.Log(err1)
		t.FailNow()
	}

	id2, data2, err2 := messageUnserialize(b)
	if err2 != nil {
		t.Log(err2)
		t.FailNow()
	}

	if id != id2 {
		t.Log("message ids not equal")
		t.FailNow()
	}

	if !bytes.Equal(data, data2) {
		t.Log("message data not equal")
		t.FailNow()
	}
}
