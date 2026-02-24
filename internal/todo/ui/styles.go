package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/toba/jig/internal/todo/config"
)

// Color palette
var (
	ColorPrimary   = lipgloss.Color("#7C3AED") // Purple
	ColorSecondary = lipgloss.Color("#6B7280") // Gray
	ColorSuccess   = lipgloss.Color("#10B981") // Green
	ColorWarning   = lipgloss.Color("#F59E0B") // Amber
	ColorDanger    = lipgloss.Color("#EF4444") // Red
	ColorMuted     = lipgloss.Color("#9CA3AF") // Light gray
	ColorSubtle    = lipgloss.Color("#555555") // Dark gray (for tree lines)
	ColorBlue      = lipgloss.Color("#3B82F6") // Blue
	ColorCyan      = lipgloss.Color("14")      // Bright Cyan (ANSI)
)

// NamedColors maps color names to lipgloss colors.
var NamedColors = map[string]lipgloss.Color{
	"green":  ColorSuccess,
	"yellow": ColorWarning,
	"red":    ColorDanger,
	"gray":   ColorSecondary,
	"grey":   ColorSecondary,
	"blue":   ColorBlue,
	"purple": ColorPrimary,
	"cyan":   ColorCyan,
}

// ResolveColor converts a color name or hex code to a lipgloss.Color.
func ResolveColor(color string) lipgloss.Color {
	if strings.HasPrefix(color, "#") {
		return lipgloss.Color(color)
	}
	if c, ok := NamedColors[strings.ToLower(color)]; ok {
		return c
	}
	return ColorMuted
}

// IsValidColor returns true if the color is a valid named color or hex code.
func IsValidColor(color string) bool {
	if strings.HasPrefix(color, "#") {
		// Valid hex: #RGB or #RRGGBB
		return len(color) == 4 || len(color) == 7
	}
	_, ok := NamedColors[strings.ToLower(color)]
	return ok
}

// Status badge styles (for inline use, like in show command)
var (
	StatusOpen = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fff")).
			Background(ColorSuccess).
			Padding(0, 1).
			Bold(true)

	StatusDone = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fff")).
			Background(ColorSecondary).
			Padding(0, 1)

	StatusInProgress = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#fff")).
				Background(ColorWarning).
				Padding(0, 1).
				Bold(true)
)

// Status text styles (for table use, no background/padding)
var (
	StatusOpenText       = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
	StatusDoneText       = lipgloss.NewStyle().Foreground(ColorSecondary)
	StatusInProgressText = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true)
)

// TagBadge is the style for tag badges - black text on gray background.
var TagBadge = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#000")).
	Background(ColorMuted).
	Padding(0, 1)

// RenderTag renders a single tag as a badge
func RenderTag(tag string) string {
	return TagBadge.Render(tag)
}

// RenderTags renders multiple tags as badges separated by spaces
func RenderTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	rendered := make([]string, len(tags))
	for i, tag := range tags {
		rendered[i] = RenderTag(tag)
	}
	return strings.Join(rendered, " ")
}

// RenderTagsCompact renders tags for list views with a max count.
// Shows up to maxTags badges, with "+N" indicator if there are more.
// Tags longer than 12 chars are truncated.
func RenderTagsCompact(tags []string, maxTags int) string {
	if len(tags) == 0 {
		return ""
	}
	if maxTags <= 0 {
		maxTags = 1
	}

	showTags := tags
	var extra int
	if len(tags) > maxTags {
		showTags = tags[:maxTags]
		extra = len(tags) - maxTags
	}

	rendered := make([]string, len(showTags))
	for i, tag := range showTags {
		// Truncate long tags
		displayTag := tag
		if len(displayTag) > maxTagDisplayLen {
			displayTag = displayTag[:truncatedTagLen] + ".."
		}
		rendered[i] = RenderTag(displayTag)
	}

	result := strings.Join(rendered, " ")
	if extra > 0 {
		result += Muted.Render(fmt.Sprintf(" +%d", extra))
	}
	return result
}

