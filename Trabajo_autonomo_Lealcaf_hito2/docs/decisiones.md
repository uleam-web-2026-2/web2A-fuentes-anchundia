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
| `repositorio.go` | Lecturas/escrituras con GORM y transacciones de persistencia. |

Flujo: **ruta → autenticación → manejador HTTP → servicio → repositorio → base de datos**. El resultado vuelve por el mismo camino; el manejador traduce los errores a HTTP. El servicio decide si se puede confirmar o cancelar y el repositorio guarda el cambio junto con el movimiento de puntos dentro de una transacción.

Para demostrar las capas, abrimos `manejadores.go` y `servicio.go`: el primero interpreta la petición; el segundo contiene la regla del cambio de estado. Si cambiamos PostgreSQL por otra base, revisamos `db.go`, el adaptador de `repositorio.go` y los detalles de mapeo GORM en los modelos. La regla de estados y la forma de la petición no deberían cambiar. Las pruebas deben confirmar ese resultado; no asumimos portabilidad automática de cualquier motor.

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
