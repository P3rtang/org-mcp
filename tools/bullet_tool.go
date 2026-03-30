package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/utils/itertools"
)

type BulletInput struct {
	Bullets      []mcp.OneOf[*BulletInputUnion] `json:"bullets" jsonschema:"description=List of bullet point operations to perform."`
	Path         string                         `json:"path,omitempty" jsonschema:"description=Optional file path; defaults to ./.tasks.org."`
	ShowDiff     bool                           `json:"show_diff,omitempty" jsonschema:"description=Whether to show the diff of changes made to the Org file.,default=false"`
	ShowAffected *bool                          `json:"show_affected,omitempty" jsonschema:"description=Whether to include the affected items in the response. This will include all items that were modified as well as their children.,default=true"`
	Columns      []*orgmcp.Column               `json:"columns,omitempty" jsonschema:"description=List of columns to include in the output. If not specified defaults to [UID | PREVIEW]."`
}

type BulletInputUnion struct {
	tag string

	Add    BulletInputAdd
	Update BulletInputUpdate
	Remove BulletInputRemove
}

func (b *BulletInputUnion) Value() any {
	switch b.tag {
	case "add":
		return b.Add
	case "update":
		return b.Update
	case "remove":
		return b.Remove
	default:
		return nil
	}
}

func (b *BulletInputUnion) Tag() string {
	return b.tag
}

func (b *BulletInputUnion) FromJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	switch raw["method"] {
	case "add":
		b.tag = "add"
		return json.Unmarshal(data, &b.Add)
	case "update":
		b.tag = "update"
		return json.Unmarshal(data, &b.Update)
	case "remove":
		b.tag = "remove"
		return json.Unmarshal(data, &b.Remove)
	default:
		return errors.New("invalid method; must be 'add', 'update' or 'remove'")
	}
}

type BulletInputAdd struct {
	Method   string `json:"method" jsonschema:"description=The action to perform on the bullet point.,enum=add,required=true"`
	Parent   string `json:"parent" jsonschema:"description=UID of the parent header or bullet under which to add the new bullet point."`
	Content  string `json:"content" jsonschema:"description=Text content of the new bullet point."`
	Checkbox string `json:"checkbox,omitempty" jsonschema:"description=Checkbox status for the new bullet.,enum=None;Unchecked;Checked"`
}

func (b *BulletInputAdd) Apply(ctx context.Context, of *orgmcp.OrgFile) (res ApplyResult) {
	res.affectedItems = make(map[orgmcp.Uid]orgmcp.Render)

	parent, ok := of.GetUid(orgmcp.NewUid(b.Parent)).Split()
	if !ok {
		res.err = fmt.Errorf("Parent uid %s not found, skipping addition.", b.Parent)
		return
	}

	bullet := orgmcp.NewBullet(parent, orgmcp.NewBulletStatus(b.Checkbox))
	bullet.SetContent(b.Content)

	res.affectedItems[bullet.Uid()] = bullet

	return
}

type BulletInputUpdate struct {
	Method   string `json:"method" jsonschema:"description=The action to perform on the bullet point.,enum=update,required=true"`
	Uid      string `json:"uid" jsonschema:"description=UID of the bullet point to modify.,required=true"`
	Content  string `json:"content,omitempty" jsonschema:"description=Text content of the bullet."`
	Checkbox string `json:"checkbox,omitempty" jsonschema:"description=Checkbox status for the bullet.,enum=None;Unchecked;Checked"`
}

func (b *BulletInputUpdate) Apply(ctx context.Context, of *orgmcp.OrgFile) (res ApplyResult) {
	res.affectedItems = make(map[orgmcp.Uid]orgmcp.Render)

	selected, ok := of.GetUid(orgmcp.NewUid(b.Uid)).Split()
	if !ok {
		res.err = fmt.Errorf("Uid %s not found", b.Uid)
		return
	}

	bullet, ok := selected.(*orgmcp.Bullet)
	if !ok {
		res.err = fmt.Errorf("UID %s does not correspond to a bullet, skipping update.", b.Uid)
		return
	}

	if b.Content != "" {
		bullet.SetContent(b.Content)
	}

	if b.Checkbox != "" {
		switch b.Checkbox {
		case "None":
			bullet.SetCheckbox(orgmcp.NoCheck)
		case "Unchecked":
			bullet.SetCheckbox(orgmcp.Unchecked)
		case "Checked":
			bullet.SetCheckbox(orgmcp.Checked)
		default:
			res.err = fmt.Errorf("invalid checkbox value: %s", b.Checkbox)
			return
		}
	}

	res.affectedItems[bullet.Uid()] = bullet

	return
}

