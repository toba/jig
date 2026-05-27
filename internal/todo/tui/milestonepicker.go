package tui

import (
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/toba/jig/internal/todo/issue"
	"github.com/toba/jig/internal/todo/ui"
)

// milestoneSelectedMsg is sent when a milestone is selected from the picker.
// An empty milestoneID clears the assignment (or, in filter mode, the filter).
type milestoneSelectedMsg struct {
	issueIDs    []string
	milestoneID string
	filterMode  bool // when true, set the list filter instead of assigning
}

// closeMilestonePickerMsg is sent when the milestone picker is cancelled.
type closeMilestonePickerMsg struct{}

// openMilestonePickerMsg requests opening the milestone picker for issue(s).
type openMilestonePickerMsg struct {
	issueIDs         []string // IDs of issues to update
	issueTitle       string   // Display title (single title or "N issues")
	currentMilestone string   // Only meaningful for single issue / current filter
	filterMode       bool     // when true, picker sets the list filter rather than assigning
}

// milestoneItem wraps a milestone to implement list.Item.
type milestoneItem struct {
	id          string // milestone ID ("" for the clear entry)
	short       string
	name        string
	description string
	isCurrent   bool
}

func (i milestoneItem) Title() string       { return i.name }
func (i milestoneItem) Description() string { return i.description }
func (i milestoneItem) FilterValue() string { return i.short + " " + i.name }

// milestoneItemDelegate handles rendering of milestone picker items.
type milestoneItemDelegate struct{}

func (d milestoneItemDelegate) Height() int                             { return 1 }
func (d milestoneItemDelegate) Spacing() int                            { return 0 }
func (d milestoneItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d milestoneItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(milestoneItem)
	if !ok {
		return
	}

	cursor := renderPickerCursor(index, &m)
	var label string
	if item.id == "" {
		label = ui.Muted.Render("(none)")
	} else {
		label = ui.Secondary.Render("["+item.short+"]") + " " + item.name
	}
	renderPickerItem(w, cursor, label, item.isCurrent)
}

// milestonePickerModel is the model for the milestone picker view.
type milestonePickerModel struct {
	list             list.Model
	issueIDs         []string
	issueTitle       string
	currentMilestone string
	filterMode       bool
	width            int
	height           int
}

func newMilestonePickerModel(issueIDs []string, issueTitle, currentMilestone string, filterMode bool, milestones []*issue.Milestone, width, height int) milestonePickerModel {
	delegate := milestoneItemDelegate{}

	items := make([]list.Item, 0, len(milestones)+1)
	selectedIndex := 0

	// First entry clears the milestone.
	items = append(items, milestoneItem{id: "", name: "(none)", isCurrent: currentMilestone == ""})

	for _, ms := range milestones {
		isCurrent := ms.ID == currentMilestone
		if isCurrent {
			selectedIndex = len(items)
		}
		items = append(items, milestoneItem{
			id:          ms.ID,
			short:       ms.Short,
			name:        ms.Name,
			description: ms.Description,
			isCurrent:   isCurrent,
		})
	}

	dims := calculatePickerDimensions(width, height, defaultPickerDimensionConfig())

	title := "Select Milestone"
	if filterMode {
		title = "Filter by Milestone"
	}

	l := list.New(items, delegate, dims.ListWidth, dims.ListHeight)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.Filter = substringFilter
	l.Styles.Title = listTitleStyle
	l.Styles.TitleBar = lipgloss.NewStyle().Padding(0, 0, 0, 0)
	l.Styles.Filter.Focused.Prompt = lipgloss.NewStyle().Foreground(ui.ColorPrimary)
	l.Styles.Filter.Blurred.Prompt = lipgloss.NewStyle().Foreground(ui.ColorPrimary)
	l.Styles.Filter.Cursor.Color = ui.ColorPrimary

	if selectedIndex < len(items) {
		l.Select(selectedIndex)
	}

	return milestonePickerModel{
		list:             l,
		issueIDs:         issueIDs,
		issueTitle:       issueTitle,
		currentMilestone: currentMilestone,
		filterMode:       filterMode,
		width:            width,
		height:           height,
	}
}

func (m milestonePickerModel) Init() tea.Cmd {
	return nil
}

func (m milestonePickerModel) Update(msg tea.Msg) (milestonePickerModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		dims := calculatePickerDimensions(msg.Width, msg.Height, defaultPickerDimensionConfig())
		m.list.SetSize(dims.ListWidth, dims.ListHeight)

	case tea.KeyPressMsg:
		if m.list.FilterState() != list.Filtering {
			switch msg.String() {
			case "enter":
				if item, ok := m.list.SelectedItem().(milestoneItem); ok {
					return m, func() tea.Msg {
						return milestoneSelectedMsg{issueIDs: m.issueIDs, milestoneID: item.id, filterMode: m.filterMode}
					}
				}
			case "esc", "backspace":
				return m, func() tea.Msg {
					return closeMilestonePickerMsg{}
				}
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m milestonePickerModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var description string
	if item, ok := m.list.SelectedItem().(milestoneItem); ok && item.description != "" {
		description = item.description
	}

	var issueID string
	if len(m.issueIDs) == 1 {
		issueID = m.issueIDs[0]
	}

	return renderPickerModal(pickerModalConfig{
		Title:       "Select Milestone",
		IssueTitle:  m.issueTitle,
		IssueID:     issueID,
		ListContent: m.list.View(),
		Description: description,
		Width:       m.width,
	})
}

// ModalView returns the picker rendered as a centered modal overlay on top of the background.
func (m milestonePickerModel) ModalView(bgView string, fullWidth, fullHeight int) string {
	modal := m.View()
	return overlayModal(bgView, modal, fullWidth, fullHeight)
}
