@echo off

setlocal enableextensions enabledelayedexpansion

if exist vendor\github.com (
  echo Please delete the vendor\github.com directory
  echo See https://github.com/golang/go/issues/19000 for details
  exit
)

call cmd\windows\env.cmd

rem set CGO_ENABLED=0
rem set EXTRA_LD=-extldflags -static
set CGO_ENABLED=1
set EXTRA_LD=

for /f %%d in ('wmic path win32_utctime get /format:list ^| findstr "="') do (
  for /f "tokens=1,2 delims==" %%a in ("%%d") do (
    if "%%a" == "Year"   set Y=%%b
    if "%%a" == "Month"  set N=%%b
    if "%%a" == "Day"    set D=%%b
    if "%%a" == "Hour"   set H=%%b
    if "%%a" == "Minute" set M=%%b
    if "%%a" == "Second" set S=%%b
  )
)
if %N% lss 10 set N=0%N%
if %D% lss 10 set D=0%D%
if %H% lss 10 set H=0%H%
if %M% lss 10 set M=0%M%
if %S% lss 10 set S=0%S%

set COPYRIGHT=(C)_Alexey_Rusov_(rolic402@mail.ru),_2017-%Y%
set BUILD_TIME=%Y%-%N%-%D%_%H%:%M%:%S%

set BUILD=
if exist BUILD_NUMBER (
  for /f %%i in ('type BUILD_NUMBER') do set "BUILD=%%i"
)
if "%BUILD%"=="" (set BUILD=1)

for /f %%i in ('type VERSION') do set "VERSION=%%i"
set VERSION=%VERSION%.%BUILD%

set TAGS=
if exist TAGS (
  for /f %%i in ('type TAGS') do (
    if not "!TAGS!"=="" set TAGS=!TAGS!_
    set TAGS=!TAGS!%%i
  )
)

go build ^
  -o cmd\windows\%APP%.exe ^
  --ldflags "%EXTRA_LD% -X github.com/alrusov/misc.appVersion=%VERSION% -X github.com/alrusov/misc.appTags=%TAGS% -X github.com/alrusov/misc.buildTime=%BUILD_TIME% -X github.com/alrusov/misc.copyright=%COPYRIGHT%"

set /a BUILD+=1
echo %BUILD% >BUILD_NUMBER
