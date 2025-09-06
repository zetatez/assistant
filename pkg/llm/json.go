package llm

import (
	"encoding/json"
	"errors"
)

// ExtractJSONObject 从任意文本中提取最外层 JSON Object
func ExtractJSONObject(input string) (string, error) {
	start := -1
	depth := 0

	for i, r := range input {
		if r == '{' {
			if depth == 0 {
				start = i
			}
			depth++
		} else if r == '}' {
			depth--
			if depth == 0 && start != -1 {
				return input[start : i+1], nil
			}
		}
	}

	return "{}", errors.New("no valid JSON object found")
}

func ExtractAndValidateJSONObject(input string) (string, error) {
	obj, err := ExtractJSONObject(input)
	if err != nil {
		return "{}", err
	}

	var tmp any
	if err := json.Unmarshal([]byte(obj), &tmp); err != nil {
		return "{}", err
	}
	return obj, nil
}
