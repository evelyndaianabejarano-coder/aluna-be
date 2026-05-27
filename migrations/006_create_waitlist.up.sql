CREATE TABLE waitlist (
    id        UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    alumno_id UUID        NOT NULL REFERENCES users(id)   ON DELETE RESTRICT,
    clase_id  UUID        NOT NULL REFERENCES classes(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_waitlist_alumno_clase UNIQUE (alumno_id, clase_id)
);

CREATE INDEX idx_waitlist_clase_id ON waitlist (clase_id);
