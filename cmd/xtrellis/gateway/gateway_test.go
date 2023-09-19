package gateway

import (
	"bytes"
	"testing"
)

func TestMessageQueue(t *testing.T) {
	id := 3
	v1 := []byte("test1")
	v2 := []byte("test2")

	enqueue(id, v1)
	enqueue(id, v2)

	_, err := dequeue(4)
	if err == nil {
		t.Log("expected index to not have a queue")
		t.FailNow()
	}

	m1, err := dequeue(id)
	if !bytes.Equal(m1, v1) {
		t.Log("bytes not equal")
		t.FailNow()
	}

	m2, err := dequeue(id)
	if !bytes.Equal(m2, v2) {
		t.Log("bytes not equal")
		t.FailNow()
	}
}
