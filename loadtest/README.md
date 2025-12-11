go-video Load Testing
=====================

This folder contains scripts to stress-test the "user follow" scenario via the
Kong gateway, from a local Windows machine (or any machine that can reach the
gateway).

The main idea: create one "star" user, create many "fan" accounts, log them in
to get JWT access tokens, then continuously call:

- `POST /api/user/v1/inner/relation/follow`

through the gateway for 1 hour at a target QPS.


Files
-----

- `follow_stress.js`  
  k6 script that sends follow requests from gateway to `user-service`.

- `prepare_fans.ps1`  
  PowerShell helper to bulk-create "fan" accounts and generate `tokens.txt`
  with their access tokens.

- `.gitignore`  
  Ignores `tokens.txt` and other local artifacts.


Prerequisites
-------------

1. You can reach the gateway from your local machine, for example:

   - `http://<gateway-host>:8000` (docker-compose)
   - or `http://<node-ip>:30080` / `https://api.xxx.com` (K8s / production)

2. k6 installed on your local machine.

   On Windows this is typically:

   - Install via package manager (choco/scoop) or download the k6 binary.
   - You should be able to run `k6 version` in PowerShell.

3. PowerShell 5+ (already available on modern Windows).


Step 1: Create the "star" user and get its UUID
-----------------------------------------------

You need one target user that everyone will follow. You can create it through
the user open APIs on the gateway.

In PowerShell:

```powershell
$Gateway = "http://your-gateway-host:30080"   # TODO: change to real gateway

# 1) Register star user (ignore error if account already exists)
$body = @{ account = "star_user_1"; password = "StarUser123" } | ConvertTo-Json
try {
    Invoke-RestMethod -Method Post `
        -Uri "$Gateway/api/user/v1/open/users/register" `
        -ContentType "application/json" `
        -Body $body | Out-Null
} catch {
    Write-Warning "Register star user may have failed or already exists: $_"
}

# 2) Login and get user_uuid and access_token
$login = Invoke-RestMethod -Method Post `
    -Uri "$Gateway/api/user/v1/open/users/login" `
    -ContentType "application/json" `
    -Body $body

$StarUserUuid   = $login.data.user_uuid
$StarAccessToken = $login.data.access_token

Write-Host "Star user UUID: $StarUserUuid"
Write-Host "Star access token: $StarAccessToken"
```

You will use `$StarUserUuid` as `TARGET_UUID` when running k6.


Step 2: Generate fan tokens (tokens.txt)
----------------------------------------

Use the helper script to create many fan accounts and log them in to generate a
`tokens.txt` file. Each line in `tokens.txt` will be one access token.

From the repository root (`D:\project\go-video`):

```powershell
cd .\loadtest

.\prepare_fans.ps1 `
    -Gateway "http://your-gateway-host:30080" `  # change to real gateway
    -FanCount 200 `                              # how many fan accounts
    -BaseAccount "fan_" `                        # prefix for fan accounts
    -Password "FanPass123" `                     # password used for all fans
    -OutputPath "tokens.txt"
```

Notes:

- The script:
  - calls `/api/user/v1/open/users/register` to create fan accounts
  - then `/api/user/v1/open/users/login` to get tokens
  - writes each `data.access_token` into `tokens.txt` (one per line)
- You can safely re-run it; duplicate accounts are allowed (register may fail,
  login will still work if the account already exists).


Step 3: Run the follow load test with k6
----------------------------------------

From the `loadtest` folder:

```powershell
cd .\loadtest

$env:GATEWAY     = "http://your-gateway-host:30080"  # same as above
$env:TARGET_UUID = "<StarUserUuid from step 1>"

# Example: 2000 QPS for 1 hour
k6 run .\follow_stress.js
```

The `follow_stress.js` script reads:

- `GATEWAY` from env: base URL of the gateway
- `TARGET_UUID` from env: the star user UUID
- `tokens.txt` from the current folder: fan access tokens

You can adjust the QPS and duration directly inside `follow_stress.js`:

- `rate`: target requests per second (QPS)
- `duration`: total test duration (e.g. `"30m"`, `"1h"`)
- `preAllocatedVUs` / `maxVUs`: concurrent virtual user capacity


Step 4: Suggested ramp-up strategy
----------------------------------

To avoid crashing everything immediately, ramp up in stages:

1. Start small:

   - `rate: 500`, `duration: "5m"`

2. If stable, increase:

   - `rate: 1000`, `duration: "10m"`

3. For heavier stress:

   - `rate: 2000–3000`, `duration: "30m"–"1h"`

On your 8-core local machine, k6 should handle a few thousand QPS as a client.
The real bottleneck will usually be the backend (user-service, MySQL, Redis,
network, etc.).


Step 5: What to watch while the test runs
-----------------------------------------

During the run, monitor:

- k6 output:
  - HTTP error rate (5xx / 4xx not caused by bad parameters)
  - p95 / p99 latency
- Gateway:
  - CPU, memory, and latency
  - Upstream error ratio
- user-service:
  - CPU, memory, goroutines, GC
  - HTTP 5xx count and latency
- MySQL / Redis:
  - QPS, slow queries, locks, connection usage

When errors or latency explode and do not recover after lowering QPS, you have
effectively reached the capacity limit for this scenario.

