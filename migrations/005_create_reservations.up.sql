CREATE TABLE reservations (
    id        UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    alumno_id UUID        NOT NULL REFERENCES users(id)   ON DELETE RESTRICT,
    clase_id  UUID        NOT NULL REFERENCES classes(id) ON DELETE RESTRICT,
    estado    VARCHAR(20) NOT NULL DEFAULT 'pendiente'
                  CHECK (estado IN ('pendiente', 'confirmada', 'cancelada')),
    asistio   BOOLEAN     DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reservations_alumno_id ON reservations (alumno_id);
CREATE INDEX idx_reservations_clase_id  ON reservations (clase_id);
CREATE INDEX idx_reservations_estado    ON reservations (estado);
