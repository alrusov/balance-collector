@echo off

call env.cmd
..\..\%APP% --config config\%APP%.toml %*
