package markdown

import (
	"bufio"
	"bytes"
	"fmt"
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

const emptyTemplate = `# {{title .Title}}

**DO NOT MODIFY THIS FILE!**
*This changelog is automatically generated by [changelog](https://github.com/moorara/changelog)*


`

const changelogTemplate = `{{range .}}## [{{.TagName}}]({{.TagURL}}) ({{time .TagTime}})

[Compare Changes]({{.CompareURL}})

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

	funcMap = template.FuncMap{
		"title": strings.Title,
		"time": func(t time.Time) string {
			return t.Format(timeLayout)
		},
	}
)

// processor implements the changelog.Processor interface for Markdown format.
type processor struct {
	logger   log.Logger
	filename string
	content  string
}

// NewProcessor creates a new changelog processor for Markdown format.
func NewProcessor(logger log.Logger, filename string) changelog.Processor {
	return &processor{
		logger:   logger,
		filename: filepath.Clean(filename),
		content:  "",
	}
}

func (p *processor) createChangelog() (*changelog.Changelog, error) {
	chlog := changelog.NewChangelog()

	tmpl, _ := template.New("changelog").Funcs(funcMap).Parse(emptyTemplate)
	buf := new(bytes.Buffer)
	_ = tmpl.Execute(buf, chlog)
	p.content = buf.String()

	p.logger.Warnf("%s not found", p.filename)
	p.logger.Info("A new changelog is created.")

	return chlog, nil
}

func (p *processor) Parse(opts changelog.ParseOptions) (*changelog.Changelog, error) {
	p.logger.Debugf("Opening %s ...", p.filename)

	f, err := os.Open(p.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return p.createChangelog()
		}
		return nil, err
	}
	defer f.Close()

	p.logger.Debugf("Parsing %s ...", p.filename)

	content := ""
	chlog := new(changelog.Changelog)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		content += fmt.Sprintln(line)

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

	p.content = content

	p.logger.Infof("Successfully parsed %s", p.filename)

	return chlog, nil
}

func (p *processor) Render(chlog *changelog.Changelog) (string, error) {
	p.logger.Debug("Updating the changelog ...")

	// ==============================> RENDER THE CONTENT FOR NEW RELEASES <==============================

	tmpl, err := template.New("changelog").Funcs(funcMap).Parse(changelogTemplate)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err = tmpl.Execute(buf, chlog.New); err != nil {
		return "", err
	}

	newContent := buf.String()

	// ==============================> UPDATE THE CHANGELOG FILE <==============================

	f, err := os.OpenFile(p.filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if i := strings.Index(p.content, "##"); i == -1 {
		p.content += newContent
	} else {
		p.content = p.content[:i] + newContent + p.content[i:]
	}

	if _, err = f.WriteString(p.content); err != nil {
		return "", err
	}

	p.logger.Infof("Successfully updated the changelog: %s", p.filename)

	return newContent, nil
}
