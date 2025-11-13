package models

import (
	"encoding/json"
	"kubecloud/kubedeployer"
	"time"
)

// Cluster represents a deployed cluster in the system
type Cluster struct {
	ID          int       `gorm:"primaryKey;autoIncrement;column:id"`
	UserID      int       `gorm:"user_id;index" json:"user_id" binding:"required"`
	ProjectName string    `gorm:"project_name;uniqueIndex:idx_user_project" json:"project_name" binding:"required"`
	Result      string    `gorm:"type:text" json:"result"` // JSON serialized kubedeployer.Cluster
	Kubeconfig  string    `gorm:"type:text" json:"kubeconfig"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GetClusterResult deserializes the Result field into a kubedeployer.Cluster
func (c *Cluster) GetClusterResult() (kubedeployer.Cluster, error) {
	var cluster kubedeployer.Cluster
	err := json.Unmarshal([]byte(c.Result), &cluster)
	return cluster, err
}

// SetClusterResult serializes a kubedeployer.Cluster into the Result field
func (c *Cluster) SetClusterResult(cluster kubedeployer.Cluster) error {
	data, err := json.Marshal(cluster)
	if err != nil {
		return err
	}
	c.Result = string(data)
	return nil
}

func (c *Cluster) GetLeaderIP() (string, error) {
	cluster, err := c.GetClusterResult()
	if err != nil {
		return "", err
	}

	for _, node := range cluster.Nodes {
		if node.Type == kubedeployer.NodeTypeLeader {
			return node.IP, nil
		}
	}
	return "", nil
}
