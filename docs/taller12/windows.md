# Incidencias de Windows en el Taller 12

Al comenzar, Docker Desktop estaba instalado y su motor Linux apagado. docker desktop start inició el motor 29.9.2 y permitió preparar el PostgreSQL de demostración.

Windows mostró el aviso de Firewall de fidelizacion al iniciar go run. Ese aviso bloqueó el control de Bruno. El usuario lo cerró y se retomó el Runner con el ejecutable original. El primer recorrido se mantuvo separado de las versiones posteriores.

El Python del sistema no incluía pypdf. Se utilizó el Python del runtime de Codex para leer el material y convertir el formato de las capturas. Los ejecutores HTTP usan solo la biblioteca estándar y funcionan con el Python del sistema.

Las evidencias se escribieron mediante Python con codificación UTF-8. La API se consultó con 127.0.0.1 para coincidir con la publicación IPv4 de Docker. Los puertos 18082 y 55434 se reservaron para esta actividad.

El primer arranque de migraciones falló con SQLSTATE 42P01 al conceder permisos a cuentas. La tabla creada por la estrategia de nombres de GORM se llama cuenta. Se comprobó con psql, se corrigió el SQL y se repitió la migración. La preparación fallida quedó revertida por la transacción. El siguiente arranque terminó correctamente.

Al incorporar el límite, una prueba antigua hacía seis logins consecutivos y recibió 429 en el sexto. Sus casos de JSON inválido se separaron en subpruebas con contadores propios. La suite completa pasó después del ajuste.
