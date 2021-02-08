@echo off

setlocal enableextensions enabledelayedexpansion

set APP=%cd%
pushd ..\
set APP=!APP:%cd%\=!
popd

set BUILD_NUMBER_FILE=BUILD_NUMBER
set TAGS_FILE=TAGS

if exist STATIC (
  set CGO_ENABLED=0 
  set EXTRA_LD=-extldflags -static
) else (
  set CGO_ENABLED=1
  set EXTRA_LD=
)

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

for /f "tokens=*" %%i in ('type COPYRIGHT') do set COPYRIGHT=%%i%Y%
set COPYRIGHT=%COPYRIGHT: =_%

set BUILD_TIME=%Y%-%N%-%D%_%H%:%M%:%S%

set BUILD=
if exist %BUILD_NUMBER_FILE% (
  for /f %%i in ('type %BUILD_NUMBER_FILE%') do set BUILD=%%i
)
if "%BUILD%"=="" set BUILD=1

for /f %%i in ('type VERSION') do set VERSION=%%i.%BUILD%

set TAGS=
if exist %TAGS_FILE% (
  for /f %%i in ('type %TAGS_FILE%') do (
    if not "!TAGS!"=="" set TAGS=!TAGS!_
    set TAGS=!TAGS!%%i
  )
)

go build --ldflags "%EXTRA_LD% -X github.com/alrusov/misc.appVersion=%VERSION% -X github.com/alrusov/misc.appTags=%TAGS% -X github.com/alrusov/misc.buildTime=%BUILD_TIME% -X github.com/alrusov/misc.copyright=%COPYRIGHT%"

set /a BUILD+=1
echo %BUILD% >%BUILD_NUMBER_FILE%
