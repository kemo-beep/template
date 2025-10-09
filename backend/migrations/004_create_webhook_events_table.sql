-- Migration: Create webhook_events table
-- Description: Creates the webhook_events table to track webhook events from payment providers
-- Version: 004
-- Date: 2024-01-01

CREATE TABLE IF NOT EXISTS webhook_events (
    id SERIAL PRIMARY KEY,
    provider VARCHAR(20) NOT NULL CHECK (provider IN ('stripe', 'polar', 'paypal')),
    event_type VARCHAR(100) NOT NULL,
    event_id VARCHAR(255) NOT NULL UNIQUE,
    processed BOOLEAN DEFAULT FALSE,
    data JSONB NOT NULL,
    processed_at TIMESTAMP NULL,
    error TEXT,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_webhook_events_provider ON webhook_events(provider);
CREATE INDEX IF NOT EXISTS idx_webhook_events_event_type ON webhook_events(event_type);
CREATE INDEX IF NOT EXISTS idx_webhook_events_event_id ON webhook_events(event_id);
CREATE INDEX IF NOT EXISTS idx_webhook_events_processed ON webhook_events(processed);
CREATE INDEX IF NOT EXISTS idx_webhook_events_created_at ON webhook_events(created_at);
CREATE INDEX IF NOT EXISTS idx_webhook_events_processed_at ON webhook_events(processed_at);

-- Create partial indexes for unprocessed events
CREATE INDEX IF NOT EXISTS idx_webhook_events_unprocessed ON webhook_events(provider, created_at) WHERE processed = FALSE;

-- Add comments for documentation
COMMENT ON TABLE webhook_events IS 'Tracks webhook events from payment providers';
COMMENT ON COLUMN webhook_events.provider IS 'Payment provider that sent the webhook';
COMMENT ON COLUMN webhook_events.event_type IS 'Type of webhook event';
COMMENT ON COLUMN webhook_events.event_id IS 'Unique event ID from the provider';
COMMENT ON COLUMN webhook_events.processed IS 'Whether the event has been processed';
COMMENT ON COLUMN webhook_events.data IS 'Raw webhook event data';
COMMENT ON COLUMN webhook_events.processed_at IS 'When the event was processed';
COMMENT ON COLUMN webhook_events.error IS 'Error message if processing failed';
