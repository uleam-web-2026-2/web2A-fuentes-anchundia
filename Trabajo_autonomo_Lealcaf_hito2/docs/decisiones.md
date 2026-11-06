# Decisiones del proyecto · LealCaf

## Semana 8 · estructura

El módulo Go de esta copia está directamente en `Trabajo_autonomo_Lealcaf_hito2/`, junto con `docs/`, `scripts/` y `redocly.yaml`. Desde esa carpeta se ejecuta `go test ./...`. No hay una subcarpeta `hito2/`.

Elegimos capas lógicas dentro de `internal/fidelizacion`, conservando un solo paquete Go para un proyecto pequeño. Los archivos separan responsabilidades; no necesitamos un paquete por cada archivo.

| Archivo | Responsabilidad |
|---|---|
| `modelos.go` | Entidades, relaciones, roles y estados del dominio. |
| `db.go` | Conexión PostgreSQL, migración conservadora de las seis entidades y Cuenta, y carga opcional de datos de demostración. |
| `autenticacion.go` | Contraseñas con hash, emisión/verificación del JWT e identidad autenticada. La autorización por rol se aplica con SoloRol en manejadores.go. |
| `rutas.go` | Registro de rutas, grupo protegido y asociación con manejadores. |
| `manejadores.go` | Recibir HTTP, leer ID/cuerpo, llamar al servicio y convertir el resultado a JSON y código HTTP. No consulta la base directamente. |
| `servicio.go` | Casos de uso y reglas de negocio: puntos vigentes, permisos de propiedad y transiciones de Canje. No conoce Gin ni códigos HTTP. |
| `repositorio.go` | En Semana 8 contenía GORM; en Semana 10 declara interfaces del negocio. Los adaptadores están en `repositorio_gorm.go` y `repositorio_sqlc.go`. |

Flujo: **ruta → autenticación → manejador HTTP → servicio → repositorio → base de datos**. El resultado vuelve por el mismo camino; el manejador traduce los errores a HTTP. El servicio decide si se puede confirmar o cancelar y el repositorio guarda el cambio junto con el movimiento de puntos dentro de una transacción.

Para demostrar las capas, abrimos `manejadores.go` y `servicio.go`: el primero interpreta la petición; el segundo contiene la regla del cambio de estado. Si cambiamos PostgreSQL por otra base, revisamos `db.go`, el adaptador de persistencia y los detalles de mapeo GORM en los modelos. La regla de estados y la forma de la petición no deberían cambiar. Las pruebas deben confirmar ese resultado; no asumimos portabilidad automática de cualquier motor.

## Estado del canje y códigos HTTP

- Nace `solicitado`, sin descontar puntos.
- Solo `administrador` puede pasar `solicitado → confirmado` o `solicitado → cancelado`.
- Al confirmar se vuelve a comprobar el saldo; se genera un movimiento negativo y se guarda el estado confirmado de manera atómica.
- `confirmado` y `cancelado` son finales. No se reabren ni se cancelan canjes confirmados; una corrección requiere otro movimiento de negocio.
- `401`: falta autenticación válida. `403`: la identidad es válida pero no tiene permiso. `400`: petición mal formada o destino desconocido. `422`: la petición se entiende y el rol puede usar la ruta, pero incumple la regla del negocio. La transición antes documentada como `409` pasa a `422` para el Hito 2. `409` se conserva para una factura duplicada.

## Dos roles y cuentas separadas

Los únicos roles son `cliente` y `administrador`. `admin` es el nombre de una cuenta de prueba con rol administrador. Una sucursal no es un rol.

`Cuenta` guarda usuario, hash de contraseña, rol y un `ClienteID` opcional. Un cliente autenticado obtiene su identidad del JWT; no puede elegirla mediante `X-Cliente-ID`. El administrador no necesita estar asociado a un Cliente. No hay registro ni CRUD de usuarios en este hito.

