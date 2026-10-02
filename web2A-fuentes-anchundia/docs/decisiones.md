# Decisiones de Diseño - Mesa de Ayuda (Web 2A)

## 1. Enrutador: `go-chi/chi/v5`
Se seleccionó `chi` por su diseño idiomático que respeta la interfaz estándar `net/http`, su excelente rendimiento y su soporte nativo para la composición de middleware mediante sub-enrutadores y el método `Use()`.

## 2. Manejo de Concurrencia
Para el almacenamiento de tickets en memoria se implementó una estructura protegida por un `sync.Mutex`. Esto previene condiciones de carrera (*race conditions*) cuando múltiples clientes realizan peticiones concurrentes de lectura y escritura.

## 3. Contrato de Respuestas JSON Unificado
Todas las salidas de la API (tanto éxitos como errores) siguen una estructura envolvente estricta (`ok`, `datos` o `error` con código y mensaje). Esto simplifica el trabajo del cliente al consumir los endpoints, garantizando que nunca se devuelvan textos planos inesperados, incluso en errores 404 o 405.

## 4. Middleware Propios
* **Registro (*Logging*):** Mide el tiempo de ejecución y registra de forma uniforme el método, la ruta y el código de estado HTTP de cada petición.
* **Recuperación (*Recovery*):** Captura pánicos imprevistos en los controladores para evitar que el proceso del servidor finalice abruptamente, respondiendo con un código HTTP 500 estructurado.