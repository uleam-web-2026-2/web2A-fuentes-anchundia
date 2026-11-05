# Colección Bruno · LealCaf con PostgreSQL

**Semana 10:** se conservan las rutas y los seis códigos HTTP. Antes de ejecutar 0 completa `clientePassword` y `adminPassword` en tu entorno Local; el paquete los deja vacíos. Con `DEMO_DATA=true`, el servidor necesita esas mismas contraseñas en `.env` como `DEMO_CLIENT_PASSWORD` y `DEMO_ADMIN_PASSWORD`. No modifica contraseñas de cuentas existentes.

1. Inicie el proyecto con **Iniciar-Proyecto.cmd**. El arranque usa PostgreSQL y la base **fidelizacion_db** del proyecto.
2. Abra esta carpeta como colección en Bruno y seleccione **Local**.
3. Antes del turno envíe **0 · Preparar tokens y canje nuevo**. Debe responder 201. El script inicia sesión con ambas cuentas, identifica al cliente demo por la respuesta del login, consulta el catálogo y encuentra un canje confirmado propio. Crea un canje solicitado nuevo para 4 y 5.
4. En el turno envíe **1 → 2 → 3 → 4 → 5 → 6**. Deben responder **401, 200 con token, 200, 403, 200 y 422**.
5. Ejecute nuevamente 0 antes de repetir. La petición 5 consume el canje nuevo al confirmarlo; la 6 no modifica el confirmado reservado.

Los usuarios son **cliente** y **admin**, con roles cliente y administrador. Sus contraseñas se configuran localmente; no se distribuyen valores por defecto. El arranque con DEMO_DATA=true prepara un cliente ficticio dedicado y sus puntos sin convertir al cliente 1 anterior en usuario demo ni cambiar sus compras o canjes.

## Variables

`baseUrl`, `clientePassword` y `adminPassword` son las variables configurables del entorno Local. Completa las contraseñas localmente y no publiques el entorno con esos valores. Tokens e identificadores (`clienteId`, `recompensaId`, `canjeConfirmadoId`, `canjeId`) son variables de ejecución en memoria. Los IDs se obtienen del servidor; no se presupone que el cliente o un canje tengan ID 1. La sucursal 2 es Norte. Si el puerto cambia, ajuste baseUrl antes de preparar.

Al cerrar Bruno o vencer el token, ejecute nuevamente 0. No hay que pegar tokens ni editar peticiones durante el turno. La petición 2 renueva el token cliente en vivo. La API no acepta los encabezados X-Rol y X-Cliente-ID como autenticación.

## Scripts y comprobación

Los scripts usan las APIs integradas oficiales [`bru.sendRequest`, `req.setUrl`, `req.setBody` y variables de ejecución](https://docs.usebruno.com/testing/script/javascript-reference). El cuerpo JSON de 0 se establece en el script con un objeto para usar la recompensa real; no depende de cuándo se interpolen las variables de la interfaz. No requieren paquetes npm, filesystem ni Developer Mode.

Comprobación opcional con CLI desde esta carpeta: `npx.cmd --yes @usebruno/cli@4.3.0 run --env Local`. El recorrido debe ejecutarse completo en una misma ejecución. Como antecedente histórico, la revisión del 29/10/2026 pasó 7/7 peticiones, 6/6 pruebas y 7/7 aserciones sobre PostgreSQL temporal. Resultados y límites en `../validacion.md`; los PNG anteriores se conservan como respaldo. Las evidencias nuevas corresponden a `../taller10/`.

Si el servidor de ensayo usa otro puerto, añadir `--env-var baseUrl=http://127.0.0.1:18082`. Para guardar un informe sin tokens ni contraseñas, usar `--reporter-json resultados.json --reporter-skip-all-headers --reporter-skip-body`. La colección crea y confirma un canje: usar los datos ficticios dedicados y recordar que se consumen puntos.
