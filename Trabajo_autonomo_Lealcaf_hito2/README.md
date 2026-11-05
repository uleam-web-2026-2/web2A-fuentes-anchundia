# LealCaf · Hito 2

API de fidelización de cafetería con Go, Gin, GORM, PostgreSQL y consultas generadas por sqlc. La entidad principal es **Canje**. Las compras generan puntos con vigencia de seis meses y los canjes se comparten entre sucursales.

**Integrantes:** Fuentes Cedeño José Miguel y Anchundia Anchundia Luis Antonio · Web II A.

## Carpeta del módulo y requisitos

Ejecuta los comandos dentro de `Trabajo_autonomo_Lealcaf_hito2/`, junto a `go.mod`, `main.go`, `sqlc.yaml` y `docs/`. La carpeta superior contiene otros trabajos: ejecutar Go desde allí no verifica esta API.

Se requiere Go 1.27.1 o posterior compatible con `go.mod`, PostgreSQL y sqlc **1.31.1**. Node.js/npm sirven para validar OpenAPI y Bruno para repetir las peticiones del Hito 2.

## Configuración y arranque

1. Inicia PostgreSQL y usa la base del proyecto. En una instalación nueva crea una base vacía; conserva las bases que ya tengan datos.
2. Copia `.env.example` a `.env` si todavía no existe:

```powershell
if (-not (Test-Path -LiteralPath '.env')) {
    Copy-Item -LiteralPath '.env.example' -Destination '.env'
}
```

3. Completa localmente `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` y `JWT_SECRET`. La clave JWT debe tener al menos 32 bytes aleatorios; el iniciador Windows puede generarla si falta. `PORT` selecciona el puerto HTTP y `JWT_TTL` la duración del token. No distribuyas `.env`.
4. Usa `REPOSITORY_ADAPTER=sqlc` para el adaptador SQL generado. `REPOSITORY_ADAPTER=gorm` selecciona el otro adaptador sin cambiar reglas, rutas ni manejadores. Ambos utilizan la conexión PostgreSQL abierta por GORM.
5. Si `DEMO_DATA=true`, completa `DEMO_CLIENT_PASSWORD` y `DEMO_ADMIN_PASSWORD` con las contraseñas locales de las cuentas `cliente` y `admin`. No hay contraseñas demo por defecto en el código. Si las cuentas ya existen, usa sus contraseñas actuales: el sembrado comprueba identidad y no sobrescribe credenciales. `DEMO_DATA=false` desactiva esos datos.
6. Genera y comprueba desde la carpeta del módulo:

```powershell
sqlc version
sqlc generate
go mod download
go build ./...
go test ./...
```

`sqlc version` debe indicar `v1.31.1`. El código generado se incluye para compilar; repite `sqlc generate` cuando cambien consultas o esquema descriptivo. Consulta la [instalación oficial de sqlc](https://docs.sqlc.dev/en/latest/overview/install.html).

7. Ejecuta `.\Iniciar-Proyecto.cmd` en Windows o `go run .` con la configuración completa. El servidor usa `http://localhost:8080` salvo que cambies `PORT`; Ctrl+C lo detiene. `ENV_FILE` permite escoger otra configuración privada y las variables del proceso tienen prioridad sobre `.env`.

La base de datos se prepara con AutoMigrate. `db/schema.sql` solo describe las tablas para sqlc: **no se ejecuta como migración**.

## Arquitectura

- El servicio depende de interfaces del negocio y no importa GORM ni `database/sql`.
- `Canje` no tiene etiquetas JSON. `dtos.go` define peticiones y respuestas, incluidas listas y datos anidados. El POST no permite controlar ID, cliente, puntos, fecha o estado desde el cuerpo: nace `solicitado` y la identidad viene del token.
- `RepositorioSqlc` realiza la búsqueda por ID y el cambio de estado del canje; las demás operaciones fuera de transacciones continúan con GORM.
- Confirmar guarda juntos estado y descuento; registrar una compra guarda juntos compra y acumulación. Cada adaptador usa su propia transacción. sqlc utiliza `WithTx` sin llamadas GORM dentro del callback.
- El estado del canje cambia mediante `PATCH /canjes/{id}/estado`.

Las alternativas y la decisión están en [docs/decisiones.md](docs/decisiones.md), sección Semana 10.

## Bruno y contrato

Abre `docs/hito2_bruno/` como colección y selecciona **Local**. Completa localmente `clientePassword` y `adminPassword` con los mismos valores de `DEMO_CLIENT_PASSWORD` y `DEMO_ADMIN_PASSWORD`; ajusta `baseUrl` si cambias el puerto. No publiques esos valores.

Con los datos demo habilitados, ejecuta **0** para obtener tokens e IDs reales y preparar un canje. Después ejecuta **1 a 6**: deben conservar **401, 200, 200, 403, 200 y 422**. Repite 0 antes de cada ensayo porque confirmar consume puntos.

```powershell
npx.cmd --yes @redocly/cli@2.54.3 lint docs/openapi.yaml
```

El contrato está en `docs/openapi.yaml` y la vista en `docs/api.html`. Las operaciones futuras conservan **(plan)**; esa documentación no crea rutas.

## Pruebas y evidencias

`go test ./...` incluye pruebas con SQLite en memoria y casos del servicio con un repositorio falso sin base. Las pruebas PostgreSQL se omiten explícitamente si falta `LEALCAF_TEST_POSTGRES_DSN`.

Para comprobar ambos adaptadores y el rollback, usa una base PostgreSQL de pruebas desechable con permiso para crear esquemas. Configura localmente `LEALCAF_TEST_POSTGRES_DSN` en formato `host=... port=... user=... password=... dbname=... sslmode=disable`, y ejecuta:

```powershell
go test -count=1 -v ./internal/fidelizacion -run 'TestTaller10|TestPostgres'
```

Las pruebas crean sus propios esquemas y los limpian al terminar. Las evidencias se guardan en `docs/taller10/`: peticiones antes/después, pruebas Go, Redocly y fallo forzado con GORM y sqlc. Los registros de `docs/hito2_capturas/` pertenecen al Hito 2 y no sustituyen las comprobaciones actuales.
