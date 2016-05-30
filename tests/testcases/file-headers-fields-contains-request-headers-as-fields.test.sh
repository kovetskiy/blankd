:run

:request /
tests:assert-no-diff-blank $_request/headers/values -w <<RAW
Accept=*/*
User-Agent=testcase
RAW
