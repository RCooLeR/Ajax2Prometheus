package forward

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestForwarderSendsRawFrameAndCapturesACK(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	received := make(chan string, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		buf := make([]byte, 128)
		n, _ := conn.Read(buf)
		received <- string(buf[:n])
		_, _ = conn.Write([]byte("\n0000000A\"ACK\"0001[]\r"))
	}()

	raw := []byte("\nABC00001X\r")
	result := New(listener.Addr().String(), time.Second).Send(context.Background(), raw)
	if !result.ACK {
		t.Fatalf("expected ACK result, got status=%s error=%s", result.Status, result.Error)
	}
	if result.Duration <= 0 {
		t.Fatalf("expected positive duration, got %s", result.Duration)
	}

	select {
	case got := <-received:
		if got != string(raw) {
			t.Fatalf("forwarded frame = %q, want %q", got, string(raw))
		}
	case <-time.After(time.Second):
		t.Fatal("server did not receive forwarded frame")
	}
}

func TestForwarderDialError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	_ = listener.Close()

	result := New(addr, 50*time.Millisecond).Send(context.Background(), []byte("test"))
	if result.Status != StatusDialError && result.Status != StatusTimeout {
		t.Fatalf("status = %s, error = %s", result.Status, result.Error)
	}
	if result.Error == "" {
		t.Fatal("expected dial error text")
	}
	if result.StartedAt.IsZero() {
		t.Fatal("expected started_at to be recorded")
	}
}

func TestGroupSendsToAllTargets(t *testing.T) {
	addr1, received1 := ackServer(t)
	addr2, received2 := ackServer(t)

	raw := []byte("\nABC00001X\r")
	results := NewGroup([]string{addr1, addr2}, time.Second).Send(context.Background(), raw)
	if len(results) != 2 {
		t.Fatalf("results length = %d", len(results))
	}
	if !AllACK(results) {
		t.Fatalf("expected all ACK, got %s", Summary(results))
	}

	for _, received := range []chan string{received1, received2} {
		select {
		case got := <-received:
			if got != string(raw) {
				t.Fatalf("forwarded frame = %q, want %q", got, string(raw))
			}
		case <-time.After(time.Second):
			t.Fatal("server did not receive forwarded frame")
		}
	}
}

func ackServer(t *testing.T) (string, chan string) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	received := make(chan string, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		buf := make([]byte, 128)
		n, _ := conn.Read(buf)
		received <- string(buf[:n])
		_, _ = conn.Write([]byte("\n0000000A\"ACK\"0001[]\r"))
	}()
	return listener.Addr().String(), received
}

func ExampleSkipped() {
	result := Skipped("127.0.0.1:4321", "parse_status=crc_invalid")
	fmt.Println(result.Enabled, result.Status, result.Target, result.Error)
	// Output: true skipped 127.0.0.1:4321 parse_status=crc_invalid
}