// Text styles
var (
	Bold      = lipgloss.NewStyle().Bold(true)
	Muted     = lipgloss.NewStyle().Foreground(ColorMuted)
	Primary   = lipgloss.NewStyle().Foreground(ColorPrimary)
	Success   = lipgloss.NewStyle().Foreground(ColorSuccess)
	Warning   = lipgloss.NewStyle().Foreground(ColorWarning)
	Danger    = lipgloss.NewStyle().Foreground(ColorDanger)
	Secondary = lipgloss.NewStyle().Foreground(ColorSecondary)
)

// ID style - distinctive for issue IDs
var ID = lipgloss.NewStyle().
	Foreground(ColorPrimary).
	Bold(true)

// TreeLine style - subtle for tree connectors
var TreeLine = lipgloss.NewStyle().Foreground(ColorSubtle)

// Title style
var Title = lipgloss.NewStyle().Bold(true)

// Path style - subdued
var Path = lipgloss.NewStyle().Foreground(ColorMuted)

// Header style for section headers
var Header = lipgloss.NewStyle().
	Foreground(ColorPrimary).
	Bold(true).
	MarginBottom(1)

// RenderStatus returns a styled status badge based on the status string (legacy, uses hardcoded colors)
func RenderStatus(status string) string {
	switch status {
	case config.StatusReady, config.StatusDraft:
		return StatusOpen.Render(status)
	case config.StatusCompleted, config.StatusScrapped:
		return StatusDone.Render(status)
	case config.StatusInProgress, config.StatusReview, "in_progress":
		return StatusInProgress.Render(status)
	default:
		return Muted.Render(status)
	}
}

// RenderStatusText returns styled status text (for tables, no background) (legacy, uses hardcoded colors)
func RenderStatusText(status string) string {
	switch status {
	case config.StatusReady, config.StatusDraft:
		return StatusOpenText.Render(status)
	case config.StatusCompleted, config.StatusScrapped:
		return StatusDoneText.Render(status)
	case config.StatusInProgress, config.StatusReview, "in_progress":
		return StatusInProgressText.Render(config.StatusInProgress)
	default:
		return Muted.Render(status)
	}
}

// StatusIcon returns a Unicode icon for the given status.
func StatusIcon(status string) string {
	switch status {
	case "draft":
		return "△" // "⬜︎"∷∴▲
	case "ready":
		return "○"
	case "in-progress":
		return "◔"
	case "review":
		return "◈"
	case "completed":
		return "✔"
	case "scrapped":
		return "✖"
	default:
		return "○"
	}
}

// RenderStatusWithColor returns a styled status badge using the specified color.
func RenderStatusWithColor(status, color string, isArchiveStatus bool) string {
	c := ResolveColor(color)
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#fff")).
		Background(c).
		Padding(0, 1)

	if !isArchiveStatus {
		style = style.Bold(true)
	}

	return style.Render(StatusIcon(status) + " " + status)
}

// RenderStatusIconWithColor returns styled status text (for tables) using the specified color.
func RenderStatusIconWithColor(status, color string, isArchiveStatus bool) string {
	c := ResolveColor(color)
	style := lipgloss.NewStyle().Foreground(c)

	if !isArchiveStatus {
		style = style.Bold(true)
	}

	return style.Render(StatusIcon(status))
}

// RenderStatusIconAndLabel returns a styled status icon followed by the status name.
// Used in contexts like the status picker where the label should be visible.
func RenderStatusIconAndLabel(status, color string, isArchiveStatus bool) string {
	c := ResolveColor(color)
	style := lipgloss.NewStyle().Foreground(c)

	if !isArchiveStatus {
		style = style.Bold(true)
	}

	return style.Render(StatusIcon(status) + " " + status)
}

// TypeAbbrev returns a two-letter abbreviation for a type name.
// The first letter is uppercase and the second is lowercase (e.g., "feature" → "Fe").
func TypeAbbrev(typeName string) string {
	if typeName == "" {
		return ""
	}
	r := []rune(typeName)
	if len(r) == 1 {
		return strings.ToUpper(string(r[0]))
	}
	return strings.ToUpper(string(r[0])) + strings.ToLower(string(r[1]))
}

