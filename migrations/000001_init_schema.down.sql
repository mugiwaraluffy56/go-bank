-- Drop triggers
DROP TRIGGER IF EXISTS update_accounts_updated_at ON accounts;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order of creation (to handle foreign key constraints)
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS transfers;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS users;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";
