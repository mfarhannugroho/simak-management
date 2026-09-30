# ─── PT INKA SIMAK Database Backup Script (PowerShell) ──────────────────────────
# Parametric database backup using pg_dump with compression

param (
    [string]$BackupDir = "D:\backups\simak",
    [string]$DbHost = "localhost",
    [string]$DbPort = "5432",
    [string]$DbName = "simak_db",
    [string]$DbUser = "simak_user",
    [string]$DbPassword = "simak_password",
    [int]$RetentionDays = 14
)

$ErrorActionPreference = "Stop"

if (!(Test-Path $BackupDir)) {
    New-Item -ItemType Directory -Force -Path $BackupDir | Out-Null
}

$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$targetFile = Join-Path $BackupDir "simak_backup_$timestamp.dump"

Write-Host "[$((Get-Date).ToString('yyyy-MM-dd HH:mm:ss'))] Starting backup of $DbName..."

$env:PGPASSWORD = $DbPassword
& pg_dump -h $DbHost -p $DbPort -U $DbUser -d $DbName -F c -b -f $targetFile

if ($LASTEXITCODE -eq 0) {
    $size = (Get-Item $targetFile).Length / 1MB
    Write-Host "[$((Get-Date).ToString('yyyy-MM-dd HH:mm:ss'))] Backup created: $targetFile ($([math]::Round($size, 2)) MB)"
} else {
    Write-Error "Backup failed with exit code $LASTEXITCODE"
}

# Cleanup older backups
$cutoff = (Get-Date).AddDays(-$RetentionDays)
Get-ChildItem -Path $BackupDir -Filter "simak_backup_*.dump" | Where-Object { $_.LastWriteTime -lt $cutoff } | ForEach-Object {
    Write-Host "Removing expired backup: $($_.FullName)"
    Remove-Item $_.FullName -Force
}
