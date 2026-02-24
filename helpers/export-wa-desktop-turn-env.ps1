param(
  [Parameter(Mandatory = $true)]
  [string]$TurnAllocateJson,
  [ValidateSet('semicolon','json-compact','json-pretty')]
  [string]$Format = 'semicolon'
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $TurnAllocateJson)) {
  throw "File not found: $TurnAllocateJson"
}

function Get-AttrHex($item, [string]$typeHex) {
  foreach ($a in $item.attrs) {
    if ($a.type -eq $typeHex) {
      return [string]$a.hex
    }
  }
  return ""
}

function Get-AttrDecodedEndpoint0016($item) {
  foreach ($a in $item.attrs) {
    if ($a.type -ne "0x0016") { continue }
    if ($null -ne $a.decoded -and $null -ne $a.decoded.endpoint) {
      $ep = [string]$a.decoded.endpoint
      if ($ep.Trim() -ne "") { return $ep.Trim() }
    }
  }
  return ""
}

$payload = Get-Content -Raw -Path $TurnAllocateJson | ConvertFrom-Json

if ($null -eq $payload.items -or $payload.items.Count -eq 0) {
  throw "No items in JSON: $TurnAllocateJson"
}

# Use the first observed 0x4024 as the session value.
$first4024 = ""
foreach ($it in $payload.items) {
  $h = Get-AttrHex $it "0x4024"
  if ($h -and $h.Trim() -ne "") { $first4024 = $h.Trim(); break }
}

# Build endpoint -> 0x4000 mapping (first seen per endpoint)
$map = @{}
foreach ($it in $payload.items) {
  $ep = Get-AttrDecodedEndpoint0016 $it
  if (-not $ep -or $ep.Trim() -eq "") { continue }
  if ($map.ContainsKey($ep)) { continue }
  $h4000 = Get-AttrHex $it "0x4000"
  if ($h4000 -and $h4000.Trim() -ne "") {
    $map[$ep] = ("hex:" + $h4000.Trim())
  }
}

if ($map.Count -eq 0) {
  throw "Could not build 0x4000 map (missing decoded 0x0016 endpoints). Re-generate the JSON with the updated helpers/parse-pktmon-stun.ps1."
}

# Output env snippet
Write-Output "# --- WhatsApp Desktop TURN Allocate-derived attributes ---"
Write-Output "# Generated from: $TurnAllocateJson"
Write-Output "QP_CALL_RELAY_TURN_AUTO_ATTR_0016=1"
Write-Output "QP_CALL_RELAY_TURN_FORCE_NO_USERNAME=1"
Write-Output "QP_CALL_RELAY_TURN_OMIT_REQUESTED_TRANSPORT=1"
Write-Output ""

if ($first4024 -and $first4024.Trim() -ne "") {
  Write-Output ("QP_CALL_RELAY_TURN_ATTR_4024=hex:" + $first4024)
} else {
  Write-Output "# QP_CALL_RELAY_TURN_ATTR_4024=hex:<fill_from_desktop_capture>"
}

if ($Format -eq 'semicolon') {
  # systemd EnvironmentFile-friendly (no quotes needed)
  # endpoint=hex:...;endpoint=hex:...
  $pairs = New-Object System.Collections.Generic.List[string]
  foreach ($k in ($map.Keys | Sort-Object)) {
    $pairs.Add(("{0}={1}" -f $k, $map[$k])) | Out-Null
  }
  $value = ($pairs -join ';')
  Write-Output ("QP_CALL_RELAY_TURN_ATTR_4000_BY_ENDPOINT=" + $value)
} elseif ($Format -eq 'json-pretty') {
  $json = ($map | ConvertTo-Json -Depth 3)
  Write-Output ("QP_CALL_RELAY_TURN_ATTR_4000_BY_ENDPOINT=" + $json)
} else {
  $json = ($map | ConvertTo-Json -Compress)
  Write-Output ("QP_CALL_RELAY_TURN_ATTR_4000_BY_ENDPOINT=" + $json)
}
Write-Output ""
Write-Output "# IMPORTANT: Do NOT set QP_CALL_RELAY_TURN_ATTR_0016 when AUTO_ATTR_0016=1"
Write-Output "# QP_CALL_RELAY_TURN_ATTR_0016="
