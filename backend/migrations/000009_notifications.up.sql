-- Migration 000009: Notification Engine
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(50) NOT NULL, -- SIMAK_SUBMITTED, APPROVAL_REQUIRED, SIMAK_APPROVED, SIMAK_REJECTED, REVISION_REQUIRED, SIMAK_ISSUED, SIMAK_EXPIRING, SIMAK_EXPIRED, INSPECTION_COMPLETED, DISCREPANCY_CREATED, DISCREPANCY_RESOLVED
    reference_id VARCHAR(100),
    reference_type VARCHAR(50), -- simak, discrepancy, inspection
    action_url VARCHAR(255),
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
