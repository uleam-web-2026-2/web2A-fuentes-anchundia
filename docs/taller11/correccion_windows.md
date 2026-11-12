# Corrección de la comprobación del contexto en Windows

Se corrigió un falso positivo del ejecutor al revisar `contexto.txt`. PowerShell en Windows une las líneas con CRLF. La expresión original no tenía en cuenta el retorno de carro antes del final de línea y, por esa razón, marcaba `.env.example` como si fuera privado. Ese archivo de ejemplo sí está permitido por `.dockerignore`.

La expresión corregida reconoce los finales de línea de Windows y Linux. Sigue rechazando `.env`, sus variantes privadas y `.git`, y permite únicamente el ejemplo exacto `.env.example` entre los archivos `.env.*`.

La comprobación reprodujo el falso positivo original y aprobó 53 casos con nombres permitidos y privados, finales LF y CRLF, ausencia de salto final y listados de varios archivos. También se comprobó la sintaxis de PowerShell del archivo actualizado. No se cambió la exclusión de secretos ni se desactivó la validación.

Para actualizar una copia ya ejecutada, reemplace solamente `scripts/taller11.ps1` por el incluido en el ZIP actualizado y vuelva a abrir `Taller11.cmd`. Las imágenes existentes pueden reutilizar su caché de construcción y los archivos privados de configuración se conservan.
