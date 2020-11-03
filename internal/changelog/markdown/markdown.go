package markdown

import (
	"bufio"
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/moorara/changelog/internal/changelog"
	"github.com/moorara/changelog/pkg/log"
)

const timeLayout = "2006-01-02"

const markdownTemplate = `
{{range .}}## [{{.TagName}}]({{.TagURL}}) ({{time .TagTime}})

{{range .IssueGroups}}**{{title .Title}}:**

{{range .Issues}}  - {{.Title}} [#{{.Number}}]({{.URL}}) ({{if ne .OpenedBy.Username .ClosedBy.Username}}[{{.OpenedBy.Username}}]({{.OpenedBy.URL}}), {{end}}[{{.ClosedBy.Username}}]({{.ClosedBy.URL}}))
{{end}}
{{end}}{{range .MergeGroups}}**{{title .Title}}:**

{{range .Merges}}  - {{.Title}} [#{{.Number}}]({{.URL}}) ({{if ne .OpenedBy.Username .MergedBy.Username}}[{{.OpenedBy.Username}}]({{.OpenedBy.URL}}), {{end}}[{{.MergedBy.Username}}]({{.MergedBy.URL}}))
{{end}}
{{end}}
{{end}}`

var (
	h1Regex = regexp.MustCompile(`^# ([0-9A-Za-z-_]+)$`)
	h2Regex = regexp.MustCompile(`^## \[([0-9A-Za-z-.]+)\]\(([0-9A-Za-z-.:/]+)\) \((\d{4}-\d{2}-\d{2})\)$`)
)

// processor implements the changelog.Processor interface for Markdown format.
type processor struct {
	logger   log.Logger
	filename string
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

		if sm := h1Regex.FindStringSubmatch(line); len(sm) == 2 {
			chlog.Title = sm[1]
		} else if sm := h2Regex.FindStringSubmatch(line); len(sm) == 4 {
			t, err := time.Parse(timeLayout, sm[3])
			if err != nil {
				return nil, err
			}

			chlog.Existing = append(chlog.Existing, changelog.Release{
				TagName: sm[1],
				TagURL:  sm[2],
				TagTime: t,
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

	// ==============================> RENDER THE CONTENT FOR NEW RELEASES <==============================

	tmpl := template.New("help")
	tmpl = tmpl.Funcs(template.FuncMap{
		"title": strings.Title,
		"time": func(t time.Time) string {
			return t.Format(timeLayout)
		},
	})

	tmpl, err := tmpl.Parse(markdownTemplate)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, chlog.New)
	if err != nil {
		return "", err
	}

	content := buf.String()

	// ==============================> UPDATE THE CHANGELOG FILE <==============================

	// TODO:

	p.logger.Infof("Successfully rendered the changelog.")

	return content, nil
}
