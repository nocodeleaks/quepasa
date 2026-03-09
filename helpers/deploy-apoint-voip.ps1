param(
  [string]$HostIP = "198.51.100.10",
  [int]$SshPort = 26492,
  [string]$User = "root",
  [string]$KeyPath = "$env:USERPROFILE\\.ssh\\id_ed25519_public_example",
  [string]$RemoteBase = "/opt/quepasa",
  [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"

function Require-Command([string]$Name) {
  $cmd = Get-Command $Name -ErrorAction SilentlyContinue
  if (-not $cmd) {
    throw "Missing required command '$Name'. Install OpenSSH client or ensure it's on PATH."
  }
}

Require-Command ssh
Require-Command scp

if (-not (Test-Path $KeyPath)) {
  throw "SSH key not found at: $KeyPath"
}

$repoRoot = Split-Path -Parent $PSScriptRoot
$srcPath = Join-Path $repoRoot "src"
$envFile = Join-Path $srcPath ".env.public-voip-example"

if (-not (Test-Path $srcPath)) { throw "src/ not found: $srcPath" }
if (-not (Test-Path $envFile)) { throw "Missing env file: $envFile" }

Write-Host "Deploying to $User@${HostIP}:$RemoteBase (ssh port $SshPort)" -ForegroundColor Cyan

$sshBase = @("-i", $KeyPath, "-p", "$SshPort", "$User@$HostIP")

# 1) Create directories
& ssh @sshBase "mkdir -p $RemoteBase; mkdir -p $RemoteBase/.dist/call_dumps" | Out-Host

# 2) Upload src/ (includes views/, swagger/, etc.)
#    Note: this overwrites remote src/ entirely.
& ssh @sshBase "rm -rf $RemoteBase/src" | Out-Host
& scp -i $KeyPath -P $SshPort -r "$srcPath" "$User@${HostIP}:$RemoteBase/" | Out-Host

# 3) Activate env file on server
& ssh @sshBase "cp -f $RemoteBase/src/.env.public-voip-example $RemoteBase/src/.env" | Out-Host

# 3.1) Fix ownership so systemd user can write sqlite DBs and dumps
& ssh @sshBase "chown -R quepasa:quepasa $RemoteBase/src; chown -R quepasa:quepasa $RemoteBase/.dist || true" | Out-Host

# 4) Build on server (needs Go + gcc for sqlite3)
if (-not $SkipBuild) {
  $buildCmd = @(
    "set -e",
    "cd $RemoteBase/src",
    "export PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
    "if ! command -v go >/dev/null 2>&1; then echo 'Go not installed (expected at /usr/local/go/bin/go)'; exit 2; fi",
    "if ! command -v gcc >/dev/null 2>&1; then echo 'gcc not installed (required for go-sqlite3)'; exit 3; fi",
    "go version",
    "go build -o $RemoteBase/quepasa",
    "echo 'Built: $RemoteBase/quepasa'"
  ) -join "; "

  & ssh @sshBase $buildCmd | Out-Host
}

Write-Host "Done. Next: run on server with: cd $RemoteBase; ./quepasa" -ForegroundColor Green
