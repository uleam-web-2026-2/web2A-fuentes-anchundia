# Pendientes honestos

## Entrega grupal

- José Miguel debe integrar el avance posterior a 220f246, revisarlo en su equipo, generar evidencias propias y registrar sus aportes reales. Los doce commits preparados en esta ejecución tienen la identidad configurada de Luis; no se atribuyeron cambios a José.
- Publicar el avance restante tras esa integración. Por autorización expresa, GitHub `taller13` contiene únicamente los seis commits iniciales: allí todavía no están la ruta /ws, la demo, el reto ni las evidencias finales del desarrollo completo. Esas partes se conservan localmente y en el ZIP.
- Ambos integrantes deben revisar y acordar la declaración del Hito 3. El archivo separado es un borrador, no una decisión grupal ya aprobada.

## Fuera del alcance del Taller 13

No se implementaron reconexión automática, distribución entre réplicas, persistencia de avisos ni pagos. Tampoco revocación activa de una conexión cuyo rol cambie o cuyo token venza después de abrirla; D3 y D4 explican el comportamiento real. El análisis de carreras con `-race` es opcional y no se presenta como ejecutado; sí hay pruebas concurrentes funcionales del mapa protegido por mutex.

Las fases locales obligatorias, las tres vueltas de Bruno, la captura simultánea y los logs se ejecutaron. Las incidencias recuperadas están en windows.md.
