@echo off

pg_ctl.exe stop -D ".\data"
rmdir /s /q ".\data"