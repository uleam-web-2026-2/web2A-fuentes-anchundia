# Verificación PostgreSQL y preparación de entrega · 29 de octubre de 2026

## Disposición de la entrega

El código actual se encuentra en **hito2/**, con un módulo Go propio. **docs/** y **redocly.yaml** están en la raíz del repositorio. El módulo Go que ya existía en esa raíz corresponde a trabajo anterior: ejecutar allí `go test ./...` no comprueba el módulo anidado de Hito 2.

Desde la raíz, los comandos actuales son:

```powershell
go -C hito2 test ./...
npx.cmd @redocly/cli lint docs/openapi.yaml --format=stylish
.\hito2\Iniciar-Proyecto.cmd
```

Para arrancar se requiere PostgreSQL, la base **fidelizacion_db** y una configuración privada **hito2/.env**, creada localmente desde **hito2/.env.example** sin reemplazar una configuración que ya exista. La plantilla está incluida; el archivo privado y los binarios están excluidos de Git. El iniciador genera una clave JWT local si falta y ejecuta Go cuando no hay un binario local.

## Migración previa del 28 de octubre

- Se respaldó la base original con pg_dump antes de modificarla y se ensayó primero sobre una copia aislada.
- La migración transaccional agregó **sucursals**, **cuenta** y **sucursal_id** en compras y canjes.
- Se conservaron todos los campos de los registros históricos: 1 cliente, 1 compra, 3 recompensas y 1 movimiento de puntos; no había canjes históricos.
- La compra histórica quedó asociada inicialmente a Centro, ID 1.
- Los datos ficticios se prepararon aparte: cliente demo con ID generado, dos cuentas, compra y puntos propios y un canje confirmado propio. No se convirtió al cliente histórico en usuario de demostración.
- Las secuencias solo avanzan si están atrasadas. Repetir la preparación no duplica créditos ni renueva vencimientos.
- No se importaron a PostgreSQL los ensayos del archivo SQLite anterior. La aplicación usa PostgreSQL; SQLite se mantiene únicamente para pruebas automáticas aisladas.

## Verificaciones automatizadas previas

Estas comprobaciones corresponden al trabajo de migración, antes de las capturas manuales del alumno:

- Pasaron 36 pruebas principales de Go, con subcasos.
- Redocly CLI **2.71.2** validó el contrato sin errores ni avisos.
- Bruno CLI **4.3.0** pasó **7/7 peticiones, 6/6 pruebas y 7/7 aserciones** sobre una copia aislada restaurada del respaldo original.
- En la base real se aplicó la migración y se comparó cada campo histórico contra la copia previa: sin cambios. Se verificaron las siete tablas y las dos nuevas referencias.
- La revisión automática bloqueó repetir la colección completa sobre la base real porque crea y confirma canjes. Esa ejecución automatizada de la colección se mantuvo en la copia aislada. Las pruebas manuales posteriores del alumno se describen aparte.

## Evidencias directas tomadas por el alumno

El 28 de octubre, después de iniciar la versión PostgreSQL, el alumno ejecutó la preparación 0 y las seis peticiones en Bruno y guardó capturas directas en **docs/hito2_capturas/**:

| Archivo | Evidencia |
|---|---|
| 0_respuesta.png | Preparación y creación del canje: 201 |
| 1_respuesta.png | Ruta protegida sin token: 401 |
| 2_respuesta.png | Login correcto: 200 y token |
| 3_respuesta.png | Consulta propia con token: 200 |
| 4_respuesta.png | Cliente intenta confirmar: 403 |
| 5_respuesta.png | Administrador confirma: 200 |
| 6_respuesta.png | Transición prohibida: 422 |
| pruebas.png | Pruebas Go aprobadas |
| contrato.png | Contrato OpenAPI válido |

El alumno compartió la salida de `go test ./...`, ejecutado entonces desde la carpeta independiente **LealCaf_Hito2**: `ok fidelizacion/internal/fidelizacion 11.201s`. La captura **pruebas.png** muestra `go test -count=1 ./...` aprobado en **8.376s**, sin caché. En esta entrega, el comando equivalente desde la raíz es **go -C hito2 test ./...**.

La captura del validador muestra **Your API description is valid**, sin errores ni advertencias del contrato. El registro de instalación compartido por el alumno y la validación final acreditan **Redocly CLI 2.72.0**. En PowerShell se usó **npx.cmd**; el anuncio informativo de Redocly y el aviso de actualización de npm no son errores de OpenAPI.

Los HTML y reportes de la ejecución anterior se conservan como históricos. Sus fechas y puertos describen aquella ejecución; los PNG actuales anteriores son las capturas directas del alumno.

## Comprobaciones pendientes de los integrantes

Repetir el arranque y el recorrido desde la carpeta de entrega en la computadora de exposición; leer el código, ensayar con reloj y explicar las decisiones. Revisar las contribuciones y commits reales de ambos y comprobar que el push esté en el repositorio de la pareja dentro de **uleam-web-2026-2** antes del cierre. Este informe no acredita participación, commits ni publicación que no se hayan verificado.

Para consultar las tablas: **pgAdmin → PostgreSQL 18 → Databases → fidelizacion_db → Schemas → public → Tables → Refresh**.

## Revisión final de la copia de entrega · 29 de octubre

- En **LealCaf_Entrega**, `go vet ./...` y `go test ./...` comprobaron el módulo anterior de la raíz sin fallos.
- `go -C hito2 vet ./...` pasó y `go -C hito2 test -count=1 ./...` terminó con **ok fidelizacion/internal/fidelizacion 10.393s**, sin usar el resultado en caché.
- Redocly CLI **2.72.0** volvió a validar **docs/openapi.yaml** sin errores ni advertencias.
- GitHub Actions conserva la comprobación anterior y añade vet y test del módulo **hito2**.
- Los enlaces de la documentación apuntan a la nueva disposición. Las nueve capturas directas del alumno están en **docs/hito2_capturas/**.
- El historial revisado conserva commits reales de ambos integrantes. La declaración de IA se encuentra en **docs/decisiones.md**.
- La configuración privada de esta computadora quedó preparada localmente en **hito2/.env**, excluida de Git. El repositorio incluye **hito2/.env.example**, sin la contraseña real ni la clave de firma. Los binarios, datos y el inventario obsoleto del paquete anterior están excluidos de la entrega.

La publicación se considera comprobada solo cuando el commit enviado aparezca en la rama main del repositorio de la organización. La lectura del código y el ensayo oral siguen a cargo de la pareja.
