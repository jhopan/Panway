# Start Panway as a Windows process that is not tied to the calling shell.
#
# A `go run` started from an agent/session shell is killed when that session
# ends, and Start-Process children can still share the caller's job object.
# Win32_Process.Create goes through the WMI service, so the server keeps
# running until someone stops it. Output goes to logs\panway-<date>.log via
# the app logger; the console stream is not redirected.
#
# Usage: powershell -NoProfile -File scripts/start-server.ps1
param(
    [string]$GoExe = 'C:\Program Files\Go\bin\go.exe'
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$logDir = Join-Path $root 'logs'
New-Item -ItemType Directory -Force -Path $logDir | Out-Null

$commandLine = '"' + $GoExe + '" run . --log-dir logs'
$result = Invoke-CimMethod -ClassName Win32_Process -MethodName Create -Arguments @{
    CommandLine      = $commandLine
    CurrentDirectory = $root
}
if ($result.ReturnValue -ne 0) {
    throw "failed to start Panway: Win32_Process.Create returned $($result.ReturnValue)"
}
Write-Output "Panway start requested, pid $($result.ProcessId)"
Write-Output "Health: http://127.0.0.1:3110/healthz"
Write-Output "Logs:   $logDir"