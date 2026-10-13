param([switch]$Development)
$ErrorActionPreference='Stop'
$projectRoot=Split-Path $PSScriptRoot -Parent
$storeRoot=Split-Path $projectRoot -Parent
$runtimeRoot=Join-Path $storeRoot 'runtime'
$secretFile=Join-Path $runtimeRoot 'local-secrets.json'
if(!(Test-Path $secretFile)){throw 'Local secrets missing. Use Docker quickstart or provision MySQL/Redis first.'}
$localSecrets=Get-Content -LiteralPath $secretFile | ConvertFrom-Json
$env:GOPATH=Join-Path $storeRoot 'cache/gopath';$env:GOCACHE=Join-Path $storeRoot 'cache/go-build';$env:GOTOOLCHAIN='local'
$env:npm_config_cache=Join-Path $storeRoot 'cache/npm';$env:PLAYWRIGHT_BROWSERS_PATH=Join-Path $storeRoot 'tools/playwright'
$env:MYSQL_DSN="mall:$($localSecrets.db_password)@tcp(127.0.0.1:13306)/mall?charset=utf8mb4&parseTime=true&loc=UTC"
$env:REDIS_ADDR='127.0.0.1:16379';$env:REDIS_PASSWORD=$localSecrets.redis_password;$env:JWT_SECRET=$localSecrets.jwt_secret
$env:ADMIN_EMAIL=$localSecrets.admin_email;$env:ADMIN_PASSWORD=$localSecrets.admin_password;$env:GIN_MODE='release'
$goBinary=Join-Path $storeRoot 'tools/go/go/bin/go.exe'
$mysqlBinary=Get-ChildItem (Join-Path $storeRoot 'tools/mysql') -Recurse -Filter mysqld.exe | Select-Object -First 1 -ExpandProperty FullName
$redisBinary=Get-ChildItem (Join-Path $storeRoot 'tools/redis') -Recurse -Filter redis-server.exe | Select-Object -First 1 -ExpandProperty FullName
function Start-LocalProcess($name,$exe,$arguments,$directory){
 $process=Start-Process -FilePath $exe -ArgumentList $arguments -WorkingDirectory $directory -WindowStyle Hidden -PassThru -RedirectStandardOutput (Join-Path $runtimeRoot "$name-out.log") -RedirectStandardError (Join-Path $runtimeRoot "$name-err.log")
 $process.Id | Set-Content (Join-Path $runtimeRoot "$name.pid")
}
function Has-Port($port){return [bool](Get-NetTCPConnection -State Listen -LocalPort $port -ErrorAction SilentlyContinue)}
if(!(Has-Port 13306)){Start-LocalProcess 'mysql' $mysqlBinary @('--no-defaults',"--basedir=$((Split-Path (Split-Path $mysqlBinary)))","--datadir=$(Join-Path $runtimeRoot 'mysql')",'--port=13306','--bind-address=127.0.0.1','--mysqlx=OFF','--console') $runtimeRoot}
if(!(Has-Port 16379)){Start-LocalProcess 'redis' $redisBinary @('redis.conf') $runtimeRoot}
$deadline=(Get-Date).AddSeconds(30);while((!(Has-Port 13306) -or !(Has-Port 16379)) -and (Get-Date)-lt $deadline){Start-Sleep -Milliseconds 250}
if(!(Has-Port 13306) -or !(Has-Port 16379)){throw 'MySQL/Redis did not start; inspect runtime logs'}
$oldMysqlPassword=$env:MYSQL_PWD;$oldRedisPassword=$env:REDISCLI_AUTH
try{
 $env:MYSQL_PWD=$localSecrets.db_password;$env:REDISCLI_AUTH=$localSecrets.redis_password
 $mysqlClient=Join-Path (Split-Path $mysqlBinary) 'mysql.exe';$redisClient=Join-Path (Split-Path $redisBinary) 'redis-cli.exe'
 $deadline=(Get-Date).AddSeconds(30)
 do{
  & $mysqlClient --no-defaults -h 127.0.0.1 -P 13306 -u mall -e 'SELECT 1' mall *> $null;$mysqlReady=$LASTEXITCODE -eq 0
  $redisReply=& $redisClient -h 127.0.0.1 -p 16379 ping 2>$null;$redisReady=$redisReply -eq 'PONG'
  if($mysqlReady -and $redisReady){break};Start-Sleep -Milliseconds 250
 }while((Get-Date)-lt $deadline)
 if(!$mysqlReady -or !$redisReady){throw 'Dependencies are listening but not ready; inspect runtime logs'}
}finally{$env:MYSQL_PWD=$oldMysqlPassword;$env:REDISCLI_AUTH=$oldRedisPassword}

if(!(Has-Port 8080)){Push-Location (Join-Path $projectRoot 'server');try{& $goBinary build -o (Join-Path $runtimeRoot 'mall-server.exe') .;if($LASTEXITCODE){throw 'Backend build failed'}}finally{Pop-Location};Start-LocalProcess 'api' (Join-Path $runtimeRoot 'mall-server.exe') @() (Join-Path $projectRoot 'server')}
$nodeBinary=(Get-Command node.exe).Source
if($Development){foreach($app in @('frontend','admin')){$port=if($app-eq'frontend'){5173}else{5174};if(!(Has-Port $port)){Start-LocalProcess $app $nodeBinary @((Join-Path $projectRoot 'node_modules/vite/bin/vite.js'),'--host','127.0.0.1','--port',"$port") (Join-Path $projectRoot $app)}}}
else{foreach($app in @('frontend','admin')){$port=if($app-eq'frontend'){8088}else{8089};if(!(Test-Path (Join-Path $projectRoot "$app/dist/index.html"))){throw 'Run npm ci and npm run build first'};if(!(Has-Port $port)){Start-LocalProcess "preview-$app" $nodeBinary @((Join-Path $PSScriptRoot 'preview.mjs'),$app,"$port") $projectRoot}}}
foreach($entry in @(@('mysql',$mysqlBinary),@('redis',$redisBinary))){
 $ownedProcess=Get-CimInstance Win32_Process | Where-Object {$_.ExecutablePath -eq $entry[1]} | Select-Object -First 1
 if($ownedProcess){$ownedProcess.ProcessId | Set-Content (Join-Path $runtimeRoot "$($entry[0]).pid")}
}
Write-Output 'Started locally. Storefront: http://127.0.0.1:8088 ; Admin: http://127.0.0.1:8089 (development: 5173/5174).'
