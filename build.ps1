$ErrorActionPreference = 'Stop'

$repoRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$frontendDir = Join-Path $repoRoot 'frontend'
$backendDir = Join-Path $repoRoot 'backend'
$embeddedDistDir = Join-Path $backendDir 'web\dist'
$releaseDir = Join-Path $repoRoot 'release'
$outputExe = Join-Path $releaseDir 'openbridge.exe'
$envExampleSource = Join-Path $backendDir '.env.example'
$envExampleTarget = Join-Path $releaseDir '.env.example'

Write-Host '[1/3] Building frontend...'
Push-Location $frontendDir
try {
    npm run build
} finally {
    Pop-Location
}

Write-Host '[2/3] Copying frontend dist into backend embed directory...'
if (Test-Path $embeddedDistDir) {
    Remove-Item $embeddedDistDir -Recurse -Force
}
New-Item -ItemType Directory -Force -Path (Join-Path $backendDir 'web') | Out-Null
Copy-Item -Recurse -Force (Join-Path $frontendDir 'dist') $embeddedDistDir

Write-Host '[3/3] Building backend executable...'
New-Item -ItemType Directory -Force -Path $releaseDir | Out-Null
Push-Location $backendDir
try {
    go build -o $outputExe .
} finally {
    Pop-Location
}

if (Test-Path $envExampleSource) {
    Copy-Item -Force $envExampleSource $envExampleTarget
}

Write-Host ''
Write-Host "Build complete: $outputExe"
if (Test-Path $envExampleTarget) {
    Write-Host "Env example copied to: $envExampleTarget"
}
