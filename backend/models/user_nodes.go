package models

import (
	"time"

	"gorm.io/gorm"
)

// UserNodes model holds info of reserved nodes of user
type UserNodes struct {
	ID         int       `gorm:"primaryKey;autoIncrement;column:id"`
	UserID     int       `gorm:"user_id" binding:"required"`
	ContractID uint64    `gorm:"contract_id" binding:"required"`
	NodeID     uint32    `gorm:"node_id;index:idx_user_node_id,unique" binding:"required"`
	CreatedAt  time.Time `json:"created_at"`
}

type GormUserNodesRepository struct {
	db *gorm.DB
}

func NewGormUserNodesRepository(db DB) *GormUserNodesRepository {
	return &GormUserNodesRepository{db: db.GetDB()}
}

// CreateUserNode creates new node record for user
func (r *GormUserNodesRepository) CreateUserNode(userNode *UserNodes) error {
	return r.db.Create(&userNode).Error
}

// DeleteUserNode deletes a node record for user by its contract ID
func (r *GormUserNodesRepository) DeleteUserNode(contractID uint64) error {
	return r.db.Where("contract_id = ?", contractID).Delete(&UserNodes{}).Error
}

// ListUserNodes returns all nodes records for user by its ID
func (r *GormUserNodesRepository) ListUserNodes(userID int) ([]UserNodes, error) {
	var userNodes []UserNodes
	return userNodes, r.db.Where("user_id = ?", userID).Find(&userNodes).Error
}

// ListAllReservedNodes returns all reserved nodes from all users
func (r *GormUserNodesRepository) ListAllReservedNodes() ([]UserNodes, error) {
	var userNodes []UserNodes
	return userNodes, r.db.Find(&userNodes).Error
}

func (r *GormUserNodesRepository) GetUserNodeByNodeID(nodeID uint64) (UserNodes, error) {
	var userNode UserNodes
	return userNode, r.db.Where("node_id = ?", nodeID).First(&userNode).Error
}

func (r *GormUserNodesRepository) GetUserNodeByContractID(contractID uint64) (UserNodes, error) {
	var userNode UserNodes
	return userNode, r.db.Where("contract_id = ?", contractID).First(&userNode).Error
}
