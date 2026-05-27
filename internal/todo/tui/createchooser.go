package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/toba/jig/internal/todo/ui"
)

// openCreateChooserMsg requests opening the "create issue or milestone" chooser.
type openCreateChooserMsg struct{}

// createChooserModel is a small modal that lets the user choose what to create.
type createChooserModel struct {
	width  int
	height int
}

func newCreateChooserModel(width, height int) createChooserModel {
	return createChooserModel{width: width, height: height}
}

func (m createChooserModel) Init() tea.Cmd { return nil }

func (m createChooserModel) Update(msg tea.Msg) (createChooserModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		switch msg.String() {
		case "i":
			return m, func() tea.Msg { return openCreateModalMsg{} }
		case "m":
			return m, func() tea.Msg { return openMilestoneCreateModalMsg{} }
		case "esc", "q":
			return m, func() tea.Msg { return closeCreateModalMsg{} }
		}
	}
	return m, nil
}

func (m createChooserModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}
	modalWidth := max(36, min(50, m.width*40/100))

	header := lipgloss.NewStyle().Bold(true).Render("Create")
	options := helpKeyStyle.Render("i") + " " + helpStyle.Render("Issue") + "    " +
		helpKeyStyle.Render("m") + " " + helpStyle.Render("Milestone")
	help := helpKeyStyle.Render("esc") + " " + helpStyle.Render("cancel")

	content := header + "\n\n" + options + "\n\n" + help

	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorPrimary).
		Padding(1, 2).
		Width(modalWidth)
	return border.Render(content)
}

// ModalView returns the chooser rendered as a centered overlay.
func (m createChooserModel) ModalView(bgView string, fullWidth, fullHeight int) string {
	return overlayModal(bgView, m.View(), fullWidth, fullHeight)
}
