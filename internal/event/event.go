package event

import "time"

type Class string

const (
	ClassUnknown Class = "unknown"
	ClassAlarm   Class = "alarm"
	ClassRestore Class = "restore"
	ClassArm     Class = "arm"
	ClassDisarm  Class = "disarm"
	ClassNight   Class = "night"
	ClassTamper  Class = "tamper"
	ClassTrouble Class = "trouble"
	ClassTest    Class = "test"
	ClassCommon  Class = "common"
)

type ParseStatus string

const (
	ParseStatusOK             ParseStatus = "ok"
	ParseStatusCRCInvalid     ParseStatus = "crc_invalid"
	ParseStatusLengthInvalid  ParseStatus = "length_invalid"
	ParseStatusFormatInvalid  ParseStatus = "format_invalid"
	ParseStatusAccountInvalid ParseStatus = "account_invalid"
	ParseStatusDecryptInvalid ParseStatus = "decrypt_invalid"
)

type Normalized struct {
	Account     string       `json:"account"`
	Protocol    string       `json:"protocol"`
	Sequence    string       `json:"sequence"`
	Receiver    string       `json:"receiver"`
	Line        string       `json:"line"`
	EventCode   string       `json:"event_code"`
	ContactID   string       `json:"contact_id"`
	EventClass  Class        `json:"event_class"`
	EventAction string       `json:"event_action"`
	EventName   string       `json:"event_name"`
	Description string       `json:"description"`
	Source      string       `json:"source"`
	Signal      string       `json:"signal"`
	Severity    string       `json:"severity"`
	Partition   string       `json:"partition"`
	Group       string       `json:"group"`
	Zone        string       `json:"zone"`
	Device      string       `json:"device"`
	User        string       `json:"user"`
	OccurredAt  time.Time    `json:"occurred_at"`
	ReceivedAt  time.Time    `json:"received_at"`
	RawData     string       `json:"raw_data"`
	RawPayload  string       `json:"raw_payload"`
	RawMessage  string       `json:"raw_message"`
	XData       []XDataEntry `json:"xdata"`
	ParseStatus ParseStatus  `json:"parse_status"`
	ParseError  string       `json:"parse_error"`
	Encrypted   bool         `json:"encrypted"`
}

func (e Normalized) IsValid() bool {
	return e.ParseStatus == ParseStatusOK
}

type XDataEntry struct {
	Identifier string `json:"identifier"`
	Value      string `json:"value"`
	Raw        string `json:"raw"`
}
