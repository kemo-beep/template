-- Migration: Create payment_methods table
-- Description: Creates the payment_methods table to store user payment methods
-- Version: 003
-- Date: 2024-01-01

CREATE TABLE IF NOT EXISTS payment_methods (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('card', 'bank_account', 'paypal', 'apple_pay', 'google_pay')),
    is_default BOOLEAN DEFAULT FALSE,
    last4 VARCHAR(4),
    brand VARCHAR(20), -- visa, mastercard, amex, etc.
    exp_month INTEGER CHECK (exp_month >= 1 AND exp_month <= 12),
    exp_year INTEGER CHECK (exp_year >= 2020),
    metadata JSONB,
    
    -- External payment provider IDs
    stripe_payment_method_id VARCHAR(255) UNIQUE,
    polar_payment_method_id VARCHAR(255) UNIQUE,
    paypal_payment_method_id VARCHAR(255) UNIQUE,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_payment_methods_user_id ON payment_methods(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_methods_type ON payment_methods(type);
CREATE INDEX IF NOT EXISTS idx_payment_methods_is_default ON payment_methods(is_default);
CREATE INDEX IF NOT EXISTS idx_payment_methods_stripe_id ON payment_methods(stripe_payment_method_id);
CREATE INDEX IF NOT EXISTS idx_payment_methods_polar_id ON payment_methods(polar_payment_method_id);
CREATE INDEX IF NOT EXISTS idx_payment_methods_paypal_id ON payment_methods(paypal_payment_method_id);
CREATE INDEX IF NOT EXISTS idx_payment_methods_created_at ON payment_methods(created_at);

-- Create partial index for default payment methods
CREATE INDEX IF NOT EXISTS idx_payment_methods_default ON payment_methods(user_id) WHERE is_default = TRUE;

-- Add comments for documentation
COMMENT ON TABLE payment_methods IS 'Stores user payment methods for recurring payments';
COMMENT ON COLUMN payment_methods.user_id IS 'Reference to the user who owns this payment method';
COMMENT ON COLUMN payment_methods.type IS 'Type of payment method';
COMMENT ON COLUMN payment_methods.is_default IS 'Whether this is the default payment method';
COMMENT ON COLUMN payment_methods.last4 IS 'Last 4 digits of the card/account';
COMMENT ON COLUMN payment_methods.brand IS 'Card brand (visa, mastercard, etc.)';
COMMENT ON COLUMN payment_methods.exp_month IS 'Expiration month (1-12)';
COMMENT ON COLUMN payment_methods.exp_year IS 'Expiration year';
COMMENT ON COLUMN payment_methods.metadata IS 'Additional payment method metadata';
COMMENT ON COLUMN payment_methods.stripe_payment_method_id IS 'Stripe payment method ID';
COMMENT ON COLUMN payment_methods.polar_payment_method_id IS 'Polar payment method ID';
COMMENT ON COLUMN payment_methods.paypal_payment_method_id IS 'PayPal payment method ID';
