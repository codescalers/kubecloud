package activities

import (
	"encoding/json"
	"fmt"

	"github.com/xmonader/ewf"
)

// getFromState is a generic helper function to extract and type-cast values from workflow state.
// It handles both direct type assertions and JSON-based conversions for serialized/deserialized values.
func getFromState[T any](state ewf.State, key string) (T, error) {
	value, ok := state[key]
	if !ok {
		var zero T
		return zero, fmt.Errorf("missing '%s' in state", key)
	}

	// Try direct type assertion first (for newly created values)
	if val, ok := value.(T); ok {
		return val, nil
	}

	// Handle the case where value was serialized/deserialized and became a map
	// Use JSON marshaling/unmarshaling to convert map to struct
	valueBytes, err := json.Marshal(value)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to marshal %s value: %w", key, err)
	}

	var result T
	if err := json.Unmarshal(valueBytes, &result); err != nil {
		var zero T
		return zero, fmt.Errorf("failed to unmarshal %s: %w", key, err)
	}

	return result, nil
}
