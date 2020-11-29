package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	flags "github.com/jessevdk/go-flags"
)

type opts struct {
	MDLocation       string `short:"l" long:"mdlocation" env:"MDLOCATION" required:"true" description:"location of the markdown file"`
	OutputLocation   string `short:"o" long:"output_location" env:"OUTPUT_LOCATION" default:"./" description:"location of the output rendered pdf file"`
	TemplateLocation string `short:"t" long:"template_location" env:"TEMPLATE_LOCATION" default:"tmpl.html" description:"location of the template file"`
	RenderHTML       bool   `short:"h" env:"RENDER_HTML" description:"render html file"`
}

func main() {
	fmt.Println("gencv")
	o := opts{}
	if _, err := flags.Parse(&o); err != nil {
		os.Exit(1)
		return
	}

	// reading and parsing source file
	var f *os.File
	var err error
	if f, err = os.Open(o.MDLocation); err != nil {
		fmt.Printf("failed to open markdown source file at location %s: %v", o.MDLocation, err)
		return
	}
	cv := parseMDFile(f)
	if err = f.Close(); err != nil {
		fmt.Printf("failed to close markdown source file at location %s: %v", o.MDLocation, err)
		return
	}
	out := renderHtml(o.TemplateLocation, cv)

	// render html if needed
	if !o.RenderHTML {
		return
	}

	if err = ioutil.WriteFile(path.Join(o.OutputLocation, "cv.html"), []byte(out), 0644); err != nil {
		fmt.Printf("failed to write the rendered html to location %s: %v", path.Join(o.OutputLocation, "cv.html"), err)
		return
	}

}

type cvTmpl struct {
	Header   string
	Avatar   string
	Contacts string
	Body     string
}

type cv struct {
	header   string
	avatar   string
	contacts string
	body     string
}

func renderHtml(tmplLoc string, cv cv) string {
	// rendering markdown
	var extensions = parser.NoIntraEmphasis |
		parser.Tables |
		parser.FencedCode |
		parser.Autolink |
		parser.Strikethrough |
		parser.SpaceHeadings |
		parser.BackslashLineBreak |
		parser.EmptyLinesBreakList

	renderer := html.NewRenderer(html.RendererOptions{})

	tmplBody := cvTmpl{}
	tmplBody.Header = string(markdown.ToHTML([]byte(cv.header), parser.NewWithExtensions(extensions), renderer))
	tmplBody.Avatar = string(markdown.ToHTML([]byte(cv.avatar), parser.NewWithExtensions(extensions), renderer))
	tmplBody.Contacts = string(markdown.ToHTML([]byte(cv.contacts), parser.NewWithExtensions(extensions), renderer))
	tmplBody.Body = string(markdown.ToHTML([]byte(cv.body), parser.NewWithExtensions(extensions), renderer))

	var f *os.File
	var err error

	// loading template
	if f, err = os.Open(tmplLoc); err != nil {
		log.Panicf("failed to open template file at location %s: %v", tmplLoc, err)
		return ""
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Panicf("failed to read template file at location %s: %v", tmplLoc, err)
		return ""
	}

	tmpl, err := template.New("tmpl").Parse(string(b))
	if err != nil {
		log.Panicf("failed to parse template file at location %s: %v", tmplLoc, err)
		return ""
	}

	out := &bytes.Buffer{}

	if err = tmpl.Execute(out, tmplBody); err != nil {
		log.Panicf("failed to execute template: %v", err)
		return ""
	}
	return out.String()
}

func parseMDFile(f *os.File) (cv cv) {
	status := 0 // 0 - reading head, 1 - reading avatar, 2 - reading contacts, 3 - reading body

	scn := bufio.NewScanner(f)
	for scn.Scan() {
		line := scn.Text()
		if strings.TrimSpace(line) == "---" {
			status++
			continue
		}

		switch status {
		case 0:
			cv.header += line // + "\n"
		case 1:
			cv.avatar += line + "\n"
		case 2:
			cv.contacts += line + "\n"
		case 3:
			cv.body += line + "\n"
		}
	}
	return cv
}
