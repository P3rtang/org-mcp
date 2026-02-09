#!/bin/bash

go build .

json='{ "method":"initialize", "params":{} }'

echo "$json" | ./org-mcp
