# Ejecución en Windows

Se verificó en Windows amd64 con PowerShell, Go 1.27.1 y PostgreSQL 18.6.

- sqlc no estaba en el PATH. Se usó el binario oficial Windows amd64 de sqlc 1.31.1 y su ruta absoluta; la generación se ejecutó desde la carpeta que contiene sqlc.yaml. Para repetir, agregue la carpeta de sqlc.exe al PATH o invoque su ruta con el operador & de PowerShell.
- Se utilizó npm.cmd/npx.cmd para evitar depender de la política de ejecución de scripts .ps1 al ejecutar herramientas Node. Redocly se fijó en 2.54.3 y Bruno CLI fue 4.3.0.
- La primera compilación fue lenta por la creación de la caché de Go. Para la verificación automatizada se ubicó GOCACHE en la carpeta temporal de trabajo; no es necesario añadir esta variable al .env de la API.
- PostgreSQL se ejecutó en un cluster temporal con puerto 55432 y bases nuevas. No se cambió la instalación ni la base habitual del alumno. El servidor de prueba se mantiene como proceso mientras se ejecutan las pruebas; los procesos auxiliares del entorno aislado terminan al cerrar su sesión.

No se modificaron reglas de negocio para resolver diferencias de Windows. Todas las rutas de los comandos con espacios se citaron. Las herramientas y el cluster temporal no se incluyen en la entrega.
