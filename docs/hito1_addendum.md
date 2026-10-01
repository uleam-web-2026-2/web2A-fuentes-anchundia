# Hito 1 · Addendum técnico

**Pareja:** Fuentes Cedeño José Miguel · Anchundia Anchundia Luis Antonio
**Paralelo:** Aplicaciones Web II A

## A. Estructura del proyecto

El repositorio del curso ya trae el servidor base de la mesa de ayuda, así que el proyecto de fidelización vive en su propia carpeta `fidelizacion/` (con su propio `go.mod`) y la documentación del hito va en `docs/`. Todos los comandos de Go se ejecutan dentro de `fidelizacion/`.

```text
web2A-fuentes-anchundia/
├── docs/
│   ├── hito1_ficha_del_negocio.md
│   ├── hito1_addendum.md
│   ├── hito1_presentacion.pdf
│   └── (capturas .png)
└── fidelizacion/
    ├── main.go
    ├── go.mod
    ├── go.sum
    ├── .gitignore
    ├── .env.example
    └── internal/
        └── fidelizacion/
            ├── modelos.go
            ├── db.go
            ├── manejadores.go
            ├── rutas.go
            └── fidelizacion_test.go
```

Rutas de la tabla relativas a `fidelizacion/`:

| Carpeta / archivo | Qué guarda |
| --- | --- |
| `main.go` | Lee `.env`, se conecta a PostgreSQL, migra las tablas, carga datos de ejemplo y arranca el servidor en el puerto 8080. |
| `go.mod`, `go.sum` | Módulo `fidelizacion` y sus dependencias (Gin, GORM, controlador de PostgreSQL, godotenv). |
| `.gitignore` | Evita subir `.env` y binarios. |
| `.env.example` | Variables de configuración con valores de ejemplo, sin credenciales reales. |
| `internal/fidelizacion/modelos.go` | Los cinco structs, las constantes de estados, tipos y roles, y las reglas del programa (puntos por dólar y vigencia). |
| `internal/fidelizacion/db.go` | Conexión a PostgreSQL, migración con GORM y datos de ejemplo. |
| `internal/fidelizacion/manejadores.go` | Los seis manejadores de endpoints, el control de rol y el cálculo del saldo. |
| `internal/fidelizacion/rutas.go` | Registro de las rutas y qué rol puede llamar a cada una. |
| `internal/fidelizacion/fidelizacion_test.go` | Las pruebas automáticas. |
| `../docs/` | Ficha, addendum, presentación y capturas del hito. |

## B. Configuración y secretos

El proyecto lee las credenciales de la base de datos desde variables de entorno, cargadas de un archivo `.env` que no se sube al repositorio (está en `.gitignore`). El archivo de ejemplo es `.env.example`.

