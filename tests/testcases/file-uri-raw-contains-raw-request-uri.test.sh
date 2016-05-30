:run

:request '/x/y.z?a=%2fb&c=f'
tests:assert-no-diff-blank $_request/uri/raw -w <<RAW
/x/y.z?a=%2fb&c=f
RAW
