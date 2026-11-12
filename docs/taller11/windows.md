# Incidencias y soluciones en Windows - Taller 11

Este registro resume los problemas encontrados al preparar y ejecutar el taller de LealCafe en Windows y los cambios utilizados para resolverlos.

## 1. Requisitos de Docker Desktop: virtualizacion y WSL

**Que fallo.** El instalador y despues Docker Desktop indicaron que la funcion Virtual Machine Platform no estaba habilitada. Tras superar ese requisito, Docker indico que WSL no estaba instalado.

**Causa.** Faltaban componentes de Windows necesarios para el backend WSL 2. La instalacion por usuario no podia habilitar por si sola la funcion que requiere permisos de administrador.

**Solucion.** Se habilito la funcion de Windows con permisos de administrador y se reinicio el equipo. Se ejecuto `wsl --install` desde PowerShell administrativo; se instalaron WSL y Ubuntu y se completo la configuracion del usuario de Linux. Despues se abrio Docker Desktop con WSL 2.

**Comprobacion.** `docker run --rm hello-world` descargo la imagen y respondio `Hello from Docker!`.

## 2. Falso positivo al comprobar el contexto

**Que fallo.** El ejecutor se detuvo con el mensaje `Se encontro un archivo privado en el contexto de compilacion`.

**Causa.** El verificador no contemplaba correctamente los finales de linea CRLF de Windows. Marcaba `.env.example` como privado, aunque ese ejemplo publico esta permitido.

**Solucion.** Se corrigio la expresion que comprueba los nombres de archivos para aceptar LF y CRLF. La comprobacion sigue rechazando `.env`, sus variantes privadas y `.git`, y permite el ejemplo exacto `.env.example`.

**Comprobacion.** La ejecucion pudo terminar el punto 1 y continuar con las pruebas de la API. La correccion tambien se comprobo con 53 casos de nombres y finales de linea.

## 3. Conexion de las pruebas a PostgreSQL

**Que fallo.** Las pruebas aisladas pasaron, pero los grupos de confirmaciones concurrentes y transacciones PostgreSQL informaron que no pudieron conectarse a la base de pruebas.

**Diagnostico.** Esos mensajes no incluian la causa original de conexion; por si solos no permitian distinguir entre una configuracion de credenciales distinta de la conservada por una base anterior y un problema con el puerto publicado.

**Solucion.** Se actualizo el ejecutor para identificar la base manual y su volumen por la configuracion local. Una configuracion diferente prepara otra base y conserva las anteriores. Tambien se agregaron una consulta SQL para comprobar usuario, contrasena y nombre de base por su interfaz de red, y una comprobacion TCP desde la computadora al puerto publicado. Las pruebas Go se ejecutan con `-count=1` para obtener resultados nuevos.

**Comprobacion.** El ejecutor informo `PostgreSQL comprobado en el puerto 55433`, paso la compilacion y las pruebas, y pudo arrancar la API y ejecutar las seis peticiones.

## 4. Registro de go run bloqueado

**Que fallo.** Despues de obtener los seis codigos correctos, Windows impidio leer `.env.gorun.stdout` porque el archivo estaba siendo utilizado por otro proceso.

**Causa.** El ejecutor intentaba recoger un archivo de redireccion que todavia permanecia abierto. El fallo ocurria durante la recogida del registro, despues de las respuestas HTTP correctas.

**Solucion.** Se sustituyo la redireccion a archivos temporales por captura asincrona en memoria de stdout y stderr. El ejecutor detiene el arbol de procesos que inicio, espera su cierre, recoge ambas salidas y escribe despues `logs_gorun.txt`.

**Comprobacion.** La comprobacion de captura conservo 1000 lineas de cada salida. La ejecucion completa de Windows pudo finalizar y producir el resultado final del taller.

## Resultado final informado por el ejecutor

En la computadora de ejecucion se encontro `docs/taller11/resultado.txt` con la confirmacion de construccion de la imagen, seis peticiones en los tres modos, tres arranques limpios y persistencia despues de `down`. `ultimo_error.txt` no estaba presente.

Las seis peticiones mantienen la secuencia `401, 200, 200, 403, 200, 422`. Las evidencias de la ejecucion estan en esta carpeta y las decisiones D1 a D9 en `docs/decisiones.md`.

## Apoyo de IA

OpenAI Codex apoyo el diagnostico, los cambios del ejecutor y la redaccion de este registro a partir de los mensajes de error y resultados compartidos durante la ejecucion. La declaracion de uso tambien debe figurar en el cuerpo de los commits correspondientes.
