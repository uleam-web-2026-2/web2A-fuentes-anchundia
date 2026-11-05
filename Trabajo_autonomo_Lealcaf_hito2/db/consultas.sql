-- Consulta de canjes y actualización condicional de estado.
-- name: BuscarCanje :one
SELECT id, sucursal_id, cliente_id, recompensa_id, puntos_utilizados, fecha, estado
FROM canjes WHERE id = $1;

-- name: BloquearCanje :one
SELECT id, sucursal_id, cliente_id, recompensa_id, puntos_utilizados, fecha, estado
FROM canjes WHERE id = $1 FOR UPDATE;

-- name: CambiarEstadoCanje :execrows
UPDATE canjes SET estado = sqlc.arg(hasta)
WHERE id = sqlc.arg(id) AND estado = sqlc.arg(desde);

-- Operaciones transaccionales de compra y confirmación.
-- name: BuscarClienteTransaccion :one
SELECT id, activo FROM clientes WHERE id = $1;

-- name: BloquearClienteTransaccion :one
SELECT id, activo FROM clientes WHERE id = $1 FOR UPDATE;

-- name: BuscarSucursalTransaccion :one
SELECT id, nombre, direccion, activa FROM sucursals WHERE id = $1;

-- name: SaldoTransaccion :one
SELECT COALESCE(SUM(puntos), 0)::bigint AS saldo
FROM movimiento_puntos
WHERE cliente_id = sqlc.arg(cliente_id)
  AND ((tipo = 'acumulacion' AND fecha_vencimiento > sqlc.arg(fecha)) OR tipo = 'canje');

-- name: ExisteFacturaTransaccion :one
SELECT EXISTS (SELECT 1 FROM compras WHERE numero_factura = $1);

-- name: CrearCompraTransaccion :one
INSERT INTO compras (sucursal_id, cliente_id, fecha, total, numero_factura)
VALUES ($1, $2, $3, $4, $5) RETURNING id;

-- name: CrearMovimientoTransaccion :one
INSERT INTO movimiento_puntos (cliente_id, compra_id, canje_id, tipo, puntos, fecha, fecha_vencimiento)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;
