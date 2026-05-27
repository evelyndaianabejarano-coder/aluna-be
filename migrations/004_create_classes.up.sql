CREATE TABLE classes (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    titulo       VARCHAR(255) NOT NULL,
    descripcion  TEXT         NOT NULL,
    profesor_id  UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    categoria_id UUID         NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    fecha_hora   TIMESTAMPTZ  NOT NULL,
    duracion     INT          NOT NULL CHECK (duracion > 0),
    cupo_maximo  INT          NOT NULL CHECK (cupo_maximo > 0),
    modalidad    VARCHAR(20)  NOT NULL DEFAULT 'presencial'
                     CHECK (modalidad IN ('presencial', 'online')),
    estado       VARCHAR(20)  NOT NULL DEFAULT 'activa'
                     CHECK (estado IN ('activa', 'cancelada')),
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_classes_profesor_id  ON classes (profesor_id);
CREATE INDEX idx_classes_categoria_id ON classes (categoria_id);
CREATE INDEX idx_classes_fecha_hora   ON classes (fecha_hora);
CREATE INDEX idx_classes_estado       ON classes (estado);
