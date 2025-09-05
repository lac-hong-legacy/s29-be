-- +goose Up
-- +goose StatementBegin

-- Users table (linked to Kratos identities)
CREATE TABLE users (
    id UUID PRIMARY KEY NOT NULL,
    kratos_identity_id UUID NOT NULL UNIQUE, -- Links to Kratos identity
    email VARCHAR(255) NOT NULL UNIQUE,
    is_active BOOLEAN DEFAULT true,

    xp_points BIGINT DEFAULT 0,
    streak_days INT DEFAULT 0,
    last_lesson_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin


DROP TABLE IF EXISTS users;

-- +goose StatementEnd 