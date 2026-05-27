CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email        VARCHAR(255) NOT NULL,
    password_hash TEXT        NOT NULL,
    role         VARCHAR(20)  NOT NULL DEFAULT 'alumno'
                     CHECK (role IN ('alumno', 'profesor', 'admin')),
    nombre       VARCHAR(255) NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_users_email ON users (email);
