#!/bin/sh
set -e

sed -i "s/{{CFG_ETCDPOINT}}/$CFG_ETCDPOINT/g" conf/app.toml
sed -i "s/{{CFG_AUTHREDISHOST}}/$CFG_AUTHREDISHOST/g" conf/app.toml
sed -i "s/{{CFG_LOGINREDISHOST}}/$CFG_LOGINREDISHOST/g" conf/app.toml

# if no command, run default
if [ ! -n "$1" ]; then
	exec ./auth allinone
fi

exec "$@"