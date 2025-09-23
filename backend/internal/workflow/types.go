package workflow

import "github.com/xmonader/ewf"

// TemplateBuilderFunc defines the signature for creating workflow templates with base hooks
type TemplateBuilderFunc func(
	steps []ewf.Step,
	afterWorkflow []ewf.AfterWorkflowHook,
	beforeWorkflow []ewf.BeforeWorkflowHook,
	afterStep []ewf.AfterStepHook,
	beforeStep []ewf.BeforeStepHook,
) ewf.WorkflowTemplate
