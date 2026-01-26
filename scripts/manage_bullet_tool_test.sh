#!/bin/bash

go build . && echo '{"method":"tools/call","params":{"name":"manage_bullet","arguments": {"bullets": [{ "method": "toggle", "uid": "99590227.b1" }]},"_meta": {"progressToken": 0} } }' | ./org-mcp
