package sia

func CRC16(value string) string {
	var crc uint16
	for _, b := range []byte(value) {
		temp := uint16(b)
		for i := 0; i < 8; i++ {
			temp ^= crc & 1
			crc >>= 1
			if temp&1 != 0 {
				crc ^= 0xA001
			}
			temp >>= 1
		}
	}
	return upperHex16(crc)
}

func upperHex16(value uint16) string {
	const digits = "0123456789ABCDEF"
	return string([]byte{
		digits[value>>12&0xF],
		digits[value>>8&0xF],
		digits[value>>4&0xF],
		digits[value&0xF],
	})
}
