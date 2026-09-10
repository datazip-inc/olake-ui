#!/bin/sh
set -eu

: "${POSTGRES_SEEDS:?ERROR: POSTGRES_SEEDS environment variable is required}"
: "${POSTGRES_USER:?ERROR: POSTGRES_USER environment variable is required}"

sqltool() {
  temporal-sql-tool \
    --plugin postgres12 \
    --ep "${POSTGRES_SEEDS}" \
    -u "${POSTGRES_USER}" \
    -p "${DB_PORT:-5432}" \
    "$@"
}

echo 'Starting PostgreSQL schema setup...'
echo 'Waiting for PostgreSQL port to be available...'
nc -z -w 10 "${POSTGRES_SEEDS}" "${DB_PORT:-5432}"
echo 'PostgreSQL port is available'

# create and setup-schema error if the DB/schema already exists (existing OLake volumes).
# update-schema applies pending versions and must succeed.
echo 'Setting up temporal database...'
sqltool --db temporal create \
  || echo '  (create skipped: database already exists)'
sqltool --db temporal setup-schema -v 0.0 \
  || echo '  (setup-schema skipped: already initialized)'
sqltool --db temporal update-schema \
  -d /etc/temporal/schema/postgresql/v12/temporal/versioned

echo 'Setting up temporal_visibility database...'
sqltool --db temporal_visibility create \
  || echo '  (create skipped: database already exists)'
sqltool --db temporal_visibility setup-schema -v 0.0 \
  || echo '  (setup-schema skipped: already initialized)'
sqltool --db temporal_visibility update-schema \
  -d /etc/temporal/schema/postgresql/v12/visibility/versioned

echo 'PostgreSQL schema setup complete'
