# Entorno del Taller 12

Integrantes de LealCafé, Anchundia Anchundia Luis Antonio y Fuentes Cedeño José Miguel.

Se creó `taller12` desde `d62846e` de `taller11`, después de actualizar las referencias remotas. Al comenzar la rama no existía localmente ni en origin.

Se usa Gin 1.12.0, Go 1.27.1 para Windows AMD64, Git 2.56.1, Docker Engine 29.9.2 y Compose 5.6.1. Git Bash y Windows PowerShell están instalados. Los registros se escriben explícitamente en UTF-8.

Se incorporó `github.com/go-chi/cors v1.2.2`, middleware estándar de net/http del kit que puede adaptarse a Gin. La adaptación conserva el framework de LealCafé.

La base de demostración del taller es `lealcaf_taller12` en el contenedor `lealcaf-t12-demo`, volumen `lealcaf_t12_demo` y puerto local 55434. La línea de base escucha en 18082. Los archivos `.env.taller12*` y el entorno Bruno `Taller12-privado` permanecen privados.