| Variable | Para qué sirve | Valor de ejemplo |
| --- | --- | --- |
| `DB_HOST` | Servidor de PostgreSQL | `localhost` |
| `DB_PORT` | Puerto de PostgreSQL | `5432` |
| `DB_USER` | Usuario de la base de datos | `usuario_prueba` |
| `DB_PASSWORD` | Contraseña de la base de datos | `cambiar_esta_clave` |
| `DB_NAME` | Nombre de la base de datos | `fidelizacion_db` |

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=usuario_prueba
DB_PASSWORD=cambiar_esta_clave
DB_NAME=fidelizacion_db
```

Para correr el proyecto (desde la carpeta `fidelizacion/`) se copia `.env.example` a `.env`, se ponen los datos reales y se ejecuta `go run .`. Las pruebas no necesitan PostgreSQL: usan una base SQLite en memoria.

## C. Pruebas

Son 10 pruebas en `internal/fidelizacion/fidelizacion_test.go`. Cada una crea su propia base en memoria.

| Prueba | Qué caso cubre |
| --- | --- |
| `TestConsultarPuntos` | El cliente consulta su saldo (acumulación menos canje) y recibe `200`. |
| `TestClienteNoPuedeVerPuntosAjenos` | Un cliente que pide los puntos de otro recibe `403`. |
| `TestPuntosVencidosNoCuentanEnElSaldo` | Los puntos con fecha de vencimiento pasada no suman al saldo. |
| `TestRegistrarCompraGeneraPuntosConVencimiento` | Una compra de $50.75 genera 50 puntos con vencimiento a seis meses. |
| `TestRegistrarCompraConFacturaDuplicada` | La misma factura registrada dos veces responde `409`. |
| `TestClienteNoPuedeRegistrarCompras` | El rol cliente no puede usar `POST /compras` (`403`). |
| `TestCanjeConPuntosInsuficientes` | Un canje sin saldo suficiente responde `400` "puntos insuficientes". |
| `TestCanjeCorrectoDescuentaPuntosAlConfirmar` | Solicitar crea el canje sin descontar; confirmar descuenta los puntos. |
| `TestTransicionProhibidaDeConfirmadoACancelado` | Pasar de `confirmado` a `cancelado` responde `409`. |
| `TestRecompensasSoloMuestraLasActivas` | El catálogo no incluye recompensas inactivas. |

Comando (desde la carpeta `fidelizacion/`):

```bash
cd fidelizacion
go test ./...
```

**Captura de `go test ./...`:**

![Pruebas](hito1_pruebas.png)

## D. Boceto de la pantalla principal

La pantalla principal es **Mis puntos** (rol cliente). Consume `GET /clientes/{id}/puntos`. Muestra el nombre, la cédula, el saldo vigente y el historial con fecha, tipo, puntos y fecha de vencimiento. El estado de los canjes del cliente se muestra en la pantalla Recompensas.

```text
┌─────────────────────────────────────────────────────┐
│               PROGRAMA DE PUNTOS · CAFETERÍA         │
├─────────────────────────────────────────────────────┤
│ Cliente: Juan Pérez                                 │
│ Cédula:  0900000001                                 │
│                                                     │
│              PUNTOS DISPONIBLES                     │
│                   50 puntos                         │
├─────────────────────────────────────────────────────┤
│ HISTORIAL DE MOVIMIENTOS                            │
├────────────┬─────────────┬────────┬─────────────────┤
│ Fecha      │ Tipo        │ Puntos │ Vence           │
├────────────┼─────────────┼────────┼─────────────────┤
│ 01/10/2026 │ Acumulación │ +50    │ 01/04/2027      │
│ 20/09/2026 │ Canje       │ -40    │ No aplica       │
└────────────┴─────────────┴────────┴─────────────────┘
```

![Boceto](hito1_boceto.png)

## E. Diagrama de secuencia del caso de uso principal

El caso de uso principal es el ciclo completo de un canje: el cliente solicita una recompensa y el administrador la confirma.

```mermaid
sequenceDiagram
    actor Cliente
    actor Admin as Administrador
    participant Pantalla as Pantallas
    participant API as API (Gin)
    participant BD as PostgreSQL

    Cliente->>Pantalla: Elige una recompensa
    Pantalla->>API: POST /canjes {recompensa_id}
    API->>BD: Busca recompensa activa y calcula saldo vigente
    BD-->>API: Recompensa y saldo
    API->>BD: Crea canje en estado solicitado
    API-->>Pantalla: 201 canje solicitado
    Pantalla-->>Cliente: Muestra "solicitud enviada"

    Admin->>Pantalla: Abre Gestión de canjes
    Pantalla->>API: GET /canjes?estado=solicitado
    API-->>Pantalla: 200 lista de canjes
    Admin->>Pantalla: Confirma el canje
    Pantalla->>API: PATCH /canjes/{id}/estado {confirmado}
    API->>BD: Revalida saldo, registra movimiento negativo y cambia estado
    BD-->>API: Guardado
    API-->>Pantalla: 200 canje confirmado
    Pantalla-->>Admin: Muestra canje confirmado
```

## F. Capturas de respuestas

Las dos capturas se toman con curl o Bruno contra el servidor corriendo. Se envía el rol en el encabezado `X-Rol` (y `X-Cliente-ID` para el cliente).

### Caso correcto

```bash
curl -i http://localhost:8080/clientes/1/puntos -H "X-Rol: cliente" -H "X-Cliente-ID: 1"
```

Resultado esperado: `HTTP/1.1 200 OK` con un JSON que incluye `cliente_id`, `cedula`, `nombre`, `saldo` y `movimientos`.

![Respuesta correcta](hito1_respuesta_ok.png)

### Caso con error de validación

El cliente pide una recompensa que cuesta más puntos de los que tiene:

```bash
curl -i -X POST http://localhost:8080/canjes \
  -H "Content-Type: application/json" -H "X-Rol: cliente" -H "X-Cliente-ID: 1" \
  -d '{"recompensa_id": 3}'
```

Resultado esperado: `HTTP/1.1 400 Bad Request` con `{"error":"puntos insuficientes","necesarios":100,"saldo":...}`.

![Respuesta con error](hito1_respuesta_error.png)
