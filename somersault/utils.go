package somersault

func getIntFromMap(m map[interface{}]interface{}, name interface{}) (int, bool) {
	if val, ok := m[name]; ok {
		if v, ok := val.(int); ok {
			return v, true
		}
	}
	return 0, false
}

func getStringFromMap(m map[interface{}]interface{}, name interface{}) (string, bool) {
	if val, ok := m[name]; ok {
		if v, ok := val.(string); ok {
			return v, true
		}
	}
	return "", false
}

func getStringFromMap(m map[interface{}]interface{}, name interface{}) (string, bool) {
	if val, ok := m[name]; ok {
		if v, ok := val.(string); ok {
			return v, true
		}
	}
	return "", false
}
