-- Migration: Create subscriptions table
-- Description: Creates the subscriptions table to track user subscriptions
-- Version: 001
-- Date: 2024-01-01

CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    plan_id INTEGER REFERENCES plans(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'canceled', 'past_due', 'incomplete', 'incomplete_expired', 'trialing', 'paused')),
    current_period_start TIMESTAMP NOT NULL,
    current_period_end TIMESTAMP NOT NULL,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    canceled_at TIMESTAMP NULL,
    trial_start TIMESTAMP NULL,
    trial_end TIMESTAMP NULL,
    quantity INTEGER DEFAULT 1 CHECK (quantity > 0),
    metadata JSONB,
    
    -- External payment provider IDs
    stripe_subscription_id VARCHAR(255) UNIQUE,
    polar_subscription_id VARCHAR(255) UNIQUE,
    payment_method VARCHAR(50) CHECK (payment_method IN ('stripe', 'polar')),
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_product_id ON subscriptions(product_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_plan_id ON subscriptions(plan_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_stripe_id ON subscriptions(stripe_subscription_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_polar_id ON subscriptions(polar_subscription_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_payment_method ON subscriptions(payment_method);
CREATE INDEX IF NOT EXISTS idx_subscriptions_current_period_end ON subscriptions(current_period_end);
CREATE INDEX IF NOT EXISTS idx_subscriptions_created_at ON subscriptions(created_at);

-- Create partial indexes for active subscriptions
CREATE INDEX IF NOT EXISTS idx_subscriptions_active ON subscriptions(user_id) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_subscriptions_trialing ON subscriptions(user_id) WHERE status = 'trialing';

-- Add comments for documentation
COMMENT ON TABLE subscriptions IS 'Tracks user subscriptions to products and plans';
COMMENT ON COLUMN subscriptions.user_id IS 'Reference to the user who owns this subscription';
COMMENT ON COLUMN subscriptions.product_id IS 'Reference to the product being subscribed to';
COMMENT ON COLUMN subscriptions.plan_id IS 'Reference to the specific plan (optional)';
COMMENT ON COLUMN subscriptions.status IS 'Current status of the subscription';
COMMENT ON COLUMN subscriptions.current_period_start IS 'Start of the current billing period';
COMMENT ON COLUMN subscriptions.current_period_end IS 'End of the current billing period';
COMMENT ON COLUMN subscriptions.cancel_at_period_end IS 'Whether subscription will cancel at period end';
COMMENT ON COLUMN subscriptions.canceled_at IS 'When the subscription was canceled';
COMMENT ON COLUMN subscriptions.trial_start IS 'Start of trial period (if applicable)';
COMMENT ON COLUMN subscriptions.trial_end IS 'End of trial period (if applicable)';
COMMENT ON COLUMN subscriptions.quantity IS 'Number of units subscribed to';
COMMENT ON COLUMN subscriptions.metadata IS 'Additional subscription metadata';
COMMENT ON COLUMN subscriptions.stripe_subscription_id IS 'Stripe subscription ID for external tracking';
COMMENT ON COLUMN subscriptions.polar_subscription_id IS 'Polar subscription ID for external tracking';
COMMENT ON COLUMN subscriptions.payment_method IS 'Payment provider used for this subscription';
