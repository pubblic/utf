package utf

import (
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"
)

// cited from package unicode/utf16
const (
	replacementChar = '\uFFFD'     // Unicode replacement character
	maxRune         = '\U0010FFFF' // Maximum valid Unicode code point.
)

const (
	// 0xd800-0xdc00 encodes the high 10 bits of a pair.
	// 0xdc00-0xe000 encodes the low 10 bits of a pair.
	// the value is those 20 bits plus 0x10000.
	surr1 = 0xd800
	surr2 = 0xdc00
	surr3 = 0xe000

	surrSelf = 0x10000
)

func UTF8Count(u []uint16) int {
	var n int
	for i := 0; i < len(u); i++ {
		switch r := u[i]; {
		case r < surr1, surr3 <= r:
			n += utf8.RuneLen(rune(r))
		case surr1 <= r && r < surr2 && i+1 < len(u) &&
			surr2 <= u[i+1] && u[i+1] < surr3:
			n += utf8.RuneLen(utf16.DecodeRune(rune(r), rune(u[i+1])))
			i++
		default:
			// length of the bytes required to encode U+FFFD (Replacement Character)
			n += 3
		}
	}
	return n
}

func UTF8Decode(u []uint16) []byte {
	buf := make([]byte, UTF8Count(u))
	j := 0
	for i := 0; i < len(u); i++ {
		switch r := u[i]; {
		case r < surr1, surr3 <= r:
			j += utf8.EncodeRune(buf[j:], rune(r))
		case surr1 <= r && r < surr2 && i+1 < len(u) &&
			surr2 <= u[i+1] && u[i+1] < surr3:
			j += utf8.EncodeRune(buf[j:], utf16.DecodeRune(rune(r), rune(u[i+1])))
			i++
		default:
			// Invalid character is decoded to replacement character
			buf[j] = 0xEF
			buf[j+1] = 0xBF
			buf[j+2] = 0xBD
			j += 3
		}
	}
	return buf
}

func UTF8DecodeToString(u []uint16) string {
	x := UTF8Decode(u)
	return *(*string)(unsafe.Pointer(&x))
}

func UTF16RuneLen(r rune) int {
	if r < surrSelf || r > maxRune {
		return 1
	}
	return 2
}

func UTF16CountInString(src string) int {
	var n int
	for _, r := range src {
		n += UTF16RuneLen(r)
	}
	return n
}

func UTF16Count(src []byte) int {
	var n int
	for len(src) > 0 {
		r, size := utf8.DecodeRune(src)
		src = src[size:]
		n += UTF16RuneLen(r)
	}
	return n
}

func UTF16EncodeRune(dst []uint16, r rune) int {
	_ = dst[0]
	if 0 <= r && r < surr1 || surr3 <= r && r < surrSelf {
		dst[0] = uint16(r)
		return 1
	}
	if surrSelf <= r && r <= maxRune {
		r1, r2 := utf16.EncodeRune(r)
		dst[0] = uint16(r1)
		dst[1] = uint16(r2)
		return 2
	}
	dst[0] = uint16(replacementChar)
	return 1
}

func UTF16EncodeString(dst []uint16, src string) int {
	var i int
	for _, r := range src {
		i += UTF16EncodeRune(dst[i:], r)
	}
	return i
}

func UTF16Encode(dst []uint16, src []byte) int {
	var i int
	for len(dst) > 0 {
		r, size := utf8.DecodeRune(src)
		src = src[size:]
		i += UTF16EncodeRune(dst[i:], r)
	}
	return i
}
