param(
    [string]$Gateway        = "http://117.50.33.177:30080",
    [string]$StarAccount    = "star_user_1",
    [string]$StarPassword   = "StarUser123",
    [int]   $FanCount       = 10000,
    [string]$FanBaseAccount = "fan_",
    [string]$FanPassword    = "FanPass123",
    [string]$TokensFile     = "tokens.txt",
    [int]   $Vus            = 300,
    [string]$MaxDuration    = "5m"
)

<#
    End-to-end "follow once" load test runner (PowerShell).

    Scenario:
      - Create (or reuse) one "star" user.
      - Create N fan accounts and generate access tokens into tokens.txt.
      - Run k6 follow_once.js so that each fan sends exactly one follow request.

    This is mainly for "no event loss" verification: if tokens.txt has 10000
    lines, there should be roughly 10000 follow records for the star user
    after Kafka consumers finish.

    Usage (from repo root in PowerShell):

        cd .\loadtest
        .\run_follow_once.ps1 `
            -Gateway "http://117.50.33.177:30080" `
            -StarAccount "star_user_1" `
            -StarPassword "StarUser123" `
            -FanCount 10000
#>

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptDir

Write-Host "== follow-once load test config =="
Write-Host "Gateway        = $Gateway"
Write-Host "StarAccount    = $StarAccount"
Write-Host "FanCount       = $FanCount"
Write-Host "FanBaseAccount = $FanBaseAccount"
Write-Host "TokensFile     = $TokensFile"
Write-Host "VUs            = $Vus"
Write-Host "MaxDuration    = $MaxDuration"
Write-Host ""

# Check k6 is available
if (-not (Get-Command k6 -ErrorAction SilentlyContinue)) {
    Write-Error "k6 is required but not found in PATH. Please install k6 and try again."
}

Write-Host "== Step 1: create / login star user =="

$starBody = @{
    account  = $StarAccount
    password = $StarPassword
} | ConvertTo-Json

Write-Host "Registering star user (ignore errors if already exists)..."
try {
    Invoke-RestMethod -Method Post `
        -Uri "$Gateway/api/user/v1/open/users/register" `
        -ContentType "application/json" `
        -Body $starBody | Out-Null
} catch {
    Write-Warning "Register star user may have failed or already exists: $_"
}

Write-Host "Logging in star user..."
$starLogin = Invoke-RestMethod -Method Post `
    -Uri "$Gateway/api/user/v1/open/users/login" `
    -ContentType "application/json" `
    -Body $starBody

$StarUserUuid   = $starLogin.data.user_uuid
$StarAccessToken = $starLogin.data.access_token

if ([string]::IsNullOrWhiteSpace($StarUserUuid)) {
    Write-Error "Failed to obtain star user UUID from login response."
}

Write-Host "Star user UUID: $StarUserUuid"
Write-Host ""

Write-Host "== Step 2: generate fan tokens ($FanCount accounts) into $TokensFile =="

if (Test-Path $TokensFile) {
    Write-Host "Removing existing $TokensFile"
    Remove-Item -Force $TokensFile
}

& .\prepare_fans.ps1 `
    -Gateway $Gateway `
    -FanCount $FanCount `
    -BaseAccount $FanBaseAccount `
    -Password $FanPassword `
    -OutputPath $TokensFile

if (-not (Test-Path $TokensFile)) {
    Write-Error "Tokens file $TokensFile was not created."
}

$fanTokenCount = (Get-Content $TokensFile | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }).Count
Write-Host "Fan tokens written to $TokensFile ($fanTokenCount tokens)."
if ($fanTokenCount -le 0) {
    Write-Error "No fan tokens found in $TokensFile"
}
Write-Host ""

Write-Host "== Step 3: run k6 follow_once.js (each fan follows once) =="

$env:GATEWAY      = $Gateway
$env:TARGET_UUID  = $StarUserUuid
$env:VUS          = "$Vus"
$env:MAX_DURATION = $MaxDuration

Write-Host "Running k6 with:"
Write-Host "  GATEWAY      = $env:GATEWAY"
Write-Host "  TARGET_UUID  = $env:TARGET_UUID"
Write-Host "  VUS          = $env:VUS"
Write-Host "  MAX_DURATION = $env:MAX_DURATION"
Write-Host ""

k6 run "$scriptDir/follow_once.js"

Write-Host ""
Write-Host "Done. Now you can verify follower count in DB or via relation API."

