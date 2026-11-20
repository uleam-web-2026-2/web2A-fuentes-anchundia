# Semana 12 · las seis peticiones del hito 2

Integrantes, Anchundia Anchundia Luis Antonio y Fuentes Cedeño José Miguel.

| Petición | Antes (taller 11) | Con las defensas | Con el usuario de la API (reto) |
| --- | --- | --- | --- |
| 1 sin token | 401 | 401 | 401 |
| 2 login | 200 | 200 | 200 |
| 3 con token | 200 | 200 | 200 |
| 4 rol sin permiso | 403 | 403 | 403 |
| 5 cambio permitido | 200 | 200 | 200 |
| 6 transición prohibida | 422 | 422 | 422 |

Las tres etapas se probaron por HTTP y dentro del Collection Runner de Bruno Desktop. En cada recorrido se ejecutaron las siete peticiones en orden y con la ejecución paralela desactivada. La petición 0 obtuvo las dos sesiones y creó un canje nuevo solicitado antes de la petición 5. El Runner mostró 7 aprobadas, 0 fallidas y 0 omitidas en los tres casos.

| Etapa | Instancia | URL | Captura real |
| --- | --- | --- | --- |
| Antes | Ejecutable original conservado de la línea de base de go run | http://127.0.0.1:18082 | [01-antes.png](bruno/01-antes.png) |
| Con defensas | docker run, lealcaf:taller12-defensas | http://127.0.0.1:18082 | [02-con-defensas.png](bruno/02-con-defensas.png) |
| Usuario restringido | Docker Compose, lealcaf_api | http://127.0.0.1:18082 | [03-usuario-restringido.png](bruno/03-usuario-restringido.png) |

El aviso del Firewall retrasó la primera captura. La primera ejecución HTTP ocurrió con go run antes de incorporar las defensas. Se conservó ese ejecutable y se volvió a iniciar para realizar la captura de la misma versión original en Bruno. El usuario cerró el aviso de Windows. Las capturas se obtuvieron de la aplicación y se convirtieron de JPEG a PNG conservando los píxeles.

Los resúmenes de Bruno pueden verse iguales al coincidir los resultados. Los registros de las instancias corresponden a recorridos independientes. El entorno Taller12-privado contiene las credenciales demo locales y su contenido se excluye del repositorio.

- [Resultados HTTP antes](peticiones_antes.json).
- [Resultados HTTP con defensas](peticiones_defensas.json).
- [Resultados HTTP del reto](peticiones_reto.json).
- [Segundo Compose con volumen nuevo](peticiones_compose.json).
- [Logs completos con defensas](logs_seis.txt).
- [Registro de Compose con usuario restringido](compose.txt).
- [Compra y puntos con el usuario restringido](reto_compra.txt).

Los códigos se conservaron. En la ráfaga deliberada de siete logins incorrectos se observaron cinco 401 y dos 429. Ese caso se registra en rafaga.txt y utiliza un proceso recién iniciado. Antes de los recorridos funcionales se reinició únicamente la API del taller para comenzar con el contador vacío.
