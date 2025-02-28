package utils

import json "github.com/bytedance/sonic"

func ToJSON(v any) string {
	s, _ := json.MarshalString(v)
	return s
}
