# Revisión de entrega del Taller 12

Integrantes, Anchundia Anchundia Luis Antonio y Fuentes Cedeño José Miguel.

| Criterio docente | Peso | Resultado local y evidencia |
| --- | --- | --- |
| Auditoría e inyección | 30 | Búsquedas inicial y final, tabla revisada y respuestas reales 200/400/400 en inyeccion.txt |
| CORS | 20 | Preflight permitido con cabecera, no permitido sin cabecera y PATCH real de curl con 200 |
| Límite en login | 20 | Cinco 401, sexto y séptimo 429, Retry-After de 60 segundos y seis peticiones conservadas |
| Decisiones D1 a D4 | 10 | Documentadas en docs/decisiones.md |
| Reto de PostgreSQL | 20 | Usuario restringido, propietario de migraciones separado, permisos por tabla, DROP rechazado, seis peticiones y arranque desde volumen nuevo |

La tabla relaciona los criterios con la evidencia obtenida. La nota corresponde al docente.

## Comprobaciones realizadas

- go build, go vet y go test del módulo vigente pasaron. go_test.txt incluye pruebas reales de PostgreSQL para los adaptadores GORM y sqlc.
- Las tres etapas conservaron los seis códigos. Bruno mostró siete peticiones aprobadas en cada recorrido, incluida la preparación.
- Las capturas originales se obtuvieron de Bruno Desktop. La conversión a PNG conservó los píxeles.
- El usuario restringido también registró una compra y consultó los puntos.
- El catálogo muestra las tablas de lealcaf_migrador y los permisos limitados de lealcaf_api.
- Otro proyecto Compose con un volumen nuevo arrancó y pasó las seis peticiones. Se conservó ese volumen al detenerlo.
- El mismo canje siguió confirmado después de down y up sin eliminar el volumen principal del taller.

## Pendientes de entrega y participación

- Publicar taller12 después de la autorización de Luis.
- Verificar GitHub Actions para el último commit que se publique.
- Los commits nuevos de esta sesión pertenecen a Luis Anchundia. El historial previo de José Miguel se conserva. Al último fetch no había una rama remota taller12 ni aportes nuevos del compañero para este taller. El criterio de participación de ambos necesita un aporte real de José Miguel con su identidad.

## Preparación de la defensa

Ambos integrantes deben poder explicar por qué los valores se pasan como parámetros, qué nombres de tablas son constantes, por qué CORS responde antes de autenticar, qué IP usa el límite y quién ejecuta las migraciones.

La prueba de navegador de CORS es ilustrativa en el PDF y no exige una captura adicional. Las comprobaciones funcionales requeridas ya se ejecutaron sobre la API propia con datos ficticios.

### Preguntas para practicar

- ¿Por qué una comilla en el filtro no altera el SQL? El marcador mantiene la expresión fija y el servicio también valida el estado.
- ¿Por qué curl pudo confirmar un canje desde otro origen? La autorización del servidor aceptó su token y su rol. CORS es una política aplicada por navegadores.
- ¿Por qué el sexto login devuelve 429? Los cinco intentos anteriores aún están dentro de la ventana móvil de 60 segundos.
- ¿Por qué se concede UPDATE de id sobre clientes? PostgreSQL exige permiso UPDATE para el bloqueo FOR UPDATE que protege confirmaciones concurrentes.
- ¿Qué ocurre al cambiar un modelo? El comando migrar debe aplicar el cambio antes de servir la nueva versión.
- ¿Qué podría hacer una inyección con el usuario limitado? Leer o modificar dentro de los permisos concedidos. El rol carece de propiedad para borrar tablas.

## Verificación de José Miguel

Ejecuté las tres pruebas del límite de inicio de sesión en Windows PowerShell 5.1. Las tres pasaron y el paquete reportó `ok` en 0.669 s.

Comprobé que el límite cuenta los intentos dentro de una ventana de tiempo, funciona con solicitudes simultáneas y se aplica al inicio de sesión. Las cabeceras que envía el cliente no sirven para cambiar la IP que usa el contador.
