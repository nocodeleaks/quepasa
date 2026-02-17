param(
  [string]$HostIP = "143.208.224.21",
  [int]$SshPort = 26492,
  [string]$User = "root",
  [string]$KeyPath = "$env:USERPROFILE\\.ssh\\id_ed25519_sufficit",
  [int]$DurationSeconds = 180,
  [string]$RemoteBase = "/opt/quepasa",
  [string]$RemoteSubdir = ".dist/pcaps",
  [string]$Interface = "eth0",
  [string]$FilePrefix = "webrtc",
  [string]$CallID = "",
  # Default captures STUN/TURN (3478/5349) + typical high UDP media ports.
  # Tip: to avoid SIP noise, set Filter to something like:
  #   "udp and host 57.144.179.54 and port 3478"
  [string]$Filter = "udp and (port 3478 or port 5349 or portrange 10000-65000)",
  [string]$LocalOutDir = ".dist/pcaps"
)

$ErrorActionPreference = "Stop"

function Require-Command([string]$Name) {
  $cmd = Get-Command $Name -ErrorAction SilentlyContinue
  if (-not $cmd) {
    throw "Missing required command '$Name'. Ensure OpenSSH client is installed and on PATH."
  }
}

Require-Command ssh
Require-Command scp

if (-not (Test-Path $KeyPath)) {
  throw "SSH key not found at: $KeyPath"
}

$remoteDir = "$RemoteBase/$RemoteSubdir"

# Ensure local out dir exists
New-Item -ItemType Directory -Force $LocalOutDir | Out-Null

Write-Host "Starting remote tcpdump capture on ${User}@${HostIP}:${SshPort} for $DurationSeconds seconds" -ForegroundColor Cyan
Write-Host "Remote dir: $remoteDir" -ForegroundColor DarkCyan
Write-Host "Interface: $Interface" -ForegroundColor DarkCyan
Write-Host "Filter: $Filter" -ForegroundColor DarkCyan

# Build a bash command that:
# - creates the remote dir
# - generates a timestamped filename
# - runs tcpdump under timeout
# - prints PID/PCAP/LOG so caller can download
#
# IMPORTANT: quoting is tricky across PowerShell -> SSH -> bash.
# Strategy:
# - Build a bash script as a list of simple statements.
# - Store the tcpdump filter in a bash variable (single-quoted), then pass it to tcpdump as "$filter".
$callIDSafe = ($CallID -replace "[^A-Za-z0-9_-]", "")
$filterEsc = $Filter -replace "'", "'\\''"

$bashLines = @(
  'set -e',
  ("mkdir -p {0}" -f $remoteDir),
  'ts=$(date +%Y%m%d_%H%M%S)',
  ("cid='{0}'" -f $callIDSafe),
  'suffix=""',
  'if [ -n "$cid" ]; then suffix="_$cid"; fi',
  ('pcap=' + $remoteDir + '/' + $FilePrefix + '${suffix}_${ts}.pcap'),
  ('log=' + $remoteDir + '/' + $FilePrefix + '${suffix}_${ts}.log'),
  ('timeout {0} tcpdump -i {1} -nn -s0 -w "$pcap" ''{2}'' >"$log" 2>&1 || true' -f $DurationSeconds, $Interface, $filterEsc),
  'echo PCAP=$pcap',
  'echo LOG=$log'
)

$bashScript = ($bashLines -join "`n") + "`n"
$bashScript = $bashScript.Replace("`r`n", "`n").Replace("`r", "")

# Execute by piping the script over SSH to avoid complex quoting issues.
# We also strip CR (Windows CRLF) on the remote side (octal 015) so bash doesn't see `$'\r'`.
$output = $bashScript | & ssh -i $KeyPath -p $SshPort "$User@$HostIP" "tr -d '\015' | bash -s"
if ($LASTEXITCODE -ne 0) {
  throw "Remote capture command failed (exit=$LASTEXITCODE)"
}
$pcapRemote = ($output | Select-String -Pattern '^PCAP=' | ForEach-Object { $_.Line.Substring(5) } | Select-Object -First 1)
$logRemote = ($output | Select-String -Pattern '^LOG=' | ForEach-Object { $_.Line.Substring(4) } | Select-Object -First 1)

if ([string]::IsNullOrWhiteSpace($pcapRemote) -or [string]::IsNullOrWhiteSpace($logRemote)) {
  throw "Remote capture did not return PCAP/LOG paths. Output: $output"
}

Write-Host "Remote capture completed." -ForegroundColor Green
Write-Host "Remote PCAP: $pcapRemote" -ForegroundColor Green
Write-Host "Remote LOG : $logRemote" -ForegroundColor Green

Write-Host "Downloading capture files..." -ForegroundColor Cyan
$pcapSrc = ("$User@{0}:{1}" -f $HostIP, $pcapRemote)
$logSrc = ("$User@{0}:{1}" -f $HostIP, $logRemote)

& scp -i $KeyPath -P $SshPort $pcapSrc $LocalOutDir | Out-Host
& scp -i $KeyPath -P $SshPort $logSrc $LocalOutDir | Out-Host

$pcapLocal = Join-Path $LocalOutDir (Split-Path -Leaf $pcapRemote)
$logLocal = Join-Path $LocalOutDir (Split-Path -Leaf $logRemote)

Write-Host "Saved local PCAP: $pcapLocal" -ForegroundColor Green
Write-Host "Saved local LOG : $logLocal" -ForegroundColor Green

Get-Item $pcapLocal, $logLocal | Format-Table Name, Length, LastWriteTime -AutoSize
