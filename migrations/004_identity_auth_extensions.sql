ALTER TABLE users
	ADD COLUMN IF NOT EXISTS group_id BIGINT,
	ADD COLUMN IF NOT EXISTS login_id TEXT,
	ADD COLUMN IF NOT EXISTS user_name TEXT,
	ADD COLUMN IF NOT EXISTS last_login TIMESTAMPTZ,
	ADD COLUMN IF NOT EXISTS updated_by TEXT,
	ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ;

DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1
		FROM pg_constraint
		WHERE conname = 'fk_users_group_id'
	) THEN
		ALTER TABLE users
			ADD CONSTRAINT fk_users_group_id
			FOREIGN KEY (group_id) REFERENCES account_groups(id);
	END IF;
END $$;

UPDATE users u
SET group_id = uag.group_id
FROM user_account_groups uag
WHERE u.id = uag.user_id
  AND u.group_id IS NULL;

UPDATE users
SET login_id = username
WHERE login_id IS NULL OR TRIM(login_id) = '';

UPDATE users
SET user_name = username
WHERE user_name IS NULL OR TRIM(user_name) = '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_login_id ON users(login_id);
CREATE INDEX IF NOT EXISTS idx_users_group_id ON users(group_id);
