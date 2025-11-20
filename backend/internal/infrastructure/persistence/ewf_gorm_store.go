package persistence

import (
	"context"
	"encoding/json"
	"fmt"

	"kubecloud/internal/core/models"

	"github.com/xmonader/ewf"
	"gorm.io/gorm"
)

type GormEWFRepository struct {
	db *gorm.DB
}

func NewGormEWFRepository(db models.DB) *GormEWFRepository {
	return &GormEWFRepository{db: db.GetDB()}
}

type gormWorkflowRecord struct {
	UUID   string `gorm:"primaryKey;column:uuid"`
	Name   string `gorm:"column:name;not null;index"`
	Status string `gorm:"column:status;not null;index"`
	Data   []byte `gorm:"column:data;not null"`
}

type gormTemplateRecord struct {
	Name string `gorm:"primaryKey;column:name"`
	Data []byte `gorm:"column:data;not null"`
}

type serializableTemplate struct {
	Steps []ewf.Step `json:"steps"`
}

type QueueMetadataRecord struct {
	Name string `gorm:"primaryKey;column:name"`
	Data []byte `gorm:"column:data;not null"`
}

func NewGormStore(db *gorm.DB) *GormEWFRepository {
	return &GormEWFRepository{db: db}
}

func (r *GormEWFRepository) Setup() error {
	return r.db.AutoMigrate(&gormWorkflowRecord{}, &gormTemplateRecord{}, &QueueMetadataRecord{})
}

func (r *GormEWFRepository) SaveWorkflow(ctx context.Context, workflow *ewf.Workflow) error {
	data, err := json.Marshal(workflow)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow: %w", err)
	}

	gormWorkflow := gormWorkflowRecord{
		UUID:   workflow.UUID,
		Name:   workflow.Name,
		Status: string(workflow.Status),
		Data:   data,
	}

	return r.db.WithContext(ctx).Save(&gormWorkflow).Error
}

func (r *GormEWFRepository) LoadWorkflowByName(ctx context.Context, name string) (*ewf.Workflow, error) {
	var gormWorkflow gormWorkflowRecord
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&gormWorkflow).Error; err != nil {
		return nil, err
	}

	var workflow ewf.Workflow
	err := json.Unmarshal(gormWorkflow.Data, &workflow)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow: %w", err)
	}
	return &workflow, nil
}

func (r *GormEWFRepository) LoadWorkflowByUUID(ctx context.Context, uuid string) (*ewf.Workflow, error) {
	var gormWorkflow gormWorkflowRecord
	if err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&gormWorkflow).Error; err != nil {
		return nil, err
	}
	var workflow ewf.Workflow
	if err := json.Unmarshal(gormWorkflow.Data, &workflow); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow: %w", err)
	}
	return &workflow, nil
}

func (r *GormEWFRepository) ListWorkflowUUIDsByStatus(ctx context.Context, status ewf.WorkflowStatus) ([]string, error) {
	var uuids []string
	err := r.db.WithContext(ctx).
		Model(&gormWorkflowRecord{}).
		Where("status = ?", status).
		Pluck("uuid", &uuids).
		Error
	return uuids, err
}

func (r *GormEWFRepository) LoadWorkflowTemplate(ctx context.Context, name string) (*ewf.WorkflowTemplate, error) {
	var gormTemplate gormTemplateRecord
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&gormTemplate).Error; err != nil {
		return nil, err
	}

	var st serializableTemplate
	if err := json.Unmarshal(gormTemplate.Data, &st); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow template: %w", err)
	}
	return &ewf.WorkflowTemplate{Steps: st.Steps}, nil
}

func (r *GormEWFRepository) LoadAllWorkflowTemplates(ctx context.Context) (map[string]*ewf.WorkflowTemplate, error) {
	var gormTemplates []gormTemplateRecord
	err := r.db.WithContext(ctx).Find(&gormTemplates).Error
	if err != nil {
		return nil, err
	}

	templates := make(map[string]*ewf.WorkflowTemplate)
	for _, record := range gormTemplates {
		var st serializableTemplate
		if err := json.Unmarshal(record.Data, &st); err != nil {
			return nil, fmt.Errorf("failed to unmarshal workflow template: %w", err)
		}
		templates[record.Name] = &ewf.WorkflowTemplate{Steps: st.Steps}
	}
	return templates, nil
}

func (r *GormEWFRepository) SaveWorkflowTemplate(ctx context.Context, name string, template *ewf.WorkflowTemplate) error {
	st := serializableTemplate{Steps: template.Steps}
	data, err := json.Marshal(st)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow template: %w", err)
	}

	gormTemplate := gormTemplateRecord{
		Name: name,
		Data: data,
	}

	return r.db.WithContext(ctx).Save(&gormTemplate).Error
}

func (r *GormEWFRepository) SaveQueueMetadata(ctx context.Context, metadata *ewf.QueueMetadata) error {
	if metadata == nil {
		return fmt.Errorf("metadata cannot be nil")
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal queue metadata: %w", err)
	}

	gormMetadata := QueueMetadataRecord{
		Name: metadata.Name,
		Data: data,
	}
	return r.db.WithContext(ctx).Save(&gormMetadata).Error
}

func (r *GormEWFRepository) LoadAllQueueMetadata(ctx context.Context) ([]*ewf.QueueMetadata, error) {
	var gormQueuesMetadata []QueueMetadataRecord
	err := r.db.WithContext(ctx).Find(&gormQueuesMetadata).Error
	if err != nil {
		return nil, err
	}

	var queues []*ewf.QueueMetadata
	for _, record := range gormQueuesMetadata {
		var metadata ewf.QueueMetadata
		if err := json.Unmarshal(record.Data, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal queue metadata: %w", err)
		}
		queues = append(queues, &metadata)
	}
	return queues, nil
}

func (r *GormEWFRepository) DeleteQueueMetadata(ctx context.Context, name string) error {
	return r.db.WithContext(ctx).Delete(&QueueMetadataRecord{}, "name = ?", name).Error
}

func (r *GormEWFRepository) Close() error {
	return nil
}