// RenderTypeText returns styled type text using the specified color.
// If color is empty, uses muted styling.
func RenderTypeText(typeName, color string) string {
	abbrev := TypeAbbrev(typeName)
	if abbrev == "" {
		return ""
	}
	if color == "" {
		return Muted.Render(abbrev)
	}
	c := ResolveColor(color)
	return lipgloss.NewStyle().Foreground(c).Render(abbrev)
}

// RenderTypeWithColor returns a styled type badge with colored background.
func RenderTypeWithColor(typeName, color string) string {
	abbrev := TypeAbbrev(typeName)
	if abbrev == "" {
		return ""
	}
	c := ResolveColor(color)
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#fff")).
		Background(c).
		Bold(true).
		Padding(0, 1)
	return style.Render(abbrev)
}

// RenderPriorityWithColor returns a styled priority badge using the specified color.
func RenderPriorityWithColor(priority, color string) string {
	if priority == "" {
		return ""
	}
	c := ResolveColor(color)
	style := lipgloss.NewStyle().
		Foreground(c).
		Bold(priority == config.PriorityCritical || priority == config.PriorityHigh)
	return style.Render("[" + priority + "]")
}

// RenderPriorityText returns styled priority text for tables.
func RenderPriorityText(priority, color string) string {
	if priority == "" {
		return ""
	}
	c := ResolveColor(color)
	style := lipgloss.NewStyle().Foreground(c)
	if priority == config.PriorityCritical || priority == config.PriorityHigh {
		style = style.Bold(true)
	}
	return style.Render(priority)
}

// GetPrioritySymbol returns the raw symbol for a priority without styling.
// Returns empty string for normal/empty priority.
func GetPrioritySymbol(priority string) string {
	switch priority {
	case config.PriorityCritical:
		return "‼"
	case config.PriorityHigh:
		return "!"
	case config.PriorityLow:
		return "↓"
	case config.PriorityDeferred:
		return "→"
	default:
		return ""
	}
}

// RenderPrioritySymbol returns a compact symbol for priority (used in TUI).
// Returns empty string for normal/empty priority.
func RenderPrioritySymbol(priority, color string) string {
	symbol := GetPrioritySymbol(priority)
	if symbol == "" {
		return ""
	}

	c := ResolveColor(color)
	style := lipgloss.NewStyle().Foreground(c)
	if priority == config.PriorityCritical || priority == config.PriorityHigh {
		style = style.Bold(true)
	}
	return style.Render(symbol)
}

// IssueRowConfig holds configuration for rendering an issue row
type IssueRowConfig struct {
	StatusColor   string
	TypeColor     string
	PriorityColor string
	Priority      string // Priority value (critical, high, normal, low, deferred)
	IsArchive     bool
	MaxTitleWidth int  // 0 means no truncation
	ShowCursor    bool // Show selection cursor
	IsSelected    bool
	IsMarked      bool     // Marked for multi-select batch operations
	Tags          []string // Tags to display (optional)
	ShowTags      bool     // Whether to show tags column
	TagsColWidth  int      // Width of tags column (0 = default)
	MaxTags       int      // Max tags to show (0 = default of 1)
	TreePrefix    string   // Tree prefix (e.g., "├─" or "  └─") to prepend to ID
	Dimmed        bool     // Render row dimmed (for unmatched ancestor issues in tree)
	IDColWidth    int      // Width of ID column (0 = default of ColWidthID)
	HasDueDate    bool     // Show hourglass indicator for issues with due dates
}

// Base column widths for issue lists (minimum sizes)
const (
	ColWidthID     = 12
	ColWidthStatus = 3
	ColWidthType   = 2
	ColWidthTags   = 24
)

