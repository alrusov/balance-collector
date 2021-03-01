@echo off

setlocal enableextensions enabledelayedexpansion

rem #-----------------------------------------------------------------------------------------------------------------------------------------------------#

set APP=%cd%
pushd ..\
set APP=!APP:%cd%\=!
popd

if exist BUILD_NUMBER (
	set BUILD_NUMBER_FILE=BUILD_NUMBER
) else (
	set BUILD_NUMBER_FILE=..\go-builds\v\%APP%
)

set BUILD_NUMBER=
if exist %BUILD_NUMBER_FILE% (
	for /f %%i in ('type %BUILD_NUMBER_FILE%') do set BUILD_NUMBER=%%i
)
if "%BUILD_NUMBER%"=="" set BUILD_NUMBER=1

if exist WITHOUT_GLOBAL_TAGS (
	set TAGS_FILE=TAGS
) else (
	set TAGS_FILE=..\go-builds\TAGS
)

rem #-----------------------------------------------------------------------------------------------------------------------------------------------------#

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

rem #-----------------------------------------------------------------------------------------------------------------------------------------------------#

for /f %%i in ('type VERSION') do set VERSION=%%i.%BUILD_NUMBER%

rem #-----------------------------------------------------------------------------------------------------------------------------------------------------#

set TAGS=
if exist %TAGS_FILE% (
	for /f %%i in ('type %TAGS_FILE%') do (
		if not "!TAGS!"=="" set TAGS=!TAGS!_
		set TAGS=!TAGS!%%i
	)
)

rem #-----------------------------------------------------------------------------------------------------------------------------------------------------#

set GO_FLAGS=

if exist STATIC (
	set CGO_ENABLED=0 
	set EXTRA_LD=-extldflags -static
) else (
	set CGO_ENABLED=1
	set EXTRA_LD=
)

rem #-----------------------------------------------------------------------------------------------------------------------------------------------------#

git pull

go build -o %APP%.exe ^
	%GO_FLAGS% ^
	--ldflags "%EXTRA_LD% -X github.com/alrusov/misc.appVersion=%VERSION% -X github.com/alrusov/misc.appTags=%TAGS% -X github.com/alrusov/misc.buildTime=%BUILD_TIME% -X github.com/alrusov/misc.copyright=%COPYRIGHT%"

set /a BUILD_NUMBER+=1
echo %BUILD_NUMBER% >%BUILD_NUMBER_FILE%

rem #-----------------------------------------------------------------------------------------------------------------------------------------------------#