El JWT usa HS256, dura una hora e incluye `sub`, `rol`, `cliente_id`, `iat`, `exp`, `iss` y `aud`. La clave se recibe por configuración y no se versiona. El contenido del token puede leerse; la firma permite detectar alteraciones. Cambiar el rol en el token sin volver a firmarlo produce `401`.

## Sexta entidad · Sucursal

La cafetería sigue siendo un único comercio con varias sucursales, no un programa de coalición. Agregamos `Sucursal` (`ID`, `Nombre`, `Direccion`, `Activa`), relacionada 1:N con Compra y 1:N con Canje mediante `SucursalID`.

Los puntos son compartidos entre sucursales: pertenecen al cliente. La sucursal registra dónde ocurrió la compra o se pidió la recompensa y debe existir y estar activa. No dividimos el saldo por sucursal. El catálogo de recompensas es común. Las rutas para listar/ver sucursales están declaradas como `(plan)` en OpenAPI, sin programar un CRUD adicional para este hito.

## Alcance del Hito 2 y conservación de datos

Canje es la entidad principal evaluada. Conservamos las rutas de compras, puntos y recompensas ya implementadas; la incorporación de Sucursal no añade endpoints implementados. El contrato incluye todo el mapa existente y distingue los planes futuros.

La configuración está fuera del código en .env. El motor de ejecución es PostgreSQL y se mantiene la base **fidelizacion_db** existente en localhost:5432. La variante SQLite anterior queda histórica para el arranque; SQLite solo se conserva en pruebas automáticas aisladas.

La preparación de esquema y datos se ejecuta de forma transaccional. Conserva los registros anteriores y agrega las tablas y columnas necesarias. Los datos ficticios se cargan solo con DEMO_DATA=true, en un cliente dedicado de correo cliente.hito2@example.invalid. No se presume que un ID anterior esté libre ni se cambian estados o saldos de clientes históricos. Reiniciar no duplica los puntos demo ni renueva su fecha de vencimiento.

Bruno obtiene clienteId del login, recompensaId del catálogo y canjeConfirmadoId del listado autorizado del administrador. Por eso las peticiones no dependen de cliente 1 ni canje 1. La petición 0 crea otro canje solicitado y guarda su canjeId. Tokens e identificadores viven solo en memoria, sin persistirse en el entorno.

## Declaración de IA · Hito 2

En la revisión del 29/10/2026 se corrigieron las referencias heredadas a `hito2/`: el módulo está directamente en la carpeta abierta. El iniciador ahora ejecuta siempre `go run .` para evitar demostrar un binario antiguo. Se conservan funcionalidades, reglas y datos.

`TestPostgresConfirmacionesConcurrentes`, en `postgres_test.go`, comprueba sobre una base desechable la migración idempotente y dos confirmaciones simultáneas con saldo para una sola: una se acepta, otra se rechaza, queda saldo 20 y un único descuento. Usa un esquema propio y limpia solo ese esquema. Se activa con `LEALCAF_TEST_POSTGRES_DSN` en formato clave=valor; sin esa variable se omite explícitamente.

La concurrencia se verificó en PostgreSQL 18.6 temporal, con API en 18082. Después de configurar el entorno privado se verificaron PostgreSQL local y la API en localhost:8080. Los registros, resultados y límites se describen en `validacion.md`.

**OpenAI Codex:** apoyo en implementación, reorganización de capas, autenticación, ampliación con Sucursal, pruebas, contrato OpenAPI, colección Bruno y revisión de documentación. Los integrantes deben leer, verificar y comprender el resultado, adaptar las decisiones y poder explicarlas. Esta declaración no acredita commits ni participación que no hayan ocurrido.

## Documentación de esta entrega

