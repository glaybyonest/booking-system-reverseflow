[CmdletBinding()]
param(
    [switch]$NoBuild,
    [switch]$SkipSeed
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$envFile = if (Test-Path -LiteralPath ".env") { ".env" } else { ".env.example" }

function Get-ConfigValue {
    param(
        [string]$Name,
        [string]$Fallback
    )

    $processValue = [Environment]::GetEnvironmentVariable($Name, "Process")
    if ($processValue) {
        return $processValue
    }

    if (Test-Path -LiteralPath $envFile) {
        $line = Get-Content -Encoding UTF8 $envFile |
            Where-Object { $_ -match "^\s*$Name\s*=" } |
            Select-Object -Last 1
        if ($line) {
            return ($line -replace "^\s*$Name\s*=\s*", "").Trim()
        }
    }

    return $Fallback
}

$composeArgs = @("compose", "--env-file", $envFile, "-f", "deploy/docker/docker-compose.yml")
$upArgs = $composeArgs + @("up", "-d")
if (-not $NoBuild) {
    $upArgs += "--build"
}

Write-Host "Starting Docker services..."
docker @upArgs
if ($LASTEXITCODE -ne 0) {
    throw "docker compose up failed"
}

Write-Host "Applying database migrations..."
$migrationsPath = (Resolve-Path "backend/migrations").Path
$previousErrorActionPreference = $ErrorActionPreference
$ErrorActionPreference = "Continue"
$migrateOutput = & docker run --rm --network reserveflow_reserveflow -v "${migrationsPath}:/migrations:ro" migrate/migrate:v4.18.2 -path=/migrations -database "postgres://reserveflow:reserveflow@postgres:5432/reserveflow?sslmode=disable" up 2>&1
$migrateCode = $LASTEXITCODE
$ErrorActionPreference = $previousErrorActionPreference
$migrateOutput | ForEach-Object { Write-Host $_ }
if ($migrateCode -ne 0 -and ($migrateOutput -notmatch "no change")) {
    throw "database migration failed"
}

if (-not $SkipSeed) {
    Write-Host "Loading local access accounts..."
    $seedFiles = @(
        @{ Source = "backend/seeds/dev-users.sql"; Target = "/tmp/reserveflow-dev-users.sql"; Label = "local access accounts" }
    )

    foreach ($seed in $seedFiles) {
        $copyArgs = $composeArgs + @("cp", $seed.Source, "postgres:$($seed.Target)")
        docker @copyArgs
        if ($LASTEXITCODE -ne 0) {
            throw "copying $($seed.Label) failed"
        }

        $execArgs = $composeArgs + @("exec", "-T", "postgres", "psql", "-U", "reserveflow", "-d", "reserveflow", "-v", "ON_ERROR_STOP=1", "-f", $seed.Target)
        docker @execArgs
        if ($LASTEXITCODE -ne 0) {
            throw "$($seed.Label) seed failed"
        }
    }
}

Write-Host ""
Write-Host "ReserveFlow is ready:"
Write-Host "Frontend:   http://localhost:$(Get-ConfigValue FRONTEND_PORT 3000)"
Write-Host "Backend:    http://localhost:$(Get-ConfigValue BACKEND_PORT 18080)/health"
Write-Host "Prometheus: http://localhost:$(Get-ConfigValue PROMETHEUS_PORT 9090)"
Write-Host "Grafana:    http://localhost:$(Get-ConfigValue GRAFANA_PORT 3001)"
