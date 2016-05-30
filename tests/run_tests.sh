#!/bin/bash

set -e

cd $(readlink -f  $(dirname $0))

export BUILD=$(mktemp -u -t buildXXXXXX)
go build -o $BUILD ../
trap "rm $BUILD" EXIT

./lib/tests.sh -d testcases -s util/setup.sh -t util/teardown.sh ${@:--Aa}
