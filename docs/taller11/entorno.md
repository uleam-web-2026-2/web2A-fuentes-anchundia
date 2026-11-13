# Entorno del Taller 11

Equipo de Luis Antonio Anchundia Anchundia.

| Componente | Versión observada |
| --- | --- |
| Windows 11 Pro | 10.0.26200.9711 |
| Windows PowerShell | 5.1.26100.9711 |
| Docker Desktop | 4.98.0 |
| Docker CLI y motor Linux | 29.9.2 |
| Docker Compose | v5.6.1 |
| Go para Windows AMD64 | 1.27.1 |
| Git para Windows | 2.55.0.windows.5 |

La configuración privada se generó con `-Paso Configurar` y se validó con `docker compose config --quiet`. Desde la raíz de `taller11` se lanzó el mismo paso que llama `Taller11.cmd`, mediante Windows PowerShell para evitar la pausa interactiva del archivo CMD.

```powershell
$env:PATH = 'C:/Program Files/Go/bin;' + $env:PATH
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\taller11.ps1 -Paso Todo
```

La API usa el puerto local 8080 y la base manual de la comparación usa el 55433. La colección de Bruno del Hito 2 usa el entorno local `Taller11-privado` con `http://localhost:8080`.
