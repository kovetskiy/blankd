#!/bin/bash

shopt -s expand_aliases

alias @var='tests:value'

:run() {
    tests:silence tests:pipe $BUILD -o /tmp/log "$@"
    return $(tests:get-exitcode)
}
