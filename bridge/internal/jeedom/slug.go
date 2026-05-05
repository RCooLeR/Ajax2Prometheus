package jeedom

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"unicode"
	"unicode/utf8"
)

func RepairText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	decoded, ok := decodeCP1251Mojibake(value)
	if !ok || decoded == value {
		return value
	}
	if mojibakeScore(decoded) < mojibakeScore(value) || cyrillicCount(decoded) > cyrillicCount(value) {
		return decoded
	}
	return value
}

func Slug(value string) string {
	original := strings.TrimSpace(value)
	value = RepairText(original)

	var b strings.Builder
	lastSeparator := false
	for _, r := range value {
		r = unicode.ToLower(foldLatin(r))
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastSeparator = false
		default:
			if part := transliterateCyrillic(r); part != "" {
				b.WriteString(part)
				lastSeparator = false
				continue
			}
			if !lastSeparator {
				b.WriteByte('_')
				lastSeparator = true
			}
		}
	}

	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "dev_" + shortHash(original)
	}
	return out
}

func MetricName(value, fallback string) string {
	slug := Slug(value)
	if strings.HasPrefix(slug, "dev_") && fallback != "" {
		return fallback
	}
	return strings.ReplaceAll(slug, "-", "_")
}

func shortHash(value string) string {
	sum := sha1.Sum([]byte(value))
	return hex.EncodeToString(sum[:])[:8]
}

func foldLatin(r rune) rune {
	switch r {
	case 'à', 'á', 'â', 'ã', 'ä', 'å', 'ā', 'ă', 'ą', 'À', 'Á', 'Â', 'Ã', 'Ä', 'Å', 'Ā', 'Ă', 'Ą':
		return 'a'
	case 'ç', 'ć', 'č', 'Ç', 'Ć', 'Č':
		return 'c'
	case 'ď', 'Đ', 'đ', 'Ď':
		return 'd'
	case 'è', 'é', 'ê', 'ë', 'ē', 'ė', 'ę', 'È', 'É', 'Ê', 'Ë', 'Ē', 'Ė', 'Ę':
		return 'e'
	case 'ì', 'í', 'î', 'ï', 'ī', 'į', 'Ì', 'Í', 'Î', 'Ï', 'Ī', 'Į':
		return 'i'
	case 'ñ', 'ń', 'Ñ', 'Ń':
		return 'n'
	case 'ò', 'ó', 'ô', 'õ', 'ö', 'ø', 'ō', 'Ò', 'Ó', 'Ô', 'Õ', 'Ö', 'Ø', 'Ō':
		return 'o'
	case 'ř', 'Ř':
		return 'r'
	case 'š', 'ś', 'Š', 'Ś':
		return 's'
	case 'ť', 'Ť':
		return 't'
	case 'ù', 'ú', 'û', 'ü', 'ū', 'Ù', 'Ú', 'Û', 'Ü', 'Ū':
		return 'u'
	case 'ý', 'ÿ', 'Ý':
		return 'y'
	case 'ž', 'ź', 'ż', 'Ž', 'Ź', 'Ż':
		return 'z'
	default:
		return r
	}
}

func transliterateCyrillic(r rune) string {
	switch r {
	case 'а':
		return "a"
	case 'б':
		return "b"
	case 'в':
		return "v"
	case 'г', 'ґ':
		return "g"
	case 'д':
		return "d"
	case 'е', 'ё', 'э':
		return "e"
	case 'є':
		return "ye"
	case 'ж':
		return "zh"
	case 'з':
		return "z"
	case 'и', 'ы':
		return "y"
	case 'і':
		return "i"
	case 'ї':
		return "yi"
	case 'й':
		return "y"
	case 'к':
		return "k"
	case 'л':
		return "l"
	case 'м':
		return "m"
	case 'н':
		return "n"
	case 'о':
		return "o"
	case 'п':
		return "p"
	case 'р':
		return "r"
	case 'с':
		return "s"
	case 'т':
		return "t"
	case 'у':
		return "u"
	case 'ф':
		return "f"
	case 'х':
		return "kh"
	case 'ц':
		return "ts"
	case 'ч':
		return "ch"
	case 'ш':
		return "sh"
	case 'щ':
		return "shch"
	case 'ю':
		return "yu"
	case 'я':
		return "ya"
	case 'ь', 'ъ':
		return ""
	default:
		return ""
	}
}

