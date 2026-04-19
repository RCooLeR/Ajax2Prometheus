package sia

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"strings"
)

var zeroIV = make([]byte, aes.BlockSize)

func decodeAESKey(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	if decoded, err := hex.DecodeString(value); err == nil {
		if isAESKeySize(len(decoded)) {
			return decoded, nil
		}
	}

	raw := []byte(value)
	if isAESKeySize(len(raw)) {
		return raw, nil
	}

	return nil, fmt.Errorf("AES key must be 16, 24, or 32 bytes, or 32/48/64 hex characters")
}

func decryptSIAContent(key []byte, encryptedHex string) (string, error) {
	ciphertext, err := hex.DecodeString(strings.TrimSpace(encryptedHex))
	if err != nil {
		return "", fmt.Errorf("decode encrypted payload: %w", err)
	}
	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("encrypted payload length %d is not an AES block multiple", len(ciphertext))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	plaintext := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, zeroIV).CryptBlocks(plaintext, ciphertext)

	return normalizeDecryptedContent(string(plaintext)), nil
}

func encryptSIAContent(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	padded := leftZeroPad([]byte(plaintext), aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, zeroIV).CryptBlocks(ciphertext, padded)
	return strings.ToUpper(hex.EncodeToString(ciphertext)), nil
}

func normalizeDecryptedContent(value string) string {
	value = strings.TrimRight(value, "\x00")
	value = strings.TrimLeft(value, "\x00")
	value = strings.TrimLeft(value, "0")
	if idx := strings.Index(value, "|"); idx >= 0 && !strings.HasPrefix(value, "#") {
		value = value[idx+1:]
	}
	return value
}

func leftZeroPad(value []byte, blockSize int) []byte {
	if len(value)%blockSize == 0 {
		return value
	}
	size := len(value) + blockSize - len(value)%blockSize
	out := make([]byte, size)
	for i := 0; i < size-len(value); i++ {
		out[i] = '0'
	}
	copy(out[size-len(value):], value)
	return out
}

func isAESKeySize(size int) bool {
	return size == 16 || size == 24 || size == 32
}
