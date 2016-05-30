:run

:request / --data AAABBB
tests:assert-no-diff-blank $_request/body/raw -w <<RAW
AAABBB
RAW
