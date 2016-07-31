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

    tests:ensure $BUILD -l "$_address" -e program "$@"
    @var _process cat $(tests:get-stdout-file)
}

:response() {
    tests:put response
}

:request() {
    local uri="$1"
    shift

    local tls=false
    if [[ "$#" -ge "1" ]]; then
        if [[ "$1" == "--tls" ]]; then
            tls=true
            shift
        fi
    fi

    tests:put-string program_args ''

    local scheme="http://"
    if $tls; then
        scheme="https://"
    fi

    tests:eval curl -sv -A "testcase" "$@" "$scheme$_address$uri" '2>&1'

    stdout_file=$(tests:get-stdout-file)
    stdout=$(cat "$stdout_file")

    local request=$(
        echo "$stdout" | sed -r '/^[\*}{] /d' | sed -r 's/^[><] //'
    )

	tests:put-string request "$request"

    _request=$(cat program_args)
}
