#!/bin/bash

go build . && echo '{"method":"tools/call","params":{"name":"view_header","arguments": {"uid": 28028254},"_meta": {"progressToken": 0} } }' | ./org-mcp
