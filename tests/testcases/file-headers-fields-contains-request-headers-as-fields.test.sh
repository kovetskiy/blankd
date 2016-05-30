:run

:request /
tests:assert-no-diff-blank $_request/headers/fields -w <<RAW
Accept=*/*
User-Agent=testcase
RAW
