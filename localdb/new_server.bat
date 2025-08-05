@echo off

if not exist data mkdir data

initdb.exe -D ".\data" -U postgres
pg_ctl.exe start -D ".\data" -l log.txt