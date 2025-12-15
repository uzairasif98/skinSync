package utils

import (
	"fmt"
	"strconv"
)

func InterfaceToUint(value interface{}) (uint, error) {
	switch v := value.(type) {
	case int:
		return uint(v), nil
	case int8, int16, int32, int64:
		return uint(v.(int64)), nil
	case uint, uint8, uint16, uint32, uint64:
		return v.(uint), nil
	case float32, float64:
		return uint(v.(float64)), nil
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid string format: %v", err)
		}
		return uint(parsed), nil
	default:
		return 0, fmt.Errorf("unsupported type: %T", v)
	}
}