El modelo y mapa vigentes están en [hito2_modelo.md](hito2_modelo.md) y la configuración técnica en [hito2_tecnico.md](hito2_tecnico.md). Se retiraron la presentación, boceto y capturas de la entrega anterior; esta limpieza no elimina entidades, reglas, pruebas de migración ni datos PostgreSQL. ChatGPT y Claude fueron declarados como apoyo en el trabajo inicial; OpenAI Codex apoyó la revisión y preparación del Hito 2.

## Semana 10 · transacciones

### D1 · El caso

Nuestro caso principal es **confirmar un canje de LealCaf**. `Servicio.ActualizarEstadoCanje` cambia una fila de `canjes` de `solicitado` a `confirmado` e inserta el descuento en `movimiento_puntos`. Si se guardara el estado y fallara el descuento, el canje parecería entregado pero el cliente conservaría sus puntos. Si se descontara primero y fallara el estado, perdería puntos por un canje todavía solicitado.

El Hito 2 ya usaba una transacción GORM, pero el servicio dependía del tipo concreto `*Repositorio`. Esta semana conservamos la atomicidad al introducir la interfaz y sqlc. No creamos la tabla Pago del ejemplo de canchas: usamos las dos tablas reales de nuestro negocio.

También `RegistrarCompra` inserta en `compras` y, cuando genera al menos un punto, en `movimiento_puntos`. Conservamos su atomicidad para evitar una factura registrada sin puntos. Solicitar un canje solo inserta la solicitud: el descuento ocurre al confirmar.

### D2 · Opciones que consideramos

1. **Abrir la transacción en el servicio con GORM o `database/sql`.** Agruparía las escrituras, pero el servicio conocería `*gorm.DB` o `*sql.Tx`, tendría que construir el repositorio transaccional y cambiar al sustituir el adaptador. Rompe la restricción de no importar esas herramientas y dificulta probarlo con un falso sin base.
2. **Añadir una operación atómica `ConfirmarYDescontar` al repositorio.** El adaptador podría guardar ambos cambios. En LealCaf también hay que coordinar el bloqueo del cliente, el saldo vigente y la transición. Trasladar todo eso al adaptador duplicaría reglas entre GORM y sqlc; guardar solo las escrituras dejaría las comprobaciones fuera de la transacción. Las compras necesitarían otra operación específica.
3. **Un callback con una interfaz de repositorio transaccional.** `EnTransaccion(func(RepositorioTransaccion) error)` abre y cierra la transacción en el adaptador. El servicio conserva las reglas y utiliza el repositorio del callback para todas las lecturas y escrituras. Un error revierte; terminar correctamente permite confirmar. La interfaz puede implementarse con un falso sin exponer herramientas de persistencia.

### D3 · La que elegimos y por qué

Elegimos el callback con `RepositorioTransaccion`: sirve para los dos casos reales y mantiene las reglas juntas en `servicio.go`. `Repositorio` declara las operaciones del servicio y la entrada `EnTransaccion`. El puerto transaccional ofrece solo los métodos que necesitan compras y confirmaciones, sin permitir abrir otra transacción desde el callback.

La confirmación conserva el orden de concurrencia del proyecto: leer el canje, bloquear su cliente, volver a leer el canje con bloqueo, validar la transición y actualizar el estado solo si sigue en `solicitado`. Para confirmar, antes del cambio se comprueba que el cliente siga activo y tenga saldo suficiente. Después del cambio se inserta el descuento. Para cancelar, se guarda el estado `cancelado` sin generar un movimiento. Los canjes de un cliente comparten el bloqueo del cliente, para que dos confirmaciones no gasten el mismo saldo. Las comprobaciones y escrituras usan el repositorio recibido en el callback.

`RepositorioGORM.EnTransaccion` utiliza `DB.Transaction` y entrega una instancia con `DB: tx`. Si el callback devuelve un error, GORM revierte la transacción y lo devuelve al servicio. El servicio no importa `gorm` ni `database/sql`. El falso de `servicio_falso_test.go` trabaja con copias y aplica los datos solo cuando el callback termina bien, sin abrir una base.

