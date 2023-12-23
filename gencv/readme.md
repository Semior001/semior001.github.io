# Gencv

# Requirements:
- wkhtmltopdf cli tool

# Overview
Generates PDF with your CV, parsed from the source markdown file

The source markdown file should contain 4 parts, divided with the line, containing `---`:
- header
- profile photo
- contacts
- body

Example:
```markdown
# Yelshat Duskaliyev

---

![avatar](https://semior001.github.io/img/me_2.png)

---

## Contact information
- **Email** - [yduskaliyev@semior.dev](mailto:yduskaliyev@semior.dev)                                   
- **Mobile phone** - +77022065472 (kz), +79656022297 (ru)                                           
- **Hometown** - Almaty, Kazakhstan                                                             
- **Website** - [https://semior001.github.io](https://swemior001.github.io)                     
- **Github** - [Semior001](https://github.com/semior001)                                      
- **LinkedIn** - [Yelshat Duskaliyev](https://www.linkedin.com/in/yelshat-duskaliev-181813139/) 
- **Telegram** - [Semior001](https://t.me/semior001)                                            

---

## Professional Experience
- **Junior developer** <br>
    Technodom Operator JSC • Full time <br>
    09/2018 - 07/2019 <br>
    Almaty, Kazakhstan <br>
    > Worked as a full-stack developer, developed and supported several projects for internal usage.

```

This tool will build a pdf with your cv like this one:
![](out.png)

Also it may build an html page, if the `-h` flag specified.

Also, you may specify your own template for your cv page, by passing it to `-t`. Template data:
```go
type cvTmpl struct {
	Header   string
	Avatar   string
	Contacts string
	Body     string
}
```

# Usage:
```text
gencv
Usage:
  gencv [OPTIONS]

Application Options:
  -l, --mdlocation=        location of the markdown file [$MDLOCATION]
  -o, --output_location=   location of the output rendered pdf file (default: ./) [$OUTPUT_LOCATION]
  -t, --template_location= location of the template file (default: tmpl.html) [$TEMPLATE_LOCATION]
  -h                       render html file [$RENDER_HTML]

Help Options:
  -h, --help               Show this help message
```
