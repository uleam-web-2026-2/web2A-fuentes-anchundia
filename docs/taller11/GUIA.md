# Taller 11 de LealCafé - ejecución y entrega

## Qué incluye

Este paquete parte de `taller10` (`5f6e6f6`) e incluye el Dockerfile de dos etapas, `.dockerignore`, Compose, las decisiones D1 a D9 y un ejecutor para obtener las evidencias. La API vigente sigue dentro de `Trabajo_autonomo_Lealcaf_hito2/`. Los archivos del taller 11 están en la raíz del repositorio y en `docs/taller11/`.

El único ajuste funcional de Go es escuchar en `:PORT`, en vez de `127.0.0.1:PORT`, para que Docker pueda reenviar las conexiones. También se corrigió el CI que apuntaba a una carpeta `hito2/` inexistente.

Los resultados de Docker aparecen como pendientes hasta ejecutar el paquete. La revisión previa del paquete se describe en `validacion_previa.md`.

## 1. Preparar la computadora

Se necesita **Windows con Docker Desktop funcionando** y **Go 1.27.1 o superior**, como pide el módulo existente. Para subir los cambios por terminal también se necesita Git. Abra Docker Desktop y espere a que indique que el motor está listo.

Abra la raíz del repositorio completo en la rama `taller11`. Antes de ejecutar, compruebe que allí estén `Dockerfile`, `compose.yaml`, `Taller11.cmd`, `scripts/` y `Trabajo_autonomo_Lealcaf_hito2/`. El ZIP parcial del reto sirve para integrar la configuración y consultar las evidencias del compañero. Genere sus propias evidencias ejecutando el taller desde el repositorio.

Cierre el servidor anterior con `Ctrl+C` si usa el puerto 8080. El ejecutor usa ese puerto para las tres comparaciones. Si otro programa debe conservarlo, cambie `PORT` en el `.env` de la raíz después de configurar, como se explica abajo.

## 2. Ejecutar

Haga doble clic en **Taller11.cmd**. También puede abrir PowerShell en la carpeta del proyecto y ejecutar:

```powershell
.\Taller11.cmd
```

La primera vez tarda por las descargas de Go, PostgreSQL, distroless y las dependencias. El ejecutor realiza lo siguiente, en orden:

1. Genera contraseñas locales y los archivos privados de configuración.
2. Arranca una base manual aislada en `55433`, con datos ficticios. Conserva su PostgreSQL anterior y el `.env` previo de la API.
3. Construye las dos etapas y guarda tamaño, capas, contexto y usuario.
4. Comprueba la conexión SQL y el puerto de la base. Ejecuta `go build ./...` y `go test -count=1 -v ./...`, arranca `go run .` y comprueba las seis peticiones. `-count=1` exige una ejecución nueva de las pruebas, sin reutilizar un resultado anterior de la caché.
5. Detiene ese servidor, registra los dos errores solicitados y prueba la API en un contenedor contra la misma base manual.
6. Compara la fecha del mismo canje entre `go run` y el contenedor.
7. Prueba tres arranques desde una base vacía en el proyecto de ensayo `lealcaf-taller11-ensayo` y elimina solo esos datos de ensayo.
8. Arranca el Compose final, comprueba las seis peticiones y verifica que el canje persiste después de `down` y `up` sin `-v`.
9. Completa la tabla y añade las mediciones reales a D1, D4 y D6.

Al terminar debe aparecer **LISTO** y existir `docs/taller11/resultado.txt`. Si aparece **DETENIDO**, esa fase necesita revisión: el mensaje y el archivo `ultimo_error.txt` indican la causa. No se sustituyen errores por resultados esperados.

El paquete crea una base de demostración independiente para que la pareja pueda reproducirlo sin intercambiar las claves de su base anterior. La API y su comportamiento son los del proyecto existente. La base manual y su volumen se identifican por la configuración: si esta cambia, se prepara otra base sin borrar las anteriores. Si una base de otro intento del taller ocupa el puerto necesario, se detiene ese contenedor conservando sus datos.

## 3. Revisar con Bruno

Abra como colección esta carpeta:

```text
Trabajo_autonomo_Lealcaf_hito2/docs/hito2_bruno
```

Seleccione el entorno local **Taller11-privado**, creado por el ejecutor. Tiene las contraseñas de esta demostración y queda excluido de Git. Ejecute primero **0 - Preparar canje** y después **1, 2, 3, 4, 5, 6**, en ese orden, sin paralelismo.

La preparación debe dar `201`; las seis siguientes deben dar `401, 200, 200, 403, 200, 422`. Antes de repetir, vuelva a ejecutar 0 para crear otro canje: la petición 5 confirma el anterior. Puede guardar una captura del Runner mostrando códigos y pruebas, sin abrir el cuerpo del login.

La comprobación automática reproduce las mismas rutas, roles y cambios de estado usando HTTP. La colección de Bruno se conserva para su revisión y exposición.

## 4. Archivos que se generan como evidencia

| Archivo en docs/taller11 | Qué demuestra |
| --- | --- |
| `imagenes.txt`, `capas.txt` | Tamaños reales y capas. |
| `contexto.txt`, `usuario.txt` | Exclusión de secretos y usuario nonroot. |
| `1_sin_variables.txt`, `2_local.txt` | Errores al omitir configuración y usar localhost en el contenedor. |
| `go_build.txt`, `go_test.txt` | Compilación y pruebas del módulo vigente. |
| `base_manual.txt` | Nombre y volumen de la base manual, validación SQL y acceso al puerto desde la computadora. |
| `peticiones.md`, `peticiones_*.json` | Los seis códigos medidos en cada modo. |
| `respuesta3_*.json` | Respuestas completas de la petición 3 para comparar las fechas. |
| `logs_gorun.txt`, `logs_contenedor.txt` | Peticiones atendidas por cada ejecución. |
| `arranques_limpios.txt`, `arranque_limpio_*` | Tres inicios con base vacía y cero reinicios de API. |
| `compose.txt`, `persistencia.txt` | Estado, logs de Compose y conservación de un canje tras reiniciar. |
| `resultado.txt` | Finalización de todas las fases en esta computadora. |

