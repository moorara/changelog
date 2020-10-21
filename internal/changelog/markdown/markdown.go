package markdown

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/moorara/changelog/internal/changelog"
	"github.com/moorara/changelog/pkg/log"
)

var (
	h1Regex = regexp.MustCompile(`^# ([0-9A-Za-z-_]+)$`)
	h2Regex = regexp.MustCompile(`^## \[([0-9A-Za-z-.]+)\]\(([0-9A-Za-z-.:/]+)\) \((\d{4}-\d{2}-\d{2})\)$`)
)

// processor implements the changelog.Processor interface for Markdown format.
type processor struct {
	logger   log.Logger
	filename string
	doc      string
}

// NewProcessor creates a new changelog processor for Markdown format.
func NewProcessor(logger log.Logger, filename string) changelog.Processor {
	return &processor{
		logger:   logger,
		filename: filepath.Clean(filename),
	}
}

func (p *processor) Parse(opts changelog.ParseOptions) (*changelog.Changelog, error) {
	p.logger.Debugf("Opening %s ...", p.filename)

	f, err := os.Open(p.filename)
	if err != nil {
		if os.IsNotExist(err) {
			p.logger.Warnf("%s not found", p.filename)
			p.logger.Info("A new changelog is created.")
			return changelog.NewChangelog(), nil
		}
		return nil, err
	}
	defer f.Close()

	p.logger.Debugf("Parsing %s ...", p.filename)

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

	p.logger.Infof("Successfully parsed %s", p.filename)

	return chlog, nil
}

func (p *processor) Render(chlog *changelog.Changelog) (string, error) {
	p.logger.Debugf("Rendering the changelog ...")

	// UPDATE THE MARKDOWN DOCUMENT

	// RENDER THE MARKDOWN DOCUMENT

	p.logger.Infof("Successfully rendered the changelog.")

	return fmt.Sprintf("%+v", chlog), nil
}
