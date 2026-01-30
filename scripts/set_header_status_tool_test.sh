#!/bin/bash

go build . && echo '{"method":"tools/call","params":{"name":"set_header_status","arguments": {"uid": 22890236},"_meta": {"progressToken": 0} } }' | ./org-mcp

