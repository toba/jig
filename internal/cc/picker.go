package cc

import (
	"errors"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// PickAlias runs an interactive picker over the configured aliases and
// returns the chosen name, or "" if the user cancelled.
func PickAlias(c *Config, preselect string) (string, error) {
	names := c.Names()
	if len(names) == 0 {
		return "", errors.New("no aliases configured")
	}
	if len(names) == 1 {
		return names[0], nil
	}

	idx := 0
	for i, n := range names {
		if n == preselect {
			idx = i
			break
		}
	}

	m := pickerModel{
		c:      c,
		names:  names,
		cursor: idx,
	}
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return "", err
	}
	fm := final.(pickerModel)
	if fm.cancelled {
		return "", nil
	}
	return fm.names[fm.cursor], nil
}

type pickerModel struct {
	c         *Config
	names     []string
	cursor    int
	cancelled bool
	done      bool
}

func (m pickerModel) Init() tea.Cmd { return nil }

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "ctrl+c", "esc", "q":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			m.done = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.names)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

var (
	pickerHeader   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	pickerSelected = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)
	pickerDim      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func (m pickerModel) View() tea.View {
	var b strings.Builder
	b.WriteString(pickerHeader.Render("Pick a Claude alias"))
	b.WriteString("\n\n")
	for i, n := range m.names {
		a := m.c.Aliases[n]
		marker := "  "
		line := fmt.Sprintf("%s (%s)", n, a.CLI)
		if a.IsSource {
			line += pickerDim.Render("  [source]")
		}
		if i == m.cursor {
			marker = "▸ "
			line = pickerSelected.Render(line)
		}
		b.WriteString(marker)
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(pickerDim.Render("↑/↓ move • enter select • esc cancel"))
	b.WriteString("\n")
	return tea.NewView(b.String())
}
