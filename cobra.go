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

	exportCmd.Flags().StringP("input", "i", ".tasks.org", "Input Org file")
	exportCmd.Flags().StringP("output", "o", "", "Output Markdown file, will default to stdout.")
	exportCmd.Flags().StringP("format", "f", "markdown", "Output format")
	rootCmd.AddCommand(&exportCmd)

	embedCommand.Flags().StringP("input", "i", ".tasks.org", "Input Org file")
	rootCmd.AddCommand(&embedCommand)
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
		if os.Getenv("SHOW_DEBUG") == "" {
			os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
		}

		ctx := cmd.Context()
		logger := ctx.Value("logger").(*slog.Logger)

		f := ".tasks.org"
		writer := os.Stdout
		format := "markdown"

		cmd.Flags().Visit(func(flag *pflag.Flag) {
			switch flag.Name {
			case "input":
				f = flag.Value.String()
			case "output":
				outFile := flag.Value.String()
				if outFile != "" {
					var err error
					writer, err = os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

					if err != nil {
						logger.Error(fmt.Sprintf("Failed to open output file %s: %v", outFile, err))
						os.Exit(1)
					}
				}
			case "format":
				format = flag.Value.String()
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

		switch format {
		case "markdown":
			orgFile.RenderMarkdown(&builder, -1)
		case "table":
			table := orgmcp.PrintTable(orgFile.ChildrenRec(-1), orgmcp.AllColumns)
			builder.WriteString(table)
		}

		_, err = writer.WriteString(builder.String())
	},
}

var embedCommand = cobra.Command{
	Use:   "embed",
	Short: "Generate embeddings for Org file content",
	Long: `
Generates vector embeddings for the content of an Org file. This can be used for tasks like semantic search or clustering.
The command reads the specified Org file, extracts relevant content, and puts the resulting embeddings into the properties of the relevant org header.
`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		logger := ctx.Value("logger").(*slog.Logger)

		// Implementation for embedding generation would go here
		file := cmd.Flags().Lookup("input").Value.String()
		if file == "" {
			file = ".tasks.org"
		}

		orgFile, err := mcp.LoadOrgFile(ctx, file)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to parse %s: %v", file, err))
			os.Exit(1)
		}

		err = orgFile.GenerateEmbeddings()

		_, err = mcp.WriteOrgFileToDisk(ctx, orgFile, file)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to write updated org file to disk: %v", err))
			os.Exit(1)
		}
	},
}
