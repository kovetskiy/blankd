:run

:request /
tests:assert-no-diff-blank $_request/raw -w <<RAW
GET / HTTP/1.1
Accept: */*
User-Agent: testcase
RAW
