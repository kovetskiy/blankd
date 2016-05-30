#!/bin/bash

shopt -s expand_aliases

alias @var='tests:value'

_address=""
_process=""
_request=""

tests:clone util/program bin/

:run() {
    local port=$((10000+$RANDOM))

    _address="$(hostname):$port"

    tests:ensure $BUILD -l "$_address" -e program -o logs
    @var _process cat $(tests:get-stdout-file)
}

:response() {
    tests:put response
}

:request() {
    local uri="$1"
    shift

    tests:put-string program_args ''

    tests:eval curl -sv -A "testcase" "$@" "http://$_address$uri" '2>&1'
	tests:put request < \
		<(cat $(tests:get-stdout-file) | sed -r '/^\* /d' | sed -r 's/^> //')

	@var _request cat program_args
}
