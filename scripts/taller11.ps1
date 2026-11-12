[CmdletBinding()]
param(
    [ValidateSet('Todo', 'Configurar', 'Imagen', 'GoRun', 'Contenedor', 'Compose', 'Estado')]
    [string]$Paso = 'Todo'
)

# Compatible con Windows PowerShell 5.1. No necesita modulos adicionales.
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$script:Root = Split-Path -Parent $PSScriptRoot
$script:Module = Join-Path $script:Root 'Trabajo_autonomo_Lealcaf_hito2'
$script:Evidence = Join-Path $script:Root 'docs/taller11'
$script:Config = @{}
$script:Utf8 = New-Object System.Text.UTF8Encoding($false)
$script:ApiName = 'lealcaf-t11-api'
$script:DbName = 'lealcaf-t11-manual'
$script:DbVolume = 'lealcaf_t11_manual'
$script:TrialProject = 'lealcaf-taller11-ensayo'
$script:Http = $null
$script:PreviousVariables = @{}

function Write-Text([string]$Path, [string]$Value) {
    $directory = Split-Path -Parent $Path
    if ($directory) { [void][IO.Directory]::CreateDirectory($directory) }
    [IO.File]::WriteAllText($Path, $Value + [Environment]::NewLine, $script:Utf8)
}

function Protect-Text([string]$Value) {
    foreach ($name in @('DB_PASSWORD', 'JWT_SECRET', 'DEMO_CLIENT_PASSWORD', 'DEMO_ADMIN_PASSWORD')) {
        if ($script:Config.ContainsKey($name) -and $script:Config[$name]) {
            $Value = $Value.Replace([string]$script:Config[$name], '[SECRETO OMITIDO]')
        }
    }
    return $Value
}

function Invoke-Native {
    param([string]$Program, [string[]]$Arguments, [string]$OutputPath = '', [switch]$AllowFailure)
    $previous = $ErrorActionPreference
    try {
        # En PS 5.1 stderr de un programa nativo no debe cortar antes de leer su codigo.
        $ErrorActionPreference = 'Continue'
        $lines = @(& $Program @Arguments 2>&1)
        $code = $LASTEXITCODE
    } finally { $ErrorActionPreference = $previous }
    $clean = Protect-Text (($lines | ForEach-Object { $_.ToString() }) -join [Environment]::NewLine)
    if ($OutputPath) { Write-Text $OutputPath $clean }
    if ($code -ne 0 -and -not $AllowFailure) {
        throw "$Program termino con codigo $code.`n$clean"
    }
    return [pscustomobject]@{ Code = $code; Text = $clean }
}

function Invoke-Docker([string[]]$Arguments, [string]$OutputPath = '', [switch]$AllowFailure) {
    return Invoke-Native -Program docker -Arguments $Arguments -OutputPath $OutputPath -AllowFailure:$AllowFailure
}

function Invoke-Compose([string[]]$Arguments, [string]$OutputPath = '') {
    return Invoke-Docker -Arguments (@('compose') + $Arguments) -OutputPath $OutputPath
}

function New-Secret {
    $bytes = New-Object byte[] 32
    $generator = [Security.Cryptography.RandomNumberGenerator]::Create()
    try { $generator.GetBytes($bytes) } finally { $generator.Dispose() }
    return ([BitConverter]::ToString($bytes)).Replace('-', '').ToLowerInvariant()
}

function Get-ManualDbKey {
    # Otra configuracion recibe otra base. Las bases anteriores se conservan.
    $parts = @()
    foreach ($name in @('DB_USER', 'DB_NAME', 'DB_PASSWORD', 'DEMO_CLIENT_PASSWORD', 'DEMO_ADMIN_PASSWORD', 'MANUAL_DB_PORT')) {
        $parts += "$name=$($script:Config[$name])"
    }
    $hasher = [Security.Cryptography.SHA256]::Create()
    try { $bytes = $hasher.ComputeHash([Text.Encoding]::UTF8.GetBytes(($parts -join "`n"))) }
    finally { $hasher.Dispose() }
    return (([BitConverter]::ToString($bytes)).Replace('-', '').ToLowerInvariant()).Substring(0, 12)
}

