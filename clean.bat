@echo off

setlocal

if exist clean.bat goto ok
echo clean.bat must be run from its folder
goto end

:ok

del /q bin\*
del /q pkg\windows_386\*

:end
echo finished
