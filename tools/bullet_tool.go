package tools

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/utils/itertools"
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

	affectedCount := 0
	affectedItems := map[orgmcp.Uid]orgmcp.Render{}

	for _, b := range input.Bullets {
		selected, ok := orgFile.GetUid(orgmcp.NewUid(b.Uid)).Split()
		if !ok {
			resp = append(resp, fmt.Sprintf("Uid %s not found in %s", b.Uid, path))
			continue
		}

		switch b.Method {
		case "add":
			currentProgress := selected.CheckProgress()

			bullet := orgmcp.NewBullet(selected, orgmcp.NewBulletStatus(b.Checkbox))
			bullet.SetContent(b.Content)

			if selected.CheckProgress().AndThen(func(p orgmcp.Progress) bool {
				return currentProgress.IsSome() && p.Equal(currentProgress.Unwrap())
			}) {
				affectedCount += 1
			}
		case "remove":
			header, ok := orgFile.GetUid(selected.ParentUid()).Split()
			if !ok {
				resp = append(resp, fmt.Sprintf("Parent header for bullet with UID %s not found in %s, skipping removal.", b.Uid, path))
				continue
			}

			currentProgress := header.CheckProgress()

			header.RemoveChildren(orgmcp.NewUid(b.Uid))
			header.CheckProgress()

			if header.CheckProgress().AndThen(func(p orgmcp.Progress) bool {
				return currentProgress.IsSome() && p.Equal(currentProgress.Unwrap())
			}) {
				affectedCount += 1
			}
		case "complete":
			bullet, ok := selected.(*orgmcp.Bullet)
			if !ok {
				resp = append(resp, fmt.Sprintf("UID %s does not correspond to a bullet in %s, skipping completion.", b.Uid, path))
				continue
			}

			header, ok := orgFile.GetUid(bullet.Uid()).Split()
			if !ok {
				resp = append(resp, fmt.Sprintf("Parent header for bullet with UID %s not found in %s, skipping completion.", b.Uid, path))
				continue
			}

			currentProgress := header.CheckProgress()
			bullet.CompleteCheckbox()

			if header.CheckProgress().AndThen(func(p orgmcp.Progress) bool {
				return currentProgress.IsSome() && p.Equal(currentProgress.Unwrap())
			}) {
				affectedCount += 1
			}
		case "toggle":
			bullet, ok := selected.(*orgmcp.Bullet)
			if !ok {
				resp = append(resp, fmt.Sprintf("UID %s does not correspond to a bullet in %s, skipping toggle.", b.Uid, path))
				continue
			}

			header, ok := orgFile.GetUid(bullet.Uid()).Split()
			if !ok {
				resp = append(resp, fmt.Sprintf("Parent header for bullet with UID %s not found in %s, skipping completion.", b.Uid, path))
				continue
			}

			currentProgress := header.CheckProgress()
			bullet.ToggleCheckbox()

			if header.CheckProgress().AndThen(func(p orgmcp.Progress) bool {
				return currentProgress.IsSome() && p.Equal(currentProgress.Unwrap())
			}) {
				affectedCount += 1
			}
		case "set_content":
			bullet, ok := selected.(*orgmcp.Bullet)
			if !ok {
				resp = append(resp, fmt.Sprintf("UID %s does not correspond to a bullet in %s, skipping content update.", b.Uid, path))
				continue
			}

			bullet.SetContent(b.Content)
			affectedItems[bullet.Uid()] = bullet
			affectedCount += 1
		default:
			return nil, errors.New("invalid method; must be 'add', 'remove', 'complete', 'toggle' or 'set_content'")
		}

		affectedItems[selected.Uid()] = selected
		for _, child := range selected.ChildrenRec(1) {
			affectedItems[child.Uid()] = child
		}

		affectedCount += 1
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
