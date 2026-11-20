# Decisiones del proyecto - Semana 11

Proyecto: **LealCafé**. Integrantes: José Miguel Fuentes Cedeño y Luis Antonio Anchundia Anchundia.

Partimos de la rama `taller10`, revisión `5f6e6f6`. El módulo vigente está en `Trabajo_autonomo_Lealcaf_hito2/`. Las decisiones de las semanas anteriores siguen en [el documento del módulo](../Trabajo_autonomo_Lealcaf_hito2/docs/decisiones.md).

## Semana 11 - el contenedor

### D1 - Qué entra en la imagen

El `Dockerfile` tiene dos etapas. La primera usa `golang:1.27`, descarga las dependencias del módulo y compila con `CGO_ENABLED=0`. Copiamos primero `go.mod` y `go.sum` para reutilizar la descarga cuando solo cambia el código. Como hay varias entregas antiguas en el repositorio, el `COPY` selecciona expresamente `Trabajo_autonomo_Lealcaf_hito2/`.

La imagen final parte de `gcr.io/distroless/static-debian13:nonroot` y recibe solamente `/servidor`. No copiamos el compilador, el código fuente, las pruebas ni las dependencias descargadas. Las rutas implementadas no necesitan leer la carpeta `docs/` del disco. El proceso arranca con el usuario nonroot de la imagen base, no como root.

<!-- D1:INICIO -->
Tamano real final: **65804244 bytes** (62.76 MiB). Etapa de compilacion: **2573030927 bytes** (2453.83 MiB). Usuario observado: **65532**. Evidencias: imagenes.txt, capas.txt, contexto.txt y usuario.txt.
<!-- D1:FIN -->

Sin `.dockerignore`, los archivos privados entrarían al contexto enviado a Docker. Si un `.env` estuviera dentro del módulo copiado, llegaría además a `/src` en la etapa de compilación y podría quedar en su caché. Un `.env` de la raíz no coincide con nuestro `COPY` de la subcarpeta, pero igualmente se enviaría en el contexto. Ninguno llega automáticamente a la imagen final porque allí solo copiamos el binario. Excluirlos desde el contexto evita exponerlos también durante la construcción.

### D2 - Dónde está la base, vista desde el contenedor

En un contenedor, `localhost` apunta al propio contenedor de la API. La base está en otro proceso y esa dirección no sirve para encontrarla.

Para comparar `go run` y `docker run`, el ejecutor prepara una base de demostración separada, cuyo nombre empieza por `lealcaf-t11-manual-`, publicada por defecto en el puerto `55433` de la computadora. Ambos modos usan esa misma base. `go run` utiliza `DB_HOST=127.0.0.1`; la API en Docker utiliza `DB_HOST=host.docker.internal` y el puerto publicado `55433`. No modificamos la base anterior del proyecto ni necesitamos sus contraseñas para ensayar el paquete. La elección de una base aislada también permite compartir las instrucciones con la pareja.

El sufijo identifica la configuración local. Si cambian las credenciales, se prepara otra base en un volumen distinto y se conservan las anteriores. Antes de las pruebas, se comprueban usuario, contraseña y nombre de base mediante una consulta SQL por la interfaz de red de PostgreSQL, y se comprueba que el puerto publicado sea accesible desde la computadora. El nombre y volumen efectivos se guardan en `docs/taller11/base_manual.txt`.

Los archivos privados `.env.gorun`, `.env.localhost` y `.env.docker` se generan por separado. El `.env` original del módulo se conserva. En Linux con Docker Engine se agrega `--add-host=host.docker.internal:host-gateway`.

Las salidas reales de los errores solicitados quedan en `docs/taller11/1_sin_variables.txt` y `2_local.txt`. En esta API la variable obligatoria de conexión es `DB_NAME`, junto con `DB_HOST`, `DB_PORT`, `DB_USER` y `DB_PASSWORD`; no se llama `DATABASE_URL` como en el ejemplo de canchas. La firma usa `JWT_SECRET` y el puerto usa `PORT`.

