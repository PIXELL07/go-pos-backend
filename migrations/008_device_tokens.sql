-- Description: FCM device token table for push notifications

CREATE TABLE IF NOT EXISTS device_tokens (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT         NOT NULL,
    platform   VARCHAR(20)  NOT NULL DEFAULT 'android',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (token)
);

CREATE INDEX IF NOT EXISTS idx_device_tokens_user_id ON device_tokens(user_id) WHERE deleted_at IS NULL;
