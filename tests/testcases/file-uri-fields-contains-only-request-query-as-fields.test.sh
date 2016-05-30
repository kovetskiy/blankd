:run

:request '/x/y.z?a=%2fb&c=f'
tests:assert-no-diff-blank $_request/uri/fields -w <<RAW
a=/b
c=f
RAW
