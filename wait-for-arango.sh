#!/bin/sh
# wait-for-arango.sh

set -e
cmd="$@"

until curl http://arango:8529/_api/version; do
>&2 echo "\nArango is unavailable - sleeping"
sleep 1
done

>&2 echo "\nArango is up - executing command"
exec $cmd
