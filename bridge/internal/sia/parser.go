package sia

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/RCooLeR/AjaxBridge/internal/event"
)

var (
	headerPattern = regexp.MustCompile(`^"(\*?[^"]+)"([0-9A-Fa-f]{4})(R[0-9A-Za-z]+)?(L[0-9A-Za-z]+)?#([^\[]+)`)
	timePattern   = regexp.MustCompile(`_(\d{2}:\d{2}:\d{2}),(\d{2})-(\d{2})-(\d{4})`)
)

type Parser struct {
	account   string
	strictCRC bool
	aesKey    []byte
}

type Frame struct {
	Raw       string
	CRC       string
	Length    string
	Payload   string
	Token     string
	Encrypted bool
	Sequence  string
	Receiver  string
	Line      string
	Account   string
}

func NewParser(account string, strictCRC bool, encryptionKey string) (*Parser, error) {
	key, err := decodeAESKey(encryptionKey)
	if err != nil {
		return nil, err
	}
	return &Parser{account: account, strictCRC: strictCRC, aesKey: key}, nil
}

func (p *Parser) Parse(raw []byte) (*event.Normalized, *Frame, error) {
	now := time.Now().UTC()
	line := strings.TrimSpace(strings.Trim(string(raw), "\x00\r\n"))
	normalized := &event.Normalized{
		ReceivedAt:  now,
		OccurredAt:  now,
		RawMessage:  line,
		ParseStatus: event.ParseStatusOK,
	}
	if line == "" {
		return normalized, nil, parseError(event.ParseStatusFormatInvalid, "empty SIA frame")
	}

	frame, err := parseFrame(line)
	if err != nil {
		normalized.ParseStatus = event.ParseStatusFormatInvalid
		normalized.ParseError = err.Error()
		return normalized, frame, err
	}
	copyFrame(normalized, frame)

	if err := validateFrame(frame, p.strictCRC); err != nil {
		status := event.ParseStatusFormatInvalid
		if strings.Contains(err.Error(), "CRC") {
			status = event.ParseStatusCRCInvalid
		}
		if strings.Contains(err.Error(), "length") {
			status = event.ParseStatusLengthInvalid
		}
		normalized.ParseStatus = status
		normalized.ParseError = err.Error()
		return normalized, frame, parseError(status, err.Error())
	}

	content, xdata, trailer, err := p.extractContent(frame)
	if err != nil {
		normalized.ParseStatus = event.ParseStatusDecryptInvalid
		normalized.ParseError = err.Error()
		return normalized, frame, parseError(event.ParseStatusDecryptInvalid, err.Error())
	}
	normalized.RawData = content
	normalized.XData = xdata
	normalized.RawPayload = rebuildRawPayload(content, xdata, trailer)

	if p.account != "" && frame.Account != p.account {
		err := fmt.Errorf("unexpected account %q, expected %q", frame.Account, p.account)
		normalized.ParseStatus = event.ParseStatusAccountInvalid
		normalized.ParseError = err.Error()
		return normalized, frame, parseError(event.ParseStatusAccountInvalid, err.Error())
	}

	parseContent(normalized, content)
	if ts, ok := parseTimestamp(trailer); ok {
		normalized.OccurredAt = ts
	}
	enrichEvent(normalized)
	return normalized, frame, nil
}

func parseFrame(line string) (*Frame, error) {
	frame := &Frame{Raw: line}
	payload := line
	if len(line) >= 8 && isHex(line[:4]) && isHex(line[4:8]) {
		frame.CRC = strings.ToUpper(line[:4])
		frame.Length = strings.ToUpper(line[4:8])
		payload = line[8:]
	}
	frame.Payload = payload

	matches := headerPattern.FindStringSubmatch(payload)
	if matches == nil {
		return frame, fmt.Errorf("SIA payload header did not match expected DC-09 shape")
	}
	frame.Token = matches[1]
	frame.Encrypted = strings.HasPrefix(frame.Token, "*")
	frame.Sequence = matches[2]
	frame.Receiver = fallback(matches[3], "R0")
	frame.Line = fallback(matches[4], "L0")
	frame.Account = matches[5]
	return frame, nil
}

func validateFrame(frame *Frame, strictCRC bool) error {
	if frame == nil {
		return fmt.Errorf("missing frame")
	}
	if frame.Length != "" {
		expected, err := strconv.ParseInt(frame.Length, 16, 32)
		if err != nil {
			return fmt.Errorf("invalid frame length %q: %w", frame.Length, err)
		}
		if int(expected) != len(frame.Payload) {
			return fmt.Errorf("frame length mismatch: got %d, expected %d", len(frame.Payload), expected)
		}
	}
	if strictCRC && frame.CRC != "" {
		calculated := CRC16(frame.Payload)
		if calculated != frame.CRC {
			return fmt.Errorf("CRC mismatch: got %s, calculated %s", frame.CRC, calculated)
		}
	}
	return nil
}

func (p *Parser) extractContent(frame *Frame) (string, []event.XDataEntry, string, error) {
	open := strings.Index(frame.Payload, "[")
	if open < 0 {
		return "", nil, "", fmt.Errorf("payload has no data opening bracket")
	}
	rest := frame.Payload[open+1:]

	if frame.Encrypted {
		if len(p.aesKey) == 0 {
			return "", nil, "", fmt.Errorf("encrypted payload received but no AES key is configured")
		}
		plain, err := decryptSIAContent(p.aesKey, rest)
		if err != nil {
			return "", nil, "", err
		}
		return splitPlainContent(plain)
	}

	close := strings.Index(rest, "]")
	if close < 0 {
		return "", nil, "", fmt.Errorf("payload has no data closing bracket")
	}
	xdata, trailer := parseXDataTrailer(rest[close+1:])
	return rest[:close], xdata, trailer, nil
}

