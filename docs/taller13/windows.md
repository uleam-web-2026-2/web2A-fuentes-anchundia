# Incidencias reales en Windows

- Docker Desktop estaba instalado con el motor apagado. Se inició la aplicación existente y se verificó el motor; se creó `lealcaf-taller13` con volumen independiente, sin borrar los anteriores.
- El sandbox inicial impidió `git fetch` y la descarga de Go. Se usó la autorización de ejecución correspondiente; más tarde el usuario habilitó acceso completo. No se modificó la identidad de Git.
- El primer intento de controlar Bruno devolvió `Computer Use was not approved to use Bruno`. Tras habilitar el acceso se retomó la colección existente y se obtuvieron las tres capturas reales.
- La extracción PDF con Python falló inicialmente por la salida CP1252 de la consola. Se repitió con `PYTHONIOENCODING=utf-8`; los documentos de entrega se escribieron explícitamente en UTF-8, sin redirecciones UTF-16 de PowerShell 5.
- La escala de pantalla hizo que los tamaños del control de ventanas no coincidieran con los píxeles físicos. Se ajustaron las ventanas con el menú de Windows y se verificó una captura real de todo el escritorio. No se presionó Escape ni se compusieron capturas.
- Al abrir un ejecutable nuevo apareció una solicitud de firewall. No se concedieron permisos. Para Go local se incorporó `BIND_HOST=127.0.0.1`, manteniendo vacío el valor en Docker. Al usar localhost, Bruno eligió `::1` y la primera vuelta del reto dio ECONNREFUSED. Se corrigió la URL privada a `http://127.0.0.1:18083`; la repetición real pasó y es la captura 03-reto.png.
- La primera subida de Compose intentó compartir 18083 con Go local y falló por puerto ocupado. Se separó la API de Compose en 18085 mediante un archivo privado de entorno; después migró y arrancó correctamente. PostgreSQL se publica solo en 127.0.0.1:15433 para Go local.
- La biblioteca WebSocket quedó en v1.8.15. sqlc no se regeneró: no cambió el esquema ni las consultas de canjes; se reutilizó `ClienteID` y se probaron ambos adaptadores con PostgreSQL.

Los fallos recuperados no se reportan como pruebas exitosas: las capturas y salidas finales proceden de las repeticiones que realmente pasaron.
