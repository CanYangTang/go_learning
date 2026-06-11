package channel

import (
	"reflect"
	"testing"
	"time"
)

func TestSendMessage(t *testing.T) {
	got := SendMessage("hello")
	want := "hello"

	if got != want {
		t.Fatalf("SendMessage() = %q, want %q", got, want)
	}
}

func TestBufferedMessages(t *testing.T) {
	messages := []string{"a", "b", "c"}
	got := BufferedMessages(messages)

	if !reflect.DeepEqual(got, messages) {
		t.Fatalf("BufferedMessages() = %v, want %v", got, messages)
	}
}

func TestBufferedMessagesEmpty(t *testing.T) {
	got := BufferedMessages(nil)

	if len(got) != 0 {
		t.Fatalf("BufferedMessages(nil) length = %v, want 0", len(got))
	}
}

func TestGenerateNumbers(t *testing.T) {
	got := GenerateNumbers(5)
	want := []int{1, 2, 3, 4, 5}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GenerateNumbers() = %v, want %v", got, want)
	}
}

func TestGenerateNumbersZero(t *testing.T) {
	got := GenerateNumbers(0)

	if len(got) != 0 {
		t.Fatalf("GenerateNumbers(0) length = %v, want 0", len(got))
	}
}

func TestReceiveWithTimeoutSuccess(t *testing.T) {
	got, ok := ReceiveWithTimeout("ok", 5*time.Millisecond, 50*time.Millisecond)

	if !ok {
		t.Fatal("ReceiveWithTimeout() ok = false, want true")
	}

	if got != "ok" {
		t.Fatalf("ReceiveWithTimeout() = %q, want %q", got, "ok")
	}
}

func TestReceiveWithTimeoutTimeout(t *testing.T) {
	got, ok := ReceiveWithTimeout("slow", 50*time.Millisecond, 5*time.Millisecond)

	if ok {
		t.Fatal("ReceiveWithTimeout() ok = true, want false")
	}

	if got != "" {
		t.Fatalf("ReceiveWithTimeout() = %q, want empty string", got)
	}
}
