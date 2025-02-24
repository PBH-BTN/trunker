package conv

import "golang.org/x/text/encoding/charmap"

func TransUTF8To9959_1(raw []byte) []byte {
	encoded, _ := charmap.ISO8859_1.NewEncoder().Bytes(raw)
	return encoded
}

func Trans9959_1ToUTF8(raw []byte) []byte {
	encoded, _ := charmap.ISO8859_1.NewDecoder().Bytes(raw)
	return encoded
}
