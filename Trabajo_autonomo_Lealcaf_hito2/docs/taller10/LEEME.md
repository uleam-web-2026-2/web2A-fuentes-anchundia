# Evidencias del Taller 10

Verificación real sobre la copia recibida en coweb.zip y su solución, el 5 de noviembre de 2026. Las pruebas nuevas se ejecutaron en un cluster PostgreSQL temporal y bases vacías; no sobre los datos del alumno.

| Archivo | Evidencia |
|---|---|
| peticiones.md | Tabla de los seis códigos antes/después. |
| bruno_antes.json / bruno_despues.json | Resumen del reporte real Bruno; sin cabeceras ni cuerpos que puedan contener JWT o contraseñas. |
| bruno_despues.txt | Salida real del CLI: 7/7 peticiones (incluye preparación), 6/6 tests y 7/7 aserciones. |
| go_test.txt | Salida completa de go test -count=1 -v ./... con conexión PostgreSQL de pruebas; 44 tests principales aprobados, además de sus subtests. |
| error_forzado.txt | Extracto de go_test.txt para las transacciones GORM/sqlc: rollback real ante CHECK 23514 y reintento correcto. |
| dto.md / dto_http.json | POST con campos extra antes/después y explicación de lo que ya protegía el Hito 2. |
| redocly.txt | Lint con Redocly 2.54.3, sin errores ni avisos. |
| sqlc.txt | Generación con sqlc 1.31.1 y comprobación de reproducibilidad. |
| compilacion.txt | Resultado de compilación y go vet para la API verificada. |
| esquema_postgres.txt | Columnas reales consultadas a information_schema; sin datos de usuarios. |
| entorno.txt / windows.md | Versiones, entorno aislado y observaciones Windows. |

Para repetir las pruebas PostgreSQL, configure una base desechable con LEALCAF_TEST_POSTGRES_DSN y ejecute el comando documentado en README.md. Sin esa variable las dos pruebas PostgreSQL se omiten de forma explícita. Los datos, puertos y resultados de esta carpeta describen esta verificación; no acreditan todavía una ejecución en la computadora del alumno ni una entrega por Git.

## Repetir únicamente la prueba de rollback

Desde `Trabajo_autonomo_Lealcaf_hito2`, usa una base PostgreSQL desechable con permisos para crear esquemas. Ajusta el DSN a tu entorno. Si requiere contraseña, complétala localmente en la variable del proceso.

```powershell
$env:LEALCAF_TEST_POSTGRES_DSN = 'host=127.0.0.1 port=5432 user=postgres dbname=lealcaf_pruebas sslmode=disable'
go test -count=1 -v ./internal/fidelizacion -run '^TestTaller10TransaccionesPostgres$'
```

Comprueba que aparezca `PASS` para los subtests `gorm` y `sqlc`. Si aparece `SKIP`, revisa que la variable esté definida en la misma terminal. El `SQLSTATE 23514` es el rechazo esperado del `CHECK` en la segunda escritura.

En la repetición local, ambos adaptadores conservaron el canje solicitado, sin descuentos y con saldo 100. Para la compra quedaron cero filas de la factura probada y saldo 60. Al quitar cada restricción, el reintento confirmó el canje con saldo 60 y la compra elevó el saldo a 110.
