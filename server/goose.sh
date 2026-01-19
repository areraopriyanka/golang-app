#!/usr/bin/env bash

set -eou pipefail

user=${DATABASE_DBUSER:-middleware}
password=${DATABASE_DBPASSWORD:-Dreamfi#1234}
dbname=${DATABASE_DBNAME:-middleware}
host=${DATABASE_HOST:-localhost}
port=${DATABASE_PORT:-5432}
sslmode=${DATABASE_SSLMODE:-disable}

# string must be kept in sync with pkg/db/db.go
export GOOSE_DBSTRING="user=${user} password=${password} host=${host} port=${port} dbname=${dbname} sslmode=${sslmode}"
export GOOSE_DRIVER=postgres
goose -dir pkg/db/scripts/ $*
