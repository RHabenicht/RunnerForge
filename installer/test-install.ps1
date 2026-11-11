$ErrorActionPreference = 'Stop'
$root = Join-Path $env:RUNNER_TEMP 'RunnerForge install test'
$setup = Join-Path $PSScriptRoot '..\dist\RunnerForge-Setup-0.2.0.exe'
$p = Start-Process $setup -ArgumentList @('/VERYSILENT', '/SUPPRESSMSGBOXES', '/NORESTART', ('/DIR="' + $root + '"')) -Wait -PassThru
if ($p.ExitCode -ne 0) { throw "Installer failed: $($p.ExitCode)" }
$exe = Join-Path $root 'RunnerForge.exe'
if (!(Test-Path $exe)) { throw 'Installed executable missing' }
$p = Start-Process $exe -PassThru
try {
  for ($i=0; $i -lt 30; $i++) {
    Start-Sleep -Milliseconds 250
    $p.Refresh()
    if ($p.HasExited) { throw 'Installed application exited early' }
    if ($p.MainWindowHandle -ne 0) { break }
  }
  if ($p.MainWindowHandle -eq 0) { throw 'Installed application has no window' }
  if (!$p.CloseMainWindow() -or !$p.WaitForExit(10000)) { throw 'Installed application did not close normally' }
} finally { if (!$p.HasExited) { Stop-Process -Id $p.Id -Force } }
# A save created by the user must survive uninstallation, even in the app folder.
$save = Join-Path $root 'preserve-character.json'
Set-Content $save '{"name":"Preserve me"}'
$uninstall = Join-Path $root 'unins000.exe'
$p = Start-Process $uninstall -ArgumentList @('/VERYSILENT','/SUPPRESSMSGBOXES','/NORESTART') -Wait -PassThru
if ($p.ExitCode -ne 0) { throw "Uninstaller failed: $($p.ExitCode)" }
if (Test-Path $exe) { throw 'Executable left behind after uninstall' }
if (!(Test-Path $save)) { throw 'Uninstall deleted a user save' }
