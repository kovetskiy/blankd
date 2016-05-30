#!/bin/bash

tests:describe "logs:"
tests:eval cat logs

if [[ "$_process" ]]; then
    tests:eval kill -USR2 "$_process"
fi
