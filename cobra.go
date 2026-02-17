package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/tools"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	rootCmd.AddCommand(&serveCmd)

	exportCmd.Flags().StringP("input", "i", ".tasks.org", "Input Org file (default: .tasks.org)")
	exportCmd.Flags().StringP("output", "o", "out.md", "Output Markdown file (default: out.md)")
	rootCmd.AddCommand(&exportCmd)
}

var rootCmd = cobra.Command{
	Use:   "org-mcp",
	Short: "A tool for managing Org files using the MCP protocol",
	Long: `
org-mcp is a command-line tool that implements the MCP protocol to manage Org files.
It can run as a server that listens for MCP messages and responds accordingly, or it can export Org files to Markdown format.
`,
}

var serveCmd = cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server",
	Long: `
Starts the MCP server which listens for incoming messages and responds accordingly.
The server uses a message sender that encodes messages as JSON and sends them to stdout.

This command can be simply run without the "serve" subcommand, as it is the default action when running "org-mcp".
`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		logger := ctx.Value("logger").(*slog.Logger)

		// Create a message sender that encodes and sends JSON to stdout
		// Using a persistent encoder for better performance and proper flushing
		encoder := json.NewEncoder(os.Stdout)
		sender := mcp.NewMessageSender(func(msg any) error {
			return encoder.Encode(msg)
		})

		// Create and run the MCP server
		server := mcp.NewServer(os.Stdin, sender, logger)

		server.AddTool(&tools.ViewTool)
		server.AddTool(&tools.HeaderTool)
		server.AddTool(&tools.BulletTool)
		server.AddTool(&tools.StatusTool)
		server.AddTool(&tools.VectorSearch)
		server.AddTool(&tools.TextTool)

		if err := server.Run(ctx); err != nil {
			logger.Error(fmt.Sprintf("Server error: %v", err))
			os.Exit(1)
		}
	},
}

var exportCmd = cobra.Command{
	Use:   "export",
	Short: "Export an Org file to Markdown format",
	Long: `
Exports the specified Org file to Markdown format. If no input file is provided, it defaults to .tasks.org.
If no output file is provided, it defaults to out.md.
`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		logger := ctx.Value("logger").(*slog.Logger)

		f := ".tasks.org"
		o := "out.md"

		cmd.Flags().Visit(func(flag *pflag.Flag) {
			switch flag.Name {
			case "input":
				f = flag.Value.String()
			case "output":
				o = flag.Value.String()
			}
		})

		file, err := os.Open(f)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to open %s: %v", f, err))
			os.Exit(1)
		}

		orgFile, err := orgmcp.OrgFileFromReader(ctx, file).Split()
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to parse %s: %v", f, err))
			os.Exit(1)
		}

		file.Close()

		builder := strings.Builder{}
		orgFile.RenderMarkdown(&builder, -1)

		out, err := os.Create(o)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to create %s: %v", o, err))
			os.Exit(1)
		}

		_, err = out.WriteString(builder.String())
	},
}
