:run

:request /
tests:assert-no-diff-blank $_request/headers/raw -w <<RAW
Accept: */*
User-Agent: testcase
RAW
