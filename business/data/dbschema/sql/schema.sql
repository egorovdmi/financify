-- Version: 1.1
-- Description: Create table users
CREATE TABLE users (
	user_id       UUID,
	name          TEXT,
	email         TEXT UNIQUE,
	roles         TEXT[],
	password_hash TEXT,
	date_created  TIMESTAMP,
	date_updated  TIMESTAMP,

	PRIMARY KEY (user_id)
);

-- Version: 1.2
-- Description: Create table calculations
CREATE TABLE calculations (
	calculation_id   UUID,
	name         TEXT,
	user_id      UUID,
	date_created TIMESTAMP,
	date_updated TIMESTAMP,

	PRIMARY KEY (calculation_id),
	FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Version: 1.3
-- Description: Create table calculation_parameters
CREATE TABLE calculation_parameters (
	parameter_id      UUID,
	title varchar,
    group_name varchar,
	calculation_id    UUID,
    operation varchar,
    repeat varchar,
    start_date varchar,
    end_date varchar,
    day_of_month int,
    dynamic_transaction_first boolean,
    amount money,
    currency varchar,
	date_created TIMESTAMP,
	date_updated TIMESTAMP,

	PRIMARY KEY (parameter_id),
	FOREIGN KEY (calculation_id) REFERENCES calculations(calculation_id) ON DELETE CASCADE
);

-- Version: 1.4
-- Description: Initial tables for payments service
CREATE TABLE scopes (
	scope_id UUID,
	user_id UUID,
	title varchar,
	amount money,
	date_created TIMESTAMP,
	date_updated TIMESTAMP,

	PRIMARY KEY (scope_id),
	FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
CREATE INDEX scopes_mm_idx ON scopes (user_id, scope_id);

CREATE TABLE wallets (
	wallet_id UUID,
	scope_id UUID,
	user_id UUID,
	title varchar,
	amount money,
	date_created TIMESTAMP,
	date_updated TIMESTAMP,

	PRIMARY KEY (wallet_id),
	FOREIGN KEY (scope_id) REFERENCES scopes(scope_id) ON DELETE CASCADE,
	FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
CREATE INDEX wallets_mm_idx ON wallets (user_id, scope_id, wallet_id);

CREATE TABLE payments (
	payment_id UUID,
	transaction_id UUID,
	user_id UUID,
	scope_id UUID,
	wallet_id UUID,
	product_name varchar,
	product_quantity int,
	product_type varchar,
	amount money,
	date_created TIMESTAMP,
	date_updated TIMESTAMP,

	PRIMARY KEY (scope_id),
	FOREIGN KEY (scope_id) REFERENCES scopes(scope_id) ON DELETE CASCADE,
	FOREIGN KEY (wallet_id) REFERENCES wallets(wallet_id) ON DELETE CASCADE,
	FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
CREATE INDEX payments_mm_idx ON payments (user_id, scope_id, wallet_id, payment_id);
