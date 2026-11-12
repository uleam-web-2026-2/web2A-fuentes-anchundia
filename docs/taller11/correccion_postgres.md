# Preparación automática de PostgreSQL

El informe recibido muestra que la imagen se construyó, pero dos grupos de pruebas no pudieron conectarse a PostgreSQL. El mensaje de esas pruebas no incluye el error original, por lo que no permite confirmar por sí solo si fue un problema de contraseña o del puerto publicado.

El ejecutor actualizado previene la reutilización accidental de una base preparada con otra configuración. Calcula un identificador local de la configuración y lo utiliza en el nombre del contenedor y de su volumen. La misma configuración reutiliza su base; otra configuración prepara una independiente. No se borran contenedores ni volúmenes antiguos durante esta reparación. Solo se detiene una base etiquetada como perteneciente al taller si ocupa el puerto que necesita la ejecución actual.

Antes de ejecutar Go, se comprueban las credenciales con `psql` mediante la IP de red del contenedor. No se usa el acceso de confianza de loopback para dar por válida una contraseña. También se comprueba la conexión TCP desde la computadora al puerto publicado. Los secretos se leen de `.env.prueba` y se excluyen de los informes y del repositorio.

Las pruebas Go siguen siendo obligatorias y se ejecutan con `-count=1`, para obtener un resultado nuevo. No se omiten las pruebas PostgreSQL para hacer avanzar el taller.

Las comprobaciones del ejecutor se describen en `verificacion_recuperacion_postgres.txt`. Los informes previos de Go y las peticiones corresponden a PostgreSQL nativo temporal; no sustituyen las evidencias que debe generar Docker en la computadora del integrante.
