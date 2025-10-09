-- Migration: Create payments table
-- Description: Creates the payments table to track payment transactions
-- Version: 002
-- Date: 2024-01-01

CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_id INTEGER REFERENCES subscriptions(id) ON DELETE SET NULL,
    product_id INTEGER REFERENCES products(id) ON DELETE SET NULL,
    amount INTEGER NOT NULL CHECK (amount > 0), -- Amount in cents
    currency VARCHAR(3) NOT NULL DEFAULT 'USD' CHECK (currency IN ('USD', 'EUR', 'GBP', 'CAD', 'AUD')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'succeeded', 'failed', 'canceled', 'refunded', 'partially_refunded')),
    payment_method VARCHAR(50) NOT NULL CHECK (payment_method IN ('stripe', 'polar', 'paypal', 'bank_transfer')),
    payment_intent_id VARCHAR(255) UNIQUE,
    transaction_id VARCHAR(255) UNIQUE,
    description TEXT,
    metadata JSONB,
    
    -- External payment provider IDs
    stripe_payment_intent_id VARCHAR(255) UNIQUE,
    stripe_charge_id VARCHAR(255) UNIQUE,
    polar_payment_id VARCHAR(255) UNIQUE,
    paypal_payment_id VARCHAR(255) UNIQUE,
    
    -- Refund information
    refunded_amount INTEGER DEFAULT 0 CHECK (refunded_amount >= 0),
    refunded_at TIMESTAMP NULL,
    refund_reason TEXT,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP NULL,
    failed_at TIMESTAMP NULL
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_subscription_id ON payments(subscription_id);
CREATE INDEX IF NOT EXISTS idx_payments_product_id ON payments(product_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_payment_method ON payments(payment_method);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at);
CREATE INDEX IF NOT EXISTS idx_payments_processed_at ON payments(processed_at);
CREATE INDEX IF NOT EXISTS idx_payments_stripe_intent_id ON payments(stripe_payment_intent_id);
CREATE INDEX IF NOT EXISTS idx_payments_stripe_charge_id ON payments(stripe_charge_id);
CREATE INDEX IF NOT EXISTS idx_payments_polar_id ON payments(polar_payment_id);
CREATE INDEX IF NOT EXISTS idx_payments_paypal_id ON payments(paypal_payment_id);

-- Create partial indexes for different payment statuses
CREATE INDEX IF NOT EXISTS idx_payments_succeeded ON payments(user_id, created_at) WHERE status = 'succeeded';
CREATE INDEX IF NOT EXISTS idx_payments_failed ON payments(user_id, created_at) WHERE status = 'failed';
CREATE INDEX IF NOT EXISTS idx_payments_pending ON payments(user_id, created_at) WHERE status = 'pending';

-- Add comments for documentation
COMMENT ON TABLE payments IS 'Tracks payment transactions for subscriptions and one-time purchases';
COMMENT ON COLUMN payments.user_id IS 'Reference to the user who made the payment';
COMMENT ON COLUMN payments.subscription_id IS 'Reference to the subscription (if applicable)';
COMMENT ON COLUMN payments.product_id IS 'Reference to the product being paid for';
COMMENT ON COLUMN payments.amount IS 'Payment amount in cents (e.g., 1000 = $10.00)';
COMMENT ON COLUMN payments.currency IS 'Currency code (ISO 4217)';
COMMENT ON COLUMN payments.status IS 'Current status of the payment';
COMMENT ON COLUMN payments.payment_method IS 'Payment method used';
COMMENT ON COLUMN payments.payment_intent_id IS 'Internal payment intent ID';
COMMENT ON COLUMN payments.transaction_id IS 'Internal transaction ID';
COMMENT ON COLUMN payments.description IS 'Payment description';
COMMENT ON COLUMN payments.metadata IS 'Additional payment metadata';
COMMENT ON COLUMN payments.stripe_payment_intent_id IS 'Stripe payment intent ID';
COMMENT ON COLUMN payments.stripe_charge_id IS 'Stripe charge ID';
COMMENT ON COLUMN payments.polar_payment_id IS 'Polar payment ID';
COMMENT ON COLUMN payments.paypal_payment_id IS 'PayPal payment ID';
COMMENT ON COLUMN payments.refunded_amount IS 'Amount refunded in cents';
COMMENT ON COLUMN payments.refunded_at IS 'When the refund was processed';
COMMENT ON COLUMN payments.refund_reason IS 'Reason for the refund';
COMMENT ON COLUMN payments.processed_at IS 'When the payment was processed';
COMMENT ON COLUMN payments.failed_at IS 'When the payment failed';
