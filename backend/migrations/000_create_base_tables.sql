-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'usd',
    is_active BOOLEAN DEFAULT TRUE,
    is_recurring BOOLEAN DEFAULT FALSE,
    interval VARCHAR(50),
    interval_count INTEGER,
    trial_days INTEGER,
    metadata JSONB,
    stripe_product_id VARCHAR(255),
    polar_product_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create plans table
CREATE TABLE IF NOT EXISTS plans (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    price BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'usd',
    interval VARCHAR(50) NOT NULL,
    interval_count INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT TRUE,
    trial_days INTEGER,
    metadata JSONB,
    stripe_price_id VARCHAR(255),
    polar_plan_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for products
CREATE INDEX IF NOT EXISTS idx_products_is_active ON products(is_active);
CREATE INDEX IF NOT EXISTS idx_products_is_recurring ON products(is_recurring);
CREATE INDEX IF NOT EXISTS idx_products_stripe_product_id ON products(stripe_product_id);
CREATE INDEX IF NOT EXISTS idx_products_polar_product_id ON products(polar_product_id);

-- Create indexes for plans
CREATE INDEX IF NOT EXISTS idx_plans_product_id ON plans(product_id);
CREATE INDEX IF NOT EXISTS idx_plans_is_active ON plans(is_active);
CREATE INDEX IF NOT EXISTS idx_plans_stripe_price_id ON plans(stripe_price_id);
CREATE INDEX IF NOT EXISTS idx_plans_polar_plan_id ON plans(polar_plan_id);
