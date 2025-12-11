param(
    [string]$Gateway    = "http://localhost:8000",
    [int]   $FanCount   = 200,
    [string]$BaseAccount = "fan_",
    [string]$Password   = "FanPass123",
    [string]$OutputPath = "tokens.txt"
)

<#
    Bulk-create "fan" accounts and generate access tokens for them via the
    gateway open APIs. The script:

    - Calls POST /api/user/v1/open/users/register for each fan account
    - Calls POST /api/user/v1/open/users/login to obtain an access token
    - Writes each data.access_token into a tokens.txt file (one per line)

    Usage example from repository root:

        cd .\loadtest
        .\prepare_fans.ps1 `
            -Gateway "http://your-gateway-host:30080" `
            -FanCount 200 `
            -BaseAccount "fan_" `
            -Password "FanPass123" `
            -OutputPath "tokens.txt"
#>

Write-Host "Gateway:     $Gateway"
Write-Host "FanCount:    $FanCount"
Write-Host "BaseAccount: $BaseAccount"
Write-Host "Password:    $Password"
Write-Host "OutputPath:  $OutputPath"

if (Test-Path $OutputPath) {
    Write-Host "Removing existing $OutputPath"
    Remove-Item -Force $OutputPath
}

for ($i = 1; $i -le $FanCount; $i++) {
    $account = "{0}{1:D4}" -f $BaseAccount, $i
    $body    = @{ account = $account; password = $Password } | ConvertTo-Json

    Write-Host "Processing account: $account"

    # Register fan account (ignore failures if already exists)
    try {
        Invoke-RestMethod -Method Post `
            -Uri "$Gateway/api/user/v1/open/users/register" `
            -ContentType "application/json" `
            -Body $body | Out-Null
    } catch {
        Write-Warning "Register failed (possibly exists) for $account: $_"
    }

    # Login to obtain access token
    try {
        $loginResp = Invoke-RestMethod -Method Post `
            -Uri "$Gateway/api/user/v1/open/users/login" `
            -ContentType "application/json" `
            -Body $body

        $token = $loginResp.data.access_token

        if ([string]::IsNullOrWhiteSpace($token)) {
            Write-Warning "No access_token in response for $account"
        } else {
            Add-Content -Path $OutputPath -Value $token
        }
    } catch {
        Write-Warning "Login failed for $account: $_"
    }
}

Write-Host "Done. Tokens written to $OutputPath"

