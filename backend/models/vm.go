package models

import (
	"encoding/json"
	"kubecloud/kubedeployer"
	"time"
)

// VM represents a deployed vm in the system
type VM struct {
	ID          int       `gorm:"primaryKey;autoIncrement;column:id"`
	UserID      string    `gorm:"index:idx_vm_user_project,unique" json:"user_id" binding:"required"`
	ProjectName string    `gorm:"index:idx_vm_user_project,unique" json:"project_name" binding:"required"`
	Result      string    `gorm:"type:text" json:"result"`
	CreatedAt   time.Time `json:"created_at"`
}

// GetVMResult deserializes the Result field into a kubedeployer.VM
func (v *VM) GetVMResult() (kubedeployer.VM, error) {
	var vm kubedeployer.VM
	err := json.Unmarshal([]byte(v.Result), &vm)
	return vm, err
}

// SetVMResult serializes a kubedeployer.VM into the Result field
func (v *VM) SetVMResult(vm kubedeployer.VM) error {
	data, err := json.Marshal(vm)
	if err != nil {
		return err
	}
	v.Result = string(data)
	return nil
}
