# Semana 11 - seis peticiones

| Peticion | Con go run | En el contenedor | Con Compose |
| --- | --- | --- | --- |
| 1 sin token | 401 | 401 | 401 |
| 2 login | 200 | 200 | 200 |
| 3 con token | 200 | 200 | 200 |
| 4 rol sin permiso | 403 | 403 | 403 |
| 5 cambio permitido | 200 | 200 | 200 |
| 6 transicion prohibida | 422 | 422 | 422 |

Secuencia esperada: 401, 200, 200, 403, 200, 422.

La preparacion crea un canje nuevo antes de cada recorrido. No se guardan tokens ni respuestas de login.

## Peticion 3 - gorun

```json
{"id":1,"sucursal_id":2,"cliente_id":1,"recompensa_id":1,"puntos_utilizados":40,"fecha":"2026-11-12T19:53:23.368194-05:00","estado":"confirmado"}
```

## Peticion 3 - contenedor

```json
{"id":1,"sucursal_id":2,"cliente_id":1,"recompensa_id":1,"puntos_utilizados":40,"fecha":"2026-11-13T00:53:23.368194Z","estado":"confirmado"}
```

## Peticion 3 - compose

```json
{"id":1,"sucursal_id":2,"cliente_id":1,"recompensa_id":1,"puntos_utilizados":40,"fecha":"2026-11-13T00:54:32.844943Z","estado":"confirmado"}
```
