# Guion de demostración · LealCaf

**Web II A · Grupo P · Fuentes Cedeño José Miguel y Anchundia Anchundia Luis Antonio.** Según el horario: viernes **30 de octubre de 2026, 09:40**, bloque 2 (09:30–10:00). Llegar antes del bloque con todo iniciado. Cierre del repositorio: jueves **29 de octubre, 23:59**; hacer push antes de esa hora.

## Antes del turno

- Mantener PostgreSQL iniciado y preparar **hito2/.env** local desde **hito2/.env.example** si falta; el archivo privado no se incluye en Git. Abrir **hito2/Iniciar-Proyecto.cmd** y comprobar la conexión a **fidelizacion_db**. Abrir la colección **docs/hito2_bruno**, el contrato **docs/openapi.yaml** y dos archivos de **hito2/internal/fidelizacion/**.
- Seleccionar **Local** en Bruno. Ejecutar **0 · Preparar tokens y canje nuevo**: hace login de ambas cuentas de prueba y elige una recompensa activa y crea un canje nuevo en Norte. Verificar `201` y el ID guardado; no escribir tokens ni editar peticiones durante el turno.
- Conservar el canje confirmado propio que la preparación identifica como **canjeConfirmadoId** para la petición 6. El saldo inicial permite ensayos, pero cada confirmación gasta puntos: si se agota, registrar una compra de prueba válida antes del próximo ensayo. No cambiar los estados finales a mano ni reutilizar como demo los registros originales de otros clientes.
- Abrir una terminal en la raíz del repositorio. Ejecutar `npx.cmd @redocly/cli lint docs/openapi.yaml --format=stylish` y `go -C hito2 test ./...`. Hito 2 tiene su propio módulo; ejecutar Go en la raíz sin `-C hito2` revisaría el proyecto anterior. Tener **docs/hito2_capturas/** disponible: el alumno tomó las capturas actuales de Bruno, pruebas y contrato.
- Ensayar los dos integrantes, con reloj, hasta completar las demostraciones en menos de cuatro minutos.

Solo hay **dos roles**: cliente y administrador. Los usuarios de prueba son `cliente` y `admin`; una sucursal indica lugar de operación y no añade un rol.

## Cinco minutos

| Tiempo | Acción y frase breve |
|---|---|
| 0:00–0:25 | Mostrar `git status` y `git log -1`: trabajo entregado antes del cierre. El docente sortea quién hace las demostraciones 1 y 2; el otro hace 3, 4 y 5. |
| 0:25–1:00 | **Contrato.** `npx.cmd @redocly/cli lint docs/openapi.yaml --format=stylish`. «Este es nuestro contrato. El validador no encuentra errores ni avisos. Aquí está security, y aquí el 401, el 403 y el 422». |
| 1:00–1:50 | **Identidad.** Enviar 1, 2 y 3. «Sin token, GET /canjes/{canjeConfirmadoId} responde 401. Iniciamos sesión como cliente y recibimos el token. Con el token, su propio canje responde 200». |
| 1:50–2:35 | **Permiso.** Enviar 4 y 5. «Cliente no puede confirmar: 403. Administrador sí puede: 200». Ambas usan el nuevo canje preparado, que al confirmar descuenta puntos. |
| 2:35–3:20 | **Regla.** Enviar 6 y ejecutar `go -C hito2 test ./...`. «Un canje confirmado no puede cancelarse: 422. La recompensa ya se entregó y los puntos se descontaron. Estas son nuestras pruebas, en verde». Decir la última frase solo si el resultado está en verde. |
| 3:20–3:50 | **Estructura.** Mostrar `hito2/internal/fidelizacion/`, `manejadores.go`, `servicio.go` y la sección «Semana 8 · estructura». «El manejador interpreta HTTP; el servicio decide la transición; el repositorio consulta y guarda». |
| 3:50–4:50 | Dos preguntas del docente: una por integrante, sin pasarla al compañero. |
| 4:50–5:00 | Registro de notas. |

Las respuestas esperadas son **401, 200, 200, 403, 200, 422**. Si una petición falla hay un reintento de 20 segundos; después se usa su captura de respaldo, que limita esa demostración a la mitad. Lo que no se muestra dentro de cinco minutos vale cero.

## Las seis preguntas publicadas, adaptadas al proyecto

### 1. ¿En qué línea se decide que esta ruta pide token?

Abrir `hito2/internal/fidelizacion/rutas.go` y señalar el middleware de autenticación aplicado al grupo que contiene `GET /canjes/:id` y `PATCH /canjes/:id/estado`. Después abrir `hito2/internal/fidelizacion/autenticacion.go`, donde se lee Bearer y se verifica el JWT. La numeración de línea puede variar; ensayar encontrar la llamada concreta, no memorizar un número.

**Respuesta:** «Todas estas rutas entran por el grupo protegido. Antes de llegar al manejador se verifica el JWT; sin uno válido la petición termina en 401. Login queda fuera del grupo».

### 2. ¿Qué lleva el token? Si cambian el rol, ¿qué responde y por qué?

**Respuesta:** «Lleva identificador de la cuenta en sub, rol, cliente_id, iat y exp, además del emisor lealcaf y audiencia lealcaf-api. Dura una hora. No lleva la contraseña. Si alguien modifica rol, la firma HS256 deja de coincidir: el servidor devuelve 401. Tener un token válido de cliente tampoco permite confirmar: ahí responde 403».

Mostrar validación de firma, método, vencimiento, emisor y audiencia en `autenticacion.go`. El JWT está firmado, no cifrado.

### 3. ¿Por qué confirmar es solo del administrador?

**Respuesta:** «Confirmar significa que la cafetería entregó la recompensa y debe descontar puntos. Si el cliente pudiera confirmarse, podría declarar una entrega que caja no realizó. El cliente solicita; el administrador comprueba y confirma o rechaza. La regla se mantiene en Centro y Norte».

### 4. Si sacan la ruta del grupo protegido, ¿qué prueba avisa?

**Respuesta:** «La prueba de integración que llama una ruta protegida sin Authorization espera 401. Al quitar la protección deja de recibirlo y falla. También verificamos token alterado y rol cliente intentando confirmar, que debe responder 403».

Abrir las pruebas y señalar el caso real de ruta sin token. No responder solo «todas las pruebas»; ubicar cuál tiene esa expectativa.

### 5. ¿Por qué 422 y no 400 ni 403? ¿Dónde se valida?

**Respuesta:** «El JSON es válido y cancelado es un estado conocido: no es un error de sintaxis 400. El token es de administrador y tiene permiso: no es 403. Falla la regla porque el canje ya está confirmado: el servicio rechaza esa transición y el manejador lo traduce a 422. No cambia estado ni puntos». Abrir la validación de transición en `servicio.go` y la traducción a HTTP en `manejadores.go`.

### 6. ¿Por qué eligieron estas capas? ¿Qué cambia al cambiar de base?

**Respuesta:** «Separamos la petición, las reglas y la persistencia para poder revisar la máquina de estados sin depender de Gin ni de SQL. El proyecto es pequeño y mantenemos un paquete con archivos de responsabilidad clara. Al cambiar de base revisamos la conexión, el repositorio y los detalles del modelo ligados a GORM; las reglas de Canje y sus peticiones deben conservarse. Lo comprobamos con pruebas».

## Pendientes que requieren a los integrantes

- Leer todo el código y ajustar estas respuestas a lo que puedan explicar con sus propias palabras.
- Hacer contribuciones reales y commits propios de ambos; no inventar historial de participación.
- Verificar que la declaración de IA refleje el uso real.
- Repetir en la computadora de exposición la conexión PostgreSQL a fidelizacion_db, preparación 0, seis respuestas, validador y pruebas; realizar el ensayo con reloj. Las capturas actuales ya fueron tomadas, pero no sustituyen la demostración en vivo.
- Revisar cambios, hacer commit y push al repositorio de la pareja en `uleam-web-2026-2` antes del cierre. Un ZIP local ni un repositorio personal sustituyen esa entrega.
