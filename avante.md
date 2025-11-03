# Project instruction for org-mcp

## Your role

You will assist in writing and brainstorming ideas for an MCP server written in golang.
The goal of the MCP server is to give ai models the ability to interact with Emacs org mode files.

## Tasks

Tasks will be added to the .tasks.org file.
Checking status progress etc. should be done through the existing org-mcp server.
While not complete yet it serves as a test for this project as well.

Always remember that .tasks.org is your main source of truth and is worked on by everyone. Programmers and models like you.

Never do more than one task at a time.
When a task is done and its progress has been saved to the tasks file, you should `attempt_completion` and wait for user feedback.

## Project overview

The MCP server will provide a RESTful API that allows AI models to read, write, and manipulate org mode files.
The server will be built using golang for performance and scalability.

## Key features

1. **Header Management**: The server will support reading and writing org mode headers, allowing changes in status TODO, PROG, DONE etc.
2. **Content Manipulation**: The server will allow adding, editing, and deleting content within org mode files.
3. **Tagging System**: The server will support adding and removing tags from org mode entries.
4. **Search Functionality**: The server will provide search capabilities to find specific entries, based on status, a uid property tag, or header tags.
5. **Version Control**: The server will maintain version history for org mode files, allowing users to revert to previous versions if needed.

## Technologies

- Golang: For building the MCP server.
- RESTful API: For communication between AI models and the MCP server.
- Emacs Org Mode: The file format that the MCP server will interact with.

## Coding standards

- Follow Go conventions and best practices.
- Write clean, maintainable, and well-documented code.
- Implement error handling and logging.
- Write unit tests to ensure code quality and reliability.
- Use version control (e.g., Git) for code management.

