:run

:request / -b 'name=value; name2=value2'
tests:assert-no-diff-blank $_request/cookies -w <<RAW
name2=value2
name=value
RAW
