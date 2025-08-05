@echo off

psql.exe -U postgres -c "DROP DATABASE IF EXISTS exploredb WITH (FORCE);"
psql.exe -U postgres -c "CREATE DATABASE exploredb;"
psql.exe -U postgres -d exploredb -f ./init.sql
