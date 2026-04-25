package sia

import (
	"fmt"
	"testing"

	"github.com/RCooLeR/AjaxBridge/internal/event"
)

func TestParserPlainSIADCS(t *testing.T) {
	payload := `"SIA-DCS"0001R0L0#0001[#0001|Nri0/BA001]_12:00:00,04-19-2026`
	raw := testFrame(payload)

	parser, err := NewParser("0001", true, "")
	if err != nil {
		t.Fatal(err)
	}
	evt, frame, err := parser.Parse([]byte(raw))
	if err != nil {
		t.Fatalf("parse frame: %v", err)
	}
	if frame.Account != "0001" {
		t.Fatalf("account = %q", frame.Account)
	}
	if evt.EventCode != "BA" {
		t.Fatalf("event code = %q", evt.EventCode)
	}
	if evt.EventClass != event.ClassAlarm {
		t.Fatalf("event class = %q", evt.EventClass)
	}
	if evt.EventAction != "burglary_alarm" {
		t.Fatalf("event action = %q", evt.EventAction)
	}
	if evt.Zone != "1" {
		t.Fatalf("zone = %q", evt.Zone)
	}
	if evt.Partition != "0" || evt.Group != "0" {
		t.Fatalf("partition/group = %q/%q", evt.Partition, evt.Group)
	}
	if evt.Device != "" {
		t.Fatalf("device = %q, want empty because ri0 is area/partition", evt.Device)
	}
}

func TestParserSIADCSAreaAndZone(t *testing.T) {
	payload := `"SIA-DCS"0004L0#0001[#0001|Nri1/BA007]_12:00:00,04-19-2026`
	raw := testFrame(payload)

	parser, err := NewParser("0001", true, "")
	if err != nil {
		t.Fatal(err)
	}
	evt, _, err := parser.Parse([]byte(raw))
	if err != nil {
		t.Fatalf("parse frame: %v", err)
	}
	if evt.Partition != "1" || evt.Group != "1" {
		t.Fatalf("partition/group = %q/%q", evt.Partition, evt.Group)
	}
	if evt.Zone != "7" {
		t.Fatalf("zone = %q", evt.Zone)
	}
	if evt.Device != "" {
		t.Fatalf("device = %q, want empty", evt.Device)
	}
}

func TestParserNightModeSIADCS(t *testing.T) {
	payload := `"SIA-DCS"0002R0L0#0001[#0001|Nri0/NL]_12:00:00,04-19-2026`
	raw := testFrame(payload)

	parser, err := NewParser("0001", true, "")
	if err != nil {
		t.Fatal(err)
	}
	evt, _, err := parser.Parse([]byte(raw))
	if err != nil {
		t.Fatalf("parse frame: %v", err)
	}
	if evt.EventClass != event.ClassNight {
		t.Fatalf("event class = %q", evt.EventClass)
	}
	if evt.EventAction != "night_mode_on" {
		t.Fatalf("event action = %q", evt.EventAction)
	}
	if evt.Signal != "night_mode" {
		t.Fatalf("signal = %q", evt.Signal)
	}
}

func TestParserCapturesExtendedData(t *testing.T) {
	payload := `"SIA-DCS"0003R0L0#0001[#0001|Nri0/PA][G50.4501,30.5234]_12:00:00,04-19-2026`
	raw := testFrame(payload)

	parser, err := NewParser("0001", true, "")
	if err != nil {
		t.Fatal(err)
	}
	evt, _, err := parser.Parse([]byte(raw))
	if err != nil {
		t.Fatalf("parse frame: %v", err)
	}
	if evt.EventAction != "panic_alarm" {
		t.Fatalf("event action = %q", evt.EventAction)
	}
	if len(evt.XData) != 1 {
		t.Fatalf("xdata length = %d", len(evt.XData))
	}
	if evt.XData[0].Identifier != "G" || evt.XData[0].Value != "50.4501,30.5234" {
		t.Fatalf("xdata = %#v", evt.XData[0])
	}
}

