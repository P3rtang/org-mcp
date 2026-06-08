#!/bin/bash

go build .

json='{ "method":"server/discover", "params":{} }'

echo "$json" | ./org-mcp serve
