:run

:response <<RAW
500 Internal Server Error
X-A: 1
RAW

:request /

tests:assert-re request "X-A: 1"