En un servidor usamos la dirección de la base disponible en ese entorno. Para los dos servicios del reto usamos el nombre de servicio `db`, explicado en D5.

### D3 - Los puertos

Con la configuración inicial, el argumento es `-p 127.0.0.1:8080:8080`. El primer `8080` pertenece a la computadora y es el que utiliza Bruno. El segundo pertenece al contenedor y debe coincidir con `PORT=8080`, donde escucha Go. La dirección `127.0.0.1` anterior a los puertos limita la publicación a esta computadora.

Si cambiamos el puerto externo, Bruno debe usar ese nuevo puerto. Si cambiamos el interno sin ajustar `PORT`, Docker reenviará las conexiones a un lugar donde la API no escucha. El ejecutor toma ambos puertos de `PORT`, definido en el `.env` de la raíz.

Encontramos una excepción concreta al supuesto de no tocar Go: el servidor existente tenía `Addr: "127.0.0.1:" + puerto`. Ese enlace escucha exclusivamente dentro del loopback del contenedor, por lo que publicar el puerto no alcanza. Lo cambiamos a `Addr: ":" + puerto`, para aceptar conexiones a través de su interfaz de red. No cambiamos rutas, autenticación, reglas de negocio ni persistencia.

### D4 - La hora

Comparamos la respuesta completa de la petición 3 sobre el mismo canje confirmado, usando la misma base en `go run` y `docker run`. El script conserva el JSON original, sin convertir las fechas antes de guardarlo, y compara ambos valores como instantes UTC.

<!-- D4:INICIO -->
Observacion real del canje 1: con go run, **2026-11-12T19:53:23.368194-05:00**; en el contenedor, **2026-11-13T00:53:23.368194Z**. Las dos fechas representan el mismo instante. Se guardaron las respuestas completas en peticiones.md.
<!-- D4:FIN -->

Por ejemplo, `13:00:00-05:00` y `18:00:00Z` pueden describir el mismo instante. Cambiar la representación no implica que se haya movido el canje cinco horas. El proceso puede usar una zona local distinta; tener zonas horarias disponibles no obliga a utilizar la de Ecuador como predeterminada. La conexión de este proyecto ya establece `TimeZone=America/Guayaquil` en PostgreSQL, pero eso no sustituye la revisión del JSON que recibe el cliente.

Decidimos mantener el instante y su offset explícito en las respuestas, y convertir a la hora de Ecuador al mostrarlo al usuario. No se modifica el código para corregir una diferencia que sea solo de representación. En LealCafé esta distinción afecta las fechas de compra, la caducidad de los puntos y los reportes por día.

Las decisiones D5 a D9 y la configuración de Compose se integraron a partir del paquete de José Miguel Fuentes Cedeño. Las mediciones de ese paquete corresponden a su equipo.

### D5 - Cómo encuentra la API a la base

`compose.yaml` define los servicios `api` y `db`. Compose los conecta a su red y permite resolver `db` por su nombre de servicio. La API recibe `DB_HOST=db` y `DB_PORT=5432`: este es el puerto de PostgreSQL dentro de esa red.

No usa `host.docker.internal`, `5433` ni `55433` para esta conexión. Esos puertos publicados en una computadora pertenecen al acceso desde fuera de la red de Compose. Aquí los dos servicios se comunican directamente.

### D6 - El orden de arranque

La API existente conecta, migra y carga datos al iniciar. Si PostgreSQL todavía no acepta conexiones, termina con error. Configuramos un `healthcheck` de la base con `pg_isready` por TCP y `depends_on: condition: service_healthy` en la API. No basta con que el proceso de la base haya empezado: debe estar listo para recibir conexiones. Usar TCP evita dar por listo al servidor temporal que PostgreSQL utiliza durante su inicialización.

El ejecutor comprueba tres arranques con una base nueva en el proyecto separado `lealcaf-taller11-ensayo`. En cada uno exige las seis respuestas correctas y cero reinicios de la API. Solamente ese proyecto de ensayo se elimina con `down -v`. Las evidencias quedan en `arranques_limpios.txt`, `arranque_limpio_*.txt` y sus logs.

