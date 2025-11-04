package models

import (
	"encoding/json"
	"kubecloud/kubedeployer"
	"time"

	"gorm.io/gorm"
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

type GormClusterRepository struct {
	db *gorm.DB
}

func NewGormClusterRepository(db DB) *GormClusterRepository {
	return &GormClusterRepository{db: db.GetDB()}
}

// CreateCluster creates a new cluster in the database
func (r *GormClusterRepository) CreateCluster(userID int, cluster *Cluster) error {
	cluster.CreatedAt = time.Now()
	cluster.UpdatedAt = time.Now()
	cluster.UserID = userID
	return r.db.Create(cluster).Error
}

// ListUserClusters returns all clusters for a specific user
func (r *GormClusterRepository) ListUserClusters(userID int) ([]Cluster, error) {
	var clusters []Cluster
	query := r.db.Where("user_id = ?", userID).Find(&clusters)
	return clusters, query.Error
}

// GetClusterByName returns a cluster by name for a specific user
func (r *GormClusterRepository) GetClusterByName(userID int, projectName string) (Cluster, error) {
	var cluster Cluster
	query := r.db.Where("user_id = ? AND project_name = ?", userID, projectName).First(&cluster)
	return cluster, query.Error
}

// UpdateCluster updates an existing cluster
func (r *GormClusterRepository) UpdateCluster(cluster *Cluster) error {
	cluster.UpdatedAt = time.Now()
	return r.db.Model(&Cluster{}).
		Where("user_id = ? AND project_name = ?", cluster.UserID, cluster.ProjectName).
		Updates(cluster).Error
}

// DeleteCluster deletes a cluster by name for a specific user
func (r *GormClusterRepository) DeleteCluster(userID int, projectName string) error {
	return r.db.Where("user_id = ? AND project_name = ?", userID, projectName).Delete(&Cluster{}).Error
}

// DeleteAllUserClusters deletes all clusters for a specific user
func (r *GormClusterRepository) DeleteAllUserClusters(userID int) error {
	return r.db.Where("user_id = ?", userID).Delete(&Cluster{}).Error
}

// CountAllClusters returns the total number of clusters in the system
func (r *GormClusterRepository) CountAllClusters() (int64, error) {
	var count int64
	err := r.db.Model(&Cluster{}).Count(&count).Error
	return count, err
}

func (r *GormClusterRepository) ListAllClusters() ([]Cluster, error) {
	var clusters []Cluster
	return clusters, r.db.Find(&clusters).Error
}
