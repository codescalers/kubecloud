package workflows

import "fmt"

// toInt safely converts various numeric types to int
// Handles int, float64, and int64 types commonly found in workflow state after JSON unmarshaling
func toInt(val interface{}) (int, error) {
	switch v := val.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case int64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", val)
	}
}

// toUint32 safely converts various numeric types to uint32
// Handles uint32, float64, int64, and int types commonly found in workflow state
func toUint32(val interface{}) (uint32, error) {
	switch v := val.(type) {
	case uint32:
		return v, nil
	case float64:
		return uint32(v), nil
	case int64:
		return uint32(v), nil
	case int:
		return uint32(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to uint32", val)
	}
}

// toUint64 safely converts various numeric types to uint64
// Handles uint64, float64, int64, and int types commonly found in workflow state
func toUint64(val interface{}) (uint64, error) {
	switch v := val.(type) {
	case uint64:
		return v, nil
	case float64:
		return uint64(v), nil
	case int64:
		return uint64(v), nil
	case int:
		return uint64(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to uint64", val)
	}
}
