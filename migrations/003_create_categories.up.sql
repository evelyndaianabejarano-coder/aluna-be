CREATE TABLE categories (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre       VARCHAR(100) NOT NULL,
    objetivo     VARCHAR(100) NOT NULL,
    descripcion  TEXT         NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_categories_nombre ON categories (nombre);
