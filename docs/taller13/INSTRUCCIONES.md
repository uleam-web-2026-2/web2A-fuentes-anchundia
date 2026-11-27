# Integración para José Miguel

La rama GitHub `taller13` publica **solo seis commits de Luis hasta 220f246**. El ZIP contiene la implementación completa, seis commits posteriores preparados por Luis con OpenAI Codex y evidencias reales de esta computadora. No presentes esos archivos ni esas ejecuciones como aportes propios. José debe revisar, ejecutar pruebas, generar sus evidencias y confirmar sus mejoras reales con su identidad.

## Actualizar e integrar conservando autoría

Trabaja en tu clon habitual, no inicialices otro Git dentro del directorio de cambios. Guarda tus modificaciones antes de integrar.

```powershell
git fetch origin
git switch taller13
git pull --ff-only origin taller13
git status
git fetch RUTA_DEL_PAQUETE/commits_pendientes.bundle backup/taller13-completo:refs/remotes/luis/taller13-completo
git log --reverse --oneline 220f246..luis/taller13-completo
git cherry-pick 220f246..luis/taller13-completo
```

El bundle requiere que el clon tenga el corte 220f246, que ya está en GitHub. Cherry-pick conserva a Luis como autor y registra a José como integrador (committer); no uses `--author`, amend ni force. Si aparece un conflicto, compara con `proyecto/` y `cambios_posteriores_al_sexto/`, resuelve, agrega los archivos específicos y usa `git cherry-pick --continue`. No borres cambios de tu compañero para resolverlo. Los archivos posteriores al corte mantienen sus rutas en ambas carpetas. `resumen_commits.md` identifica cada commit y cuáles ya están publicados.

## Entorno y ejecución

Necesitas Git, Go compatible con go.mod, Docker Desktop con motor Linux activo y Bruno. El proyecto usa Go/Gin, PostgreSQL 16, GORM y sqlc ya generado. No hace falta instalar sqlc para ejecutar ni para este taller. La biblioteca coder/websocket se descarga con las dependencias de Go. El ZIP no contiene secretos ni binarios.

Desde la raíz del clon copia `.env.example` a `.env.taller13`. Completa DB_MIGRATION_PASSWORD, DB_API_PASSWORD, JWT_SECRET aleatorio de al menos 32 bytes, DEMO_CLIENT_PASSWORD y DEMO_ADMIN_PASSWORD con valores privados de tu equipo. Usa DB_NAME=lealcaf_taller13, DB_MIGRATION_USER=lealcaf_migrador y PORT=18085. No copies contraseñas de otro integrante. Comprueba puertos libres antes de iniciar:

```powershell
docker compose --env-file .env.taller13 -f compose.yaml -f compose.taller13.yaml config --quiet
docker compose --env-file .env.taller13 -f compose.yaml -f compose.taller13.yaml up -d --build
docker compose --env-file .env.taller13 -f compose.yaml -f compose.taller13.yaml ps -a
go -C Trabajo_autonomo_Lealcaf_hito2 build ./...
go -C Trabajo_autonomo_Lealcaf_hito2 test -count=1 ./...
go -C Trabajo_autonomo_Lealcaf_hito2 vet ./...
```

El migrador prepara tablas, usuarios ficticios y permisos; la API usa lealcaf_api sin permisos de creación de esquema. No uses `down -v`. Para Go local consulta README.md de esta carpeta: requiere el entorno del módulo, DB_PASSWORD del rol restringido, migración separada y `BIND_HOST=127.0.0.1`. Mantén diferentes puertos si ejecutas Go y Compose a la vez.

## Bruno y navegador

Abre la colección existente `Trabajo_autonomo_Lealcaf_hito2/docs/hito2_bruno`. Crea un entorno privado `Taller13-Jose-privado.bru` con `baseUrl=http://127.0.0.1:18085`, `clientePassword` y `adminPassword` de tu entorno. El patrón de .gitignore lo excluye. Ejecuta la petición 0 para preparar tokens y un canje nuevo; ejecuta 1–6 en orden. Debes obtener 401, 200, 200, 403, 200 y 422. Espera un minuto antes de otro recorrido para no activar el límite de cinco logins. El canje consumido por la petición 5 no puede volver a confirmarse.

Arranca la demo desde `Trabajo_autonomo_Lealcaf_hito2/demo/taller13` con `go run servir.go`. Abre `http://localhost:5173` en dos ventanas, usando cliente y cliente2, ambos de rol cliente y con identidades distintas. Configura la URL de la API en el formulario. Ejecuta 0 y 1–6 desde Bruno con un canje nuevo. Comprueba que solo cliente recibe `estado_cambiado`. Captura las dos ventanas simultáneamente con usuarios e ID legibles. Cierra ambas conexiones y guarda los logs de esa misma ejecución; el ID del PATCH debe coincidir con la captura. Ejecuta la sustitución sed de tokens del docente antes de confirmar los logs, aunque el middleware ya los oculta. No edites, compongas ni simules las capturas.

## Verificar el reto y aportar

```powershell
python scripts/taller13_pruebas.py servicio
python scripts/taller13_pruebas.py websocket
python scripts/taller13_pruebas.py reto
```

`TestRetoAvisaADosAdministradoresYNingunCliente` conecta dos administradores ficticios y dos clientes a un servidor real y verifica `canje_nuevo` después de crear; otra prueba conserva el cambio de estado exclusivo al propietario. Revisa la separación de `AvisadorNuevos` y `Avisos.NuevoCanje`. Puedes agregar una mejora propia y su prueba o una revisión documentada con resultados de tu equipo. Usa tus datos reales de Git y declara la IA realmente utilizada:

```powershell
git config user.name
git config user.email
git status
git add RUTAS_ESPECIFICAS_REVISADAS
git commit -m "Taller 13: verificar integracion y evidencias de José Miguel" -m "IA: OpenAI Codex; describir aquí únicamente la colaboración realmente utilizada."
git push origin taller13
```

Antes de publicar revisa que no aparezcan .env privados, tokens ni credenciales de Bruno. Consulta el resultado de Actions y verifica que GitHub tenga los archivos completos. La entrega remota actual es parcial por solicitud expresa de Luis; para la entrega grupal falta integrar y publicar el resto. Revisa además `docs/hito3/declaracion_borrador.md` en pareja: es una propuesta separada del taller.
