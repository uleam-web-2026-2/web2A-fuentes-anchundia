# Semana 13 · las seis peticiones del hito 2

| Petición | Antes (taller 12) | Con el WebSocket | Con el reto |
| --- | --- | --- | --- |
| 1 sin token | 401 | Pendiente | Pendiente |
| 2 login | 200 | Pendiente | Pendiente |
| 3 con token | 200 | Pendiente | Pendiente |
| 4 rol sin permiso | 403 | Pendiente | Pendiente |
| 5 cambio permitido | 200 | Pendiente | Pendiente |
| 6 transición prohibida | 422 | Pendiente | Pendiente |

Ejecución real en Bruno 4.3.0, colección existente LealCaf · Hito 2, API local en puerto 18083 y PostgreSQL 16 aislado. La petición 0 prepara tokens y un canje nuevo; no cuenta entre las seis. Los tokens permanecen en memoria y las credenciales en un entorno privado ignorado por Git. Se espera al menos un minuto entre recorridos para respetar el límite del login.

- Antes: [captura real](bruno/01-antes.png), [texto de accesibilidad del resultado](bruno/01-antes.txt). Las siete peticiones del runner pasaron.
