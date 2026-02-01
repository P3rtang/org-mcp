package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/utils/option"
)

type BulletInput struct {
	Bullets  []BulletValue `json:"bullets,omitempty"`
	Path     string        `json:"path,omitempty"`
	ShowDiff bool          `json:"show_diff,omitempty"`
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
		"Bullets are hierarchical meaning that bullets can have sub-bullets. Sub-bullets will use parent_bullet_uid + '.b' + bullet_sub_index like header_uid.b0.b1\n" +
		"The add method is special as it requires the header_uid to be passed directly without any bullet index. The index will be determined by the tool itself.",

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
							"description": "UID of the bullet point to modify. The uid is constructed as `header_uid + '.b' + bullet_index`. For the 'add' method you should add the header uid instead",
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
			"show_diff": map[string]any{
				"type":        "boolean",
				"description": "Whether to show the diff of changes made to the Org file.",
			},
		},
	},

	Callback: bulletFunc,
}

func bulletFunc(args map[string]any, options mcp.FuncOptions) (resp []any, err error) {
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
		switch b.Method {
		case "add":
			header, ok := of.GetUid(orgmcp.NewUid(b.Uid)).Split()
			if !ok {
				return nil, errors.New("Invalid header uid for adding bullet, do not include bullet index.")
			}

			bullet := orgmcp.NewBullet(header, orgmcp.NewBulletStatus(b.Checkbox))
			bullet.SetContent(b.Content)
		case "remove":
			bullet := of.GetUid(orgmcp.NewUid(b.Uid))

			header_uid, ok := option.Map(bullet, func(b orgmcp.Render) orgmcp.Uid { return b.ParentUid() }).Split()
			if !ok {
				return nil, errors.New("invalid bullet uid; cannot find parent header")
			}

			header := of.GetUid(header_uid).Unwrap()

			header.RemoveChildren(orgmcp.NewUid(b.Uid))
			header.CheckProgress()
		case "complete":
			bullet, ok := option.Cast[orgmcp.Render, *orgmcp.Bullet](of.GetUid(orgmcp.NewUid(b.Uid))).Split()
			if !ok {
				err = errors.New("invalid bullet uid for set_content")
				return
			}

			bullet.CompleteCheckbox()

			of.GetUid(bullet.ParentUid()).Then(func(r orgmcp.Render) {
				r.CheckProgress()
			})
		case "toggle":
			bullet, ok := option.Cast[orgmcp.Render, *orgmcp.Bullet](of.GetUid(orgmcp.NewUid(b.Uid))).Split()
			if !ok {
				err = errors.New("invalid bullet uid for set_content")
				return
			}

			bullet.ToggleCheckbox()

			of.GetUid(bullet.ParentUid()).Then(func(r orgmcp.Render) {
				r.CheckProgress()
			})
		case "set_content":
			bullet, ok := option.Cast[orgmcp.Render, *orgmcp.Bullet](of.GetUid(orgmcp.NewUid(b.Uid))).Split()
			if !ok {
				err = errors.New("invalid bullet uid for set_content")
				return
			}

			bullet.SetContent(b.Content)
		default:
			return nil, errors.New("invalid method; must be 'add', 'remove', 'complete', 'toggle' or 'set_content'")
		}

		builder.Reset()
	}

	diff, err := writeOrgFileToDisk(of, path)

	if input.ShowDiff {
		resp = append(resp, diff)
	}

	return
}
