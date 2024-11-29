#!/bin/bash


gofumpt -w ./..
pg_format db/*/*.sql
sqlfluff lint db/*/*.sql --dialect postgres
