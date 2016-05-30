:run

:request / -X HEAD
tests:assert-no-diff $_request/method HEAD
