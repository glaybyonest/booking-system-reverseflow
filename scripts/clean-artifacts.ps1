[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$targets = @(
    "frontend/.next",
    "frontend/tsconfig.tsbuildinfo",
    "frontend/coverage",
    "backend/bin",
    "backend/coverage.out",
    "tmp",
    "temp"
)

foreach ($target in $targets) {
    $path = Join-Path $root $target
    if (Test-Path -LiteralPath $path) {
        Remove-Item -LiteralPath $path -Recurse -Force
        Write-Host "Removed $target"
    }
}
