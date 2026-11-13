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

## Evidencias en Bruno Desktop

Se ejecutó el Runner de la colección del Hito 2 en orden, de la petición 0 a la 6 y con el entorno privado del taller. La petición 0 respondió `201` en los tres recorridos y creó el canje que se confirmó en la petición 5. Las seis respuestas coincidieron con la tabla anterior.

| Modo de la API | URL de Bruno | Códigos HTTP de 1 a 6 | Captura del Runner |
| --- | --- | --- | --- |
| `go run` | `http://localhost:8080` | 401, 200, 200, 403, 200, 422 | [01-go-run.jpg](bruno/01-go-run.jpg) |
| `docker run` | `http://127.0.0.1:8080` | 401, 200, 200, 403, 200, 422 | [02-docker-run.jpg](bruno/02-docker-run.jpg) |
| Docker Compose | `http://127.0.0.1:8080` | 401, 200, 200, 403, 200, 422 | [03-compose.jpg](bruno/03-compose.jpg) |

En el primer intento con `docker run`, Bruno resolvió `localhost` a `::1` y recibió `ECONNREFUSED ::1:8080`. Ese intento no produjo códigos HTTP. Docker publicaba la API en `127.0.0.1:8080`. Se ajustó solo la URL del entorno privado a la dirección IPv4 y se repitió el recorrido completo, incluida la preparación. Se usó esa dirección también con Compose. Las capturas muestran el resumen con las respuestas cerradas.

El resumen del Runner se ve igual cuando coinciden los resultados. Los logs de cada instancia confirmaron ejecuciones separadas: la petición 5 confirmó el canje 4 con `go run`, el 5 con `docker run` y el 3 con Compose. Los identificadores corresponden a la base usada por cada modo.

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