// Responsive tag layout thresholds
const (
	minWidthForTags = 140 // Minimum terminal width to show tags column
	minTitleWidth   = 50  // Minimum reserved title width before allocating to tags

	tagColWidthXL     = 70 // Tag column width at 80+ space
	tagColWidthLarge  = 55 // Tag column width at 60+ space
	tagColWidthMedium = 42 // Tag column width at 45+ space
	tagColWidthSmall  = 32 // Tag column width at 35+ space

	maxTagsXL     = 5 // Max tags shown at 80+ space
	maxTagsLarge  = 4 // Max tags shown at 60+ space
	maxTagsMedium = 3 // Max tags shown at 45+ space
	maxTagsSmall  = 2 // Max tags shown at 35+ space

	maxTagDisplayLen = 12 // Truncate tags longer than this
	truncatedTagLen  = 10 // Display length after truncation (+ "..")
)

// ResponsiveColumns holds calculated column widths based on available space
type ResponsiveColumns struct {
	ID       int
	Status   int
	Type     int
	Tags     int
	MaxTags  int // How many tags to show
	ShowTags bool
}

// CalculateResponsiveColumns determines column widths based on available width.
// Prioritizes title width - tags are only shown when there's plenty of room.
func CalculateResponsiveColumns(totalWidth int, hasTags bool) ResponsiveColumns {
	cols := ResponsiveColumns{
		ID:       ColWidthID,
		Status:   ColWidthStatus,
		Type:     ColWidthType,
		Tags:     0,
		MaxTags:  0,
		ShowTags: false,
	}

	// Don't show tags in narrow viewports - prioritize title space
	// Only consider showing tags if terminal is wide enough
	if !hasTags || totalWidth < minWidthForTags {
		return cols
	}

	// Base usage: cursor (2) + ID (12) + status (5) + type (2) = 21
	cursorWidth := 2
	baseWidth := cursorWidth + cols.ID + cols.Status + cols.Type
	available := totalWidth - baseWidth

	// Reserve generous space for title, then allocate remaining to tags
	spaceForTags := available - minTitleWidth

	if spaceForTags >= ColWidthTags {
		cols.ShowTags = true

		if spaceForTags >= 80 {
			cols.Tags = tagColWidthXL
			cols.MaxTags = maxTagsXL
		} else if spaceForTags >= 60 {
			cols.Tags = tagColWidthLarge
			cols.MaxTags = maxTagsLarge
		} else if spaceForTags >= 45 {
			cols.Tags = tagColWidthMedium
			cols.MaxTags = maxTagsMedium
		} else if spaceForTags >= 35 {
			cols.Tags = tagColWidthSmall
			cols.MaxTags = maxTagsSmall
		} else {
			cols.Tags = ColWidthTags
			cols.MaxTags = 1
		}
	}

	return cols
}