function Initialize-Config {
    $path = Join-Path $script:Root '.env'
    if (-not (Test-Path $path)) {
        $values = @('DB_USER=postgres', 'DB_NAME=lealcaf_taller11', 'PORT=8080', 'MANUAL_DB_PORT=55433')
        foreach ($name in @('DB_PASSWORD', 'JWT_SECRET', 'DEMO_CLIENT_PASSWORD', 'DEMO_ADMIN_PASSWORD')) {
            $values += "$name=$(New-Secret)"
        }
        Write-Text $path ($values -join "`n")
    }
    foreach ($line in Get-Content $path) {
        if ($line -match '^([A-Z_]+)=(.*)$') { $script:Config[$matches[1]] = $matches[2] }
    }
    foreach ($name in @('DB_USER', 'DB_NAME', 'DB_PASSWORD', 'JWT_SECRET', 'DEMO_CLIENT_PASSWORD', 'DEMO_ADMIN_PASSWORD', 'PORT', 'MANUAL_DB_PORT')) {
        if (-not $script:Config.ContainsKey($name) -or -not $script:Config[$name]) {
            throw "Falta $name en el .env de la raiz. Complete esa variable localmente."
        }
    }
    foreach ($name in @('PORT', 'MANUAL_DB_PORT')) {
        $number = 0
        if (-not [int]::TryParse($script:Config[$name], [ref]$number) -or $number -lt 1 -or $number -gt 65535) {
            throw "$name debe ser un puerto entre 1 y 65535."
        }
    }
    if ($script:Config.PORT -eq $script:Config.MANUAL_DB_PORT) { throw 'Los dos puertos deben ser distintos.' }
    if ($script:Config.JWT_SECRET.Length -lt 32) { throw 'JWT_SECRET debe tener al menos 32 caracteres.' }
    # Para estos archivos generados se usan valores sin comillas ni saltos de linea.
    foreach ($name in @('DB_USER', 'DB_NAME', 'DB_PASSWORD', 'JWT_SECRET', 'DEMO_CLIENT_PASSWORD', 'DEMO_ADMIN_PASSWORD')) {
        if ($script:Config[$name] -notmatch '^[A-Za-z0-9_-]+$') {
            throw "$name debe usar letras, numeros, guion o guion bajo en esta configuracion del taller."
        }
    }
    $manualKey = Get-ManualDbKey
    $script:DbName = "lealcaf-t11-manual-$manualKey"
    $script:DbVolume = "lealcaf_t11_manual_$manualKey"
    $common = @(
        'DB_DRIVER=postgres', "DB_PORT=$($script:Config.MANUAL_DB_PORT)",
        "DB_USER=$($script:Config.DB_USER)", "DB_NAME=$($script:Config.DB_NAME)",
        "DB_PASSWORD=$($script:Config.DB_PASSWORD)", "PORT=$($script:Config.PORT)",
        "JWT_SECRET=$($script:Config.JWT_SECRET)", 'JWT_TTL=1h', 'DEMO_DATA=true',
        "DEMO_CLIENT_PASSWORD=$($script:Config.DEMO_CLIENT_PASSWORD)",
        "DEMO_ADMIN_PASSWORD=$($script:Config.DEMO_ADMIN_PASSWORD)",
        'GIN_MODE=release', 'REPOSITORY_ADAPTER=sqlc'
    )
    Write-Text (Join-Path $script:Root '.env.gorun') ((@('DB_HOST=127.0.0.1') + $common) -join "`n")
    Write-Text (Join-Path $script:Root '.env.localhost') ((@('DB_HOST=localhost') + $common) -join "`n")
    Write-Text (Join-Path $script:Root '.env.docker') ((@('DB_HOST=host.docker.internal') + $common) -join "`n")
    Write-Text (Join-Path $script:Root '.env.postgres') (@(
        "POSTGRES_USER=$($script:Config.DB_USER)", "POSTGRES_DB=$($script:Config.DB_NAME)",
        "POSTGRES_PASSWORD=$($script:Config.DB_PASSWORD)"
    ) -join "`n")
    Write-Text (Join-Path $script:Root '.env.prueba') (@(
        "PGUSER=$($script:Config.DB_USER)", "PGDATABASE=$($script:Config.DB_NAME)",
        "PGPASSWORD=$($script:Config.DB_PASSWORD)", 'PGPORT=5432', 'PGSSLMODE=disable', 'PGCONNECT_TIMEOUT=5'
    ) -join "`n")
    $privateBruno = Join-Path $script:Module 'docs/hito2_bruno/environments/Taller11-privado.bru'
    Write-Text $privateBruno (@(
        'vars {', "  baseUrl: http://localhost:$($script:Config.PORT)",
        "  clientePassword: $($script:Config.DEMO_CLIENT_PASSWORD)",
        "  adminPassword: $($script:Config.DEMO_ADMIN_PASSWORD)", '}'
    ) -join "`n")
    $script:BaseUrl = "http://127.0.0.1:$($script:Config.PORT)"
    Write-Host 'Configuracion privada lista. El .env anterior de su API se conserva.'
}

function Assert-PortFree([int]$Port) {
    $listener = [Net.Sockets.TcpListener]::new([Net.IPAddress]::Loopback, $Port)
    try { $listener.Start() } catch {
        throw "El puerto $Port esta ocupado. Detenga su go run anterior o cambie PORT/MANUAL_DB_PORT en el .env de la raiz."
    } finally { $listener.Stop() }
}

function Set-DecisionBlock([string]$Name, [string]$Value) {
    $path = Join-Path $script:Root 'docs/decisiones.md'
    $text = [IO.File]::ReadAllText($path)
    $begin = "<!-- ${Name}:INICIO -->"
    $end = "<!-- ${Name}:FIN -->"
    $pattern = [regex]::Escape($begin) + '[\s\S]*?' + [regex]::Escape($end)
    $match = [regex]::Match($text, $pattern)
    if (-not $match.Success) { throw "No se encontro el bloque de mediciones $Name en docs/decisiones.md." }
    Write-Text $path ($text.Replace($match.Value, "$begin`n$Value`n$end").TrimEnd())
}

