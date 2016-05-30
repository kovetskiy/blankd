:run

:response <<RAW
500 Internal Server Error
Date: 1

body
RAW

:request /

tests:assert-no-diff-blank request -w <<REQUEST
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