<!-- D6:INICIO -->
Intento 1 : seis codigos correctos; reinicios de API = 0; fecha UTC 2026-11-13T00:53:49.8811766Z. Intento 2 : seis codigos correctos; reinicios de API = 0; fecha UTC 2026-11-13T00:54:04.5192282Z. Intento 3 : seis codigos correctos; reinicios de API = 0; fecha UTC 2026-11-13T00:54:19.7601488Z.
<!-- D6:FIN -->

### D7 - Dónde quedan los datos

PostgreSQL guarda sus datos en el volumen nombrado `datos_postgres`, montado en `/var/lib/postgresql/data`. `docker compose down` elimina contenedores y red, pero conserva ese volumen. Al volver a subir, los canjes siguen disponibles. El ejecutor comprueba un canje nuevo confirmado antes de `down` y vuelve a consultar exactamente ese ID después del reinicio; el resultado se guarda en `persistencia.txt`.

`docker compose down -v` sí elimina el volumen del proyecto seleccionado. En el ensayo se usa deliberadamente para comprobar bases vacías. No lo utilizamos para reiniciar el proyecto final ni el contenedor manual. La base manual también tiene su volumen propio, con el prefijo `lealcaf_t11_manual_` y un sufijo que identifica la configuración. Repetir con la misma configuración reutiliza ese volumen.

### D8 - Los secretos

Las contraseñas y la clave JWT no están escritas en `compose.yaml`. Se obtienen por interpolación del `.env` privado de la raíz. `Taller11.cmd` genera valores aleatorios locales con un generador criptográfico; si el archivo ya existe, lo reutiliza. El ejemplo público `.env.example` deja vacíos esos campos.

`.gitignore` excluye `.env`, sus variantes y el entorno `Taller11-privado.bru` de Bruno. `.dockerignore` tiene patrones recursivos para excluir secretos también dentro de subcarpetas. Las variables de la API se pasan al arrancar mediante `--env-file` o `environment` de Compose. Los tokens de las peticiones solo viven en memoria y los informes no guardan respuestas de login. Antes del commit verificamos con `git status` que los archivos privados no estén incluidos.

### D9 - Lo que se publica en la máquina

Compose publica solamente la API, por defecto en `127.0.0.1:8080`. La base de Compose no tiene `ports`; usa `5432` únicamente dentro de su red. Así puede convivir con el PostgreSQL previo y con la base manual publicada en `55433`.

Para consultar la base desde su propio contenedor, con los nombres predeterminados:

```powershell
docker compose exec db psql -U postgres -d lealcaf_taller11
```

Si se cambian `DB_USER` o `DB_NAME`, se sustituyen en ese comando. No hace falta publicar otro puerto para ejecutar `psql` dentro del servicio.

## Fuentes consultadas

