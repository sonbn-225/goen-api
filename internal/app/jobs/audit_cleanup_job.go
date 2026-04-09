package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
)

// AuditCleanupJob manages the background removal of old audit logs.
type AuditCleanupJob struct {
	auditRepo interfaces.AuditRepository
	interval  time.Duration
}

func NewAuditCleanupJob(auditRepo interfaces.AuditRepository) *AuditCleanupJob {
	return &AuditCleanupJob{
		auditRepo: auditRepo,
		interval:  24 * time.Hour, // Run once a day
	}
}

func (j *AuditCleanupJob) Start(ctx context.Context) {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	// Run once immediately
	j.run(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			j.run(ctx)
		}
	}
}

func (j *AuditCleanupJob) run(ctx context.Context) {
	slog.Info("running audit cleanup job")
	
	// Default retention: 30 days (this could be fetched from global config or per-user settings)
	// For simplicity in this unified architecture, we use a conservative default.
	retentionDays := 30
	before := time.Now().AddDate(0, 0, -retentionDays)

	deleted, err := j.auditRepo.DeleteOldLogs(ctx, nil, before)
	if err != nil {
		slog.Error("failed to cleanup old audit logs", "error", err)
		return
	}

	if deleted > 0 {
		slog.Info("cleaned up old audit logs", "count", deleted, "before", before.Format("2006-01-02"))
	}
}
