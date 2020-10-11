package markdown

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/moorara/changelog/internal/changelog"
)

var (
	h1Regex = regexp.MustCompile(`^# ([0-9A-Za-z-_]+)$`)
	h2Regex = regexp.MustCompile(`^## \[([0-9A-Za-z-.]+)\]\(([0-9A-Za-z-.:/]+)\) \((\d{4}-\d{2}-\d{2})\)$`)
)

// processor implements the changelog.Processor interface for Markdown format.
type processor struct {
	logger   *log.Logger
	filename string
	doc      string
}

// NewProcessor creates a new changelog processor for Markdown format.
func NewProcessor(logger *log.Logger, filename string) changelog.Processor {
	return &processor{
		logger:   logger,
		filename: filepath.Clean(filename),
	}
}

func (p *processor) Parse(opts changelog.ParseOptions) (*changelog.Changelog, error) {
	f, err := os.Open(p.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return changelog.NewChangelog(), nil
		}
		return nil, err
	}
	defer f.Close()

	chlog := new(changelog.Changelog)
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		p.doc += fmt.Sprintln(line)

		if sm := h1Regex.FindStringSubmatch(line); len(sm) == 2 {
			chlog.Title = sm[1]
		} else if sm := h2Regex.FindStringSubmatch(line); len(sm) == 4 {
			ts, err := time.Parse("2006-01-02", sm[3])
			if err != nil {
				return nil, err
			}

			chlog.Releases = append(chlog.Releases, changelog.Release{
				GitTag:    sm[1],
				URL:       sm[2],
				Timestamp: ts,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return chlog, nil
}

func (p *processor) Render(chlog *changelog.Changelog) (string, error) {
	// UPDATE THE MARKDOWN DOCUMENT

	// RENDER THE MARKDOWN DOCUMENT

	return fmt.Sprintf("%+v", chlog), nil
}
