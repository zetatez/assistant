package memory

import (
	"testing"
)

func TestShortTerm_AddAndGet(t *testing.T) {
	st := NewShortTerm(5)

	st.Add("session1", "user", "hello")
	st.Add("session1", "assistant", "hi there")

	msgs := st.GetAll("session1")
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}

	if msgs[0].Role != "user" || msgs[0].Content != "hello" {
		t.Errorf("first message mismatch: %+v", msgs[0])
	}
}

func TestShortTerm_RingBufferOverflow(t *testing.T) {
	st := NewShortTerm(3)

	for i := 0; i < 5; i++ {
		st.Add("session1", "user", "msg")
	}

	msgs := st.GetAll("session1")
	if len(msgs) != 3 {
		t.Errorf("expected 3 messages (ring buffer capacity), got %d", len(msgs))
	}
}

func TestShortTerm_Clear(t *testing.T) {
	st := NewShortTerm(10)
	st.Add("session1", "user", "hello")

	st.Clear("session1")

	msgs := st.GetAll("session1")
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages after clear, got %d", len(msgs))
	}
}

func TestShortTerm_Len(t *testing.T) {
	st := NewShortTerm(10)

	if st.Len("session1") != 0 {
		t.Errorf("expected 0 length, got %d", st.Len("session1"))
	}

	st.Add("session1", "user", "hello")
	if st.Len("session1") != 1 {
		t.Errorf("expected 1 length, got %d", st.Len("session1"))
	}
}

func TestShortTerm_MultipleSessions(t *testing.T) {
	st := NewShortTerm(5)

	st.Add("session1", "user", "hello1")
	st.Add("session2", "user", "hello2")

	msgs1 := st.GetAll("session1")
	msgs2 := st.GetAll("session2")

	if len(msgs1) != 1 || len(msgs2) != 1 {
		t.Errorf("expected 1 message each, got session1: %d, session2: %d", len(msgs1), len(msgs2))
	}

	if msgs1[0].Content != "hello1" || msgs2[0].Content != "hello2" {
		t.Errorf("message content mismatch")
	}
}
