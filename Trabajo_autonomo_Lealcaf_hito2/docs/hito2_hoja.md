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

Comprobado: ☑ structs y tablas de seis entidades y Cuenta · ☑ validador sin errores ni avisos · ☑ colección 0–6 · ☑ respuestas 401/200/200/403/200/422 · ☑ pruebas en verde, con integración · ☑ capas y decisiones · ☑ capturas directas actuales. Historial revisado: ☑ commits reales de ambos integrantes. A cargo de la pareja: ☐ lectura y ensayo de ambos. Entrega: publicar y comprobar el repositorio de la organización antes del cierre.

Desde la raíz: servidor **hito2/Iniciar-Proyecto.cmd**, pruebas **go -C hito2 test ./...**, contrato **npx.cmd @redocly/cli lint docs/openapi.yaml --format=stylish**. Evidencias actuales del alumno: **docs/hito2_capturas/**; detalle e historial: **docs/validacion.md**. Cierre Web II A: **29/10/2026, 23:59**; repositorio de **uleam-web-2026-2**.
