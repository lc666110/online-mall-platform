$ErrorActionPreference='Stop'
$projectRoot=Split-Path $PSScriptRoot -Parent
$storeRoot=Split-Path $projectRoot -Parent
$backupRoot=Join-Path $storeRoot 'backups'
New-Item -ItemType Directory -Force $backupRoot | Out-Null
$stamp=Get-Date -Format 'yyyyMMdd-HHmmss'
Push-Location $projectRoot
try{
 git bundle create (Join-Path $backupRoot "source-$stamp.bundle") --all
 if($LASTEXITCODE){throw 'Git bundle failed; commit the source first'}
 $sourceArchive=Join-Path $backupRoot "source-$stamp.zip"
 git archive --format=zip "--output=$sourceArchive" HEAD
 if($LASTEXITCODE){throw 'Source archive failed'}
 $privateConfig=Get-Content -LiteralPath (Join-Path $storeRoot 'runtime/local-secrets.json') | ConvertFrom-Json
 $dump=Get-ChildItem (Join-Path $storeRoot 'tools/mysql') -Recurse -Filter mysqldump.exe | Select-Object -First 1 -ExpandProperty FullName
 $dataFile=Join-Path $backupRoot "mall-$stamp.sql"
 $previousPassword=$env:MYSQL_PWD
 try{
  $env:MYSQL_PWD=$privateConfig.db_password
  & $dump --no-defaults -h 127.0.0.1 -P 13306 -u mall --single-transaction --no-tablespaces --set-gtid-purged=OFF --result-file=$dataFile mall
  if($LASTEXITCODE){throw 'Database backup failed'}
 }finally{$env:MYSQL_PWD=$previousPassword}
 icacls $dataFile /inheritance:r /grant:r Administrator:F SYSTEM:F | Out-Null
 Get-ChildItem $backupRoot -File | Where-Object {$_.Name -like "*$stamp*"} | ForEach-Object {Get-FileHash -LiteralPath $_.FullName -Algorithm SHA256} | Format-Table Path,Hash
}finally{Pop-Location}
