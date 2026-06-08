param(
  [string]$Aria2Path = "aria2c",
  [string]$EnvPath = "",
  [string]$DownloadDir = "",
  [string]$Secret = "",
  [string]$RpcUrl = "",
  [string]$SessionFile = "",
  [switch]$ListenAll,
  [switch]$Background
)

$ErrorActionPreference = "Stop"

function Resolve-RepoRoot {
  $scriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
  return (Resolve-Path (Join-Path $scriptRoot "..")).Path
}

function Find-EnvFile {
  param([string]$RepoRoot, [string]$ExplicitPath)

  if ($ExplicitPath) {
    if (-not (Test-Path -LiteralPath $ExplicitPath)) {
      throw "指定的 .env 文件不存在: $ExplicitPath"
    }
    return (Resolve-Path -LiteralPath $ExplicitPath).Path
  }

  $candidates = @(
    (Join-Path $RepoRoot ".env"),
    (Join-Path $RepoRoot "backend\.env"),
    (Join-Path $RepoRoot "release\openbridge-windows-v1.0\.env")
  )

  foreach ($candidate in $candidates) {
    if (Test-Path -LiteralPath $candidate) {
      return (Resolve-Path -LiteralPath $candidate).Path
    }
  }

  return $null
}

function Read-EnvMap {
  param([string]$FilePath)

  $result = @{}
  if (-not $FilePath) {
    return $result
  }

  foreach ($line in Get-Content -LiteralPath $FilePath -Encoding UTF8) {
    $trimmed = $line.Trim()
    if (-not $trimmed -or $trimmed.StartsWith("#")) {
      continue
    }
    $pair = $trimmed -split "=", 2
    if ($pair.Count -lt 2) {
      continue
    }
    $key = $pair[0].Trim()
    $value = $pair[1].Trim()
    if ($value.Length -ge 2) {
      if (($value.StartsWith('"') -and $value.EndsWith('"')) -or ($value.StartsWith("'") -and $value.EndsWith("'"))) {
        $value = $value.Substring(1, $value.Length - 2)
      }
    }
    $result[$key] = $value
  }

  return $result
}

function Get-MapValue {
  param(
    [hashtable]$Map,
    [string[]]$Keys,
    [string]$Fallback = ""
  )

  foreach ($key in $Keys) {
    if ($Map.ContainsKey($key) -and $Map[$key]) {
      return $Map[$key]
    }
  }
  return $Fallback
}

function Parse-RpcUrl {
  param([string]$Value)

  $normalized = $Value.Trim()
  if (-not $normalized) {
    $normalized = "http://127.0.0.1:6800/jsonrpc"
  }

  if ($normalized -notmatch "^[a-zA-Z]+://") {
    $normalized = "http://$normalized"
  }

  $uri = [System.Uri]$normalized
  $port = if ($uri.Port -gt 0) { $uri.Port } else { 6800 }
  return @{
    Host = $uri.Host
    Port = $port
    Url  = $normalized
  }
}

function Test-Aria2Executable {
  param([string]$Value)

  $command = Get-Command $Value -ErrorAction SilentlyContinue
  if ($command) {
    return $command.Source
  }

  if (Test-Path -LiteralPath $Value) {
    return (Resolve-Path -LiteralPath $Value).Path
  }

  throw "找不到 aria2c 可执行文件，请通过 -Aria2Path 指定路径。"
}

$repoRoot = Resolve-RepoRoot
$envFile = Find-EnvFile -RepoRoot $repoRoot -ExplicitPath $EnvPath
$envMap = Read-EnvMap -FilePath $envFile

$resolvedDownloadDir = if ($DownloadDir) { $DownloadDir } else { Get-MapValue -Map $envMap -Keys @("ARIA2_DOWNLOAD_DIR") -Fallback (Join-Path $repoRoot "downloads") }
$resolvedSecret = if ($Secret) { $Secret } else { Get-MapValue -Map $envMap -Keys @("ARIA2_SECRET", "ARIA2_RPC_SECRET") }
$resolvedRpcUrl = if ($RpcUrl) { $RpcUrl } else { Get-MapValue -Map $envMap -Keys @("ARIA2_RPC_URL") -Fallback "http://127.0.0.1:6800/jsonrpc" }
$rpc = Parse-RpcUrl -Value $resolvedRpcUrl
$resolvedSessionFile = if ($SessionFile) { $SessionFile } else { Join-Path $repoRoot "data\aria2.session" }
$resolvedAria2Path = Test-Aria2Executable -Value $Aria2Path

$resolvedDownloadDir = [System.IO.Path]::GetFullPath($resolvedDownloadDir)
$resolvedSessionFile = [System.IO.Path]::GetFullPath($resolvedSessionFile)

$downloadDirParent = Split-Path -Parent $resolvedDownloadDir
$sessionDir = Split-Path -Parent $resolvedSessionFile

if ($downloadDirParent -and -not (Test-Path -LiteralPath $downloadDirParent)) {
  New-Item -ItemType Directory -Path $downloadDirParent -Force | Out-Null
}
if (-not (Test-Path -LiteralPath $resolvedDownloadDir)) {
  New-Item -ItemType Directory -Path $resolvedDownloadDir -Force | Out-Null
}
if ($sessionDir -and -not (Test-Path -LiteralPath $sessionDir)) {
  New-Item -ItemType Directory -Path $sessionDir -Force | Out-Null
}
if (-not (Test-Path -LiteralPath $resolvedSessionFile)) {
  New-Item -ItemType File -Path $resolvedSessionFile -Force | Out-Null
}

$listenHost = if ($ListenAll) { "0.0.0.0" } else { "127.0.0.1" }
$args = @(
  "--enable-rpc=true",
  "--rpc-listen-port=$($rpc.Port)",
  "--rpc-allow-origin-all=true",
  "--continue=true",
  "--max-concurrent-downloads=5",
  "--max-connection-per-server=16",
  "--min-split-size=10M",
  "--split=16",
  "--dir=$resolvedDownloadDir",
  "--input-file=$resolvedSessionFile",
  "--save-session=$resolvedSessionFile",
  "--save-session-interval=30",
  "--file-allocation=none",
  "--daemon=false"
)

if ($listenHost -eq "127.0.0.1") {
  $args += "--rpc-listen-all=false"
} else {
  $args += "--rpc-listen-all=true"
}

if ($resolvedSecret) {
  $args += "--rpc-secret=$resolvedSecret"
}

Write-Host "Aria2 executable : $resolvedAria2Path"
Write-Host "Env file         : $(if ($envFile) { $envFile } else { '(none)' })"
Write-Host "Download dir     : $resolvedDownloadDir"
Write-Host "Session file     : $resolvedSessionFile"
Write-Host "RPC URL          : http://${listenHost}:$($rpc.Port)/jsonrpc"
Write-Host "Listen all       : $($ListenAll.IsPresent)"
Write-Host ""

if ($Background) {
  $process = Start-Process -FilePath $resolvedAria2Path -ArgumentList $args -PassThru -WindowStyle Hidden
  Write-Host "aria2c 已在后台启动，PID=$($process.Id)"
  exit 0
}

& $resolvedAria2Path @args
