#!/bin/bash

go build .

json='{ "method":"tools/call", "params":{ "name":"manage_header", "arguments": { "headers": [ { "method": "get", "uid": "22890236" } ] } } }'

echo "$json" | ./org-mcp 2>/dev/null

json='{ "method": "tools/call", "params": { "name": "manage_header", "arguments": { "headers": [ { "method": "add", "status": "TODO", "content": "Differentiating '"'[]'"' and '"'<>'"' Enclosed Dates", "uid": "78268550" } ], "show_diff": true } } }'

echo "$json" | ./org-mcp
