# Decisiones del proyecto - Semana 11

Proyecto: **LealCafé**. Integrantes: José Miguel Fuentes Cedeño y Luis Antonio Anchundia Anchundia.

Partimos de la rama `taller10`, revisión `5f6e6f6`. El módulo vigente está en `Trabajo_autonomo_Lealcaf_hito2/`. Las decisiones de las semanas anteriores siguen en [el documento del módulo](../Trabajo_autonomo_Lealcaf_hito2/docs/decisiones.md).

## Semana 11 - el contenedor

### D1 - Qué entra en la imagen

El `Dockerfile` tiene dos etapas. La primera usa `golang:1.27`, descarga las dependencias del módulo y compila con `CGO_ENABLED=0`. Copiamos primero `go.mod` y `go.sum` para reutilizar la descarga cuando solo cambia el código. Como hay varias entregas antiguas en el repositorio, el `COPY` selecciona expresamente `Trabajo_autonomo_Lealcaf_hito2/`.

La imagen final parte de `gcr.io/distroless/static-debian13:nonroot` y recibe solamente `/servidor`. No copiamos el compilador, el código fuente, las pruebas ni las dependencias descargadas. Las rutas implementadas no necesitan leer la carpeta `docs/` del disco. El proceso arranca con el usuario nonroot de la imagen base, no como root.

<!-- D1:INICIO -->
Tamano real final: **65804243 bytes** (62.76 MiB). Etapa de compilacion: **2573010783 bytes** (2453.81 MiB). Usuario observado: **65532**. Evidencias: imagenes.txt, capas.txt, contexto.txt y usuario.txt.
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
Observacion real del canje 1: con go run, **2026-11-12T15:15:56.02458+01:00**; en el contenedor, **2026-11-12T14:15:56.02458Z**. Las dos fechas representan el mismo instante. Se guardaron las respuestas completas en peticiones.md.
<!-- D4:FIN -->

Por ejemplo, `13:00:00-05:00` y `18:00:00Z` pueden describir el mismo instante. Cambiar la representación no implica que se haya movido el canje cinco horas. El proceso puede usar una zona local distinta; tener zonas horarias disponibles no obliga a utilizar la de Ecuador como predeterminada. La conexión de este proyecto ya establece `TimeZone=America/Guayaquil` en PostgreSQL, pero eso no sustituye la revisión del JSON que recibe el cliente.

Decidimos mantener el instante y su offset explícito en las respuestas, y convertir a la hora de Ecuador al mostrarlo al usuario. No se modifica el código para corregir una diferencia que sea solo de representación. En LealCafé esta distinción afecta las fechas de compra, la caducidad de los puntos y los reportes por día.

## Fuentes consultadas

- Kit y taller de la semana 11 proporcionados por el docente.
- [Docker: construcción en varias etapas](https://docs.docker.com/build/building/multi-stage/).
- [Docker: orden de arranque y service_healthy](https://docs.docker.com/compose/how-tos/startup-order/).
- [Docker: variables de entorno e interpolación](https://docs.docker.com/compose/how-tos/environment-variables/variable-interpolation/).
- [Docker: volúmenes](https://docs.docker.com/engine/storage/volumes/).

## Declaración de IA

OpenAI Codex apoyó la preparación del Dockerfile, Compose, la exclusión de secretos, el ajuste de escucha HTTP, el ejecutor para Windows, la documentación y la revisión de rutas del CI. Las mediciones de Docker y los resultados HTTP se completan solamente al ejecutar el paquete; no se presentan resultados ficticios. Cada integrante debe comprender y poder explicar las decisiones y registrar su aporte real. La herramienta y el alcance de su uso también se indican en el cuerpo de los commits.
