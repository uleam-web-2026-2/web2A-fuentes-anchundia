# Seis peticiones del Hito 2

Ejecutadas el 5 de noviembre de 2026 con Bruno CLI 4.3.0, colección `docs/hito2_bruno`, preparación 0 incluida.

Antes: copia original recibida en coweb.zip, GORM, PostgreSQL 18.6, base temporal lealcaf_antes, API 18080.
Después: solución del taller, adaptador sqlc, base temporal lealcaf_despues, API 18081.
La preparación creó un canje nuevo en cada ejecución. No se utilizó la base habitual del alumno.

| Petición | Antes | Después |
|---|---:|---:|
| 1 sin token | 401 | 401 |
| 2 login | 200 | 200 |
| 3 con token | 200 | 200 |
| 4 rol sin permiso | 403 | 403 |
| 5 cambio permitido | 200 | 200 |
| 6 transición prohibida | 422 | 422 |

La preparación 0 devolvió 201 en ambas ejecuciones. Bruno informó 7/7 peticiones, 6/6 pruebas y 7/7 aserciones en cada una. No cambió ningún código.

Evidencia: `bruno_antes.json`, `bruno_despues.json` (resumen del reporte real; se excluyen cuerpos y cabeceras para no guardar contraseñas ni JWT) y salida CLI `bruno_despues.txt`.
