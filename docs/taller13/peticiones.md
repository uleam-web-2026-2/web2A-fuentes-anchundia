# Semana 13 · las seis peticiones del hito 2

| Petición | Antes (taller 12) | Con el WebSocket | Con el reto |
| --- | --- | --- | --- |
| 1 sin token | 401 | 401 | Pendiente |
| 2 login | 200 | 200 | Pendiente |
| 3 con token | 200 | 200 | Pendiente |
| 4 rol sin permiso | 403 | 403 | Pendiente |
| 5 cambio permitido | 200 | 200 | Pendiente |
| 6 transición prohibida | 422 | 422 | Pendiente |

Ejecución real en Bruno 4.3.0, colección existente LealCaf · Hito 2, API local en puerto 18083 y PostgreSQL 16 aislado. La petición 0 prepara tokens y un canje nuevo; no cuenta entre las seis. Los tokens permanecen en memoria y las credenciales en un entorno privado ignorado por Git. Se espera al menos un minuto entre recorridos para respetar el límite del login.

- Antes: [captura real](bruno/01-antes.png), [texto de accesibilidad del resultado](bruno/01-antes.txt). Las siete peticiones del runner pasaron.
- WebSocket: [captura real](bruno/02-websocket.png), [texto observado](bruno/02-websocket.txt). Las siete pasaron; canje nuevo 3.
- Navegador: otra ejecución real del mismo runner creó y confirmó el canje 4 a las 09:49:13 (America/Guayaquil). [Dos ventanas simultáneas](aviso_navegador.png): cliente ID 1 recibe `estado_cambiado`, cliente2 ID 2 permanece sin avisos. [Logs de la misma ejecución](logs_avisos.txt), incluidos ambos cierres con 101. Las conexiones sobrevivieron varios minutos sin mensajes.