func TestParserDeviceBypassSIADCS(t *testing.T) {
	payload := `"SIA-DCS"0005R0L0#0001[#0001|Nri1/QB007]_12:00:00,04-19-2026`
	raw := testFrame(payload)

	parser, err := NewParser("0001", true, "")
	if err != nil {
		t.Fatal(err)
	}
	evt, _, err := parser.Parse([]byte(raw))
	if err != nil {
		t.Fatalf("parse frame: %v", err)
	}
	if evt.EventClass != event.ClassTrouble {
		t.Fatalf("event class = %q", evt.EventClass)
	}
	if evt.EventAction != "device_bypassed" {
		t.Fatalf("event action = %q", evt.EventAction)
	}
	if evt.EventName != "Device bypassed or deactivated" {
		t.Fatalf("event name = %q", evt.EventName)
	}
	if evt.Signal != "bypass" {
		t.Fatalf("signal = %q", evt.Signal)
	}
	if evt.Zone != "7" {
		t.Fatalf("zone = %q", evt.Zone)
	}
}

func TestParserDeviceBypassRestoreSIADCS(t *testing.T) {
	payload := `"SIA-DCS"0006R0L0#0001[#0001|Nri1/QU007]_12:00:00,04-19-2026`
	raw := testFrame(payload)

	parser, err := NewParser("0001", true, "")
	if err != nil {
		t.Fatal(err)
	}
	evt, _, err := parser.Parse([]byte(raw))
	if err != nil {
		t.Fatalf("parse frame: %v", err)
	}
	if evt.EventClass != event.ClassRestore {
		t.Fatalf("event class = %q", evt.EventClass)
	}
	if evt.EventAction != "device_bypass_restored" {
		t.Fatalf("event action = %q", evt.EventAction)
	}
	if evt.Signal != "bypass" {
		t.Fatalf("signal = %q", evt.Signal)
	}
}

func TestParserTamperBypassSIADCS(t *testing.T) {
	payload := `"SIA-DCS"0007R0L0#0001[#0001|Nri1/TB007]_12:00:00,04-19-2026`
	raw := testFrame(payload)

	parser, err := NewParser("0001", true, "")
	if err != nil {
		t.Fatal(err)
	}
	evt, _, err := parser.Parse([]byte(raw))
	if err != nil {
		t.Fatalf("parse frame: %v", err)
	}
	if evt.EventClass != event.ClassTrouble {
		t.Fatalf("event class = %q", evt.EventClass)
	}
	if evt.EventAction != "tamper_bypassed" {
		t.Fatalf("event action = %q", evt.EventAction)
	}
	if evt.Signal != "tamper_bypass" {
		t.Fatalf("signal = %q", evt.Signal)
	}
}

func TestParserTamperBypassRestoreSIADCS(t *testing.T) {
	payload := `"SIA-DCS"0008R0L0#0001[#0001|Nri1/TU007]_12:00:00,04-19-2026`
	raw := testFrame(payload)

	parser, err := NewParser("0001", true, "")
	if err != nil {
		t.Fatal(err)
	}
	evt, _, err := parser.Parse([]byte(raw))
	if err != nil {
		t.Fatalf("parse frame: %v", err)
	}
	if evt.EventClass != event.ClassRestore {
		t.Fatalf("event class = %q", evt.EventClass)
	}
	if evt.EventAction != "tamper_bypass_restored" {
		t.Fatalf("event action = %q", evt.EventAction)
	}
	if evt.Signal != "tamper_bypass" {
		t.Fatalf("signal = %q", evt.Signal)
	}
}

func TestBuildResponseHasValidCRCAndLength(t *testing.T) {
	frame := &Frame{
		Sequence: "0001",
		Receiver: "R0",
		Line:     "L0",
		Account:  "0001",
	}
	responder, err := NewResponder("")
	if err != nil {
		t.Fatal(err)
	}
	response := responder.Build(ResponseACK, frame)
	parsed, err := parseFrame(string(response[1 : len(response)-1]))
	if err != nil {
		t.Fatalf("parse response frame: %v", err)
	}
	if err := validateFrame(parsed, true); err != nil {
		t.Fatalf("response is invalid: %v", err)
	}
}

func TestParserRejectsUnexpectedAccount(t *testing.T) {
	payload := `"NULL"0001R0L0#0001[]_12:00:00,04-19-2026`
	raw := testFrame(payload)

	parser, err := NewParser("9999", true, "")
	if err != nil {
		t.Fatal(err)
	}
	evt, _, err := parser.Parse([]byte(raw))
	if err == nil {
		t.Fatal("expected account validation error")
	}
	if evt.ParseStatus != event.ParseStatusAccountInvalid {
		t.Fatalf("parse status = %q", evt.ParseStatus)
	}
}

func testFrame(payload string) string {
	return fmt.Sprintf("\n%s%04X%s\r", CRC16(payload), len(payload), payload)
}
