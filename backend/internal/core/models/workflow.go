package models

// GormWorkflowRecord holds the workflow record model
type GormWorkflowRecord struct {
	UUID      string `gorm:"primaryKey;column:uuid"`
	Name      string `gorm:"column:name;not null;index"`
	Status    string `gorm:"column:status;not null;index"`
	Data      []byte `gorm:"column:data;not null"`
	QueueName string `gorm:"column:queue_name;index"`
	UserID    int    `gorm:"column:user_id;index"`
}
