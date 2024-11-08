#!/bin/bash


gofumpt -w ./..
pg_format db/*/*.sql
