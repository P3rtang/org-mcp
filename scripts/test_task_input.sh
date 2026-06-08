#!/bin/bash

go build .

json='{ "method":"tools/call", "params":{ "name":"test_task_input", "arguments": {} } }'

echo "$json" | ./org-mcp serve 2>/dev/null
