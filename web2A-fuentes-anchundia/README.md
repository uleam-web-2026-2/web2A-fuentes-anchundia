# Mesa de Ayuda - Servidor Base (Web 2A)

Servidor HTTP desarrollado en Go utilizando el enrutador `chi`, con un contrato de respuesta JSON unificado, almacenamiento en memoria protegido por concurrencia y middleware propios de registro y recuperación de pánicos.

## Tabla de Rutas

| Método | Ruta | Éxito | Errores posibles |
| --- | --- | --- | --- |
| GET | /tickets | 200 OK (Lista de tickets) | — |
| POST | /tickets | 201 Created (Ticket creado) | 400 cuerpo_malformado · 422 datos_invalidos |
| GET | /tickets/{id} | 200 OK (Ticket encontrado) | 400 id_invalido · 404 no_encontrado |
| GET | /explotar | — | 500 error_interno |
| * | * (Ruta inexistente) | — | 404 ruta_inexistente |
| * | * (Método no soportado) | — | 405 metodo_no_permitido |

