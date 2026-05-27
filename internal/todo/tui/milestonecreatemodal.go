package tui

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/toba/jig/internal/todo/ui"
)

// openMilestoneCreateModalMsg requests opening the milestone create modal.
type openMilestoneCreateModalMsg struct{}

// milestoneCreatedMsg is sent when a new milestone is submitted.
type milestoneCreatedMsg struct {
	short string
	name  string
	due   string
}

// milestoneCreateModalModel is a multi-field modal for creating a milestone.
type milestoneCreateModalModel struct {
	inputs  []textinput.Model
	focus   int
	width   int
	height  int
	errText string
}

const (
	msFieldShort = 0
	msFieldName  = 1
	msFieldDue   = 2
)

func newMilestoneCreateModalModel(width, height int) milestoneCreateModalModel {
	mk := func(placeholder string, limit, w int) textinput.Model {
		ti := textinput.New()
		ti.Placeholder = placeholder
		ti.CharLimit = limit
		ti.SetWidth(w)
		ti.Prompt = ""
		styles := ti.Styles()
		styles.Focused.Prompt = lipgloss.NewStyle().Foreground(ui.ColorPrimary)
		styles.Focused.Text = lipgloss.NewStyle()
		styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(ui.ColorMuted)
		styles.Blurred.Prompt = lipgloss.NewStyle().Foreground(ui.ColorPrimary)
		styles.Blurred.Text = lipgloss.NewStyle()
		styles.Blurred.Placeholder = lipgloss.NewStyle().Foreground(ui.ColorMuted)
		ti.SetStyles(styles)
		return ti
	}

	inputs := []textinput.Model{
		mk("v1", 3, 6),
		mk("Milestone name...", 200, 44),
		mk("YYYY-MM-DD (optional)", 10, 12),
	}
	inputs[msFieldShort].Focus()

	return milestoneCreateModalModel{
		inputs: inputs,
		focus:  msFieldShort,
		width:  width,
		height: height,
	}
}

func (m milestoneCreateModalModel) Init() tea.Cmd { return textinput.Blink }

func (m *milestoneCreateModalModel) focusField(i int) tea.Cmd {
	m.focus = (i + len(m.inputs)) % len(m.inputs)
	var cmd tea.Cmd
	for j := range m.inputs {
		if j == m.focus {
			cmd = m.inputs[j].Focus()
		} else {
			m.inputs[j].Blur()
		}
	}
	return cmd
}

func (m milestoneCreateModalModel) Update(msg tea.Msg) (milestoneCreateModalModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return closeCreateModalMsg{} }
		case "tab", "down":
			cmd := m.focusField(m.focus + 1)
			return m, cmd
		case "shift+tab", "up":
			cmd := m.focusField(m.focus - 1)
			return m, cmd
		case "enter":
			short := m.inputs[msFieldShort].Value()
			name := m.inputs[msFieldName].Value()
			due := m.inputs[msFieldDue].Value()
			if short == "" || name == "" {
				m.errText = "short name and name are required"
				return m, nil
			}
			return m, func() tea.Msg {
				return milestoneCreatedMsg{short: short, name: name, due: due}
			}
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
	return m, cmd
}

func (m milestoneCreateModalModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}
	modalWidth := max(48, min(64, m.width*60/100))

	header := lipgloss.NewStyle().Bold(true).Render("Create New Milestone")

	field := func(label string, idx int) string {
		labelStyle := ui.Muted
		if m.focus == idx {
			labelStyle = ui.Primary
		}
		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ui.ColorMuted).
			Padding(0, 1).
			Render(m.inputs[idx].View())
		return labelStyle.Render(label) + "\n" + box
	}

	body := field("Short (2-3 chars)", msFieldShort) + "\n" +
		field("Name", msFieldName) + "\n" +
		field("Due date", msFieldDue)

	help := helpKeyStyle.Render("tab") + " " + helpStyle.Render("next field") + "  " +
		helpKeyStyle.Render("enter") + " " + helpStyle.Render("create") + "  " +
		helpKeyStyle.Render("esc") + " " + helpStyle.Render("cancel")

	content := header + "\n\n" + body + "\n"
	if m.errText != "" {
		content += "\n" + ui.Danger.Render(m.errText) + "\n"
	}
	content += "\n" + help

	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorPrimary).
		Padding(1, 2).
		Width(modalWidth)
	return border.Render(content)
}

// ModalView returns the modal rendered as a centered overlay.
func (m milestoneCreateModalModel) ModalView(bgView string, fullWidth, fullHeight int) string {
	return overlayModal(bgView, m.View(), fullWidth, fullHeight)
}
