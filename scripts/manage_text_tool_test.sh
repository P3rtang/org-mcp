#!/bin/bash

go build .

json='{ "method":"tools/call", "params":{ "name":"manage_text", "arguments": { "values": [ { "method": "add", "uid": "56493203", "content": "This is a text entry in the header" } ], "show_diff": true } } }'

echo "$json" | ./org-mcp 2>/dev/null

json='{ "method":"tools/call", "params":{ "name":"manage_text", "arguments": { "values": [ { "method": "update", "uid": "56493203.t0", "content": "This is the updated text" } ], "show_diff": true } } }'

echo "$json" | ./org-mcp 2>/dev/null
