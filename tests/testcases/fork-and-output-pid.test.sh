#!/bin/bash

tests:ensure :run -l :65002 -e /bin/true
tests:assert-stdout-re '^\d+$'

@var process cat $(tests:get-stdout-file)

tests:ensure ps -p $process
tests:ensure kill -9 $process