func splitPlainContent(value string) (string, []event.XDataEntry, string, error) {
	close := strings.Index(value, "]")
	if close < 0 {
		return value, nil, "", nil
	}
	xdata, trailer := parseXDataTrailer(value[close+1:])
	return value[:close], xdata, trailer, nil
}

func parseContent(normalized *event.Normalized, content string) {
	if content == "" {
		normalized.EventCode = "RP"
		return
	}

	accountPart, dataPart, ok := strings.Cut(content, "|")
	if ok && strings.HasPrefix(accountPart, "#") && normalized.Account == "" {
		normalized.Account = strings.TrimPrefix(accountPart, "#")
	}
	if !ok {
		dataPart = content
	}

	switch strings.TrimPrefix(normalized.Protocol, "*") {
	case "ADM-CID":
		parseContactID(normalized, dataPart)
	default:
		parseSIADCS(normalized, dataPart)
	}
}

func parseSIADCS(normalized *event.Normalized, data string) {
	data = strings.TrimPrefix(data, "N")
	for {
		token, rest, ok := strings.Cut(data, "/")
		if !ok {
			break
		}
		if parseSIAPrefix(normalized, token) {
			data = rest
			continue
		}
		break
	}

	for i := 0; i+1 < len(data); i++ {
		if isUpperAlpha(data[i]) && isUpperAlpha(data[i+1]) {
			normalized.EventCode = data[i : i+2]
			rest := data[i+2:]
			normalized.Zone = leadingDigits(rest)
			return
		}
	}
	normalized.EventCode = "YN"
}

func parseSIAPrefix(normalized *event.Normalized, token string) bool {
	token = strings.TrimSpace(token)
	if token == "" {
		return true
	}
	lower := strings.ToLower(token)
	switch {
	case strings.HasPrefix(lower, "ri"):
		value := strings.TrimLeft(token[2:], "0")
		if value == "" {
			value = "0"
		}
		normalized.Partition = value
		normalized.Group = value
		return true
	default:
		return false
	}
}

func parseContactID(normalized *event.Normalized, data string) {
	digits := onlyDigits(data)
	if len(digits) >= 9 {
		qualifier := digits[0:1]
		normalized.ContactID = contactIDQualifier(qualifier) + digits[1:4]
		normalized.EventCode = "CID" + digits[1:4]
		normalized.Partition = strings.TrimLeft(digits[4:6], "0")
		normalized.Group = normalized.Partition
		normalized.Zone = strings.TrimLeft(digits[6:9], "0")
		return
	}
	normalized.EventCode = "CID"
}

func parseXDataTrailer(value string) ([]event.XDataEntry, string) {
	var entries []event.XDataEntry
	rest := value
	for strings.HasPrefix(rest, "[") {
		close := strings.Index(rest, "]")
		if close < 0 {
			break
		}
		raw := rest[1:close]
		entry := event.XDataEntry{Raw: raw}
		if raw != "" {
			entry.Identifier = raw[:1]
			if len(raw) > 1 {
				entry.Value = raw[1:]
			}
		}
		entries = append(entries, entry)
		rest = rest[close+1:]
	}
	return entries, rest
}

func rebuildRawPayload(content string, xdata []event.XDataEntry, trailer string) string {
	var b strings.Builder
	b.WriteString(content)
	for _, entry := range xdata {
		b.WriteString("[")
		b.WriteString(entry.Raw)
		b.WriteString("]")
	}
	b.WriteString(trailer)
	return b.String()
}

func contactIDQualifier(value string) string {
	switch value {
	case "1":
		return "E"
	case "3":
		return "R"
	default:
		return value
	}
}

func parseTimestamp(value string) (time.Time, bool) {
	match := timePattern.FindStringSubmatch(value)
	if match == nil {
		return time.Time{}, false
	}
	parsed, err := time.Parse("15:04:05,01-02-2006", match[1]+","+match[2]+"-"+match[3]+"-"+match[4])
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func copyFrame(normalized *event.Normalized, frame *Frame) {
	normalized.Account = frame.Account
	normalized.Protocol = frame.Token
	normalized.Sequence = frame.Sequence
	normalized.Receiver = frame.Receiver
	normalized.Line = frame.Line
	normalized.Encrypted = frame.Encrypted
}

type typedParseError struct {
	status event.ParseStatus
	err    string
}

func (e typedParseError) Error() string {
	return e.err
}

func parseError(status event.ParseStatus, message string) error {
	return typedParseError{status: status, err: message}
}

func ParseStatusFromError(err error) event.ParseStatus {
	if err == nil {
		return event.ParseStatusOK
	}
	if typed, ok := err.(typedParseError); ok {
		return typed.status
	}
	return event.ParseStatusFormatInvalid
}

func isHex(value string) bool {
	for _, r := range value {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

func isUpperAlpha(b byte) bool {
	return b >= 'A' && b <= 'Z'
}

func fallback(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func leadingDigits(value string) string {
	var b strings.Builder
	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			break
		}
		b.WriteByte(value[i])
	}
	return strings.TrimLeft(b.String(), "0")
}

func onlyDigits(value string) string {
	var b strings.Builder
	for i := 0; i < len(value); i++ {
		if value[i] >= '0' && value[i] <= '9' {
			b.WriteByte(value[i])
		}
	}
	return b.String()
}
