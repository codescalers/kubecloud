package activities

import (
	"fmt"
	"kubecloud/internal/notification"
	"kubecloud/internal/workflow"

	"github.com/xmonader/ewf"
)

// CreateTemplateBuilder creates a template builder function with predefined notification and logging hooks
// gives priority to the hooks defined in the template itself over the predefined ones
func CreateTemplateBuilder(ns *notification.NotificationService) workflow.TemplateBuilderFunc {
	return func(
		steps []ewf.Step,
		afterWorkflow []ewf.AfterWorkflowHook,
		beforeWorkflow []ewf.BeforeWorkflowHook,
		afterStep []ewf.AfterStepHook,
		beforeStep []ewf.BeforeStepHook,
	) ewf.WorkflowTemplate {
		return ewf.WorkflowTemplate{
			Steps: steps,
			BeforeWorkflowHooks: append(beforeWorkflow,
				hookNotifyWorkflowStarted(ns),
				hookLogWorkflowStarted,
			),
			AfterWorkflowHooks: append(afterWorkflow,
				notifyWorkflowProgress(ns),
				hookLogWorkflowDone,
			),
			BeforeStepHooks: append(beforeStep,
				hookLogStepStarted,
			),
			AfterStepHooks: append(afterStep,
				notifyStepHook(ns),
				hookLogStepDone,
			),
		}
	}
}

func newKubecloudWorkflowTemplate(n *notification.NotificationService) ewf.WorkflowTemplate {
	return ewf.WorkflowTemplate{
		BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{
			hookNotifyWorkflowStarted(n),
		},
		AfterWorkflowHooks: []ewf.AfterWorkflowHook{
			hookLogWorkflowDone,
			notifyWorkflowProgress(n),
		},
		BeforeStepHooks: []ewf.BeforeStepHook{
			hookLogStepStarted,
		},
		AfterStepHooks: []ewf.AfterStepHook{
			hookLogStepDone,
		},
	}
}

func getOrdinalSuffix(n int) string {
	switch n % 100 {
	case 11, 12, 13:
		return "th"
	default:
		switch n % 10 {
		case 1:
			return "st"
		case 2:
			return "nd"
		case 3:
			return "rd"
		default:
			return "th"
		}
	}
}

func getDeployNodeStepName(index int) string {
	return fmt.Sprintf("deploy-%d%s-node", index, getOrdinalSuffix(index))
}
