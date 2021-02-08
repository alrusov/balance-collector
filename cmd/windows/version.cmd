@echo off

setlocal enabledelayedexpansion

call env.cmd

..\..\%APP% --version