// RenderIssueRow renders an issue as a single row with ID, Type, Status, Tags (optional), Title
func RenderIssueRow(id, status, typeName, title string, cfg IssueRowConfig) string {
	// Column styles - use responsive widths if provided
	idColWidth := ColWidthID
	if cfg.IDColWidth > 0 {
		idColWidth = cfg.IDColWidth
	}
	typeStyle := lipgloss.NewStyle().Width(ColWidthType)
	statusStyle := lipgloss.NewStyle().Width(ColWidthStatus)

	tagsColWidth := ColWidthTags
	if cfg.TagsColWidth > 0 {
		tagsColWidth = cfg.TagsColWidth
	}
	tagsStyle := lipgloss.NewStyle().Width(tagsColWidth)

	maxTags := 1
	if cfg.MaxTags > 0 {
		maxTags = cfg.MaxTags
	}

	// Highlight style for marked rows
	highlightStyle := lipgloss.NewStyle().Foreground(ColorWarning)

	// Build ID column with manual padding
	// (lipgloss Width() doesn't correctly handle Unicode box-drawing characters)
	var idCol string
	// Calculate visual width: tree prefix (in runes) + ID length
	visualWidth := len([]rune(cfg.TreePrefix)) + len(id)
	padding := ""
	if idColWidth > visualWidth {
		padding = strings.Repeat(" ", idColWidth-visualWidth)
	}
	if cfg.Dimmed {
		idCol = Muted.Render(cfg.TreePrefix) + Muted.Render(id) + padding
	} else if cfg.IsMarked {
		// Only highlight the ID when marked
		idCol = highlightStyle.Render(cfg.TreePrefix) + highlightStyle.Render(id) + padding
	} else {
		idCol = TreeLine.Render(cfg.TreePrefix) + ID.Render(id) + padding
	}

	var typeCol string
	if typeName != "" {
		if cfg.Dimmed {
			typeCol = typeStyle.Render(Muted.Render(typeName))
		} else {
			typeCol = typeStyle.Render(RenderTypeText(typeName, cfg.TypeColor))
		}
	} else {
		typeCol = typeStyle.Render("")
	}

	var statusCol string
	if cfg.Dimmed {
		statusCol = statusStyle.Render(Muted.Render(StatusIcon(status)))
	} else {
		statusCol = statusStyle.Render(RenderStatusIconWithColor(status, cfg.StatusColor, cfg.IsArchive))
	}

	// Tags column (optional)
	var tagsCol string
	if cfg.ShowTags {
		if cfg.Dimmed {
			if len(cfg.Tags) > 0 {
				tagsCol = tagsStyle.Render(Muted.Render(cfg.Tags[0]))
			} else {
				tagsCol = tagsStyle.Render("")
			}
		} else {
			tagsCol = tagsStyle.Render(RenderTagsCompact(cfg.Tags, maxTags))
		}
	}

	// Priority symbol (prepended to title)
	var prioritySymbol string
	if !cfg.Dimmed {
		prioritySymbol = RenderPrioritySymbol(cfg.Priority, cfg.PriorityColor)
		if prioritySymbol != "" {
			prioritySymbol += " "
		}
	}

	// Due date hourglass indicator (after priority symbol, before title)
	var dueDateSymbol string
	if !cfg.Dimmed && cfg.HasDueDate {
		dueDateSymbol = lipgloss.NewStyle().Foreground(ColorDanger).Render("⏳") + " "
	}

	// Title (truncate if needed, accounting for priority symbol and due date width)
	displayTitle := title
	titleColWidth := cfg.MaxTitleWidth // Save original for padding
	maxWidth := cfg.MaxTitleWidth
	if maxWidth > 0 && prioritySymbol != "" {
		maxWidth -= 2 // Account for symbol + space
	}
	if maxWidth > 0 && dueDateSymbol != "" {
		maxWidth -= 3 // Account for hourglass (2 cells wide) + space
	}
	if maxWidth > 3 && len(title) > maxWidth {
		displayTitle = title[:maxWidth-3] + "..."
	} else if maxWidth > 0 && maxWidth <= 3 && len(title) > maxWidth {
		displayTitle = title[:maxWidth]
	}

	// Cursor and title styling
	var cursor string
	var titleStyled string
	if cfg.ShowCursor {
		if cfg.IsSelected {
			cursor = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("▌")
			titleStyled = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(displayTitle)
		} else {
			cursor = " "
			if cfg.Dimmed {
				titleStyled = Muted.Render(displayTitle)
			} else {
				titleStyled = displayTitle
			}
		}
	} else {
		cursor = ""
		if cfg.Dimmed {
			titleStyled = Muted.Render(displayTitle)
		} else {
			titleStyled = displayTitle
		}
	}

	if cfg.ShowTags {
		// Pad title column to fixed width so tags align in a column
		// Calculate padding needed: titleColWidth - (priority symbol width + title length)
		titleLen := len(displayTitle)
		if prioritySymbol != "" {
			titleLen += 2 // symbol + space
		}
		if dueDateSymbol != "" {
			titleLen += 3 // hourglass (2 cells wide) + space
		}
		padding := ""
		if titleColWidth > titleLen {
			padding = strings.Repeat(" ", titleColWidth-titleLen)
		}
		return cursor + idCol + " " + typeCol + " " + statusCol + " " + prioritySymbol + dueDateSymbol + titleStyled + padding + " " + tagsCol
	}
	return cursor + idCol + " " + typeCol + " " + statusCol + " " + prioritySymbol + dueDateSymbol + titleStyled
}
