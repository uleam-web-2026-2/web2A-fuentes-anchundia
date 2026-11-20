# Integración del bloque pendiente del Taller 12

Este material corresponde a la continuación del Taller 12 de LealCafé. La rama remota `taller12` contiene el bloque inicial hasta el commit `e3ed429` (auditoría y CORS). El paquete adjunto contiene el bloque pendiente: límite de solicitudes, reto de PostgreSQL, separación de migraciones, Docker Compose, documentación y evidencias.

## Autoría e integración

Los commits incluidos en `commits_pendientes.bundle` fueron desarrollados por **Anchundia Anchundia Luis Antonio**. No se deben copiar y confirmar como si fueran autoría de otra persona. La vía recomendada conserva esa atribución y permite que Git registre a José como quien integró el trabajo:

```powershell
git switch taller12
git pull --ff-only origin taller12
git fetch .\commits_pendientes.bundle backup/taller12-completo:refs/remotes/luis/taller12-completo
git log --reverse --oneline e3ed429..luis/taller12-completo
git cherry-pick e3ed429..luis/taller12-completo
```

El último comando mantiene a Luis como autor de los commits e identifica a José como *committer* de la integración. Si apareciera un conflicto, resuélvelo revisando el archivo equivalente de la carpeta incluida en este paquete. Ejecuta `git add <archivo>` y `git cherry-pick --continue`; no uses `--force` ni reescribas el historial.

José debe revisar el código, ejecutar las comprobaciones y crear sus propios commits únicamente para aportes que realice él, por ejemplo una mejora, corrección, revisión documentada o evidencia ejecutada en su equipo. Todos sus commits deben usar su identidad configurada con `git config user.name` y `git config user.email`, e incluir una declaración de IA ajustada a lo que realmente utilizó:

```powershell
git commit -m "Taller 12: documentar revisión de José Miguel" -m "IA: OpenAI Codex; se utilizó para orientar la revisión y se verificaron localmente las pruebas indicadas."
git push origin taller12
```

No confirmes ni publiques archivos `.env`, contraseñas, tokens, entornos privados de Bruno ni configuraciones personales.

## Configuración privada y pruebas

Desde la raíz, crea dos archivos privados a partir de `.env.example`:

```powershell
Copy-Item .env.example .env.taller12.compose
Copy-Item .env.example .env.taller12
```

Completa en ambos las contraseñas PostgreSQL, `JWT_SECRET`, `DEMO_CLIENT_PASSWORD` y `DEMO_ADMIN_PASSWORD`. Para ejecutar Go directamente, crea además `Trabajo_autonomo_Lealcaf_hito2/.env` desde su `.env.example`, ajusta host, puerto y las credenciales de una base de demostración propia.

Ejecuta las comprobaciones desde una terminal PowerShell:

```powershell
Push-Location Trabajo_autonomo_Lealcaf_hito2
go build ./...
go vet ./...
go test -count=1 ./...
Pop-Location

docker compose --env-file .env.taller12.compose config --quiet
docker compose --env-file .env.taller12.compose up -d --build
python scripts/taller12.py reto --port 18082
```

Usa el puerto que hayas definido en `PORT`. La ráfaga de login debe ejecutarse con una API recién iniciada y después de esperar un minuto desde cualquier login anterior:

```powershell
python scripts/taller12.py rafaga --port 18082
```

Para detener sin perder la evidencia ni los datos, usa `docker compose --env-file .env.taller12.compose down`, sin `-v`. Los detalles, evidencias y resultados esperados se encuentran en `docs/taller12/README.md`, `docs/taller12/peticiones.md` y `docs/taller12/revision.md`.

## Contenido del paquete

- `commits_pendientes.bundle`: historial de Luis posterior a `e3ed429`, para preservar la autoría real.
- `commits_pendientes.md`: listado y alcance de esos commits.
- Las carpetas del proyecto con los archivos posteriores al corte y toda la evidencia de `docs/taller12`.
- Los archivos `.env.example` públicos. No hay archivos `.env`, secretos, tokens ni entornos privados de Bruno.
