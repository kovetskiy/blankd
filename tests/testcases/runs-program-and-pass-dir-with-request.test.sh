#!/bin/bash

:run

:request /

tests:eval tree "$_request" \| tail -n+2
tests:assert-no-diff stdout <<TREE
├── body
│   ├── fields
│   └── raw
├── cookies
│   ├── fields
│   └── raw
├── headers
│   ├── fields
│   └── raw
├── host
├── _id
├── method
├── raw
└── uri
    ├── fields
    ├── path
    ├── query
    └── raw

4 directories, 14 files
TREE
