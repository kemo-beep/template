# Database Migrations

This directory contains SQL migration files for the database schema. These migrations are designed to be run in order and are idempotent.

## Migration Files

1. **001_create_subscriptions_table.sql** - Creates the `subscriptions` table
2. **002_create_payments_table.sql** - Creates the `payments` table  
3. **003_create_payment_methods_table.sql** - Creates the `payment_methods` table
4. **004_create_webhook_events_table.sql** - Creates the `webhook_events` table
5. **005_add_subscription_fields_to_users.sql** - Adds subscription fields to the `users` table

## Running Migrations

### Option 1: Using Make (Recommended)

```bash
# Run SQL migrations (recommended for production)
make db-migrate-sql

# Check migration status
make db-status

# Reset database and run migrations
make db-reset
```

### Option 2: Direct Go Command

```bash
cd backend
go run scripts/run_migrations.go
```

### Option 3: GORM AutoMigrate (Development Only)

```bash
# This will auto-migrate all models (less control)
make db-migrate
```

## Migration Script

The `scripts/run_migrations.go` script:
- Reads all `.sql` files from the `migrations/` directory
- Sorts them alphabetically to ensure correct order
- Executes each migration in sequence
- Provides detailed logging of the migration process
- Stops on first error to prevent partial migrations

## Database Schema

### Subscriptions Table
- Tracks user subscriptions to products/plans
- Links to users, products, and plans
- Supports both Stripe and Polar payment providers
- Includes trial period support

### Payments Table
- Records all payment transactions
- Links to users, products, and subscriptions
- Supports multiple payment methods
- Tracks payment status and metadata

### Payment Methods Table
- Stores user payment methods (cards, etc.)
- Links to users
- Supports both Stripe and Polar payment providers
- Includes card details (last4, brand, expiry)

### Webhook Events Table
- Logs all webhook events from payment providers
- Tracks processing status
- Stores event data and any errors
- Supports both Stripe and Polar webhooks

### Users Table (Updated)
- Added subscription status fields
- Links to active subscription
- Tracks trial periods
- Includes pro user status

## Best Practices

1. **Always backup** your database before running migrations in production
2. **Test migrations** on a copy of production data first
3. **Run migrations** during maintenance windows
4. **Monitor** the migration process for any errors
5. **Verify** the schema after migration completion

## Troubleshooting

### Migration Fails
- Check database connection
- Verify SQL syntax in migration files
- Check for conflicting data
- Review error logs for specific issues

### Partial Migration
- The script stops on first error
- Fix the issue and re-run
- Consider manual cleanup if needed

### Rollback
- SQL migrations don't include rollback scripts
- Use database backups for rollback
- Consider implementing rollback scripts for critical changes

## Development vs Production

- **Development**: Use either SQL migrations or GORM AutoMigrate
- **Production**: Always use SQL migrations for better control and audit trail
- **CI/CD**: Use SQL migrations for consistent deployments

## Adding New Migrations

1. Create a new `.sql` file with the next number (e.g., `006_...`)
2. Use descriptive names for the migration
3. Test the migration on a copy of production data
4. Update this README if needed
5. Consider adding rollback scripts for critical changes
