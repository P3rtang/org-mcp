package tools

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/utils/itertools"
	"github.com/p3rtang/org-mcp/utils/option"
)

type BulletInput struct {
	Bullets      []BulletValue    `json:"bullets,omitempty" jsonschema:"description=List of bullet point operations to perform."`
	Path         string           `json:"path,omitempty" jsonschema:"description=Optional file path; defaults to ./.tasks.org."`
	ShowDiff     bool             `json:"show_diff,omitempty" jsonschema:"description=Whether to show the diff of changes made to the Org file."`
	ShowAffected *bool            `json:"show_affected,omitempty" jsonschema:"description=Whether to include the affected items in the response. This will include all items that were modified as well as their children.,default=true"`
	Columns      []*orgmcp.Column `json:"columns,omitempty" jsonschema:"description=List of columns to include in the output. If not specified defaults to [UID | PREVIEW]."`
}

type BulletValue struct {
	Uid      string `json:"uid,omitempty" jsonschema:"description=UID of the bullet point to modify. For the 'add' method; this must be the parent header UID.,required=true"`
	Method   string `json:"method" jsonschema:"description=The action to perform on the bullet point.,enum=add;remove;complete;toggle;set_content,required=true"`
	Content  string `json:"content,omitempty" jsonschema:"description=Text content of the bullet."`
	Checkbox string `json:"checkbox,omitempty" jsonschema:"description=Checkbox status for the new bullet.,enum=None;Unchecked;Checked"`
}

var BulletTool = mcp.GenericTool[BulletInput]{
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
	Callback: bulletFunc,
}

func bulletFunc(input BulletInput, options mcp.FuncOptions) (resp []any, err error) {
	var path string

	if input.Path == "" {
		path = options.DefaultPath
	} else {
		path = input.Path
	}

	orgFile, err := loadOrgFile(path)
	if err != nil {
		return nil, fmt.Errorf("error loading org file: %v", err)
	}

	builder := strings.Builder{}

	affectedCount := 0
	affectedItems := map[orgmcp.Uid]orgmcp.Render{}

	for _, b := range input.Bullets {
		switch b.Method {
		case "add":
			header, ok := orgFile.GetUid(orgmcp.NewUid(b.Uid)).Split()
			if !ok {
				return nil, errors.New("Invalid header uid for adding bullet, do not include bullet index.")
			}

			currentProgress := header.CheckProgress()

			bullet := orgmcp.NewBullet(header, orgmcp.NewBulletStatus(b.Checkbox))
			bullet.SetContent(b.Content)

			if header.CheckProgress().AndThen(func(p orgmcp.Progress) bool {
				return currentProgress.IsSome() && p.Equal(currentProgress.Unwrap())
			}) {
				affectedCount += 1
			}

			affectedItems[header.Uid()] = header
			for _, child := range header.ChildrenRec(1) {
				affectedItems[child.Uid()] = child
			}

			affectedCount += 1
		case "remove":
			bullet := orgFile.GetUid(orgmcp.NewUid(b.Uid))

			header_uid, ok := option.Map(bullet, func(b orgmcp.Render) orgmcp.Uid { return b.ParentUid() }).Split()
			if !ok {
				return nil, errors.New("invalid bullet uid; cannot find parent header")
			}

			header := orgFile.GetUid(header_uid).Unwrap()
			currentProgress := header.CheckProgress()

			header.RemoveChildren(orgmcp.NewUid(b.Uid))
			header.CheckProgress()

			if header.CheckProgress().AndThen(func(p orgmcp.Progress) bool {
				return currentProgress.IsSome() && p.Equal(currentProgress.Unwrap())
			}) {
				affectedCount += 1
			}

			affectedItems[header.Uid()] = header
			for _, child := range header.ChildrenRec(1) {
				affectedItems[child.Uid()] = child
			}

			affectedCount += 1
		case "complete":
			bullet, ok := option.Cast[orgmcp.Render, *orgmcp.Bullet](orgFile.GetUid(orgmcp.NewUid(b.Uid))).Split()
			if !ok {
				err = errors.New("invalid bullet uid for set_content")
				return
			}

			header, ok := orgFile.GetUid(bullet.Uid()).Split()
			if !ok {
				return nil, errors.New("invalid bullet uid; cannot find parent header")
			}

			currentProgress := header.CheckProgress()
			bullet.CompleteCheckbox()

			if header.CheckProgress().AndThen(func(p orgmcp.Progress) bool {
				return currentProgress.IsSome() && p.Equal(currentProgress.Unwrap())
			}) {
				affectedCount += 1
			}

			affectedItems[header.Uid()] = header
			for _, child := range header.ChildrenRec(1) {
				affectedItems[child.Uid()] = child
			}

			affectedCount += 1
		case "toggle":
			bullet, ok := option.Cast[orgmcp.Render, *orgmcp.Bullet](orgFile.GetUid(orgmcp.NewUid(b.Uid))).Split()
			if !ok {
				err = errors.New("invalid bullet uid for set_content")
				return
			}

			bullet.ToggleCheckbox()

			orgFile.GetUid(bullet.ParentUid()).Then(func(r orgmcp.Render) {
				r.CheckProgress()

				affectedItems[r.Uid()] = r
				for _, child := range r.ChildrenRec(1) {
					affectedItems[child.Uid()] = child
				}
			})
		case "set_content":
			bullet, ok := option.Cast[orgmcp.Render, *orgmcp.Bullet](orgFile.GetUid(orgmcp.NewUid(b.Uid))).Split()
			if !ok {
				err = errors.New("invalid bullet uid for set_content")
				return
			}

			bullet.SetContent(b.Content)
			affectedItems[bullet.Uid()] = bullet
			affectedCount += 1
		default:
			return nil, errors.New("invalid method; must be 'add', 'remove', 'complete', 'toggle' or 'set_content'")
		}

		builder.Reset()
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

	diff, err := writeOrgFileToDisk(orgFile, path)

	if input.ShowDiff {
		resp = append(resp, diff)
	}

	return
}
