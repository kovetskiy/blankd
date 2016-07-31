:run --tls

:response <<RAW
500 Internal Server Error
Date: 1

body
RAW

:request / --tls --insecure

tests:assert-no-diff-blank request -w <<REQUEST
CApath: none
GET / HTTP/1.1
Host: $_address
User-Agent: testcase
Accept: */*

HTTP/1.1 500 Internal Server Error
Date: 1
Content-Length: 5
Content-Type: text/plain; charset=utf-8

body
REQUEST