function Update-Table {
    $names = @('1 sin token', '2 login', '3 con token', '4 rol sin permiso', '5 cambio permitido', '6 transicion prohibida')
    $lines = @('# Semana 11 - seis peticiones', '', '| Peticion | Con go run | En el contenedor | Con Compose |', '| --- | --- | --- | --- |')
    $reports = @{}
    foreach ($stage in @('gorun', 'contenedor', 'compose')) {
        $path = Join-Path $script:Evidence "peticiones_$stage.json"
        if (Test-Path $path) { $reports[$stage] = Get-Content $path -Raw | ConvertFrom-Json }
    }
    for ($i = 0; $i -lt 6; $i++) {
        $values = @()
        foreach ($stage in @('gorun', 'contenedor', 'compose')) {
            if ($reports.ContainsKey($stage)) { $values += [string]$reports[$stage].peticiones[$i].codigo }
            else { $values += 'Pendiente' }
        }
        $lines += "| $($names[$i]) | $($values[0]) | $($values[1]) | $($values[2]) |"
    }
    $lines += @('', 'Secuencia esperada: 401, 200, 200, 403, 200, 422.', '', 'La preparacion crea un canje nuevo antes de cada recorrido. No se guardan tokens ni respuestas de login.')
    foreach ($stage in @('gorun', 'contenedor', 'compose')) {
        $path = Join-Path $script:Evidence "respuesta3_$stage.json"
        if (Test-Path $path) {
            $lines += @('', "## Peticion 3 - $stage", '', '```json', ([IO.File]::ReadAllText($path)).Trim(), '```')
        }
    }
    Write-Text (Join-Path $script:Evidence 'peticiones.md') ($lines -join "`n")
}

function Invoke-Api([string]$Method, [string]$Route, [string]$Token = '', $Body = $null) {
    $message = [Net.Http.HttpRequestMessage]::new([Net.Http.HttpMethod]::new($Method), "$script:BaseUrl$Route")
    if ($Token) { $message.Headers.Authorization = [Net.Http.Headers.AuthenticationHeaderValue]::new('Bearer', $Token) }
    if ($null -ne $Body) {
        $message.Content = [Net.Http.StringContent]::new(($Body | ConvertTo-Json -Depth 8 -Compress), [Text.Encoding]::UTF8, 'application/json')
    }
    try {
        $response = $script:Http.SendAsync($message).GetAwaiter().GetResult()
        try {
            $raw = $response.Content.ReadAsStringAsync().GetAwaiter().GetResult()
            $data = $null
            if ($raw.TrimStart() -match '^[\{\[]') { $data = $raw | ConvertFrom-Json }
            return [pscustomobject]@{ Code = [int]$response.StatusCode; Raw = $raw; Data = $data }
        } finally { $response.Dispose() }
    } finally { $message.Dispose() }
}

function Wait-Api($Process = $null) {
    for ($i = 0; $i -lt 180; $i++) {
        if ($null -ne $Process) {
            $Process.Refresh()
            if ($Process.HasExited) { throw 'go run se detuvo antes de atender. Revise docs/taller11/logs_gorun.txt.' }
        }
        try {
            # GET /login devuelve 404: comprueba que el servidor HTTP ya escucha.
            $answer = Invoke-Api GET '/login'
            if ($answer.Code -eq 404) { return }
        } catch { }
        Start-Sleep -Seconds 1
    }
    throw 'La API no estuvo lista en 180 segundos. Revise los logs de esta fase.'
}

function Assert-Code($Response, [int]$Expected, [string]$Name) {
    if ($Response.Code -ne $Expected) { throw "$Name devolvio $($Response.Code), se esperaba $Expected. $($Response.Raw)" }
}

