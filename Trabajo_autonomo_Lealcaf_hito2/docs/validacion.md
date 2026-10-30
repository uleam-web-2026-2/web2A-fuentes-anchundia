# Verificación de LealCaf · Hito 2

## Estado de la entrega

El proyecto mantiene Go, Gin, GORM, PostgreSQL, JWT y la arquitectura por capas. La limpieza documental deja esta carpeta centrada en el Hito 2: se retiraron la presentación, el boceto, las capturas y el reporte exclusivos de la entrega anterior. La base de datos, el código y las pruebas se conservan.

La documentación útil que tenía nombres anteriores se adaptó a [hito2_modelo.md](hito2_modelo.md) y [hito2_tecnico.md](hito2_tecnico.md). El modelo incluye las seis entidades, sus relaciones, estados y endpoints; la guía técnica explica configuración, pruebas y el flujo HTTP. El código es la fuente de las definiciones completas de los structs.

El README principal del repositorio enlaza a estos archivos y a las evidencias. El workflow de GitHub Actions comprueba exclusivamente el módulo `Trabajo_autonomo_Lealcaf_hito2`, con la versión de su `go.mod` y los comandos build, vet y test.

## Matriz de requisitos

| Requisito | Estado inicial revisado | Actuación | Estado actual | Archivo | Evidencia |
|---|---|---|---|---|---|
| Entidades y relaciones | Seis entidades y Cuenta implementadas | Conservadas; catálogo real comprobado | Siete tablas y nueve FK | `internal/fidelizacion/modelos.go`, `db.go` | `hito2_capturas/auditoria_final_postgres.txt` |
| Cuentas y JWT | bcrypt, emisión y validación implementados | Verificadas firma, caducidad, claims y cuentas vigentes | Aprobado | `autenticacion.go`, `autenticacion_test.go` | Login 200; tokens alterados/vencidos rechazados en pruebas |
| Protección y dos roles | Grupo protegido y SoloRol existentes | Probadas todas las rutas sin token y acciones por rol | 401/403/200 correctos | `rutas.go`, `manejadores.go` | `auditoria_final_api.json` y capturas Bruno |
| CRUD mínimo de Canje y estados | Crear, listar, consultar y cambiar estado implementados | Recorrido real y pruebas de dominio | Aprobado; transición prohibida 422 | `servicio.go`, `dominio_test.go` | `bruno_ui_0.jpg` a `bruno_ui_6.jpg` |
| Atomicidad y concurrencia | Bloqueos y transacción implementados | Añadida prueba simultánea PostgreSQL | Un éxito, un rechazo y único descuento | `postgres_test.go` | `verificacion_actual_pruebas.txt` |
| Capas | HTTP, reglas y persistencia separados | Verificada delegación; documentación adaptada | Coincide con estructura actual | `manejadores.go`, `servicio.go`, `repositorio.go` | [Semana 8 · estructura](decisiones.md#semana-8--estructura) |
| Contrato y vista | Contrato completo, HTML local | Validación y comparación del contenido embebido | Sin errores ni avisos; ocho operaciones y tres planes | `openapi.yaml`, `api.html`, `../redocly.yaml` | Registros de Redocly en capturas |
| Hoja y colección | Usuarios, matriz, preparación y seis peticiones existentes | Corregidas rutas y comprobados códigos | Aprobado | `hito2_hoja.md`, `hito2_bruno/` | 7/7 peticiones, 6/6 pruebas y 7/7 aserciones |
| Capturas de respaldo | Evidencias previas disponibles | Ocho capturas nuevas de Bruno Desktop | Imágenes reales, sin JWT activo en login | `hito2_capturas/bruno_ui_*.jpg` | [Índice](hito2_capturas/index.html) |
| Guion y seis preguntas | Guion existente con rutas incorrectas | Comandos y funciones reales, reparto por sorteo | Preparado para ambos; ensayo humano pendiente | `guion_hito2.md` | Cronograma de cinco minutos y preguntas 1–6 |
| Configuración y entrega | .env local; README y CI con rutas ajenas | Corregidos iniciador, enlaces y workflow; plantilla disponible | Configuración privada excluida de Git | `../README.md`, `../scripts/iniciar-demo.ps1`, `.github/workflows/ci.yml` en la raíz Git | Arranque y recorrido local aprobados; CI remoto pendiente de publicación |

## Verificaciones registradas

- **Go 1.27.1:** descarga de módulos, build y vet aprobados. La última auditoría sin caché aprobó 36 pruebas principales con subcasos en 6.067 s y omitió la prueba opcional PostgreSQL sin DSN.
- **Concurrencia PostgreSQL 18.6:** una ejecución previa en instancia temporal aprobó las 37 pruebas principales, incluida la prueba adicional. No se debe presentar el SKIP posterior como una nueva ejecución de esa prueba.
- **OpenAPI:** Redocly CLI 2.72.1 validó sin errores ni advertencias. El HTML contiene el mismo contrato y usa el recurso JavaScript local.
- **PostgreSQL real:** consultas de solo lectura confirmaron siete tablas, nueve relaciones y una cuenta de cada rol. La auditoría de API confirmó 401 en las siete rutas protegidas, login y consultas autenticadas 200. El saldo demo entonces fue 99840 y la recompensa mínima costaba 40.
- **Privacidad:** `.env` está excluido de Git; se verificaron variables requeridas y una clave JWT de al menos 32 bytes sin registrar sus valores. No se encontraron coincidencias de los secretos privados en los archivos de entrega revisados.
- **Enlaces:** se verifican README, documentos activos e índice de evidencias después de la limpieza. No hay dependencias del servidor en los documentos retirados.

Los resultados nuevos de la limpieza se guardan por separado con el prefijo `limpieza_hito2_` en `hito2_capturas/`; permiten distinguirlos de las ejecuciones anteriores.

## Comprobación posterior a la limpieza

El acceso de arranque quedó unificado en **Iniciar-Proyecto.cmd**: se retiró el CMD duplicado y se actualizaron sus referencias en README, guía técnica e instrucciones previas. El script interno `scripts/iniciar-demo.ps1` se conserva porque el iniciador principal lo utiliza. Su sintaxis PowerShell es válida y se comprobó que el destino existe.

Después de retirar ese acceso duplicado, `go build ./...` y `go vet ./...` pasaron y `go test -count=1 ./...` terminó correctamente en **7.536 s**. Redocly volvió a validar sin errores ni avisos. La API local respondió 401 en las siete rutas sin token, 200 en login y consultas autorizadas y 403 en el listado global con rol cliente. No se crearon ni confirmaron canjes en esta repetición. La prueba opcional de concurrencia PostgreSQL no se repitió sobre la base original.

También se verificó que no queden archivos ni carpetas con nombre de Hito 1 dentro de este proyecto, que los enlaces locales existan y que `.env` continúe excluido de Git. El código, la configuración privada, los datos PostgreSQL y las capturas reales se conservaron.

`go mod download`, `go build ./...` y `go vet ./...` terminaron correctamente. `go test -count=1 -v ./...` aprobó 36 pruebas principales con subcasos en 9.608 s; la prueba opcional PostgreSQL se omitió al no configurar su DSN. No se alteró el código para limpiar documentación. Redocly volvió a validar sin errores ni avisos y se comprobó que la vista HTML contiene exactamente el mismo contrato.

La comprobación local utilizó únicamente login y consultas: las siete rutas sin token devolvieron 401; login, catálogo, puntos, listado del administrador y canje propio devolvieron 200; el listado global con rol cliente devolvió 403. El resultado está en `limpieza_hito2_api.json`, sin tokens ni contraseñas. No se ejecutaron compras, confirmaciones ni migraciones para esta comprobación. El YAML de CI es válido y sus pasos apuntan exclusivamente al módulo del Hito 2.

## Ejecución real en Bruno Desktop

Se abrió la colección actual en **Bruno Desktop 4.3.0**, se seleccionó **Local** y se enviaron **0–6 en orden** contra **localhost:8080**. Respuestas observadas: **201, 401, 200, 200, 403, 200 y 422**. Después se ejecutó el Runner de la interfaz: **All 7, Passed 7, Failed 0, Skipped 0**, con seis pruebas y siete aserciones aprobadas.

Las capturas directas, sin edición, están en `bruno_ui_0.jpg` a `bruno_ui_6.jpg` y `bruno_ui_resumen.jpg`. La imagen del login muestra la pestaña Tests y el estado 200, acreditando token Bearer y duración de una hora sin guardar un JWT activo.

Cada recorrido creó y confirmó un canje demo de 40 puntos: dos canjes y 80 puntos consumidos durante esas capturas. La colección CLI sobre el servidor local también pasó; su informe sin headers ni cuerpos privados está en `verificacion_local_bruno.json`. Las capturas previas y registros de otros ensayos del Hito 2 siguen diferenciados en el índice.

## Cambios conservados y limpieza documental

La revisión respetó el avance existente: modelos, migración conservadora, cuentas, JWT, permisos, reglas, rutas y capas. Se corrigieron rutas documentales, el lanzador de VS Code, el iniciador que podía ejecutar un binario antiguo y las rutas de GitHub Actions. Se añadió la prueba PostgreSQL de concurrencia y se prepararon evidencias reales.

La limpieza actual retira seis archivos exclusivos de la presentación anterior y adapta dos documentos a nombres de Hito 2. Se realizó un respaldo de los documentos antes de retirarlos, fuera de la carpeta del proyecto. Se mantienen las pruebas de compatibilidad con esquemas anteriores porque protegen los datos; no se eliminaron tablas ni registros PostgreSQL. No se hicieron commits ni push.

## Pendientes reales

1. Los dos integrantes deben comprender el código y ensayar el guion con reloj. La documentación preparada no acredita el ensayo oral.
2. Revisar los cambios, hacer commits con aportes reales y publicar únicamente con autorización del usuario; comprobar los archivos y el CI en GitHub después de publicar.
3. Se conserva `docs/` dentro de la carpeta del proyecto por decisión del alumno. El README principal permite localizarla, aunque el enunciado pide `docs/` en la raíz Git. La aceptación de esa diferencia corresponde al docente.

La prueba opcional PostgreSQL requiere una base desechable y `LEALCAF_TEST_POSTGRES_DSN`, en formato clave=valor. El servidor temporal usado previamente está detenido. Esta limpieza no necesita repetir migraciones ni reiniciar la base real.

## Defensa oral breve

«LealCaf acumula puntos por compras y permite solicitar recompensas en distintas sucursales. El cliente solicita; el administrador confirma o cancela. JWT identifica al usuario: 401 indica sesión inválida y 403 falta de permiso. Un estado final no se modifica, por eso la transición prohibida responde 422. Separar HTTP, reglas y persistencia permite probar el negocio sin servidor. Confirmar bloquea el saldo del cliente y guarda estado y descuento en una transacción para evitar doble gasto».

Las seis preguntas oficiales se responden con funciones reales en [el guion](guion_hito2.md).
