# Semana 11 - seis peticiones

| Peticion | Con go run | En el contenedor | Con Compose |
| --- | --- | --- | --- |
| 1 sin token | 401 | 401 | Pendiente |
| 2 login | 200 | 200 | Pendiente |
| 3 con token | 200 | 200 | Pendiente |
| 4 rol sin permiso | 403 | 403 | Pendiente |
| 5 cambio permitido | 200 | 200 | Pendiente |
| 6 transicion prohibida | 422 | 422 | Pendiente |

Secuencia esperada: 401, 200, 200, 403, 200, 422.

La preparacion crea un canje nuevo antes de cada recorrido. No se guardan tokens ni respuestas de login.

## Peticion 3 - gorun

```json
{"id":1,"sucursal_id":2,"cliente_id":1,"recompensa_id":1,"puntos_utilizados":40,"fecha":"2026-11-12T15:15:56.02458+01:00","estado":"confirmado"}
```

## Peticion 3 - contenedor

```json
{"id":1,"sucursal_id":2,"cliente_id":1,"recompensa_id":1,"puntos_utilizados":40,"fecha":"2026-11-12T14:15:56.02458Z","estado":"confirmado"}
```
