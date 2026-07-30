-- Add optional unique email and phone fields to existing users tables.

BEGIN;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS email VARCHAR(254)
        CHECK (email IS NULL OR BTRIM(email) <> '');

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS phone VARCHAR(32)
        CHECK (phone IS NULL OR BTRIM(phone) <> '');

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_email
    ON users(LOWER(email)) WHERE email IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_phone
    ON users(phone) WHERE phone IS NOT NULL;

COMMENT ON COLUMN users.email IS '用户邮箱，可为空，填写后忽略大小写全局唯一';
COMMENT ON COLUMN users.phone IS '用户手机号，可为空，填写后全局唯一';

COMMIT;
