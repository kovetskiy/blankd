:run

:request / --data 'aaa=bbb&ccc=ddd'
tests:assert-no-diff-blank "$_request/body/values" -w <<RAW
aaa=bbb
ccc=ddd
RAW
