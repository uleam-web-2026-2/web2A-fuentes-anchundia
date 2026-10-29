# LealCaf · Hito 2 de Aplicaciones Web II

API de fidelización para una cafetería con sucursales Centro y Norte. El cliente acumula un punto por dólar completo, consulta su saldo vigente por seis meses y solicita recompensas. El administrador registra compras y confirma o cancela solicitudes. Los puntos se comparten entre sucursales del mismo comercio.

**Pareja:** Fuentes Cedeño José Miguel · Anchundia Anchundia Luis Antonio · Web II A · Grupo P.

## Ubicación y comandos

El código de esta versión está en **hito2/** y la documentación está en **docs/**, en la raíz del repositorio. Los comandos de este documento se ejecutan desde esa raíz. El módulo Go de la raíz conserva trabajo anterior; para verificar Hito 2 use siempre **go -C hito2**.

## Base de datos y arranque

El proyecto actual utiliza **PostgreSQL**, con la misma base del Hito 1: **fidelizacion_db**, en **localhost:5432**. La conexión se configura en el archivo privado **.env**. La aplicación ya no arranca contra el archivo SQLite de la demostración anterior. SQLite se mantiene únicamente como herramienta para pruebas automáticas aisladas.

1. Mantenga PostgreSQL iniciado y la base **fidelizacion_db** disponible.
2. Prepare el archivo privado **hito2/.env** a partir de **hito2/.env.example** si aún no existe. El repositorio incluye únicamente la plantilla. Complete DB_USER y DB_PASSWORD con su conexión local; revise DB_HOST, DB_PORT y DB_NAME. Las credenciales reales se mantienen fuera de Git.
3. Ejecute **.\hito2\Iniciar-Proyecto.cmd** o haga doble clic en **hito2/Iniciar-Proyecto.cmd**. El iniciador genera una clave JWT local aleatoria si falta; JWT_TTL=1h corresponde al ensayo. **hito2/Iniciar-Demo.cmd** también ejecuta ese iniciador.
4. Mantenga abierta la ventana del servidor mientras prueba en Bruno. Dirección predeterminada: **http://localhost:8080**.
5. Si el puerto HTTP cambia, actualice PORT en **hito2/.env** y baseUrl en Bruno antes de preparar el ensayo. DB_PORT identifica PostgreSQL, no la API.

Para crear la configuración sin reemplazar una que ya exista, desde la raíz:

```powershell
if (-not (Test-Path -LiteralPath '.\hito2\.env')) {
    Copy-Item -LiteralPath '.\hito2\.env.example' -Destination '.\hito2\.env'
}
```

Después de completar la conexión privada, inicie con `.\hito2\Iniciar-Proyecto.cmd`. El repositorio entrega el código fuente y requiere Go para compilarlo; binarios y configuración privada se excluyen. La clave JWT configurada manualmente debe tener al menos 32 bytes aleatorios. Para ejecutar directamente con Go entre primero en **hito2/**, prepare JWT_SECRET y use `go run .`.

## Migración que conserva el Hito 1

El arranque aplica la preparación de la base de manera transaccional e idempotente. Mantiene los registros originales de clientes, compras, recompensas, puntos y canjes, agrega Sucursal y Cuenta, y añade sucursal_id a Compra y Canje. Los registros anteriores se atribuyen a la sucursal principal Centro (ID 1); la dirección inicial debe revisarse como dato académico.

Con **DEMO_DATA=true** prepara los datos ficticios del ensayo aparte del historial anterior. Crea un cliente dedicado de correo **cliente.hito2@example.invalid**, cuentas de prueba, puntos de demostración y un canje confirmado propio. Utiliza identificadores generados por la base: no presume que el cliente ni el canje tengan ID 1. Conserva las recompensas ya existentes y no cambia estados ni puntos de los clientes anteriores. Repetir el arranque no renueva fechas de acumulación ni vuelve a acreditar los mismos puntos.

El nombre de la base es **fidelizacion_db**; no debe crearse una segunda base final para la entrega. Los respaldos, las verificaciones previas y las capturas actuales del alumno se describen en [validacion.md](../docs/validacion.md).

## Ver las tablas en pgAdmin

Abra **PostgreSQL 18 → Databases → fidelizacion_db → Schemas → public → Tables** y actualice la lista después del arranque. Las tablas del modelo son clientes, compras, recompensas, canjes, movimiento_puntos, sucursals y cuenta. El nombre físico sucursals corresponde al struct Sucursal. En compras y canjes debe existir sucursal_id.

## Cuentas de demostración: dos roles

| Usuario | Contraseña de prueba | Rol | Identidad |
|---|---|---|---|
| cliente | cliente-2026 | cliente | Cliente ficticio dedicado; ID real devuelto por login |
| admin | admin-2026 | administrador | Sin Cliente asociado |

Son datos ficticios. El login entrega un JWT Bearer de una hora. La API obtiene la identidad del token firmado, no de X-Rol ni X-Cliente-ID. Las contraseñas se almacenan como hash bcrypt y no se devuelven en JSON. Una cuenta existente que entre en conflicto requiere revisión; la preparación no debe reemplazar silenciosamente su identidad o clave.

## Preparar las seis peticiones

Abra **docs/hito2_bruno**, en la raíz del repositorio, como colección Bruno y elija el entorno **Local**. Antes del turno envíe **0 · Preparar tokens y canje nuevo**. Hace ambos logins, obtiene una recompensa activa y localiza el canje confirmado propio del cliente demo; después crea un canje solicitado nuevo para las peticiones 4 y 5.

En el turno envíe **1 a 6**: **401 → 200 con token → 200 → 403 → 200 → 422**. Las peticiones 1, 3 y 6 usan **canjeConfirmadoId**, obtenido del servidor. Las peticiones 4 y 5 usan **canjeId**, creado por 0. Repita la preparación antes de cada ensayo. Tokens e identificadores permanecen en memoria, sin editar archivos durante el turno. Consulte el [detalle de la colección](../docs/hito2_bruno/README.md).

## Rutas implementadas y plan

Todas requieren token salvo login.

| Ruta | Acceso |
|---|---|
| POST /login | Público |
| GET /clientes/{id}/puntos | Cliente propio / administrador |
| GET /recompensas | Cliente / administrador |
| POST /compras | Administrador |
| POST /canjes | Cliente |
| GET /canjes | Administrador; filtro estado opcional |
| GET /canjes/{id} | Cliente propio / administrador |
| PATCH /canjes/{id}/estado | Administrador |

GET /clientes/{id}, GET /sucursales y GET /sucursales/{id} están declarados como **(plan)** en el contrato, sin ruta implementada. Se conservan las operaciones existentes del Hito 1; esta actualización cambia el motor de ejecución y la preparación de datos. Canje sigue siendo la entidad principal evaluada.

Canje nace solicitado. El administrador confirma con puntos suficientes o cancela. Confirmado y cancelado son finales. Una transición prohibida produce 422; falta de autenticación, 401; falta de permiso, 403. Factura duplicada produce 409.

## Verificaciones y evidencias

Desde la raíz del repositorio:

```powershell
go -C hito2 test ./...
npx.cmd @redocly/cli lint docs/openapi.yaml --format=stylish
```

Node.js es necesario para el validador. En PowerShell se usa npx.cmd para ejecutar la entrada de Windows. Ejecútelo antes del turno para descargar Redocly si hace falta. El archivo redocly.yaml de la raíz mantiene security-defined activa; el resultado debe tener cero errores y cero avisos. El alumno ejecutó las pruebas y Redocly CLI 2.72.0 y tomó las capturas actuales; ver [validacion.md](../docs/validacion.md).

- [Contrato OpenAPI](../docs/openapi.yaml) y [vista del contrato](../docs/api.html).
- [Hoja del hito](../docs/hito2_hoja.md): usuarios, peticiones, permisos y cumplimiento.
- [Estructura y decisiones](../docs/decisiones.md): capas, Sucursal, PostgreSQL y declaración de IA.
- [Guion y seis preguntas](../docs/guion_hito2.md): ensayo de cinco minutos.
- **docs/hito2_capturas/**, en la raíz: capturas directas tomadas por el alumno el 28 de octubre: 0_respuesta.png a 6_respuesta.png, pruebas.png y contrato.png. Los HTML y reportes anteriores se conservan como históricos y se distinguen de estas capturas actuales.

La ficha y el addendum describen la versión vigente; el PDF del Hito 1 y sus imágenes se conservan como documentos históricos.

## Entrega

Cierre Web II A: **jueves 29 de octubre de 2026, 23:59**. Grupo P: **viernes 30, 09:40**, bloque 09:30–10:00. La entrega válida es el repositorio de la pareja dentro de **uleam-web-2026-2**. Los integrantes deben revisar el código, realizar contribuciones y commits reales, declarar IA, ensayar las preguntas y hacer push cuando decidan entregar. Esta actualización no realiza commits ni subidas a Git.

**IA en Hito 2:** OpenAI Codex apoyó implementación, reorganización, autenticación, pruebas, contrato, Bruno, documentación y adaptación conservadora a la base PostgreSQL original. Los integrantes deben comprender y verificar el resultado antes de exponer.
