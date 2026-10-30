# Verificación local · 27 de octubre de 2026

## Resultado

- Go 1.27.1: 32 pruebas principales, con subcasos, pasan con go test ./.... Se mantienen las once pruebas de negocio originales adaptadas.
- Redocly CLI 2.71.2: contrato válido con cero errores y cero avisos; recommended, security-defined y no-unused-components activas. También se comprobó con CLI 1.34.5.
- Bruno CLI 4.3.0 con SQLite: 7/7 peticiones, 6/6 pruebas y 7/7 aserciones correctas.
- Bruno CLI 4.3.0 con PostgreSQL 18.6: 7/7 peticiones, 6/6 pruebas y 7/7 aserciones correctas. Base temporal aislada, sin tocar la base original.
- El iniciador Windows generó .env.demo y el login respondió 200 con un token válido en el puerto 8080.
- Se compiló bin/lealcaf.exe para Windows de 64 bits desde el código final.
- Sucursal está en los modelos y en ambas bases: Centro y Norte. Compra y Canje llevan sucursal_id; los puntos se comparten.
- Las pruebas incluyen migrar datos de esquemas anteriores sin sucursal_id, conservarlos y asignarlos a Centro; datos demo idempotentes; tokens alterados/vencidos; rechazos sin doble descuento.

## Evidencias

hito2_capturas contiene resultados JSON con JWT ocultos, HTML, nueve PNG de respaldo y las salidas de pruebas y contrato. Los PNG son representaciones del contenido real registrado; no son capturas de la interfaz gráfica de Bruno. Se conservaron por separado los resultados de SQLite (puerto 18080) y PostgreSQL (puerto 18082). El puerto predeterminado de la entrega es 8080.

api.html fue generado desde el contrato y usa el archivo local assets/redoc.standalone.js. Se comprobaron contenido y referencias; la vista gráfica del navegador no pudo verificarse en este entorno. Los PNG de respaldo sí se revisaron visualmente.

## Antes de exponer

Abrir el proyecto en el equipo de exposición, usar Iniciar-Proyecto.cmd, abrir Bruno y entorno Local, ejecutar preparación 0 y las seis peticiones, go test ./... y el validador. Ensayar con reloj y tomar también capturas propias de Bruno si se desean evidencias de su interfaz.

Los integrantes deben revisar el código y poder explicar las seis preguntas. El modelo y la guía técnica vigentes están en hito2_modelo.md y hito2_tecnico.md. Solo las operaciones rotuladas (plan) están pendientes de programar.

## Estado de Git y datos privados

No se hicieron commits, push ni cambios en GitHub. Los commits de ambos integrantes deben corresponder a contribuciones reales. El paquete excluye .env, .env.demo, bases de datos y tokens activos; el iniciador genera una clave nueva en cada instalación.
