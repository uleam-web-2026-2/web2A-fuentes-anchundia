# Verificación final · Taller 10

**Fecha:** 5 de noviembre de 2026. **Base revisada:** `4639fc2` de `origin/taller10`.

## Entorno

Windows amd64, Go 1.27.1, PostgreSQL 18.6, Bruno CLI 4.3.0, sqlc 1.31.1 y Redocly CLI 2.54.3. Compilé la API de este módulo y la inicié con configuraciones locales derivadas de `.env.example` y datos de demostración. Usé un clúster temporal en el puerto 55439, con una base vacía por adaptador y otra para las pruebas Go. La API escuchó en 18091 para sqlc y 18092 para GORM.

## Bruno

Ejecuté la colección `docs/hito2_bruno` completa, en orden y con la preparación 0. La referencia es la columna «Antes» de [peticiones.md](peticiones.md), registrada en la verificación previa del taller.

| Petición | Referencia | sqlc | GORM |
| --- | ---: | ---: | ---: |
| 0 · Crear canje para la prueba | 201 | 201 | 201 |
| 1 · Consultar sin token | 401 | 401 | 401 |
| 2 · Iniciar sesión como cliente | 200 | 200 | 200 |
| 3 · Consultar con token de cliente | 200 | 200 | 200 |
| 4 · Confirmar con rol cliente | 403 | 403 | 403 |
| 5 · Confirmar con rol administrador | 200 | 200 | 200 |
| 6 · Intentar cancelar un canje confirmado | 422 | 422 | 422 |

Cada corrida terminó con **7/7 peticiones, 6/6 pruebas y 7/7 aserciones** aprobadas. Las dos corridas usaron el código de `taller10` con distinto `REPOSITORY_ADAPTER`. La línea de base histórica se consultó en la evidencia previa.

### Captura en Bruno Desktop

Repetí el flujo desde el Collection Runner de Bruno Desktop 4.3.0, con el adaptador sqlc y PostgreSQL temporal. El Runner mostró **7 aprobadas, 0 fallidas y 0 omitidas**, con las seis pruebas y las siete aserciones visibles. La [captura del Runner](bruno_runner_final.jpg) muestra los códigos y los resultados de cada petición.

## Pruebas y revisión del código

- `go build ./...` y `go vet ./...` terminaron correctamente.
- `go test -count=1 -v ./...` aprobó **44 pruebas principales** con `LEALCAF_TEST_POSTGRES_DSN` apuntando a la base temporal. Incluyó DTO, servicio con repositorio falso y las dos pruebas PostgreSQL. La prueba de concurrencia pasó con los dos adaptadores.
- `sqlc generate` terminó correctamente y el paquete generado quedó sin diferencias en Git. Los tres SHA-256 coincidieron con [sqlc.txt](sqlc.txt).
- Redocly validó `docs/openapi.yaml` sin errores ni avisos.
- Revisé que `Servicio` dependa de `Repositorio`, que `Canje` se mapee a DTO para HTTP y que el adaptador sqlc use sus consultas para buscar y cambiar el estado. Actualicé [decisiones.md](../decisiones.md#semana-10--transacciones) con las condiciones de la confirmación y cómo se provoca el fallo transaccional.

## Fallo forzado en PostgreSQL

`TestTaller10TransaccionesPostgres` fuerza un `CHECK` en la inserción del movimiento, después de actualizar el canje o insertar la compra. PostgreSQL devolvió `SQLSTATE 23514` para ambos adaptadores.

| Caso | Resultado con GORM y sqlc |
| --- | --- |
| Confirmación con fallo | El canje siguió `solicitado`, con 0 descuentos y saldo 100. |
| Reintento de confirmación | Se registró 1 descuento y el saldo quedó en 60. |
| Compra con fallo | Quedaron 0 compras para la factura de prueba y el saldo siguió en 60. |
| Reintento de compra | Se acreditaron 50 puntos y el saldo quedó en 110. |

## Uso de IA

OpenAI Codex apoyó la comparación con la consigna, la ejecución local de Bruno y las pruebas, y la revisión de la documentación y sus enlaces.

## Canje inexistente con sqlc

En una base temporal de PostgreSQL con datos demo y `REPOSITORY_ADAPTER=sqlc`, inicié sesión como cliente y envié `GET /canjes/2147483647` con su token. El ID positivo era ajeno a los registros de la base de ensayo. La API devolvió **404** con `{"error":"canje no encontrado"}`.

Esta petición complementa las seis de Bruno. Verifica por HTTP que la ausencia de una fila de sqlc se traduce a un error de dominio y luego a un 404.
