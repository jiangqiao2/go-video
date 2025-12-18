param(
    [string]$Gateway     = "http://localhost:8000",
    # 默认 2000 个粉丝账号，避免一上来就压 1 万；如需更多可通过 -FanCount 覆盖
    [int]   $FanCount    = 2000,
    [string]$BaseAccount = "fan_",
    [string]$Password    = "FanPass123",
    [string]$OutputPath  = "tokens.txt",
    # 运行批次后缀，避免 account 重复；默认使用当前时间戳，例如 20251218T184500
    [string]$RunSuffix   = ""
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

if ([string]::IsNullOrWhiteSpace($RunSuffix)) {
    $RunSuffix = Get-Date -Format "yyyyMMddTHHmmss"
}

Write-Host "Gateway:     $Gateway"
Write-Host "FanCount:    $FanCount"
Write-Host "BaseAccount: $BaseAccount"
Write-Host "RunSuffix:   $RunSuffix"
Write-Host "Password:    $Password"
Write-Host "OutputPath:  $OutputPath"

if (Test-Path $OutputPath) {
    Write-Host "Removing existing $OutputPath"
    Remove-Item -Force $OutputPath
}

for ($i = 1; $i -le $FanCount; $i++) {
    $account = "{0}{1}_{2:D4}" -f $BaseAccount, $RunSuffix, $i
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