- Kit y taller de la semana 11 proporcionados por el docente.
- [Docker: construcción en varias etapas](https://docs.docker.com/build/building/multi-stage/).
- [Docker: orden de arranque y service_healthy](https://docs.docker.com/compose/how-tos/startup-order/).
- [Docker: variables de entorno e interpolación](https://docs.docker.com/compose/how-tos/environment-variables/variable-interpolation/).
- [Docker: volúmenes](https://docs.docker.com/engine/storage/volumes/).

## Declaración de IA

OpenAI Codex apoyó la preparación del Dockerfile, Compose, la exclusión de secretos, el ajuste de escucha HTTP, el ejecutor para Windows, la documentación y la revisión de rutas del CI. Las mediciones de Docker y los resultados HTTP se completan solamente al ejecutar el paquete; no se presentan resultados ficticios. Cada integrante debe comprender y poder explicar las decisiones y registrar su aporte real. La herramienta y el alcance de su uso también se indican en el cuerpo de los commits.


## Semana 12 · la seguridad de la API

Integrantes, Anchundia Anchundia Luis Antonio y Fuentes Cedeño José Miguel. El punto de partida es taller11, d62846e. Los ejemplos de Chi se adaptaron a Gin, el framework vigente.

### D1 · Lo que encontramos y cerramos

La búsqueda inicial produjo 48 coincidencias en el módulo vigente. Se revisaron 39 coincidencias que construyen o ejecutan consultas, tres declaraciones de interfaz y seis líneas de mensajes, URI y llamadas de transporte. De las coincidencias SQL, 22 reciben valores derivados de HTTP o de la identidad autenticada mediante GORM o sqlc. Los valores se envían separados con ? o $1. Las órdenes de ordenación y selección son constantes.

No encontramos SQL de negocio que concatene texto recibido del cliente. idDeRuta ya convierte a entero positivo y el servicio valida los estados. La concatenación de un nombre de secuencia pertenece al mantenimiento y usa un identificador obtenido y escapado por PostgreSQL. Cada coincidencia está justificada en docs/taller12/auditoria.md.

El filtro normal devolvió 200 con canjes confirmados. El estado con comilla y el ID con texto SQL devolvieron 400. Se agregaron pruebas de regresión del filtro parametrizado y del ID inválido. La búsqueda final dio 49 coincidencias e incluye Exec(statement), que ejecuta exclusivamente el archivo fijo de permisos embebido.

La API ya verifica que un cliente consulte sus propios canjes. Las pruebas existentes comprueban el rechazo de consultas de un cliente a registros ajenos.

### D2 · El origen que puede leer nuestras respuestas

Se autorizó únicamente http://localhost:5173, el origen previsto del frontend. Se permiten GET, POST y PATCH, los métodos de la API, y las cabeceras Authorization y Content-Type. La caché del preflight dura 300 segundos.

CORS se coloca antes del registro, la recuperación y la autenticación. Se adaptó el middleware net/http go-chi/cors del kit a Gin. El preflight permitido respondió 200 con Access-Control-Allow-Origin. El no permitido respondió 200 sin esa cabecera.

El PATCH con token de administrador enviado por curl desde http://malo.example confirmó un canje nuevo con 200. CORS regula la lectura y el envío que permite el navegador. Los clientes HTTP directos pueden enviar la petición y el servidor debe comprobar token, rol y reglas del negocio. Cuando se publique el frontend, su origen exacto debe incorporarse a la lista autorizada.

### D3 · El límite del login

Se cuentan cinco solicitudes por minuto por IP únicamente en POST /login. Cada solicitud aceptada se conserva durante 60 segundos en una ventana móvil. Un mutex protege el contador frente a solicitudes concurrentes y la limpieza periódica retira IP inactivas.

Se implementó el contador en Go para integrarlo con Gin y probar el tiempo sin esperas mediante un reloj controlado. Se verificó el límite con 30 llamadas concurrentes y con cabeceras de IP falsificadas. Todas las solicitudes al login cuentan, incluidas las correctas y las inválidas.

La ráfaga real respondió 401, 401, 401, 401, 401, 429 y 429. El primer Remaining fue 4 y el sexto intento devolvió Retry-After de 60 segundos. El contador devuelve los segundos restantes hasta que caduque el intento más antiguo. Las seis peticiones funcionales conservaron 401, 200, 200, 403, 200 y 422 tanto por HTTP como en Bruno.

El contador vive en memoria de cada proceso. Reiniciar la API vacía la ventana. Varias réplicas necesitarían un contador compartido para imponer un límite global.

### D4 · Detrás de un proxy

El límite obtiene la IP de Request.RemoteAddr, separa el puerto y normaliza IPv4 e IPv6. No lee X-Forwarded-For ni X-Real-IP. Las pruebas cambiaron esas cabeceras y la sexta solicitud siguió recibiendo 429. Gin conserva SetTrustedProxies(nil).

Detrás de un proxy, RemoteAddr identifica al proxy y varios clientes podrían compartir el mismo cupo. Docker también puede presentar una dirección de su red al servidor. El despliegue deberá establecer qué proxy se considera confiable y cómo entrega la IP real. Confiar directamente en una cabecera editable permitiría cambiar de identidad para evadir el límite.

## Semana 12 · reto de mínimo privilegio en la base

### D5 · Con qué usuario se conecta la API

La API usa lealcaf_api, sin superusuario, CREATEDB, CREATEROLE, REPLICATION ni BYPASSRLS y sin ser dueño de las tablas. Tiene USAGE sobre public, sin CREATE ni permiso para tablas temporales. Las tablas pertenecen a lealcaf_migrador.

Los permisos por tabla están en permisos.sql y se comprobaron en el catálogo de PostgreSQL. sucursals, recompensas y cuenta son de solo lectura. compras y movimiento_puntos permiten SELECT e INSERT. canjes permite SELECT, INSERT y UPDATE de estado. clientes permite SELECT y UPDATE de id, el permiso necesario para obtener el bloqueo SELECT FOR UPDATE que serializa las confirmaciones. Las secuencias de las tres tablas con INSERT tienen USAGE.

Las seis peticiones, una compra adicional y la consulta de puntos pasaron con el usuario restringido. El intento de DROP TABLE canjes dentro de BEGIN y ROLLBACK produjo ERROR must be owner of table canjes. Los resultados están en reto_drop.txt, reto_usuarios.txt y reto_compra.txt.

### D6 · Quién crea y cambia las tablas

lealcaf_migrador prepara las tablas mediante el mismo binario con el argumento migrar. En Windows, desde el módulo se ejecuta go run . migrar con DB_MIGRATION_USER y DB_MIGRATION_PASSWORD. El arranque normal usa DB_USER y DB_PASSWORD y atiende peticiones sin ejecutar AutoMigrate ni sembrar datos.

Compose inicializa el rol API al crear la base. El servicio migrar espera el healthcheck TCP de db y la API espera a que migrar termine con código cero. Las credenciales de migración se pasan solo al servicio de migración. En esta demostración el usuario migrador creado por la imagen PostgreSQL tiene privilegios administrativos.

La preparación del esquema, la semilla y los permisos se aplican en una transacción. Si cambia un modelo, primero se debe ejecutar migrar. Arrancar únicamente el servidor no actualiza el esquema y una operación que necesite el nuevo modelo puede fallar.

Se comprobó el arranque con otro proyecto Compose y un volumen nuevo, y la conservación del mismo canje después de down y up en el proyecto principal del taller. Se conservaron todos los volúmenes existentes. El caso de recreación desde una base vacía se probó creando un volumen nuevo.

### D7 · El alcance del mínimo privilegio

Si quedara una inyección, los permisos de lealcaf_api permitirían leer las tablas autorizadas e insertar o actualizar dentro del alcance concedido. Eso todavía puede causar exposición o alteración de datos.

Ese rol carece de propiedad y de privilegios para borrar tablas, crear roles o modificar el esquema. El mínimo privilegio reduce el alcance de una falla. La parametrización, la validación y la autorización de cada ruta siguen siendo necesarias.

Las pruebas usan exclusivamente datos ficticios en bases de demostración propias. Las contraseñas PostgreSQL, JWT y cuentas demo se obtienen del entorno privado.

### Fuentes de la semana 12

- Taller y código por pasos de la semana 12 proporcionados por el docente.
- [Middleware CORS utilizado](https://github.com/go-chi/cors).
- [PostgreSQL 16, CREATE ROLE](https://www.postgresql.org/docs/16/sql-createrole.html).
- [PostgreSQL 16, GRANT](https://www.postgresql.org/docs/16/sql-grant.html).

OpenAI Codex colaboró en la adaptación a Gin, la auditoría, las pruebas locales, la ejecución del Runner de Bruno, las capturas, los permisos y la documentación. Los commits de esta sesión corresponden a Luis Anchundia. El historial anterior de José Miguel Fuentes Cedeño se conserva.
