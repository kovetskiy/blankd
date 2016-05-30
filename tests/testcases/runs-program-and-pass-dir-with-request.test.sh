#!/bin/bash

:run

:request /

tests:eval tree "$_request" \| tail -n+2
tests:assert-no-diff stdout <<TREE
├── body
│   ├── raw
│   └── values
├── cookies
├── headers
│   ├── raw
│   └── values
├── host
├── _id
├── method
├── raw
└── uri
    ├── path
    ├── query
    ├── raw
    └── values

3 directories, 13 files
TREE
