package kubedeployer

import (
	"encoding/json"
	"testing"
)

func TestNodeUnmarshalBackwardCompatibility(t *testing.T) {
	// Old format JSON with disk_size
	oldJSON := `{
		"name": "testnode",
		"type": "worker",
		"node_id": 123,
		"cpu": 2,
		"memory": 4096,
		"root_size": 10240,
		"disk_size": 20480
	}`

	var node Node
	err := json.Unmarshal([]byte(oldJSON), &node)
	if err != nil {
		t.Fatalf("Failed to unmarshal old JSON format: %v", err)
	}

	if len(node.DataDisks) != 1 {
		t.Errorf("Expected 1 data disk, got %d", len(node.DataDisks))
	}

	if len(node.DataDisks) > 0 && node.DataDisks[0] != 20480 {
		t.Errorf("Expected disk size 20480, got %d", node.DataDisks[0])
	}

	// New format JSON with data_disks
	newJSON := `{
		"name": "testnode",
		"type": "worker",
		"node_id": 123,
		"cpu": 2,
		"memory": 4096,
		"root_size": 10240,
		"data_disks": [10240, 20480]
	}`

	var newNode Node
	err = json.Unmarshal([]byte(newJSON), &newNode)
	if err != nil {
		t.Fatalf("Failed to unmarshal new JSON format: %v", err)
	}

	if len(newNode.DataDisks) != 2 {
		t.Errorf("Expected 2 data disks, got %d", len(newNode.DataDisks))
	}

	if len(newNode.DataDisks) > 0 && newNode.DataDisks[0] != 10240 {
		t.Errorf("Expected first disk size 10240, got %d", newNode.DataDisks[0])
	}
}
