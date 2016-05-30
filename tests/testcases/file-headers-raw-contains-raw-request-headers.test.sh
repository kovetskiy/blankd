:run

:request /
tests:assert-no-diff-blank $_request/headers/raw -w <<RAW
Host: $_address
User-Agent: testcase
Accept: */*
RAW
