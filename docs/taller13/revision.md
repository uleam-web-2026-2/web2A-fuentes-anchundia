# Revisión frente al material oficial

Se leyeron ambos PDF y todos los archivos del ZIP antes de modificar funcionalidades. Los nombres del ejemplo de reservas se adaptaron a Canje, ClienteID, solicitado/confirmado/cancelado, RolCliente y RolAdministrador de LealCafé. Las instrucciones de publicación parcial provienen de la solicitud del usuario, no del PDF: el docente evalúa lo que esté publicado al cierre y por eso la parte remota sigue pendiente de integración grupal.

| Requisito y condición «Queda listo cuando» | Resultado local y evidencia |
| --- | --- |
| Fase 0: dependencia, build, test, línea de base | coder/websocket v1.8.15; compilacion.txt; pruebas_finales.txt; Bruno 01 y peticiones.md |
| Fase 1: dueño del token y segundo cliente | ClienteID existente; propietario_test.go; cliente2 con identidad propia; no se alteró esquema ni sqlc |
| Servicio avisa después de persistir | aviso_test.go prueba éxito, transición prohibida, fallo de movimiento y fallo de COMMIT; prueba_servicio.txt |
| Fase 2: /ws fuera del guardia, sesión antes de Accept | rutas.go y avisos.go; curl real 401/403/101 en ws_*.txt |
| Origen común y conexiones individuales | OrigenFront compartido con CORS; prueba_websocket.txt incluye varios canales, token inválido/vencido, otro esquema/origen, limpieza, cola lenta y concurrencia |
| Fase 3: dos ventanas simultáneas | aviso_navegador.png, clientes ID 1 y 2; canje 4 solo al propietario |
| Logs de la misma vuelta y sin tokens | logs_avisos.txt: PATCH canje 4 y ambos cierres 101 tras varios minutos; token BORRADO |
| Seis peticiones antes, WebSocket y reto | Las tres vueltas reales de Bruno dieron 401, 200, 200, 403, 200, 422; capturas 01–03 enlazadas |
| Reto: avisar creación a administradores | TestReto con dos administradores y dos clientes; prueba_reto.txt; servicio independiente de rol y WebSocket |
| D1 a D4 y todas las preguntas del anexo A | docs/decisiones.md y su archivo histórico dentro del módulo, sin borrar semanas previas |
| Defensas Taller 12 | CORS, límite login, JWT, consultas parametrizadas, rol API restringido y migración separada; tests previos pasan |
| PostgreSQL real | pruebas_postgres.txt: concurrencia, rollback y commit con GORM y sqlc; docker.txt muestra permisos restringidos |
| Docker y Compose | Imagen lealcaf:taller13, migrador exit 0, DB healthy y API 18085; http_compose.json conserva los seis códigos |
| Corte de seis commits compilable | verificacion_corte6.txt desde una extracción independiente del commit 220f246 |
| Publicación parcial autorizada | publicacion.json verifica exactamente seis commits sobre taller12; GitHub Actions completed/success |
| Entrega completa a José | INSTRUCCIONES.md; el ZIP externo se genera a partir del historial completo, con cambios posteriores al corte y bundle que conserva autoría |
| Declaración adicional del kit | docs/hito3/declaracion_borrador.md, separada y pendiente de acuerdo de la pareja |

## Defensa oral

LealCafé permite que un cliente solicite un canje y reciba en tiempo real el cambio que confirmó un administrador. El propietario es ClienteID del JWT, nunca un dato que el navegador pueda elegir en el JSON. El servicio confirma primero la transacción de estado y puntos, y después llama a una interfaz de avisos. El transporte mantiene canales por identidad, protegidos por mutex. Así varias pestañas del mismo cliente reciben el aviso y otro cliente del mismo rol no.

Al crear un canje, el servicio usa otra interfaz sin conocer roles. Avisos selecciona las conexiones administrativas según la identidad validada al abrirlas. El reto no se resuelve enviando todo a todos ni filtrando en JavaScript.

Preguntas previsibles del docente: ¿por qué no SSE? Hoy bastaría para estos avisos; WebSocket cumple el contrato del hito. ¿Qué pasa si falla COMMIT? No se avisa; lo demuestra un falso. ¿Y con dos réplicas? Falta un bus compartido. ¿Por qué hay token en query? El navegador no permite Authorization personalizada en WebSocket; se valida antes de aceptar y se borra del log. ¿Qué pasa si cambia el rol o vence el token durante la conexión? Se conserva la instantánea de apertura hasta reconectar; está documentado como límite real. ¿Puede perderse un aviso? Sí, si no hay conexión o se llena la cola; se consulta REST para recuperar el estado.
