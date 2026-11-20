# Taller 12 de LealCafé

Integrantes, Anchundia Anchundia Luis Antonio y Fuentes Cedeño José Miguel.

El módulo vigente es Trabajo_autonomo_Lealcaf_hito2. La rama de trabajo es taller12.

## Preparar la configuración local

Copia el `.env.example` de la raíz como `.env.taller12.compose` y completa DB_MIGRATION_PASSWORD, DB_API_PASSWORD, JWT_SECRET, DEMO_CLIENT_PASSWORD y DEMO_ADMIN_PASSWORD con valores privados. Las dos contraseñas PostgreSQL deben ser diferentes. Puedes elegir PORT=18082 para convivir con el Taller 11. El proyecto Compose se llama lealcaf-taller12 y crea su propio volumen.

Para los ejecutores Python, copia ese mismo ejemplo como `.env.taller12`. Para `go run`, copia `Trabajo_autonomo_Lealcaf_hito2/.env.example` como `Trabajo_autonomo_Lealcaf_hito2/.env`; conserva este último archivo solo en tu equipo. Ninguna de esas configuraciones privadas se versiona ni se incluye en el paquete del compañero.

En esta ejecución se utilizó .env.taller12.compose como archivo privado. El entorno privado de Bruno apunta a http://127.0.0.1:18082 y usa las contraseñas demo correspondientes.

## Arrancar desde la raíz

    docker compose --env-file .env.taller12.compose up -d --build

El servicio db inicializa el usuario lealcaf_migrador y el rol lealcaf_api desde db/crear_rol_api.sql. Cuando el healthcheck TCP confirma PostgreSQL, migrar ejecuta el mismo binario con el argumento migrar. Ese comando migra, carga opcionalmente los datos ficticios y asigna los permisos. La API arranca después con lealcaf_api. Su contenedor recibe únicamente sus credenciales de ejecución.

Para ejecutar las migraciones desde Windows, configura DB_HOST, DB_PORT, DB_NAME, DB_MIGRATION_USER y DB_MIGRATION_PASSWORD, y desde el módulo ejecuta:

    go run . migrar
    go run .

El arranque normal atiende con DB_USER y DB_PASSWORD y requiere que el esquema exista. Los cambios de modelos se aplican mediante el comando migrar antes de arrancar la nueva API.

## Evidencias y repetición

La tabla está en [peticiones.md](peticiones.md). Las capturas del Runner están en bruno/. Antes de cada recorrido se ejecuta la petición 0 para preparar tokens y un canje nuevo.

El ejecutor scripts/taller12.py usa Python estándar y lee .env.taller12 desde la raíz. El puerto y las contraseñas demo deben coincidir con la instancia en prueba.

    python scripts/taller12.py antes
    python scripts/taller12.py inyeccion
    python scripts/taller12.py cors
    python scripts/taller12.py rafaga
    python scripts/taller12.py defensas
    python scripts/taller12.py reto

La ráfaga necesita un proceso recién iniciado. Antes de repetir la colección, espera al menos un minuto o reinicia únicamente la API del taller. Cada recorrido completo usa tres logins. Las otras rutas mantienen sus reglas de autenticación y autorización.

## Permisos del usuario de ejecución

| Tabla | Permisos |
| --- | --- |
| sucursals, recompensas, cuenta | SELECT |
| clientes | SELECT y UPDATE de id para el bloqueo FOR UPDATE al confirmar canjes |
| compras, movimiento_puntos | SELECT e INSERT |
| canjes | SELECT, INSERT y UPDATE de estado |
| Secuencias de compras, canjes y movimiento_puntos | USAGE |

El bloqueo de cliente serializa confirmaciones concurrentes. PostgreSQL exige permiso UPDATE para obtener FOR UPDATE, aunque esa operación solo bloquee la fila. La columna id limita el permiso concedido. Las rutas de la API mantienen su autorización.

Los permisos de lectura/escritura concedidos permitirían que una inyección dañara información dentro de ese alcance. La parametrización, la validación, los roles HTTP y el usuario restringido se complementan.

## Volúmenes

Para detener y conservar datos:

    docker compose --env-file .env.taller12.compose down

Para volver a arrancar:

    docker compose --env-file .env.taller12.compose up -d --build

La prueba desde cero se realiza con otro nombre de proyecto y otro volumen nuevo. En esta actividad se conservaron los volúmenes existentes.

## Referencias

- [CORS del kit y adaptación net/http](https://github.com/go-chi/cors).
- [PostgreSQL 16, CREATE ROLE](https://www.postgresql.org/docs/16/sql-createrole.html).
- [PostgreSQL 16, GRANT](https://www.postgresql.org/docs/16/sql-grant.html).
