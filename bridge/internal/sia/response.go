package sia

import (
	"fmt"
	"time"
)

type ResponseType string

const (
	ResponseACK ResponseType = "ACK"
	ResponseNAK ResponseType = "NAK"
	ResponseDUH ResponseType = "DUH"
)

type Responder struct {
	aesKey []byte
}

func NewResponder(encryptionKey string) (*Responder, error) {
	key, err := decodeAESKey(encryptionKey)
	if err != nil {
		return nil, err
	}
	return &Responder{aesKey: key}, nil
}

func (r *Responder) Build(kind ResponseType, frame *Frame) []byte {
	if frame == nil {
		return buildPlainResponse(`"NAK"0000R0L0A0[]` + timestamp())
	}
	if frame.Encrypted && kind == ResponseACK && len(r.aesKey) > 0 {
		if payload, err := r.buildEncryptedACK(frame); err == nil {
			return buildPlainResponse(payload)
		}
	}
	account := frame.Account
	if account == "" {
		account = "0"
	}
	payload := fmt.Sprintf(`"%s"%s%s%s#%s[]`, kind, frame.Sequence, frame.Receiver, frame.Line, account)
	return buildPlainResponse(payload)
}

func (r *Responder) buildEncryptedACK(frame *Frame) (string, error) {
	encrypted, err := encryptSIAContent(r.aesKey, "]"+timestamp())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`"*ACK"%s%s%s#%s[%s`, frame.Sequence, frame.Receiver, frame.Line, frame.Account, encrypted), nil
}

func buildPlainResponse(payload string) []byte {
	return []byte(fmt.Sprintf("\n%s%04X%s\r", CRC16(payload), len(payload), payload))
}

func timestamp() string {
	return time.Now().UTC().Format("_15:04:05,01-02-2006")
}
