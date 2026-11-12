# Verificación previa del paquete

Fecha: **12 de noviembre de 2026**. Base del trabajo: rama `taller10`, revisión `5f6e6f6`.

## Comprobado antes de entregar el ZIP

| Comprobación | Resultado |
| --- | --- |
| Pruebas existentes del módulo vigente | **44 pruebas principales aprobadas, ninguna omitida**, con PostgreSQL temporal. Se ejecutaron también las variantes GORM y sqlc que requieren PostgreSQL. |
| Compilación de la API con `CGO_ENABLED=0` | Correcta; el ejecutable Linux se identificó como **statically linked**. |
| `go vet ./...` del módulo vigente | Correcto. |
| Ejecución real de `go run .` y peticiones por el ejecutor de PowerShell | **401, 200, 200, 403, 200, 422**, en orden. |
| Manejo de fallos de programas nativos en el ejecutor | Se conserva el código de salida; el ejecutor se detiene ante un fallo y permite los errores deliberados del punto 2. |
| Sustitución de bloques de mediciones | Comprobada sobre una copia temporal; las mediciones de Docker del paquete siguen pendientes. |
| Validación de `compose.yaml` con Compose | Correcta. Dos servicios, `DB_HOST=db`, puerto interno `5432`, espera por `service_healthy`, volumen nombrado y publicación exclusiva del puerto HTTP en loopback. |
| Exclusión de archivos privados en Git | `.env`, sus variantes y `Taller11-privado.bru` quedan ignorados. |
| Preparación de PostgreSQL tras un intento anterior | Comprobaciones unitarias de identidad por configuración, conservación de bases, selección del puerto y rechazo de credenciales inválidas. Estas comprobaciones simulan las respuestas Docker; la conexión TCP del ensayo es real. |
| Captura de salida del proceso Go | Captura simultánea en memoria de stdout y stderr, espera del cierre del árbol de procesos y escritura posterior del registro. Ensayo de 1000 líneas por cada salida y comprobación completa con las seis peticiones. |

Entorno de estas comprobaciones: Go **1.27.4**, PowerShell **7.5.4** sobre Linux, PostgreSQL **16.10** temporal y Compose **v5.7.1** para validar la configuración. La revisión posterior del arranque repitió las 44 pruebas sin caché, con `-count=1`, y las seis peticiones sobre PostgreSQL configurado con autenticación de contraseña **SCRAM-SHA-256**. No se omitieron pruebas. El ejecutor está escrito con las funciones disponibles en Windows PowerShell 5.1, utilizado por `Taller11.cmd`; la ejecución nativa de Windows se comprueba en la computadora del integrante.

Los informes `pruebas_previas_postgresql.txt`, `peticiones_previas_postgresql.json`, `logs_previos_postgresql.txt`, `ejecutor_previo.txt`, `build_estatico_previo.txt`, `vet_previo.txt` y `compose_config_previo.txt` corresponden a esta revisión. Se usan datos ficticios de una base temporal, sin modificar una base de los integrantes. Los secretos se omiten de los informes.

## Qué se completa al ejecutar Taller11.cmd

Este entorno no dispone de un motor Docker. Por ello no se han construido las imágenes ni arrancado contenedores aquí. La validación de Compose comprueba el archivo, no acredita el arranque de sus servicios.

Permanecen pendientes los tamaños y el usuario observado de la imagen, las seis peticiones en el contenedor, la comparación de fechas entre máquina y contenedor, los tres arranques de Compose desde cero y la conservación de datos tras `down`. El ejecutor realiza esas comprobaciones y guarda las salidas reales, actualizando `peticiones.md` y las mediciones de D1, D4 y D6.

La entrega de la asignatura queda completa cuando el integrante ejecuta el paquete correctamente, revisa las evidencias y explicaciones, y las publica en la rama `taller11`. Estos informes previos no sustituyen las evidencias pendientes de Docker.
