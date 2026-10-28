$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
Set-Location -LiteralPath $projectRoot
$configPath = Join-Path $projectRoot '.env'
if (-not (Test-Path -LiteralPath $configPath)) {
    throw 'Falta .env. Copia .env.example como .env y configura la conexion a tu PostgreSQL existente.'
}
$settings = [System.IO.File]::ReadAllLines($configPath)
$secretLine = $settings | Where-Object { $_ -match '^\s*JWT_SECRET\s*=' } | Select-Object -First 1
if (-not $secretLine -or ($secretLine -replace '^\s*JWT_SECRET\s*=','').Trim().Length -eq 0) {
    $randomBytes = New-Object byte[] 48
    $generator = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    try { $generator.GetBytes($randomBytes) } finally { $generator.Dispose() }
    $newLine = 'JWT_SECRET=' + [Convert]::ToBase64String($randomBytes)
    $settings = @($settings | Where-Object { $_ -notmatch '^\s*JWT_SECRET\s*=' }) + $newLine
    [System.IO.File]::WriteAllLines($configPath, $settings, (New-Object System.Text.UTF8Encoding($false)))
    Write-Host 'Clave de firma generada localmente en .env.'
}
$env:ENV_FILE = '.env'
$env:DB_DRIVER = 'postgres'
Write-Host 'LealCaf - Hito 2 - PostgreSQL'
Write-Host 'La conexion se lee de .env. La base SQLite anterior no se utiliza.'
Write-Host 'Mantén esta ventana abierta. Ctrl+C detiene el servidor.'
$executable = Join-Path $projectRoot 'bin\lealcaf-postgres.exe'
if (Test-Path -LiteralPath $executable) {
    & $executable
} else {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) { throw 'Se necesita Go o bin/lealcaf-postgres.exe.' }
    & go run .
}
if ($LASTEXITCODE -ne 0) { throw 'El servidor termino con un error. Revisa el mensaje anterior.' }
