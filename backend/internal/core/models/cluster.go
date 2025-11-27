package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"

	"kubecloud/internal/deployment/kubedeployer"
)

// Cluster represents a deployed cluster in the system
type Cluster struct {
	ID          int            `gorm:"primaryKey;autoIncrement;column:id"`
	UserID      int            `gorm:"column:user_id;index" json:"user_id" binding:"required"`
	ProjectName string         `gorm:"column:project_name;uniqueIndex:idx_user_project,where:deleted_at IS NULL" json:"project_name" binding:"required"`
	Result      string         `gorm:"type:text" json:"result"` // JSON serialized kubedeployer.Cluster
	Kubeconfig  string         `gorm:"type:text" json:"kubeconfig"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
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
