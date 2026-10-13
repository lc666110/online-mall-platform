$storeRoot=Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
$runtimeRoot=Join-Path $storeRoot 'runtime'
$expectedRoot=[IO.Path]::GetFullPath($storeRoot).TrimEnd('\')+'\'
foreach($name in @('preview-frontend','preview-admin','frontend','admin','api','redis','mysql')){
 $pidFile=Join-Path $runtimeRoot "$name.pid"
 if(!(Test-Path $pidFile)){continue}
 $taskPid=[int](Get-Content $pidFile)
 $process=Get-CimInstance Win32_Process -Filter "ProcessId=$taskPid" -ErrorAction SilentlyContinue
 if(!$process){continue}
 $exe=$process.ExecutablePath
 $command=$process.CommandLine
 if(($exe -and [IO.Path]::GetFullPath($exe).StartsWith($expectedRoot,[StringComparison]::OrdinalIgnoreCase)) -or ($exe -like '*node.exe' -and $command -like "*$storeRoot*")){
  if($name -eq 'mysql'){
   $privateConfig=Get-Content -LiteralPath (Join-Path $runtimeRoot 'local-secrets.json') | ConvertFrom-Json
   $previousPassword=$env:MYSQL_PWD
   try{$env:MYSQL_PWD=$privateConfig.db_password;& (Join-Path (Split-Path $exe) 'mysqladmin.exe') --no-defaults -h 127.0.0.1 -P 13306 -u root shutdown;if($LASTEXITCODE){throw 'MySQL graceful shutdown failed; retained process'}}finally{$env:MYSQL_PWD=$previousPassword}
  }elseif($name -eq 'redis'){
   $privateConfig=Get-Content -LiteralPath (Join-Path $runtimeRoot 'local-secrets.json') | ConvertFrom-Json
   $previousPassword=$env:REDISCLI_AUTH
   try{$env:REDISCLI_AUTH=$privateConfig.redis_password;& (Join-Path (Split-Path $exe) 'redis-cli.exe') -h 127.0.0.1 -p 16379 shutdown save;if($LASTEXITCODE){throw 'Redis graceful shutdown failed; retained process'}}finally{$env:REDISCLI_AUTH=$previousPassword}
  }else{Stop-Process -Id $taskPid -ErrorAction Stop}
  Write-Output "Stopped $name"
 }else{Write-Warning "Retained PID ${taskPid}: executable does not belong to this project"}
}
