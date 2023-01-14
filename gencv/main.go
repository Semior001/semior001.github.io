package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"

	html2pdf "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	flags "github.com/jessevdk/go-flags"
)

type opts struct {
	MDLocation       string `short:"l" long:"mdlocation" env:"MDLOCATION" required:"true" description:"location of the markdown file"`
	PDFLocation      string `short:"o" long:"output_pdf_location" env:"OUTPUT_PDF_LOCATION" required:"true" description:"location of the output rendered pdf file"`
	HTMLLocation     string `short:"r" long:"output_html_location" env:"OUTPUT_HTML_LOCATION" description:"location of the html render file"`
	TemplateLocation string `short:"t" long:"template_location" env:"TEMPLATE_LOCATION" default:"tmpl.html" description:"location of the template file"`
}

func main() {
	fmt.Println("gencv")
	o := opts{}
	if _, err := flags.Parse(&o); err != nil {
		os.Exit(1)
	}

	// reading and parsing source file
	var f *os.File
	var err error
	if f, err = os.Open(o.MDLocation); err != nil {
		log.Fatalf("failed to open markdown source file at location %s: %v", o.MDLocation, err)
	}
	cv := parseMDFile(f)
	if err = f.Close(); err != nil {
		log.Fatalf("failed to close markdown source file at location %s: %v", o.MDLocation, err)
	}

	templater, err := NewTemplater(o.TemplateLocation)
	if err != nil {
		log.Fatalf("failed to initialize templater: %v", err)
	}

	rd, err := templater.RenderHTML(cv)
	if err != nil {
		log.Fatalf("failed to render html: %v", err)
	}

	out, err := io.ReadAll(rd)
	if err != nil {
		log.Fatalf("failed to read generated html file: %v", err)
	}

	// generating pdf
	pdfg, err := html2pdf.NewPDFGenerator()
	if err != nil {
		log.Fatalf("failed to instantiate pdf generator: %v", err)
	}

	pdfg.Dpi.Set(400)
	pdfg.Orientation.Set(html2pdf.OrientationPortrait)
	pdfg.Grayscale.Set(false)
	pdfg.PageSize.Set(html2pdf.PageSizeA4)

	page := html2pdf.NewPageReader(bytes.NewReader(out))

	page.Zoom.Set(1.3)

	pdfg.AddPage(page)

	if err = pdfg.Create(); err != nil {
		log.Fatalf("failed to create pdf: %v", err)
	}

	if err = pdfg.WriteFile(o.PDFLocation); err != nil {
		log.Fatalf("failed to write pdf to file at location %s: %v", o.PDFLocation, err)
	}

	// render html if needed
	if o.HTMLLocation == "" {
		return
	}

	if err = ioutil.WriteFile(o.HTMLLocation, out, 0666); err != nil {
		log.Fatalf("failed to write the rendered html to location %s: %v", o.PDFLocation, err)
	}
}

type Templater struct {
	tmpl     *template.Template
	renderer *html.Renderer
	pdfg     *html2pdf.PDFGenerator
}

func NewTemplater(tmplLoc string) (*Templater, error) {
	res := &Templater{}

	// loading template
	f, err := os.Open(tmplLoc)
	if err != nil {
		return nil, fmt.Errorf("open template file at location %s: %w", tmplLoc, err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read template file at location: %s, %w", tmplLoc, err)
	}

	if res.tmpl, err = template.New("tmpl").Parse(string(b)); err != nil {
		return nil, fmt.Errorf("parse template at location %s: %w", tmplLoc, err)
	}

	// making renderers and generators
	if res.pdfg, err = html2pdf.NewPDFGenerator(); err != nil {
		return nil, fmt.Errorf("make pdf generator: %w", err)
	}

	res.renderer = html.NewRenderer(html.RendererOptions{})

	return res, nil
}

var extensions = parser.NoIntraEmphasis |
	parser.Tables |
	parser.FencedCode |
	parser.Autolink |
	parser.Strikethrough |
	parser.SpaceHeadings |
	parser.BackslashLineBreak |
	parser.EmptyLinesBreakList

func (t *Templater) RenderHTML(cv cv) (io.Reader, error) {
	tmplBody := struct {
		Header   string
		Avatar   string
		Contacts string
		Body     string
	}{}
	tmplBody.Header = string(markdown.ToHTML([]byte(cv.header), parser.NewWithExtensions(extensions), t.renderer))
	tmplBody.Avatar = string(markdown.ToHTML([]byte(cv.avatar), parser.NewWithExtensions(extensions), t.renderer))
	tmplBody.Contacts = string(markdown.ToHTML([]byte(cv.contacts), parser.NewWithExtensions(extensions), t.renderer))
	tmplBody.Body = string(markdown.ToHTML([]byte(cv.body), parser.NewWithExtensions(extensions), t.renderer))

	out := &bytes.Buffer{}

	if err := t.tmpl.Execute(out, tmplBody); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return out, nil
}

type cv struct {
	header   string
	avatar   string
	contacts string
	body     string
}

type mdParserState uint8

func (s *mdParserState) next() {
	*s++
}

const (
	head mdParserState = iota
	avatar
	contacts
	body
)

func parseMDFile(rd io.Reader) (cv cv) {
	state := head

	scn := bufio.NewScanner(rd)
	for scn.Scan() {
		line := scn.Text()
		if strings.TrimSpace(line) == "---" {
			state.next()
			continue
		}

		switch state {
		case head:
			cv.header += line
		case avatar:
			cv.avatar += line + "\n"
		case contacts:
			cv.contacts += line + "\n"
		case body:
			cv.body += line + "\n"
		}
	}
	return cv
}
