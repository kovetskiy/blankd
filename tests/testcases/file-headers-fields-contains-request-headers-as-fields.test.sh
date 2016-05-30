:run

:request /
tests:assert-no-diff-blank $_request/headers/fields -w <<RAW
Accept=*/*
Host=$_address
User-Agent=testcase
RAW
