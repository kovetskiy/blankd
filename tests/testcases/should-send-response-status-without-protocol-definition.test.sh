:run

:response <<RAW
200 OK
RAW

:request /

tests:assert-re request "HTTP/1.1 200 OK"