Revise también **docs/decisiones.md**: D1 debe contener tamaños reales, D4 las fechas observadas y D6 los tres arranques. Si quedan como pendientes, falta ejecutar alguna fase. Los integrantes pueden ajustar la redacción para explicar lo que observaron y comprendieron.

## 5. Preparar el repositorio para subir

El ZIP no incluye la carpeta privada de Git. Para trabajar con el historial original, abra PowerShell en la carpeta donde guarda sus proyectos:

```powershell
git clone --branch taller10 https://github.com/uleam-web-2026-2/web2A-fuentes-anchundia.git LealCafe-Taller11
cd LealCafe-Taller11
git switch -c taller11
```

Copie **el contenido** de la carpeta extraída del ZIP dentro de este clon. Debe reemplazar archivos en el mismo nivel; no cree otra carpeta del proyecto dentro. Si ya ejecutó el taller en la carpeta extraída, copie también sus evidencias. Puede ejecutar `Taller11.cmd` directamente en el clon después de copiar.

Los archivos privados `.env` y sus variantes pueden quedarse en su computadora: `.gitignore` los excluye. El entorno `Taller11-privado.bru` también se excluye. Compruébelo antes de confirmar:

```powershell
git status --short
git check-ignore .env .env.docker .env.gorun .env.postgres
git check-ignore Trabajo_autonomo_Lealcaf_hito2/docs/hito2_bruno/environments/Taller11-privado.bru
```

**La revisión de la asignatura pide un commit por punto y otro para el reto**, con herramienta y alcance de IA en el cuerpo. Después de ejecutar y revisar las evidencias, puede separar los cambios así:

```powershell
git add Dockerfile .dockerignore docs/taller11/imagenes.txt docs/taller11/capas.txt docs/taller11/contexto.txt docs/taller11/usuario.txt docs/taller11/build.txt docs/taller11/build_compilacion.txt
git commit -m "Taller 11 punto 1: imagen multietapa de LealCafe" -m "IA: OpenAI Codex, apoyo en Dockerfile, exclusiones y automatizacion de evidencias. Archivos y resultados revisados localmente."

git add .gitignore .env.example .github/workflows/ci.yml README.md LEEME_PRIMERO.md Taller11.cmd scripts/taller11.ps1 docs Trabajo_autonomo_Lealcaf_hito2/main.go
git commit -m "Taller 11 punto 2: API en contenedor y decisiones" -m "IA: OpenAI Codex, apoyo en escucha HTTP, ejecutor, comprobaciones y documentacion. Resultados ejecutados y revisados localmente."

git add compose.yaml
git commit -m "Taller 11 reto: API y PostgreSQL con Compose" -m "IA: OpenAI Codex, apoyo en Compose, espera de la base y persistencia. Revisados los tres arranques y la conservacion de datos."

git status --short
git push -u origin taller11
```

Use estas descripciones después de haber realizado la revisión que declaran. Cada integrante debe registrar sus propios aportes. El paquete no realiza commits ni publica cambios automáticamente.

Si ya tiene una rama `taller11`, cambie a ella con `git switch taller11` y revise su estado antes de copiar. Si Git muestra conflictos, revise esos cambios antes de confirmar. La entrega cuenta cuando los archivos y evidencias están en GitHub, en `taller11`.

## Repetir una fase o cambiar el puerto

Desde la raíz, puede ejecutar solo una fase:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\taller11.ps1 -Paso Configurar
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\taller11.ps1 -Paso Imagen
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\taller11.ps1 -Paso GoRun
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\taller11.ps1 -Paso Contenedor
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\taller11.ps1 -Paso Compose
```

Después de `Configurar`, el `.env` de la raíz ya existe. Allí puede ajustar `PORT` y `MANUAL_DB_PORT`. Las contraseñas generadas se conservan. Si cambia la configuración de la base manual, el ejecutor prepara otro contenedor y volumen; los anteriores quedan conservados. El contenedor y volumen usados se anotan en `base_manual.txt`.

Para parar el Compose final conservando los datos:

```powershell
docker compose down
```

Para volver a usarlo:

```powershell
docker compose up -d --build
```

Para encontrar la base manual activa del taller:

```powershell
docker ps --filter label=lealcaf.taller11=manual
```

Para detenerla, use `docker stop` seguido del nombre indicado en `docs/taller11/base_manual.txt`.

## Errores comunes

| Mensaje o situación | Revisión |
| --- | --- |
| Docker no responde | Abra Docker Desktop; pruebe `docker info`. En Windows puede requerir WSL 2 activo. |
| No se encuentra Go o la versión es insuficiente | Compruebe `go version`; este módulo pide al menos 1.27.1. |
| Puerto ocupado | Cierre el servidor anterior o cambie el puerto de esta configuración. |
| La descarga falla | Revise conexión a Docker Hub, gcr.io y los módulos de Go; vuelva a ejecutar la fase. |
| Login demo falla al reutilizar una base | Use las contraseñas con las que se creó esa base. Cambiar `.env` no cambia contraseñas ya almacenadas. |
| Docker accede al sistema de archivos con permisos denegados | Extraiga el proyecto en una carpeta local de su usuario y revise el acceso de Docker Desktop a esa unidad. |

Si tuvo un problema de Windows, añada `docs/taller11/windows.md` con el mensaje real, la causa y lo que hizo para resolverlo, como pide el taller.
