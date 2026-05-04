#!/usr/bin/env sh
set -eu

cd backend
go test ./...