func decodeCP1251Mojibake(value string) (string, bool) {
	bytes := make([]byte, 0, len(value))
	for _, r := range value {
		if r < utf8.RuneSelf {
			bytes = append(bytes, byte(r))
			continue
		}
		b, ok := cp1251Byte(r)
		if !ok {
			return "", false
		}
		bytes = append(bytes, b)
	}
	if !utf8.Valid(bytes) {
		return "", false
	}
	return string(bytes), true
}

func cp1251Byte(r rune) (byte, bool) {
	if r >= 'А' && r <= 'я' {
		return byte(r-'А') + 0xC0, true
	}
	switch r {
	case 'Ђ':
		return 0x80, true
	case 'Ѓ':
		return 0x81, true
	case '‚':
		return 0x82, true
	case 'ѓ':
		return 0x83, true
	case '„':
		return 0x84, true
	case '…':
		return 0x85, true
	case '†':
		return 0x86, true
	case '‡':
		return 0x87, true
	case '€':
		return 0x88, true
	case '‰':
		return 0x89, true
	case 'Љ':
		return 0x8A, true
	case '‹':
		return 0x8B, true
	case 'Њ':
		return 0x8C, true
	case 'Ќ':
		return 0x8D, true
	case 'Ћ':
		return 0x8E, true
	case 'Џ':
		return 0x8F, true
	case 'ђ':
		return 0x90, true
	case '‘':
		return 0x91, true
	case '’':
		return 0x92, true
	case '“':
		return 0x93, true
	case '”':
		return 0x94, true
	case '•':
		return 0x95, true
	case '–':
		return 0x96, true
	case '—':
		return 0x97, true
	case '™':
		return 0x99, true
	case 'љ':
		return 0x9A, true
	case '›':
		return 0x9B, true
	case 'њ':
		return 0x9C, true
	case 'ќ':
		return 0x9D, true
	case 'ћ':
		return 0x9E, true
	case 'џ':
		return 0x9F, true
	case '\u00A0':
		return 0xA0, true
	case 'Ў':
		return 0xA1, true
	case 'ў':
		return 0xA2, true
	case 'Ј':
		return 0xA3, true
	case '¤':
		return 0xA4, true
	case 'Ґ':
		return 0xA5, true
	case '¦':
		return 0xA6, true
	case '§':
		return 0xA7, true
	case 'Ё':
		return 0xA8, true
	case '©':
		return 0xA9, true
	case 'Є':
		return 0xAA, true
	case '«':
		return 0xAB, true
	case '¬':
		return 0xAC, true
	case '\u00AD':
		return 0xAD, true
	case '®':
		return 0xAE, true
	case 'Ї':
		return 0xAF, true
	case '°':
		return 0xB0, true
	case '±':
		return 0xB1, true
	case 'І':
		return 0xB2, true
	case 'і':
		return 0xB3, true
	case 'ґ':
		return 0xB4, true
	case 'µ':
		return 0xB5, true
	case '¶':
		return 0xB6, true
	case '·':
		return 0xB7, true
	case 'ё':
		return 0xB8, true
	case '№':
		return 0xB9, true
	case 'є':
		return 0xBA, true
	case '»':
		return 0xBB, true
	case 'ј':
		return 0xBC, true
	case 'Ѕ':
		return 0xBD, true
	case 'ѕ':
		return 0xBE, true
	case 'ї':
		return 0xBF, true
	default:
		return 0, false
	}
}

func mojibakeScore(value string) int {
	score := 0
	for _, r := range value {
		switch r {
		case 'Ð', 'Ñ', 'Â', 'Ã', 'Р', 'С', 'В', 'Г', 'µ', '©', '°', 'Ў', 'Ђ', 'І', 'Ѕ':
			score++
		}
	}
	return score
}

func cyrillicCount(value string) int {
	count := 0
	for _, r := range value {
		if unicode.Is(unicode.Cyrillic, r) {
			count++
		}
	}
	return count
}
