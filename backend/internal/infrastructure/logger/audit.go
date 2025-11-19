package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
)

// AuditActorType represents who performed the action that we audit.
type AuditActorType string

const (
	AuditActorUser   AuditActorType = "user"
	AuditActorAdmin  AuditActorType = "admin"
	AuditActorSystem AuditActorType = "system"
)

// AuditActor captures information about the initiator of the action.
type AuditActor struct {
	Type     AuditActorType `json:"type"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// AuditActionType enumerates the supported audit events.
type AuditActionType string

const (
	// Authentication / user lifecycle
	AuditActionUserRegister             AuditActionType = "user.register"
	AuditActionUserLogin                AuditActionType = "user.login"
	AuditActionUserLogout               AuditActionType = "user.logout"
	AuditActionUserVerify               AuditActionType = "user.verify"
	AuditActionUserPasswordChange       AuditActionType = "user.password_change"
	AuditActionUserPasswordResetRequest AuditActionType = "user.password_reset.request"
	AuditActionUserPasswordResetVerify  AuditActionType = "user.password_reset.verify"
	AuditActionUserTokenRefresh         AuditActionType = "user.token_refresh"

	// Notifications
	AuditActionNotificationList       AuditActionType = "notification.list"
	AuditActionNotificationRead       AuditActionType = "notification.read"
	AuditActionNotificationDelete     AuditActionType = "notification.delete"
	AuditActionNotificationMarkUnread AuditActionType = "notification.mark.unread"
	AuditActionNotificationSend       AuditActionType = "notification.send"

	// Billing / balance
	AuditActionBalanceCharge      AuditActionType = "balance.charge"
	AuditActionBalanceGet         AuditActionType = "balance.get"
	AuditActionVoucherRedeem      AuditActionType = "voucher.redeem"
	AuditActionVoucherCreate      AuditActionType = "voucher.create"
	AuditActionInvoiceList        AuditActionType = "invoice.list"
	AuditActionInvoiceDownload    AuditActionType = "invoice.download"
	AuditActionPendingRecordsList AuditActionType = "pending_records.list"

	// Deployments
	AuditActionDeploymentList       AuditActionType = "deployment.list"
	AuditActionDeploymentGet        AuditActionType = "deployment.get"
	AuditActionDeploymentKubeconfig AuditActionType = "deployment.kubeconfig"
	AuditActionDeploymentDeploy     AuditActionType = "deployment.deploy"
	AuditActionDeploymentDelete     AuditActionType = "deployment.delete"
	AuditActionDeploymentDeleteAll  AuditActionType = "deployment.delete_all"
	AuditActionDeploymentAddNode    AuditActionType = "deployment.node.add"
	AuditActionDeploymentRemoveNode AuditActionType = "deployment.node.remove"

	// Nodes
	AuditActionNodeReserve      AuditActionType = "node.reserve"
	AuditActionNodeUnreserve    AuditActionType = "node.unreserve"
	AuditActionNodeListRentable AuditActionType = "node.list.rentable"

	// Workers
	AuditActionWorkerInvoiceProcess AuditActionType = "worker.invoice.process"
	AuditActionWorkerInvoiceCreate  AuditActionType = "worker.invoice.create"
	AuditActionWorkerDebtUpdate     AuditActionType = "worker.debt.update"
	AuditActionWorkerClusterHealth  AuditActionType = "worker.health.cluster"
	AuditActionWorkerNodeHealth     AuditActionType = "worker.health.node"
	AuditActionWorkerBalanceSettle  AuditActionType = "worker.balance.settle"
	AuditActionWorkerPendingNotify  AuditActionType = "worker.pending.notify"

	// Admin operations
	AuditActionAdminUsersList     AuditActionType = "admin.users.list"
	AuditActionAdminUserDelete    AuditActionType = "admin.user.delete"
	AuditActionAdminVoucherGen    AuditActionType = "admin.voucher.generate"
	AuditActionAdminVoucherList   AuditActionType = "admin.voucher.list"
	AuditActionAdminCreditUser    AuditActionType = "admin.user.credit"
	AuditActionAdminPendingList   AuditActionType = "admin.pending.list"
	AuditActionAdminMailSend      AuditActionType = "admin.mail.send"
	AuditActionAdminDrainUser     AuditActionType = "admin.user.drain"
	AuditActionAdminDrainAllUsers AuditActionType = "admin.users.drain_all"
	AuditActionAdminCreditBalance AuditActionType = "admin.credit.balance"
	AuditActionMaintenanceModeSet AuditActionType = "maintenance.mode.set"
	AuditActionMaintenanceModeGet AuditActionType = "maintenance.mode.get"
	// SSH Keys
	AuditActionSSHKeyList   AuditActionType = "ssh_key.list"
	AuditActionSSHKeyAdd    AuditActionType = "ssh_key.add"
	AuditActionSSHKeyDelete AuditActionType = "ssh_key.delete"

	// Workflows
	AuditActionWorkflowStatusGet AuditActionType = "workflow.status.get"
	AuditActionWorkflowList      AuditActionType = "workflow.user.list"
)

// AuditAction wraps the action type together with optional metadata.
type AuditAction struct {
	Type     AuditActionType `json:"type"`
	Metadata map[string]any  `json:"metadata,omitempty"`
}

// AuditSeverity represents severity level of the audit entry.
type AuditSeverity string

const (
	AuditSeverityInfo     AuditSeverity = "info"
	AuditSeverityWarning  AuditSeverity = "warning"
	AuditSeverityError    AuditSeverity = "error"
	AuditSeverityCritical AuditSeverity = "critical"
)

// AuditEntryOption configures optional fields on an audit event.
type AuditEntryOption func(e *auditEvent)

type auditEvent struct {
	actorMetadata  map[string]any
	actionMetadata map[string]any
	severity       AuditSeverity
	expiresAt      *time.Time
	customFields   map[string]any
}

// WithAuditSeverity overrides the default severity value.
func WithAuditSeverity(severity AuditSeverity) AuditEntryOption {
	return func(e *auditEvent) {
		e.severity = severity
	}
}

// WithAuditActorMetadata adds metadata about the actor.
func WithAuditActorMetadata(metadata map[string]any) AuditEntryOption {
	return func(e *auditEvent) {
		if len(metadata) == 0 {
			return
		}
		if e.actorMetadata == nil {
			e.actorMetadata = map[string]any{}
		}
		for k, v := range metadata {
			e.actorMetadata[k] = v
		}
	}
}

// WithAuditActionMetadata merges additional metadata into the action.
func WithAuditActionMetadata(metadata map[string]any) AuditEntryOption {
	return func(e *auditEvent) {
		if len(metadata) == 0 {
			return
		}
		if e.actionMetadata == nil {
			e.actionMetadata = map[string]any{}
		}
		for k, v := range metadata {
			e.actionMetadata[k] = v
		}
	}
}

// WithAuditExpiry sets an explicit expiration time.
func WithAuditExpiry(t time.Time) AuditEntryOption {
	return func(e *auditEvent) {
		expiry := t
		e.expiresAt = &expiry
	}
}

// LogAudit writes the audit event using the dedicated audit logger.
// This logger writes to separate sinks with higher retention than regular logs.
func LogAudit(actorType AuditActorType, actionType AuditActionType, ip string, userAgent string, opts ...AuditEntryOption) {
	log := GetAuditLogger()
	logAuditWithLogger(log, actorType, actionType, ip, userAgent, opts...)
}

func logAuditWithLogger(log *zerolog.Logger, actorType AuditActorType, actionType AuditActionType, ip string, userAgent string, opts ...AuditEntryOption) {
	if log == nil {
		return
	}

	ev := &auditEvent{
		severity: AuditSeverityInfo,
	}

	for _, opt := range opts {
		opt(ev)
	}

	event := selectAuditEvent(log, ev.severity).
		Str("audit.actor_type", string(actorType)).
		Str("audit.action_type", string(actionType)).
		Str("audit.ip_address", ip).
		Str("audit.user_agent", userAgent).
		Str("audit.severity", string(ev.severity)).
		Time("audit.created_at", time.Now().UTC())

	if ev.actorMetadata != nil {
		event = event.Interface("audit.actor_metadata", ev.actorMetadata)
	}
	if ev.actionMetadata != nil {
		event = event.Interface("audit.action_metadata", ev.actionMetadata)
	}
	if ev.expiresAt != nil {
		event = event.Time("audit.expires_at", *ev.expiresAt)
	}
	if ev.customFields != nil {
		for k, v := range ev.customFields {
			event = event.Interface(k, v)
		}
	}

	event.Msg("audit_event")
}

func selectAuditEvent(log *zerolog.Logger, severity AuditSeverity) *zerolog.Event {
	switch severity {
	case AuditSeverityWarning:
		return log.Warn()
	case AuditSeverityError:
		return log.Error()
	case AuditSeverityCritical:
		return log.Fatal()
	default:
		return log.Info()
	}
}

func setupAuditLogger(cfg *AuditLogConfig) zerolog.Logger {
	if cfg == nil {
		return zerolog.New(io.Discard)
	}
	if !cfg.Enabled {
		return zerolog.New(io.Discard)
	}
	if cfg.Sink == nil {
		return zerolog.New(io.Discard)
	}

	return zerolog.New(cfg.Sink).With().
		Timestamp().
		Str("log_type", "audit").
		Logger()
}

// NewFileAuditSink creates a file-based audit sink with higher retention.
func NewFileAuditSink(dir string, maxSize, maxBackups, maxAge int, compress bool) (io.Writer, error) {
	if dir == "" {
		dir = filepath.Join("logs", "audit")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory %s: %w", dir, err)
	}

	auditLogFile := filepath.Join(dir, "audit.log")
	rotator := &lumberjack.Logger{
		Filename:   auditLogFile,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   compress,
	}

	// Ensure the file is writable
	if _, err := os.OpenFile(auditLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err != nil {
		return nil, fmt.Errorf("failed to open audit log file %s: %w", auditLogFile, err)
	}

	fmt.Fprintf(os.Stderr, "Audit logger initialized with file output to: %s\n", auditLogFile)
	return rotator, nil
}
