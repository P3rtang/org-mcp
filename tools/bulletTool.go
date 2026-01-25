package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/utils/option"
)

type BulletInput struct {
	Bullets []BulletValue `json:"bullets,omitempty"`
	Path    string        `json:"path,omitempty"`
}

type BulletValue struct {
	Uid      string `json:"uid,omitempty"`
	Method   string `json:"method"`
	Content  string `json:"content,omitempty"`
	Checkbox string `json:"checkbox,omitempty"`
}

var BulletTool = mcp.Tool{
	Name: "manage_bullet",
	Description: "Add, remove or complete bullet points.\n" +
		"The method parameter defines the action to take: 'add', 'remove', 'complete', 'toggle' or 'set_content'.\n" +
		"- 'add': Adds a new bullet point at the specified index under the given header_uid. Requires 'content' and 'checkbox' parameters.\n" +
		"- 'remove': Removes the bullet point identified by its uid.\n" +
		"- 'complete': Marks the bullet point as completed (Checked).\n" +
		"- 'toggle': Toggles the checkbox status of the bullet point between Checked and Unchecked.\n" +
		"- 'set_content': Updates the content of the bullet point. Requires 'content' parameter.\n\n" +
		"The 'header_uid' parameter specifies the parent header under which the bullet point resides.\n" +
		"The 'bullet_index' parameter specifies the position of the bullet point under the header (0-based index).\n\n" +
		"When targeting a bullet, the uid is constructed as `header_uid + '.b' + bullet_index`.\n" +
		"Bullets are hierarchical meaning that bullets can have sub-bullets. Sub-bullets will use parent_bullet_uid + '.b' + bullet_sub_index like h1.b0.b1\n",

	InputSchema: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"bullets": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":     "object",
					"required": []string{"method", "uid"},
					"properties": map[string]any{
						"uid": map[string]any{
							"type":        "string",
							"description": "UID of the bullet point to modify. The uid is constructed as `header_uid + 'b' + bullet_index`",
						},
						"method": map[string]any{
							"type":        "string",
							"enum":        []string{"add", "remove", "complete", "toggle", "set_content"},
							"description": "The method by which to manage the bullet point.",
						},
						"content": map[string]any{
							"type":        "string",
							"description": "Text content of the bullet.",
						},
						"checkbox": map[string]any{
							"type":        "string",
							"description": "Checkbox status for the new bullet.",
							// TODO: add partial checked
							"enum": []string{"None", "Unchecked", "Checked"},
						},
					},
				},
			},
			"path": map[string]any{
				"type":        "string",
				"description": "Optional file path, defaults to ./.tasks.org.",
			},
		},
	},
}

func BulletFunc(args map[string]any, options mcp.FuncOptions) (resp any, err error) {
	var response []map[string]any

	bytes, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("error marshalling bullet input: %v", err)
	}

	var input BulletInput
	err = json.Unmarshal(bytes, &input)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling bullet input: %v", err)
	}

	var path string

	if input.Path == "" {
		path = options.DefaultPath
	} else {
		path = input.Path
	}

	of, err := loadOrgFile(path)
	if err != nil {
		return nil, fmt.Errorf("error loading org file: %v", err)
	}

	builder := strings.Builder{}

	for _, b := range input.Bullets {
		bullet := of.GetUid(orgmcp.NewUid(b.Uid))
		fmt.Fprintf(os.Stderr, "%v", bullet)

		header_uid, ok := option.Map(bullet, func(b orgmcp.Render) orgmcp.Uid { return b.ParentUid() }).Split()
		if !ok {
			return nil, errors.New("invalid bullet uid; cannot find parent header")
		}

		header := of.GetUid(header_uid).Unwrap()

		switch b.Method {
		case "add":
			content, ok := args["content"].(string)
			if !ok || strings.TrimSpace(content) == "" {
				return nil, errors.New("content is required for add method")
			}

			var c string
			if c, ok = args["checkbox"].(string); !ok {
				return nil, errors.New("invalid or missing checkbox parameter")
			}

			bullet := orgmcp.NewBullet(header, orgmcp.NewBulletStatus(c))
			bullet.SetContent(content)
		case "remove":
			header.RemoveChildren(orgmcp.NewUid(b.Uid))
		case "complete":
			bullet.Then(func(r orgmcp.Render) {
				if b, ok := r.(*orgmcp.Bullet); ok {
					b.CompleteCheckbox()
				}
			})
		case "toggle":
			bullet.Then(func(r orgmcp.Render) {
				if b, ok := r.(*orgmcp.Bullet); ok {
					b.ToggleCheckbox()
				}
			})
		case "set_content":
			content, ok := args["content"].(string)
			if !ok || strings.TrimSpace(content) == "" {
				return nil, errors.New("content is required for set_content method")
			}

			bullet.Then(func(r orgmcp.Render) {
				if b, ok := r.(*orgmcp.Bullet); ok {
					b.SetContent(content)
				}
			})
		default:
			return nil, errors.New("invalid method; must be 'add', 'remove', 'complete', 'toggle' or 'set_content'")
		}

		header.CheckProgress()
		header.Render(&builder, 1)

		response = append(response, map[string]any{
			"header": builder.String(),
		})

		builder.Reset()
	}

	err = writeOrgFileToDisk(of, path)
	resp = response

	return
}
