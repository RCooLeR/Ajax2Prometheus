package forward

import (
	"bufio"
	"context"
	"encoding/hex"
	"errors"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	StatusDisabled      = "disabled"
	StatusSkipped       = "skipped"
	StatusACK           = "ack"
	StatusNAK           = "nak"
	StatusDUH           = "duh"
	StatusOtherResponse = "other_response"
	StatusEmptyResponse = "empty_response"
	StatusDialError     = "dial_error"
	StatusWriteError    = "write_error"
	StatusReadError     = "read_error"
	StatusTimeout       = "timeout"
)

type Forwarder struct {
	addr    string
	timeout time.Duration
}

type Group struct {
	forwarders []*Forwarder
}

type Result struct {
	Enabled       bool
	Target        string
	Status        string
	ACK           bool
	StartedAt     time.Time
	Duration      time.Duration
	RequestBytes  int
	ResponseBytes int
	ResponseASCII string
	ResponseHex   string
	Error         string
}

func New(addr string, timeout time.Duration) *Forwarder {
	return &Forwarder{addr: addr, timeout: timeout}
}

func NewGroup(addrs []string, timeout time.Duration) *Group {
	group := &Group{forwarders: make([]*Forwarder, 0, len(addrs))}
	for _, addr := range addrs {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}
		group.forwarders = append(group.forwarders, New(addr, timeout))
	}
	return group
}

func Disabled() Result {
	return Result{Enabled: false, Status: StatusDisabled}
}

func Skipped(target, reason string) Result {
	return Result{Enabled: true, Target: target, Status: StatusSkipped, Error: reason}
}

func (f *Forwarder) Addr() string {
	return f.addr
}

func (g *Group) Enabled() bool {
	return g != nil && len(g.forwarders) > 0
}

func (g *Group) Addrs() []string {
	if !g.Enabled() {
		return nil
	}
	out := make([]string, 0, len(g.forwarders))
	for _, forwarder := range g.forwarders {
		out = append(out, forwarder.Addr())
	}
	return out
}

func (g *Group) Send(ctx context.Context, raw []byte) []Result {
	if !g.Enabled() {
		return []Result{Disabled()}
	}
	results := make([]Result, len(g.forwarders))
	var wg sync.WaitGroup
	for i, target := range g.forwarders {
		wg.Go(func() {
			results[i] = target.Send(ctx, raw)
		})
	}
	wg.Wait()
	return results
}

func SkippedAll(targets []string, reason string) []Result {
	if len(targets) == 0 {
		return []Result{Disabled()}
	}
	results := make([]Result, 0, len(targets))
	for _, target := range targets {
		results = append(results, Skipped(target, reason))
	}
	return results
}

func AllACK(results []Result) bool {
	for _, result := range results {
		if result.Enabled && !result.ACK {
			return false
		}
	}
	return true
}

func Summary(results []Result) string {
	if len(results) == 0 {
		return StatusDisabled
	}
	var parts []string
	for _, result := range results {
		if !result.Enabled {
			parts = append(parts, StatusDisabled)
			continue
		}
		parts = append(parts, result.Target+"="+result.Status)
	}
	return strings.Join(parts, ",")
}

func (f *Forwarder) Send(ctx context.Context, raw []byte) (result Result) {
	startedAt := time.Now().UTC()
	result = Result{
		Enabled:      true,
		Target:       f.addr,
		StartedAt:    startedAt,
		RequestBytes: len(raw),
	}
	defer func() {
		result.Duration = time.Since(startedAt)
	}()

	dialer := net.Dialer{Timeout: f.timeout}
	conn, err := dialer.DialContext(ctx, "tcp", f.addr)
	if err != nil {
		result.Status = classifyNetworkError(err, StatusDialError)
		result.Error = err.Error()
		return result
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(f.timeout))
	if _, err := conn.Write(raw); err != nil {
		result.Status = classifyNetworkError(err, StatusWriteError)
		result.Error = err.Error()
		return result
	}

	response, err := bufio.NewReader(conn).ReadBytes('\r')
	if err != nil && len(response) == 0 {
		result.Status = classifyNetworkError(err, StatusReadError)
		result.Error = err.Error()
		return result
	}

	result.ResponseBytes = len(response)
	result.ResponseASCII = string(response)
	result.ResponseHex = strings.ToUpper(hex.EncodeToString(response))
	result.Status, result.ACK = classifyResponse(response)
	if err != nil {
		result.Error = err.Error()
	}
	return result
}

func classifyResponse(response []byte) (string, bool) {
	upper := strings.ToUpper(string(response))
	switch {
	case len(response) == 0:
		return StatusEmptyResponse, false
	case strings.Contains(upper, `"ACK"`) || strings.Contains(upper, `"*ACK"`):
		return StatusACK, true
	case strings.Contains(upper, `"NAK"`) || strings.Contains(upper, `"*NAK"`):
		return StatusNAK, false
	case strings.Contains(upper, `"DUH"`) || strings.Contains(upper, `"*DUH"`):
		return StatusDUH, false
	default:
		return StatusOtherResponse, false
	}
}

func classifyNetworkError(err error, fallback string) string {
	if err == nil {
		return fallback
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return StatusTimeout
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return StatusTimeout
	}
	return fallback
}