### D4 · La que descartamos y por qué

Descartamos la transacción concreta en el servicio porque habría que cambiar `ActualizarEstadoCanje`, `RegistrarCompra` y sus pruebas al seleccionar sqlc. También descartamos mover toda la confirmación al adaptador: duplicaría las decisiones sobre estados y puntos entre las dos implementaciones.

La opción elegida exige implementar realmente todos los métodos transaccionales. No basta con incrustar GORM y reutilizarlo dentro de una transacción SQL: esas llamadas podrían ejecutarse en otra conexión y quedar fuera de su rollback.

### D5 · Con sqlc

Fuera de las transacciones, `RepositorioSqlc` incrusta `RepositorioGORM` y sustituye buscar el canje por ID y cambiar su estado. Lo demás sigue con GORM. `main.go` selecciona el adaptador: `REPOSITORY_ADAPTER=sqlc` es el valor por defecto y `gorm` permite comparar, sin cambiar el servicio ni los manejadores.

En operaciones atómicas, sqlc usa `BeginTx` sobre la conexión `*sql.DB` que ya abre GORM. `q.WithTx(tx)` enlaza las consultas generadas a esa transacción. El callback recibe `repositorioTransaccionSqlc`, que **no incrusta GORM**: sus lecturas y escrituras usan sqlc. Se añadieron consultas de bloqueo de cliente/canje, sucursal, saldo, factura e inserción de compra y movimiento. Las comprobaciones y las dos escrituras comparten una transacción SQL; un error retorna sin `Commit` y el `Rollback` diferido descarta los cambios.

El diseño del servicio no cambia al seleccionar sqlc: cambia la implementación del puerto. El adaptador traduce `sql.ErrNoRows` al error de dominio y comprueba las filas afectadas en la actualización. Se conserva AutoMigrate. `db/schema.sql` describe las tablas para generar Go; no se ejecuta como migración. `internal/almacen/filas` se genera con sqlc 1.31.1 y no se edita a mano.

La prueba PostgreSQL de `taller10_postgres_test.go` agrega una restricción `CHECK` sobre `movimiento_puntos` para que falle la segunda escritura con `SQLSTATE 23514`. Comprueba que el canje siga solicitado y que la factura de prueba no quede registrada. Luego quita cada restricción y verifica que el reintento guarde ambas escrituras. La [verificación final](taller10/verificacion_final.md) registra los resultados de GORM y sqlc.

### D6 · Archivos

- `internal/fidelizacion/repositorio.go`: interfaces `Repositorio` y `RepositorioTransaccion`.
- `internal/fidelizacion/repositorio_gorm.go` y `repositorio_sqlc.go`: adaptadores y sus transacciones.
- `internal/fidelizacion/servicio.go`: dependencia de interfaces y callbacks; conserva reglas de estados, saldo y vigencia.
- `sqlc.yaml`, `db/schema.sql`, `db/consultas.sql` e `internal/almacen/filas/`: configuración, esquema, consultas y código generado.
- `main.go` y `internal/fidelizacion/rutas.go`: selección e inyección del adaptador, conservando las rutas.
- `internal/fidelizacion/modelos.go`, `dtos.go` y `manejadores.go`: Canje sin JSON y DTO para peticiones, listas y respuestas anidadas.
- `servicio_falso_test.go`, `dtos_test.go`, `taller10_postgres_test.go` y adaptación de `postgres_test.go`: pruebas de servicio, HTTP y adaptadores.
- `.env.example`, `db.go`, documentación y colección Bruno: configuración de contraseñas demo por entorno y contrato actualizado.

Los manejadores no cambian por elegir GORM/sqlc ni por el mecanismo transaccional; sus cambios de esta semana corresponden a DTO. `autenticacion.go` conserva JWT y permisos. No añadimos PUT ni endpoints de los planes futuros. Las evidencias del Hito 2 se conservan como históricas.
