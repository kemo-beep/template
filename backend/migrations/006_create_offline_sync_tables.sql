-- Migration: Create offline sync tables
-- Description: Tables for offline-first data synchronization

-- Offline operations queue
CREATE TABLE IF NOT EXISTS offline_operations (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    operation_id VARCHAR(255) NOT NULL,
    operation_type VARCHAR(50) NOT NULL, -- 'create', 'update', 'delete'
    table_name VARCHAR(100) NOT NULL,
    record_id VARCHAR(255),
    data JSONB,
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed'
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    
    UNIQUE(user_id, operation_id),
    INDEX idx_offline_operations_user_status (user_id, status),
    INDEX idx_offline_operations_created_at (created_at),
    INDEX idx_offline_operations_operation_type (operation_type)
);

-- Sync conflicts tracking
CREATE TABLE IF NOT EXISTS sync_conflicts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    table_name VARCHAR(100) NOT NULL,
    record_id VARCHAR(255) NOT NULL,
    local_data JSONB,
    server_data JSONB,
    conflict_type VARCHAR(50) NOT NULL, -- 'version_mismatch', 'concurrent_edit', 'deleted_modified'
    resolution_strategy VARCHAR(50), -- 'server_wins', 'client_wins', 'merge', 'manual'
    resolved_data JSONB,
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'resolved', 'ignored'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE,
    
    UNIQUE(user_id, table_name, record_id),
    INDEX idx_sync_conflicts_user_status (user_id, status),
    INDEX idx_sync_conflicts_created_at (created_at)
);

-- Data versioning for conflict detection
CREATE TABLE IF NOT EXISTS data_versions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    table_name VARCHAR(100) NOT NULL,
    record_id VARCHAR(255) NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    last_modified_by VARCHAR(50) NOT NULL, -- 'client' or 'server'
    last_modified_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    checksum VARCHAR(64), -- For data integrity
    
    UNIQUE(user_id, table_name, record_id),
    INDEX idx_data_versions_user_table (user_id, table_name),
    INDEX idx_data_versions_last_modified (last_modified_at)
);

-- Sync status tracking
CREATE TABLE IF NOT EXISTS sync_status (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_sync_at TIMESTAMP WITH TIME ZONE,
    sync_token VARCHAR(255), -- For incremental sync
    pending_operations_count INTEGER DEFAULT 0,
    conflicts_count INTEGER DEFAULT 0,
    is_online BOOLEAN DEFAULT true,
    last_online_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id),
    INDEX idx_sync_status_user (user_id),
    INDEX idx_sync_status_last_sync (last_sync_at)
);

-- Sync history for audit trail
CREATE TABLE IF NOT EXISTS sync_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sync_type VARCHAR(50) NOT NULL, -- 'full', 'incremental', 'conflict_resolution'
    operations_processed INTEGER DEFAULT 0,
    conflicts_resolved INTEGER DEFAULT 0,
    duration_ms INTEGER,
    success BOOLEAN DEFAULT true,
    error_message TEXT,
    sync_token VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_sync_history_user (user_id),
    INDEX idx_sync_history_created_at (created_at),
    INDEX idx_sync_history_sync_type (sync_type)
);

-- Add triggers for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_offline_operations_updated_at 
    BEFORE UPDATE ON offline_operations 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sync_status_updated_at 
    BEFORE UPDATE ON sync_status 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
