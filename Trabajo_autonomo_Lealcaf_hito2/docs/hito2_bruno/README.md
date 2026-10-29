# Colección Bruno · LealCaf con PostgreSQL

1. Inicie el proyecto con **Iniciar-Proyecto.cmd**. El arranque usa PostgreSQL y la misma base **fidelizacion_db** del Hito 1.
2. Abra esta carpeta como colección en Bruno y seleccione **Local**.
3. Antes del turno envíe **0 · Preparar tokens y canje nuevo**. Debe responder 201. El script inicia sesión con ambas cuentas, identifica al cliente demo por la respuesta del login, consulta el catálogo y encuentra un canje confirmado propio. Crea un canje solicitado nuevo para 4 y 5.
4. En el turno envíe **1 → 2 → 3 → 4 → 5 → 6**. Deben responder **401, 200 con token, 200, 403, 200 y 422**.
5. Ejecute nuevamente 0 antes de repetir. La petición 5 consume el canje nuevo al confirmarlo; la 6 no modifica el confirmado reservado.

Los usuarios son **cliente / cliente-2026** y **admin / admin-2026**, con roles cliente y administrador. El arranque con DEMO_DATA=true prepara un cliente ficticio dedicado y sus puntos sin convertir al cliente 1 anterior en usuario demo ni cambiar sus compras o canjes.

## Variables

`baseUrl` es la única variable persistente del entorno. Tokens e identificadores (`clienteId`, `recompensaId`, `canjeConfirmadoId`, `canjeId`) son variables de ejecución en memoria. Los IDs se obtienen del servidor; no se presupone que el cliente o un canje tengan ID 1. La sucursal 2 es Norte. Si el puerto cambia, ajuste baseUrl antes de preparar.

Al cerrar Bruno o vencer el token, ejecute nuevamente 0. No hay que pegar tokens ni editar peticiones durante el turno. La petición 2 renueva el token cliente en vivo. La API no acepta los encabezados X-Rol y X-Cliente-ID como autenticación.

## Scripts y comprobación

Los scripts usan las APIs integradas oficiales [`bru.sendRequest`, `req.setUrl`, `req.setBody` y variables de ejecución](https://docs.usebruno.com/testing/script/javascript-reference). El cuerpo JSON de 0 se establece en el script con un objeto para usar la recompensa real; no depende de cuándo se interpolen las variables de la interfaz. No requieren paquetes npm, filesystem ni Developer Mode.

Comprobación opcional con CLI desde esta carpeta: `bru run --env Local`. El recorrido debe ejecutarse completo en una misma ejecución. Los resultados de la migración actual se registrarán en docs/validacion.md después de comprobarlos; las evidencias SQLite antiguas se conservan como históricas.
