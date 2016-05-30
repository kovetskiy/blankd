:run

:request '/x/y.z?a=%2fb&c=f'
tests:assert-no-diff-blank $_request/uri/query -w <<RAW
a=%2fb&c=f
RAW
