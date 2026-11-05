-- Descripción para sqlc; AutoMigrate crea las tablas reales.
-- Los nombres son los de GORM (incluidos sucursals y movimiento_puntos).
CREATE TABLE clientes (
    id bigserial PRIMARY KEY,
    cedula varchar(10) NOT NULL UNIQUE,
    nombre text NOT NULL,
    correo text NOT NULL UNIQUE,
    telefono text,
    fecha_registro timestamptz,
    activo boolean NOT NULL DEFAULT true
);
CREATE TABLE sucursals (
    id bigserial PRIMARY KEY,
    nombre text NOT NULL,
    direccion text NOT NULL,
    activa boolean NOT NULL DEFAULT true
);
CREATE TABLE compras (
    id bigserial PRIMARY KEY,
    sucursal_id bigint NOT NULL DEFAULT 1,
    cliente_id bigint NOT NULL,
    fecha timestamptz NOT NULL,
    total numeric NOT NULL,
    numero_factura text NOT NULL UNIQUE
);
CREATE TABLE canjes (
    id bigserial PRIMARY KEY,
    sucursal_id bigint NOT NULL DEFAULT 1,
    cliente_id bigint NOT NULL,
    recompensa_id bigint NOT NULL,
    puntos_utilizados bigint NOT NULL,
    fecha timestamptz NOT NULL,
    estado text NOT NULL DEFAULT 'solicitado'
);
CREATE TABLE movimiento_puntos (
    id bigserial PRIMARY KEY,
    cliente_id bigint NOT NULL,
    compra_id bigint,
    canje_id bigint,
    tipo text NOT NULL,
    puntos bigint NOT NULL,
    fecha timestamptz NOT NULL,
    fecha_vencimiento timestamptz
);
