@echo off

call %~dp0\env.cmd
%APP% --config config\%APP%.toml %*
