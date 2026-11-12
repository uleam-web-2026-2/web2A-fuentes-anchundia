# Captura de registros de go run

Se recibió una ejecución de Windows en la que las seis peticiones nativas respondieron correctamente, pero el ejecutor falló al leer `.env.gorun.stdout`: el archivo todavía estaba abierto por otro proceso.

La captura ahora utiliza las dos salidas del proceso en memoria, mediante lecturas asíncronas de stdout y stderr. Al terminar las peticiones se detiene el árbol de procesos iniciado por el ejecutor, se espera su cierre y se completan las dos lecturas antes de escribir `logs_gorun.txt`. Los archivos temporales `.env.gorun.stdout` y `.env.gorun.stderr` ya no se usan.

La comprobación de captura conservó 1000 líneas de stdout y 1000 de stderr, sin pérdidas. La comprobación completa con PostgreSQL temporal volvió a aprobar las 44 pruebas sin caché y las seis peticiones, y verificó que el registro final contiene tanto las peticiones HTTP como los mensajes de arranque de PostgreSQL. Estos ensayos se realizaron en PowerShell sobre Linux; las evidencias finales de Docker se obtienen al ejecutar el paquete en la computadora del integrante.

Para actualizar una ejecución existente, reemplace únicamente `scripts/taller11.ps1` por la versión actual y vuelva a ejecutar `Taller11.cmd`.
