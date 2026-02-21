package github

import "time"

// Commit represents a commit from the GitHub API.
type Commit struct {
	SHA     string       `json:"sha"`
	Message string       `json:"-"` // Extracted from commit.message
	Author  string       `json:"-"` // Extracted from author.login or commit.author.name
	Date    time.Time    `json:"-"` // Extracted from commit.author.date
	RawCommit rawCommit  `json:"commit"`
	RawAuthor *rawAuthor `json:"author"`
	Files     []File     `json:"files,omitempty"`
}

type rawCommit struct {
	Message   string    `json:"message"`
	RawAuthor rawDate   `json:"author"`
}

type rawDate struct {
	Name string    `json:"name"`
	Date time.Time `json:"date"`
}

type rawAuthor struct {
	Login string `json:"login"`
}

// Normalize populates the top-level fields from raw API data.
func (c *Commit) Normalize() {
	c.Message = firstLine(c.RawCommit.Message)
	c.Date = c.RawCommit.RawAuthor.Date
	if c.RawAuthor != nil && c.RawAuthor.Login != "" {
		c.Author = c.RawAuthor.Login
	} else {
		c.Author = c.RawCommit.RawAuthor.Name
	}
}

// CompareResponse represents the response from the compare API endpoint.
type CompareResponse struct {
	Status      string   `json:"status"`
	AheadBy     int      `json:"ahead_by"`
	TotalCommits int    `json:"total_commits"`
	Commits     []Commit `json:"commits"`
	Files       []File   `json:"files"`
}

// File represents a changed file from the GitHub API.
type File struct {
	Filename string `json:"filename"`
	Status   string `json:"status"`
}

func firstLine(s string) string {
	for i, c := range s {
		if c == '\n' {
			return s[:i]
		}
	}
	return s
}
