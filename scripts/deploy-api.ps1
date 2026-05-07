param(
  [Parameter(Mandatory = $true)]
  [string]$HostName,

  [Parameter(Mandatory = $true)]
  [string]$UserName,

  [Parameter(Mandatory = $true)]
  [string]$RemotePath,

  [int]$Port = 22,
  [string]$ServiceName = 'zero-api',
  [string]$RemoteBinaryName = 'zero-api',
  [string]$RemoteConfigPath = '',
  [string]$RemoteUploadPath = '/tmp/zero-api-deploy',
  [string]$GoArch = 'amd64',
  [string]$HealthCheckUrl = 'http://127.0.0.1:8888/api/v1/health',
  [switch]$SkipTests
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Assert-CommandExists {
  param([Parameter(Mandatory = $true)][string]$Name)

  if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
    throw "Missing required command: $Name"
  }
}

function Assert-RequiredPath {
  param([Parameter(Mandatory = $true)][string]$Path)

  if (-not (Test-Path -LiteralPath $Path)) {
    throw "Missing required path: $Path"
  }
}

function Escape-ShSingleQuoted {
  param([Parameter(Mandatory = $true)][string]$Value)

  return $Value.Replace("'", "'\''")
}

Assert-CommandExists -Name 'go'
Assert-CommandExists -Name 'scp'
Assert-CommandExists -Name 'ssh'

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$apiRoot = (Resolve-Path (Join-Path $scriptDir '..')).Path
$remoteTarget = '{0}@{1}' -f $UserName, $HostName
$remoteBinaryPath = ('{0}/{1}' -f $RemotePath.TrimEnd('/'), $RemoteBinaryName)

$requiredPaths = @(
  'zero.go',
  'go.mod',
  'go.sum',
  'internal',
  'model'
)

foreach ($relativePath in $requiredPaths) {
  Assert-RequiredPath -Path (Join-Path $apiRoot $relativePath)
}

if ($RemoteConfigPath -eq '') {
  $RemoteConfigPath = ('{0}/etc/zero-api.yaml' -f $RemotePath.TrimEnd('/'))
}

$stagingDir = Join-Path ([System.IO.Path]::GetTempPath()) ('zero-api-deploy-' + [System.Guid]::NewGuid().ToString('N'))
$binaryPath = Join-Path $stagingDir $RemoteBinaryName

New-Item -ItemType Directory -Path $stagingDir | Out-Null

$previousGoos = $env:GOOS
$previousGoarch = $env:GOARCH
$previousCgo = $env:CGO_ENABLED

Push-Location $apiRoot
try {
  if (-not $SkipTests) {
    & go test ./...
    if ($LASTEXITCODE -ne 0) {
      throw 'API tests failed.'
    }
  }

  Write-Host ('Building API binary with GOOS=linux GOARCH={0}...' -f $GoArch)
  $env:GOOS = 'linux'
  $env:GOARCH = $GoArch
  $env:CGO_ENABLED = '0'

  & go build -trimpath -ldflags '-s -w' -o $binaryPath ./zero.go
  if ($LASTEXITCODE -ne 0) {
    throw 'Failed to build API binary.'
  }

  & scp -P $Port $binaryPath ('{0}:{1}' -f $remoteTarget, $RemoteUploadPath)
  if ($LASTEXITCODE -ne 0) {
    throw 'Failed to upload API binary.'
  }

  $remotePathQ = Escape-ShSingleQuoted -Value $RemotePath
  $remoteUploadPathQ = Escape-ShSingleQuoted -Value $RemoteUploadPath
  $remoteBinaryPathQ = Escape-ShSingleQuoted -Value $remoteBinaryPath
  $remoteConfigPathQ = Escape-ShSingleQuoted -Value $RemoteConfigPath
  $serviceNameQ = Escape-ShSingleQuoted -Value $ServiceName
  $healthCheckUrlQ = Escape-ShSingleQuoted -Value $HealthCheckUrl

  $remoteSteps = @(
    'set -e',
    "mkdir -p '$remotePathQ'",
    "test -f '$remoteConfigPathQ'",
    "if [ -f '$remoteBinaryPathQ' ]; then cp '$remoteBinaryPathQ' '$remoteBinaryPathQ.bak.'`$(date +%Y%m%d%H%M%S); fi",
    "mv '$remoteUploadPathQ' '$remoteBinaryPathQ'",
    "chmod 755 '$remoteBinaryPathQ'",
    "systemctl restart '$serviceNameQ'",
    "systemctl status '$serviceNameQ' --no-pager",
    "curl -fsS '$healthCheckUrlQ' >/dev/null"
  )

  $remoteCommand = [string]::Join('; ', $remoteSteps)

  & ssh -p $Port $remoteTarget $remoteCommand
  if ($LASTEXITCODE -ne 0) {
    throw 'Failed to install API binary or restart remote service.'
  }
}
finally {
  Pop-Location
  $env:GOOS = $previousGoos
  $env:GOARCH = $previousGoarch
  $env:CGO_ENABLED = $previousCgo

  if (Test-Path -LiteralPath $stagingDir) {
    Remove-Item -LiteralPath $stagingDir -Recurse -Force
  }
}

Write-Host ''
Write-Host 'API deploy finished.' -ForegroundColor Green
Write-Host ('Remote binary: {0}' -f $remoteBinaryPath)
Write-Host ('Remote config: {0}' -f $RemoteConfigPath)
Write-Host ('Service: {0}' -f $ServiceName)