function Test-Api([string]$Stage, [switch]$Save) {
    # Equivale a preparar la peticion 0 de Bruno. Credenciales y JWT solo en memoria.
    $admin = Invoke-Api POST '/login' '' @{ usuario = 'admin'; password = $script:Config.DEMO_ADMIN_PASSWORD }
    Assert-Code $admin 200 'Login de preparacion administrador'
    $client = Invoke-Api POST '/login' '' @{ usuario = 'cliente'; password = $script:Config.DEMO_CLIENT_PASSWORD }
    Assert-Code $client 200 'Login de preparacion cliente'
    $catalog = Invoke-Api GET '/recompensas' $client.Data.token
    Assert-Code $catalog 200 'Catalogo'
    $rewards = @($catalog.Data.recompensas | Where-Object { $_.activa -and $_.puntos_necesarios -gt 0 } | Sort-Object puntos_necesarios, id)
    if (-not $rewards.Count) { throw 'No hay recompensas activas para preparar el canje.' }
    $list = Invoke-Api GET '/canjes' $admin.Data.token
    Assert-Code $list 200 'Listado de canjes'
    $confirmed = @($list.Data.canjes | Where-Object { $_.cliente_id -eq $client.Data.cliente_id -and $_.estado -eq 'confirmado' } | Sort-Object id)
    if (-not $confirmed.Count) { throw 'No hay canje confirmado demo. Revise DEMO_DATA y las contrasenas de demostracion.' }
    $oldId = $confirmed[0].id
    $created = Invoke-Api POST '/canjes' $client.Data.token @{ recompensa_id = $rewards[0].id; sucursal_id = 2 }
    Assert-Code $created 201 'Preparacion de un canje nuevo'
    $newId = $created.Data.id
    $responses = @()
    $responses += Invoke-Api GET "/canjes/$oldId"
    $responses += Invoke-Api POST '/login' '' @{ usuario = 'cliente'; password = $script:Config.DEMO_CLIENT_PASSWORD }
    Assert-Code $responses[1] 200 'Peticion 2'
    $responses += Invoke-Api GET "/canjes/$oldId" $responses[1].Data.token
    $responses += Invoke-Api PATCH "/canjes/$newId/estado" $responses[1].Data.token @{ estado = 'confirmado' }
    $responses += Invoke-Api PATCH "/canjes/$newId/estado" $admin.Data.token @{ estado = 'confirmado' }
    $responses += Invoke-Api PATCH "/canjes/$oldId/estado" $admin.Data.token @{ estado = 'cancelado' }
    $expected = @(401, 200, 200, 403, 200, 422)
    $records = @()
    for ($i = 0; $i -lt 6; $i++) {
        $records += [pscustomobject]@{ peticion = $i + 1; codigo = $responses[$i].Code; esperado = $expected[$i]; correcto = ($responses[$i].Code -eq $expected[$i]) }
    }
    $passed = @($records | Where-Object { -not $_.correcto }).Count -eq 0
    $report = [pscustomobject]@{
        etapa = $Stage; fecha_utc = [DateTime]::UtcNow.ToString('o'); peticiones = $records
        correcto = $passed; canje_confirmado = $oldId; canje_nuevo = $newId
    }
    if ($Save) {
        Write-Text (Join-Path $script:Evidence "peticiones_$Stage.json") ($report | ConvertTo-Json -Depth 8)
        Write-Text (Join-Path $script:Evidence "respuesta3_$Stage.json") $responses[2].Raw
        Update-Table
    }
    Write-Host "$Stage : $(($records | ForEach-Object { $_.codigo }) -join ', ')"
    if (-not $passed) { throw 'Los codigos no coinciden. Se conservaron las respuestas reales; no se da esta fase por aprobada.' }
    if ($responses[2].Data.cliente_id -ne $client.Data.cliente_id -or $responses[4].Data.estado -ne 'confirmado') {
        throw 'Los codigos coincidieron, pero la identidad o el estado del canje no son correctos.'
    }
    return $report
}

function Get-OwnedContainer([string]$Name, [string]$Kind) {
    $found = Invoke-Docker @('ps', '-aq', '--filter', "name=^/$Name$")
    if (-not $found.Text.Trim()) { return '' }
    $label = Invoke-Docker @('inspect', '-f', '{{json .Config.Labels}}', $Name)
    $labels = $label.Text | ConvertFrom-Json
    $ownedKind = ''
    if ($null -ne $labels) {
        $property = $labels.PSObject.Properties['lealcaf.taller11']
        if ($null -ne $property) { $ownedKind = [string]$property.Value }
    }
    if ($ownedKind -ne $Kind) { throw "Ya existe $Name y no pertenece a este taller. No se modifico." }
    return $found.Text.Trim()
}

function Remove-ManualApi {
    if (Get-OwnedContainer $script:ApiName 'api') { [void](Invoke-Docker @('rm', '-f', $script:ApiName)) }
}

function Stop-ConflictingManualDb([string]$CurrentId) {
    $running = Invoke-Docker @('ps', '-q', '--filter', 'label=lealcaf.taller11=manual')
    foreach ($id in ($running.Text -split '\r?\n')) {
        $id = $id.Trim()
        if (-not $id -or $id -eq $CurrentId) { continue }
        $bindings = Invoke-Docker @('inspect', '-f', '{{json .HostConfig.PortBindings}}', $id)
        $mapping = $bindings.Text | ConvertFrom-Json
        if ($null -eq $mapping) { continue }
        $property = $mapping.PSObject.Properties['5432/tcp']
        if ($null -eq $property) { continue }
        foreach ($binding in @($property.Value)) {
            if ([string]$binding.HostPort -eq [string]$script:Config.MANUAL_DB_PORT) {
                Write-Host 'Liberando el puerto de una base anterior del taller; sus datos se conservan.'
                [void](Invoke-Docker @('stop', $id))
                break
            }
        }
    }
}

