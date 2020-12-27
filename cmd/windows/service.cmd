@echo off

if "%1" == "" (
  echo Use:
  echo %0 COMMAND
  echo where COMMAND one of: start, stop, restart, install, uninstall
  exit 1
)

call %~dp0\run.cmd --service %1
if %errorlevel% == 211 (
  echo ****
  echo Try to run as administrator
  echo ****
)