type BulletInputRemove struct {
	Method string `json:"method" jsonschema:"description=The action to perform on the bullet point.,enum=remove,required=true"`
	Uid    string `json:"uid" jsonschema:"description=UID of the bullet point to remove."`
}

func (b *BulletInputRemove) Apply(ctx context.Context, of *orgmcp.OrgFile) (res ApplyResult) {
	res.affectedItems = make(map[orgmcp.Uid]orgmcp.Render)

	selected, ok := of.GetUid(orgmcp.NewUid(b.Uid)).Split()
	if !ok {
		res.err = fmt.Errorf("Uid %s not found", b.Uid)
		return
	}

	header, ok := of.GetUid(selected.ParentUid()).Split()
	if !ok {
		res.err = fmt.Errorf("Parent header for bullet with UID %s not found, skipping removal.", b.Uid)
		return
	}

	header.RemoveChildren(orgmcp.NewUid(b.Uid))

	res.affectedItems[header.Uid()] = header

	return
}

var BulletTool = mcp.GenericTool[BulletInput]{
	Name: "manage_bullet",
	Description: "Add, remove or complete bullet points.\n" +
		"The method parameter defines the action to take: 'add', 'remove' or 'update'.\n" +
		"- 'add': Adds a new bullet point at the specified index under the given header_uid. Requires 'content' and 'checkbox' parameters.\n" +
		"- 'remove': Removes the bullet point identified by its uid.\n" +
		"- 'update': Updates the content of the bullet point. Requires 'content' parameter.\n\n" +
		"When targeting a bullet, the uid is constructed as `header_uid + '.b' + bullet_index`.\n" +
		"The 'bullet_index' specifies the position of the bullet point under the header (0-based index).\n\n" +
		"Bullets are hierarchical meaning that bullets can have sub-bullets. Sub-bullets will use parent_bullet_uid + '.b' + bullet_sub_index like header_uid.b0.b1\n",
	Callback: bulletFunc,
}

func bulletFunc(ctx context.Context, input BulletInput, options mcp.FuncOptions) (resp []any, err error) {
	var path string

	if input.Path == "" {
		path = options.DefaultPath
	} else {
		path = input.Path
	}

	orgFile, err := mcp.LoadOrgFile(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("error loading org file: %v", err)
	}

	affectedCount := 0
	affectedItems := map[orgmcp.Uid]orgmcp.Render{}

	for _, mt := range input.Bullets {
		var res ApplyResult

		switch mt.Value.Tag() {
		case "add":
			res = mt.Value.Add.Apply(ctx, &orgFile)
		case "update":
			res = mt.Value.Update.Apply(ctx, &orgFile)
		case "remove":
			res = mt.Value.Remove.Apply(ctx, &orgFile)
		}

		if res.err != nil {
			resp = append(resp, res.err)
		}

		maps.Copy(affectedItems, res.affectedItems)
		affectedCount += len(res.affectedItems)
	}

	ordered := []orgmcp.Render{}

	if input.ShowAffected == nil || *input.ShowAffected == true {
		locationTable := orgFile.BuildLocationTable()
		ordered = append(ordered, itertools.Collect(maps.Values(affectedItems))...)

		slices.SortFunc(ordered, func(a, b orgmcp.Render) int {
			return (*locationTable)[a.Uid()] - (*locationTable)[b.Uid()]
		})

		resp = append(resp, orgmcp.PrintCsv(ordered, input.Columns))
		resp = append(resp, map[string]any{
			"affected_count": affectedCount,
		})
	}

	diff, err := mcp.WriteOrgFileToDisk(ctx, orgFile, path)
	if input.ShowDiff {
		resp = append(resp, diff)
	}

	return
}
