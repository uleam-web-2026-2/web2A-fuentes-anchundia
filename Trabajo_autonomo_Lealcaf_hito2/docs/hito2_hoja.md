# Hito 2 · LealCaf

Fuentes Cedeño José Miguel · Anchundia Anchundia Luis Antonio · Web II A · Grupo P

## 1 · Usuarios de ejemplo

| Usuario | Rol | Contraseña de ejemplo |
|---|---|---|
| cliente | cliente | cliente-2026 |
| admin | administrador | admin-2026 |

Solo cuentas de prueba. PostgreSQL: **fidelizacion_db**. Preparación **0**: obtiene tokens, `clienteId`, `recompensaId` y `canjeConfirmadoId`; crea un canje solicitado en Norte y guarda `canjeId`. Los IDs se obtienen del servidor, sin asumir cliente o canje 1.

## 2 · Las seis peticiones

| # | Petición | Con qué credencial | Debe responder |
|---|---|---|---|
| 1 | GET /canjes/{{canjeConfirmadoId}} | Ninguna | 401 |
| 2 | POST /login (`cliente`) | Ninguna | 200 y token |
| 3 | GET /canjes/{{canjeConfirmadoId}} | Token cliente | 200 |
| 4 | PATCH /canjes/{{canjeId}}/estado → confirmado | Token cliente | 403 |
| 5 | La misma petición 4 | Token administrador | 200 |
| 6 | PATCH /canjes/{{canjeConfirmadoId}}/estado → cancelado | Token administrador | 422 |

## 3 · Matriz ruta × rol

| Ruta | Sin sesión | Cliente | Administrador |
|---|---|---|---|
| POST /login | Sí | Sí | Sí |
| GET /clientes/{id}/puntos | No (401) | Propio; ajeno 403 | Sí |
| GET /recompensas | No (401) | Sí | Sí |
| POST /compras | No (401) | No (403) | Sí |
| POST /canjes | No (401) | Sí | No (403) |
| GET /canjes | No (401) | No (403) | Sí |
| GET /canjes/{id} | No (401) | Propio; ajeno 403 | Sí |
| PATCH /canjes/{id}/estado | No (401) | No (403) | Sí |

## 4 · Recuadro de cumplimiento

☑ seis entidades y Cuenta: siete tablas y nueve FK verificadas en PostgreSQL real · ☑ OpenAPI sin errores ni avisos · ☑ Bruno 0–6 aprobado en interfaz y Runner · ☑ pruebas en verde con integración; concurrencia PostgreSQL aprobada en instancia aislada · ☑ capas y decisiones · ☑ ocho capturas nuevas bruno_ui_*.jpg · ☑ .env privado excluido de Git.

Pendientes: ☐ ensayo de ambos con reloj · ☐ commit y push autorizados · ☐ comprobar publicación y CI. Se conserva docs/ dentro del proyecto por decisión del alumno; su ubicación difiere de la raíz Git pedida por el enunciado. El README principal enlaza a todos los entregables.

Desde **Trabajo_autonomo_Lealcaf_hito2/**: **Iniciar-Proyecto.cmd**, **go test ./...** y **npx.cmd @redocly/cli lint docs/openapi.yaml --format=stylish**. Evidencias: **docs/hito2_capturas/index.html**; detalle: **docs/validacion.md**. Cierre Web II A: **29/10/2026, 23:59**, repositorio institucional.
