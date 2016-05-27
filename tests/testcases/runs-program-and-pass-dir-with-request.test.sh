#!/bin/bash

tests:ensure :run -l :65002 -e /bin/true
tests:assert-stdout-re '^\d+$'

curl localhost:65002

tests:ensure kill -9 $(cat $(tests:get-stdout-file))
