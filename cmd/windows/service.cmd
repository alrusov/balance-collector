@echo off

if "%1" == "" (
  echo Use:
  echo %0 COMMAND
  echo where COMMAND one of: start, stop, restart, install, uninstall
  exit 1
)

call run.cmd --service %1
if %errorlevel% == 77 (
  echo ****
  echo Try to run as the Administrator
  echo ****
)
