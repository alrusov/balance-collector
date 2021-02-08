@echo off

setlocal enabledelayedexpansion

call env.cmd

..\..\%APP% --config config\%APP%.toml %*