function Test-ManualDbAccess {
    # La IP de red obliga a comprobar la contrasena, evitando el trust de loopback.
    $network = Invoke-Docker @('inspect', '-f', '{{range .NetworkSettings.Networks}}{{println .IPAddress}}{{end}}', $script:DbName)
    $addresses = @($network.Text -split '\r?\n' | ForEach-Object { $_.Trim() } | Where-Object { $_ -match '^\d+\.\d+\.\d+\.\d+$' })
    if (-not $addresses.Count) { throw 'No se pudo obtener la direccion de red de la base aislada.' }
    $probe = Invoke-Docker -Arguments @('exec', '--env-file', '.env.prueba', $script:DbName,
        'psql', '-h', $addresses[0], '-w', '-tA', '-v', 'ON_ERROR_STOP=1', '-c', 'SELECT 1') -AllowFailure
    if ($probe.Code -ne 0 -or $probe.Text.Trim() -ne '1') {
        throw "La comprobacion de usuario, base y contrasena de PostgreSQL fallo.`n$($probe.Text)"
    }
    $client = [Net.Sockets.TcpClient]::new()
    try {
        $connection = $client.ConnectAsync('127.0.0.1', [int]$script:Config.MANUAL_DB_PORT)
        if (-not $connection.Wait(5000) -or -not $client.Connected) { throw 'Puerto no accesible' }
    } catch {
        throw "Windows no pudo alcanzar PostgreSQL en 127.0.0.1:$($script:Config.MANUAL_DB_PORT). Revise Docker Desktop y el puerto publicado."
    } finally { $client.Dispose() }
    $details = @(
        "Contenedor: $script:DbName", "Volumen: $script:DbVolume",
        "Puerto de la computadora: $($script:Config.MANUAL_DB_PORT)",
        'Comprobacion SQL con usuario, base y contrasena: correcta.',
        'Conexion TCP desde la computadora al puerto publicado: correcta.',
        'Las bases de intentos anteriores no se borraron.'
    ) -join "`n"
    Write-Text (Join-Path $script:Evidence 'base_manual.txt') $details
    Write-Host "PostgreSQL comprobado en el puerto $($script:Config.MANUAL_DB_PORT)."
}

function Start-ManualDb {
    Write-Host 'Preparando y comprobando PostgreSQL del taller.'
    $currentId = Get-OwnedContainer $script:DbName 'manual'
    Stop-ConflictingManualDb $currentId
    if ($currentId) {
        [void](Invoke-Docker @('start', $script:DbName))
    } else {
        Assert-PortFree ([int]$script:Config.MANUAL_DB_PORT)
        [void](Invoke-Docker @('run', '-d', '--name', $script:DbName, '--label', 'lealcaf.taller11=manual',
            '-p', "$($script:Config.MANUAL_DB_PORT):5432", '--env-file', '.env.postgres',
            '-v', "${script:DbVolume}:/var/lib/postgresql/data", 'postgres:16'))
    }
    for ($i = 0; $i -lt 90; $i++) {
        $ready = Invoke-Docker -Arguments @('exec', $script:DbName, 'pg_isready', '-h', '127.0.0.1', '-U', $script:Config.DB_USER, '-d', $script:Config.DB_NAME) -AllowFailure
        if ($ready.Code -eq 0) { Test-ManualDbAccess; return }
        Start-Sleep -Seconds 1
    }
    throw "La base aislada no estuvo lista. Ejecute docker logs $script:DbName."
}

function Build-Image {
    Write-Host 'Punto 1: construyendo las dos etapas. La primera descarga puede tardar.'
    [void](Invoke-Docker @('build', '-t', 'lealcaf:taller11', '.') (Join-Path $script:Evidence 'build.txt'))
    [void](Invoke-Docker @('build', '--target', 'compilacion', '-t', 'lealcaf:compilacion', '.') (Join-Path $script:Evidence 'build_compilacion.txt'))
    [void](Invoke-Docker @('image', 'ls', 'lealcaf') (Join-Path $script:Evidence 'imagenes.txt'))
    [void](Invoke-Docker @('history', 'lealcaf:taller11') (Join-Path $script:Evidence 'capas.txt'))
    $context = Invoke-Docker @('run', '--rm', 'lealcaf:compilacion', 'ls', '-Ra', '/src') (Join-Path $script:Evidence 'contexto.txt')
    if ($context.Text -match '(?m)^\.(env(\r?$|\.(?!example\r?$))|git\r?$)') { throw 'Se encontro un archivo privado en el contexto de compilacion.' }
    $user = Invoke-Docker @('image', 'inspect', '-f', '{{.Config.User}}', 'lealcaf:taller11') (Join-Path $script:Evidence 'usuario.txt')
    if ($user.Text.Trim().Split(':')[0] -ne '65532') { throw 'La imagen no esta configurada con el usuario nonroot esperado.' }
    $final = Invoke-Docker @('image', 'inspect', '-f', '{{.Size}}', 'lealcaf:taller11')
    $builder = Invoke-Docker @('image', 'inspect', '-f', '{{.Size}}', 'lealcaf:compilacion')
    $facts = "Tamano real final: **$($final.Text.Trim()) bytes** ($([math]::Round([double]$final.Text.Trim() / 1MB, 2)) MiB). " +
        "Etapa de compilacion: **$($builder.Text.Trim()) bytes** ($([math]::Round([double]$builder.Text.Trim() / 1MB, 2)) MiB). " +
        "Usuario observado: **$($user.Text.Trim())**. Evidencias: imagenes.txt, capas.txt, contexto.txt y usuario.txt."
    Set-DecisionBlock 'D1' $facts
}

