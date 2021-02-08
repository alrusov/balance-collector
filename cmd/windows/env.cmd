@echo off

pushd ..\..\
set APP=%cd%
cd ..
set APP=!APP:%cd%\=!
popd

for /f "delims=" %%x in (..\env) do (set "%%x")
