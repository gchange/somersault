package somersault

import "strconv"

func getIntFromMap(m map[string]interface{}, name string) (int, bool) {
	if val, ok := m[name]; ok {
		switch val.(type) {
		case float64:
			val, _ := val.(float64)
			return int(val), true
		case float32:
			val, _ := val.(float32)
			return int(val), true
		case int:
			val, _ := val.(int)
			return val, true
		case int8:
			val, _ := val.(int8)
			return int(val), true
		case int16:
			val, _ := val.(int16)
			return int(val), true
		case int32:
			val, _ := val.(int32)
			return int(val), true
		case int64:
			val, _ := val.(int64)
			return int(val), true
		case uint:
			val, _ := val.(uint)
			return int(val), true
		case uint8:
			val, _ := val.(uint8)
			return int(val), true
		case uint16:
			val, _ := val.(uint16)
			return int(val), true
		case uint32:
			val, _ := val.(uint32)
			return int(val), true
		case uint64:
			val, _ := val.(uint64)
			return int(val), true
		case string:
			val, _ := val.(string)
			if val, err := strconv.Atoi(val); err != nil {
				return val, true
			}
			return 0, false
		}
	}
	return 0, false
}

func getStringFromMap(m map[string]interface{}, name string) (string, bool) {
	if val, ok := m[name]; ok {
		if v, ok := val.(string); ok {
			return v, true
		}
	}
	return "", false
}
