param(
  [string]$HostIP = "198.51.100.10",
  [int]$SshPort = 26492,
  [string]$User = "root",
  [string]$KeyPath = "$env:USERPROFILE\\.ssh\\id_ed25519_public_example",
  [string]$RemoteBase = "/opt/quepasa",
  [string]$RemoteDumpSubdir = ".dist/call_dumps",
  [string]$RemoteUnit = "quepasa",
  [string]$CallID = "",
  [switch]$Latest,
  [int]$JournalTail = 6000,
  [int]$WaitForNewDumpsSeconds = 8,
  [int]$WaitForNewDumpsPollMs = 800,
  [string]$LocalBaseDir = ".dist/server_artifacts"
)

$ErrorActionPreference = "Stop"

function RequireCommand([string]$Name) {
  $cmd = Get-Command $Name -ErrorAction SilentlyContinue
  if (-not $cmd) {
    throw "Missing required command '$Name'. Ensure OpenSSH client is installed and on PATH."
  }
}

RequireCommand ssh
RequireCommand scp

if (-not (Test-Path $KeyPath)) {
  throw "SSH key not found at: $KeyPath"
}

if ($Latest -and -not [string]::IsNullOrWhiteSpace($CallID)) {
  throw "Use -Latest OR -CallID, not both."
}

if (-not $Latest -and [string]::IsNullOrWhiteSpace($CallID)) {
  throw "Provide -CallID or use -Latest to auto-detect the most recent call."
}

$dumpDir = "$RemoteBase/$RemoteDumpSubdir"

function Invoke-RemoteBash([string[]]$Lines) {
  $script = ($Lines -join "`n") + "`n"
  $script = $script.Replace("`r`n", "`n").Replace("`r", "")
  return $script | & ssh -i $KeyPath -p $SshPort "$User@$HostIP" "tr -d '\015' | bash -s"
}

if ($Latest) {
  Write-Host "Auto-detecting latest CallID from journalctl..." -ForegroundColor Cyan
  $out = Invoke-RemoteBash @(
    'set -e',
    ("journalctl -u {0} --no-pager -n {1} | grep -F '[CALL] Offer:' | tail -n 1" -f $RemoteUnit, $JournalTail),
    'true'
  )

  $line = ($out | Select-Object -Last 1)
  if ([string]::IsNullOrWhiteSpace($line)) {
    throw "Could not find a recent '[CALL] Offer:' line in journalctl. Increase -JournalTail or provide -CallID."
  }

  $m = [regex]::Match($line, 'callID=([A-Fa-f0-9]{16,64})')
  if (-not $m.Success) {
    throw "Could not parse callID from line: $line"
  }

  $CallID = $m.Groups[1].Value.ToUpperInvariant()
  Write-Host "Detected CallID=$CallID" -ForegroundColor Green
}

$callIDSafe = ($CallID -replace '[^A-Za-z0-9_-]', '')
if ([string]::IsNullOrWhiteSpace($callIDSafe)) {
  throw "Invalid CallID after sanitization: '$CallID'"
}

$localOutDir = Join-Path $LocalBaseDir $callIDSafe
New-Item -ItemType Directory -Force $localOutDir | Out-Null

Write-Host "Fetching artifacts for CallID=$CallID" -ForegroundColor Cyan
Write-Host "Remote dump dir: $dumpDir" -ForegroundColor DarkCyan
Write-Host "Local out dir : $localOutDir" -ForegroundColor DarkCyan

# 1) Fetch matching dump files
Write-Host "Listing remote dump files..." -ForegroundColor Cyan
$list = Invoke-RemoteBash @(
  'set -e',
  ("cid='{0}'" -f $callIDSafe),
  ("d='{0}'" -f $dumpDir),
  'shopt -s nullglob',
  'for f in "$d"/*"$cid"*; do',
  '  if [ -f "$f" ]; then echo "$f"; fi',
  'done'
)

$remoteFiles = @($list | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
$downloaded = New-Object 'System.Collections.Generic.HashSet[string]'
if ($remoteFiles.Count -eq 0) {
  Write-Warning "No files matched '$dumpDir/*$callIDSafe*'"
} else {
  Write-Host ("Found {0} dump file(s). Downloading..." -f $remoteFiles.Count) -ForegroundColor Green
  foreach ($f in $remoteFiles) {
    $src = "$User@${HostIP}:$f"
    & scp -i $KeyPath -P $SshPort $src $localOutDir | Out-Host
    [void]$downloaded.Add($f)
  }
}

# 1.1) Some dumps are produced a couple seconds after the first ones (e.g. TURN probe dumps).
# Re-scan for a short window and download anything new.
if ($WaitForNewDumpsSeconds -gt 0) {
  $deadline = (Get-Date).AddSeconds([math]::Max(1, $WaitForNewDumpsSeconds))
  do {
    Start-Sleep -Milliseconds ([math]::Max(200, $WaitForNewDumpsPollMs))

    $list2 = Invoke-RemoteBash @(
      'set -e',
      ("cid='{0}'" -f $callIDSafe),
      ("d='{0}'" -f $dumpDir),
      'shopt -s nullglob',
      'for f in "$d"/*"$cid"*; do',
      '  if [ -f "$f" ]; then echo "$f"; fi',
      'done'
    )

    $remoteFiles2 = @($list2 | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
    $newOnes = @($remoteFiles2 | Where-Object { -not $downloaded.Contains($_) })
    if ($newOnes.Count -gt 0) {
      Write-Host ("Found {0} new dump file(s). Downloading..." -f $newOnes.Count) -ForegroundColor Yellow
      foreach ($f in $newOnes) {
        $src = "$User@${HostIP}:$f"
        & scp -i $KeyPath -P $SshPort $src $localOutDir | Out-Host
        [void]$downloaded.Add($f)
      }
      break
    }
  } while ((Get-Date) -lt $deadline)
}

# 2) Capture journal excerpts for this CallID
$journalLocal = Join-Path $localOutDir ("journal_{0}.log" -f $callIDSafe)
Write-Host "Saving filtered journalctl logs..." -ForegroundColor Cyan
$journalOut = Invoke-RemoteBash @(
  'set -e',
  ("journalctl -u $RemoteUnit --no-pager -n $JournalTail | grep -F CallID=$callIDSafe || true"),
  ("journalctl -u $RemoteUnit --no-pager -n $JournalTail | grep -F callID=$callIDSafe || true")
)
$journalOut | Out-File -FilePath $journalLocal -Encoding utf8

Write-Host "Saved journal: $journalLocal" -ForegroundColor Green

# 3) Print quick pointers
Write-Host "Done." -ForegroundColor Green
Get-ChildItem -Path $localOutDir | Sort-Object LastWriteTime | Format-Table Name, Length, LastWriteTime -AutoSize
