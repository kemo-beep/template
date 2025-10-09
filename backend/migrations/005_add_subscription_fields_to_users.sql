-- Migration: Add subscription fields to users table
-- Description: Adds subscription-related fields to the users table
-- Version: 005
-- Date: 2024-01-01

-- Add subscription status fields to users table
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS subscription_status VARCHAR(20) DEFAULT 'free' CHECK (subscription_status IN ('free', 'trial', 'active', 'canceled', 'past_due')),
ADD COLUMN IF NOT EXISTS is_pro BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS subscription_id INTEGER REFERENCES subscriptions(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS subscription_ends_at TIMESTAMP NULL,
ADD COLUMN IF NOT EXISTS trial_ends_at TIMESTAMP NULL;

-- Create indexes for the new fields
CREATE INDEX IF NOT EXISTS idx_users_subscription_status ON users(subscription_status);
CREATE INDEX IF NOT EXISTS idx_users_is_pro ON users(is_pro);
CREATE INDEX IF NOT EXISTS idx_users_subscription_id ON users(subscription_id);
CREATE INDEX IF NOT EXISTS idx_users_subscription_ends_at ON users(subscription_ends_at);
CREATE INDEX IF NOT EXISTS idx_users_trial_ends_at ON users(trial_ends_at);

-- Create partial indexes for active subscriptions
CREATE INDEX IF NOT EXISTS idx_users_pro_users ON users(id) WHERE is_pro = TRUE;
CREATE INDEX IF NOT EXISTS idx_users_active_subscriptions ON users(id) WHERE subscription_status IN ('active', 'trial');

-- Add comments for documentation
COMMENT ON COLUMN users.subscription_status IS 'Current subscription status of the user';
COMMENT ON COLUMN users.is_pro IS 'Whether the user has pro access';
COMMENT ON COLUMN users.subscription_id IS 'Reference to the active subscription';
COMMENT ON COLUMN users.subscription_ends_at IS 'When the subscription ends';
COMMENT ON COLUMN users.trial_ends_at IS 'When the trial period ends';