function Stop-StartedProcess($Process) {
    if ($null -eq $Process) { return }
    if ([Environment]::OSVersion.Platform -eq [PlatformID]::Win32NT) {
        [void](Invoke-Native -Program taskkill.exe -Arguments @('/PID', [string]$Process.Id, '/T', '/F') -AllowFailure)
    } else {
        # Incluye todos los descendientes de go run en PowerShell sobre Unix.
        $Process.Refresh()
        if (-not $Process.HasExited) { $Process.Kill($true) }
    }
    if (-not $Process.WaitForExit(10000)) { throw 'El proceso de go run no termino en diez segundos.' }
}

function Read-ProcessCapture($ReadTask) {
    if ($null -eq $ReadTask) { return '' }
    if (-not $ReadTask.Wait(10000)) { throw 'El registro del proceso no termino de cerrarse en diez segundos.' }
    return $ReadTask.GetAwaiter().GetResult()
}

function Run-GoBaseline {
    Assert-PortFree ([int]$script:Config.PORT)
    $go = Get-Command go -ErrorAction Stop
    Write-Host 'Verificando la API actual: go build y go test.'
    $previousTestDsn = [Environment]::GetEnvironmentVariable('LEALCAF_TEST_POSTGRES_DSN', 'Process')
    try {
        $testDsn = "host=127.0.0.1 port=$($script:Config.MANUAL_DB_PORT) user=$($script:Config.DB_USER) password=$($script:Config.DB_PASSWORD) dbname=$($script:Config.DB_NAME) sslmode=disable"
        [Environment]::SetEnvironmentVariable('LEALCAF_TEST_POSTGRES_DSN', $testDsn, 'Process')
        [void](Invoke-Native go @('-C', $script:Module, 'build', './...') (Join-Path $script:Evidence 'go_build.txt'))
        [void](Invoke-Native go @('-C', $script:Module, 'test', '-count=1', '-v', './...') (Join-Path $script:Evidence 'go_test.txt'))
    } finally { [Environment]::SetEnvironmentVariable('LEALCAF_TEST_POSTGRES_DSN', $previousTestDsn, 'Process') }
    $previousGoEnv = @{}
    $server = $null
    $serverStarted = $false
    $stdoutTask = $null
    $stderrTask = $null
    try {
        foreach ($line in Get-Content (Join-Path $script:Root '.env.gorun')) {
            if ($line -match '^([A-Z_]+)=(.*)$') {
                $name = $matches[1]
                $previousGoEnv[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
                [Environment]::SetEnvironmentVariable($name, $matches[2], 'Process')
            }
        }
        $previousGoEnv['ENV_FILE'] = [Environment]::GetEnvironmentVariable('ENV_FILE', 'Process')
        [Environment]::SetEnvironmentVariable('ENV_FILE', (Join-Path $script:Root '.env.gorun'), 'Process')
        Write-Host 'Linea de base: arrancando go run .'
        # Capturar ambas salidas en memoria evita archivos bloqueados en Windows.
        $startInfo = [Diagnostics.ProcessStartInfo]::new()
        $startInfo.FileName = $go.Source
        $startInfo.Arguments = 'run .'
        $startInfo.WorkingDirectory = $script:Module
        $startInfo.UseShellExecute = $false
        $startInfo.CreateNoWindow = $true
        $startInfo.RedirectStandardOutput = $true
        $startInfo.RedirectStandardError = $true
        $startInfo.StandardOutputEncoding = [Text.Encoding]::UTF8
        $startInfo.StandardErrorEncoding = [Text.Encoding]::UTF8
        $server = [Diagnostics.Process]::new()
        $server.StartInfo = $startInfo
        if (-not $server.Start()) { throw 'No se pudo arrancar go run.' }
        $serverStarted = $true
        $stdoutTask = $server.StandardOutput.ReadToEndAsync()
        $stderrTask = $server.StandardError.ReadToEndAsync()
        Wait-Api $server
        [void](Test-Api 'gorun' -Save)
    } finally {
        foreach ($name in $previousGoEnv.Keys) { [Environment]::SetEnvironmentVariable($name, $previousGoEnv[$name], 'Process') }
        try {
            if ($serverStarted) { Stop-StartedProcess $server }
            $logs = @((Read-ProcessCapture $stdoutTask), (Read-ProcessCapture $stderrTask))
            Write-Text (Join-Path $script:Evidence 'logs_gorun.txt') (Protect-Text ($logs -join "`n"))
        } finally {
            if ($null -ne $server) { $server.Dispose() }
        }
    }
}

function Update-TimeDecision {
    $goPath = Join-Path $script:Evidence 'respuesta3_gorun.json'
    $dockerPath = Join-Path $script:Evidence 'respuesta3_contenedor.json'
    if (-not (Test-Path $goPath) -or -not (Test-Path $dockerPath)) { return }
    $goRaw = [IO.File]::ReadAllText($goPath)
    $dockerRaw = [IO.File]::ReadAllText($dockerPath)
    $goDate = [regex]::Match($goRaw, '"fecha"\s*:\s*"([^"]+)"').Groups[1].Value
    $dockerDate = [regex]::Match($dockerRaw, '"fecha"\s*:\s*"([^"]+)"').Groups[1].Value
    $goRecord = $goRaw | ConvertFrom-Json
    $dockerRecord = $dockerRaw | ConvertFrom-Json
    if ($goRecord.id -ne $dockerRecord.id) { throw 'La comparacion de hora requiere el mismo canje en ambas ejecuciones.' }
    $same = [DateTimeOffset]::Parse($goDate).UtcDateTime -eq [DateTimeOffset]::Parse($dockerDate).UtcDateTime
    $result = if ($same) { 'representan el mismo instante' } else { 'NO representan el mismo instante; investigar antes de entregar' }
    Set-DecisionBlock 'D4' "Observacion real del canje $($goRecord.id): con go run, **$goDate**; en el contenedor, **$dockerDate**. Las dos fechas $result. Se guardaron las respuestas completas en peticiones.md."
    if (-not $same) { throw 'La fecha cambio de instante, no solamente de zona horaria.' }
}

function Run-Container {
    Write-Host 'Punto 2: errores controlados y seis peticiones contra el contenedor.'
    Remove-ManualApi
    Assert-PortFree ([int]$script:Config.PORT)
    $missing = Invoke-Docker -Arguments @('run', '--rm', 'lealcaf:taller11') -OutputPath (Join-Path $script:Evidence '1_sin_variables.txt') -AllowFailure
    if ($missing.Code -eq 0) { throw 'La API arranco sin configuracion. Revise los secretos en el codigo.' }
    $wrong = Invoke-Docker -Arguments @('run', '--rm', '--env-file', '.env.localhost', 'lealcaf:taller11') -OutputPath (Join-Path $script:Evidence '2_local.txt') -AllowFailure
    if ($wrong.Code -eq 0) { throw 'El ensayo con localhost no fallo como se esperaba. Revise la direccion de la base.' }
    $arguments = @('run', '-d', '--name', $script:ApiName, '--label', 'lealcaf.taller11=api',
        '-p', "127.0.0.1:$($script:Config.PORT):$($script:Config.PORT)", '--env-file', '.env.docker')
    if ([Environment]::OSVersion.Platform -eq [PlatformID]::Unix) { $arguments += @('--add-host', 'host.docker.internal:host-gateway') }
    $arguments += 'lealcaf:taller11'
    try {
        [void](Invoke-Docker $arguments)
        Wait-Api
        [void](Test-Api 'contenedor' -Save)
        Update-TimeDecision
    } finally {
        [void](Invoke-Docker -Arguments @('logs', $script:ApiName) -OutputPath (Join-Path $script:Evidence 'logs_contenedor.txt') -AllowFailure)
        Remove-ManualApi
    }
}

function Get-ComposeContainer([string]$Project = '') {
    $arguments = @()
    if ($Project) { $arguments += @('-p', $Project) }
    $arguments += @('ps', '-q', 'api')
    $result = Invoke-Compose $arguments
    if (-not $result.Text.Trim()) { throw 'Compose no muestra el contenedor api en ejecucion.' }
    return $result.Text.Trim()
}

function Run-Compose {
    Write-Host 'Reto: tres arranques desde cero en el proyecto aislado lealcaf-taller11-ensayo.'
    [void](Invoke-Compose @('stop', 'api'))
    Assert-PortFree ([int]$script:Config.PORT)
    $cold = @('Tres arranques con base vacia. Solo se borran los datos del proyecto de ensayo.')
    try {
        for ($attempt = 1; $attempt -le 3; $attempt++) {
            [void](Invoke-Compose @('-p', $script:TrialProject, 'down', '-v', '--remove-orphans'))
            [void](Invoke-Compose @('-p', $script:TrialProject, 'up', '-d', '--build') (Join-Path $script:Evidence "arranque_limpio_$attempt.txt"))
            Wait-Api
            [void](Test-Api "compose-ensayo-$attempt")
            $id = Get-ComposeContainer $script:TrialProject
            $restart = Invoke-Docker @('inspect', '-f', '{{.RestartCount}}', $id)
            if ($restart.Text.Trim() -ne '0') { throw 'La API se reinicio: el arranque limpio no fue correcto.' }
            $cold += "Intento $attempt : seis codigos correctos; reinicios de API = 0; fecha UTC $([DateTime]::UtcNow.ToString('o'))."
            [void](Invoke-Compose @('-p', $script:TrialProject, 'logs', 'api') (Join-Path $script:Evidence "arranque_limpio_${attempt}_logs.txt"))
        }
    } finally {
        [void](Invoke-Compose @('-p', $script:TrialProject, 'down', '-v', '--remove-orphans'))
        Write-Text (Join-Path $script:Evidence 'arranques_limpios.txt') ($cold -join "`n")
    }
    Set-DecisionBlock 'D6' (($cold | Select-Object -Skip 1) -join ' ')
    Write-Host 'Arrancando el proyecto final con volumen persistente.'
    [void](Invoke-Compose @('up', '-d', '--build'))
    Wait-Api
    $before = Test-Api 'compose' -Save
    [void](Invoke-Compose @('down'))
    [void](Invoke-Compose @('up', '-d', '--build'))
    Wait-Api
    $admin = Invoke-Api POST '/login' '' @{ usuario = 'admin'; password = $script:Config.DEMO_ADMIN_PASSWORD }
    Assert-Code $admin 200 'Login tras reiniciar Compose'
    $persisted = Invoke-Api GET "/canjes/$($before.canje_nuevo)" $admin.Data.token
    Assert-Code $persisted 200 'Consulta del canje guardado antes de down'
    if ($persisted.Data.id -ne $before.canje_nuevo -or $persisted.Data.estado -ne 'confirmado') { throw 'El canje creado antes de down no se conservo.' }
    Write-Text (Join-Path $script:Evidence 'persistencia.txt') "Tras down sin -v y up se conservo el canje $($before.canje_nuevo) en estado confirmado. Fecha UTC: $([DateTime]::UtcNow.ToString('o'))."
    $ps = Invoke-Compose @('ps')
    $logs = Invoke-Compose @('logs', '--no-color', 'api')
    Write-Text (Join-Path $script:Evidence 'compose.txt') ($ps.Text + "`n`n" + $logs.Text)
}

function Show-Status {
    Update-Table
    Write-Host ([IO.File]::ReadAllText((Join-Path $script:Evidence 'peticiones.md')))
}

$previousLocation = Get-Location
try {
    Set-Location $script:Root
    if ($Paso -eq 'Estado') { Show-Status; exit 0 }
    Initialize-Config
    if ($Paso -eq 'Configurar') { Update-Table; exit 0 }
    foreach ($name in $script:Config.Keys) {
        $script:PreviousVariables[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
        [Environment]::SetEnvironmentVariable($name, $script:Config[$name], 'Process')
    }
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) { throw 'Instale Docker Desktop, abralo y espere a que este listo.' }
    [void](Invoke-Docker @('info'))
    [void](Invoke-Compose @('version') (Join-Path $script:Evidence 'version_compose.txt'))
    [void](Invoke-Compose @('config', '--quiet'))
    Add-Type -AssemblyName System.Net.Http
    $handler = [Net.Http.HttpClientHandler]::new()
    $handler.UseProxy = $false
    $script:Http = [Net.Http.HttpClient]::new($handler)
    $script:Http.Timeout = [TimeSpan]::FromSeconds(3)
    switch ($Paso) {
        'Imagen' { Build-Image }
        'GoRun' { Start-ManualDb; Run-GoBaseline }
        'Contenedor' { Start-ManualDb; Run-Container }
        'Compose' { Run-Compose }
        'Todo' {
            # Un resultado anterior no acredita esta nueva ejecucion.
            foreach ($name in @('resultado.txt', 'ultimo_error.txt', 'peticiones_gorun.json', 'peticiones_contenedor.json', 'peticiones_compose.json', 'respuesta3_gorun.json', 'respuesta3_contenedor.json', 'respuesta3_compose.json')) {
                $oldPath = Join-Path $script:Evidence $name
                if (Test-Path $oldPath) { Remove-Item $oldPath -Force }
            }
            Update-Table
            if (-not (Get-Command go -ErrorAction SilentlyContinue)) { throw 'Instale Go 1.27.1 o superior antes de ejecutar la linea de base.' }
            [void](Invoke-Native go @('version') (Join-Path $script:Evidence 'version_go.txt'))
            Remove-ManualApi
            [void](Invoke-Compose @('stop', 'api'))
            Start-ManualDb
            Build-Image
            Run-GoBaseline
            Run-Container
            Run-Compose
            Write-Text (Join-Path $script:Evidence 'resultado.txt') "Taller completado en esta computadora el $([DateTime]::UtcNow.ToString('o')) UTC: imagen, seis peticiones en tres modos, tres arranques limpios y persistencia tras down."
        }
    }
    Update-Table
    Write-Host "LISTO. Revise docs/taller11 y docs/decisiones.md. API: $script:BaseUrl"
    Write-Host 'Bruno: abra la coleccion del Hito 2 y seleccione Taller11-privado. No publique ese entorno.'
} catch {
    $errorText = Protect-Text $_.Exception.Message
    Write-Host "DETENIDO: $errorText" -ForegroundColor Red
    Write-Text (Join-Path $script:Evidence 'ultimo_error.txt') $errorText
    exit 1
} finally {
    if ($null -ne $script:Http) { $script:Http.Dispose() }
    foreach ($name in $script:PreviousVariables.Keys) { [Environment]::SetEnvironmentVariable($name, $script:PreviousVariables[$name], 'Process') }
    Set-Location $previousLocation
}
