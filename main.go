/*
TODO:
- future (maybe):
    - toggle/select view-source
    - local http server for auto-updating preview
*/
package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday"
)

func main() {
	for _, path := range os.Args[1:] {
		markdown, err := ioutil.ReadFile(path)
		checkErr(err)
		html := blackfriday.MarkdownCommon(markdown)

		filename := strings.TrimSuffix(path, filepath.Ext(path)) + ".html"
		f, err := os.Create(filename)
		checkErr(err)
		_, err = io.WriteString(f, boilerplate)
		checkErr(err)
		_, err = f.Write(html)
		checkErr(err)
		_, err = io.WriteString(f, "</body></html>")
		checkErr(err)
		checkErr(f.Close())

		log.Println("wrote", filename)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

const boilerplate = `<!DOCTYPE html>
<html>
    <head>
        <meta charset='utf-8'>
        <style>
body {
    max-width: 800px;
    margin: auto;
    font-family: "Noto Sans", sans-serif;
    background: #fff;
    color: #272727;
}
::selection {
    background: #272727;
    color: #fff;
}
:root {
    font-size: 14px;
    line-height: 1.6875;
}
h1,h2,h3,h4,h5,h6 {
    margin: 0;
}
h1 {
    font-size: 3rem;
    line-height: 1.2638;
    border-bottom: 1px ridge;
}
h2 {
    font-size: 2.5rem;
    line-height: 1.3697;
}
h3 {
    font-size: 2rem;
    line-height: 1.4757;
}
h4 {
    font-size: 1.5rem;
    line-height: 1.5816;
}
h5 {
    font-size: 1.25rem;
    line-height: 1.6875;
}
h6 {
    font-size: 1rem;
    line-height: 1.6875;
}
hr {
    border: 1px ridge;
}
blockquote {
    background: #eee;
    margin: 0;
    padding: 1px 1em;
    border-left: 5px solid #ccc;
}
pre, :not(pre) > code {
    background: #272727;
    color: #fff;
    padding: 0 0.3333em;
    border-radius: 2px;
}
code {
    font-family: "Hack", monospace;
    font-size: 9pt;
}
:not(pre) > code {
    display: inline-block;
    vertical-align: calc(5%);
}
img {
    max-width: 100%;
}
a {
    color: #2492ff;
    text-decoration: underline;
}
a:hover {
    text-decoration: none;
}
ul {
    list-style-type: circle;
}
ol {
    list-style-type: decimal-leading-zero;
}
ul,ol {
    margin: 0;
    padding: 0;
    list-style-position: outside;
}
ul ul, ul ol, ol ul, ol ol {
    margin-left: 2rem;
}

a:not([href]) {
    color: inherit;
    text-decoration: none;
    position: relative;
}
a:not([href])::before {
    content: '#';
    position: absolute;
    left: -1em;
    pointer-events: none;
}
table {
    border-radius: 5px;
}

h1[id],h2[id],h3[id],h4[id],h5[id],h6[id] {
    position: relative;
}
a.hashbang {
    position: absolute;
    left: -6rem;
    padding-left: 3rem;
    color: inherit;
    text-decoration: none;
    opacity: 0;
    width: 100%;
}
h1:hover a.hashbang,
h2:hover a.hashbang,
h3:hover a.hashbang,
h4:hover a.hashbang,
h5:hover a.hashbang,
h6:hover a.hashbang {
    opacity: 1;
}
a.hashbang:hover {
    text-decoration: underline;
}
@media (max-width: 800px) {
    ul,ol {
        list-style-position: inside;
    }
    h1[id],h2[id],h3[id],h4[id],h5[id],h6[id] {
        left: 3rem;
        width: calc(100% - 3rem);
    }
    a.hashbang {
        opacity: 1;
        width: calc(100% - 3rem);
    }
}
        </style>
        <script>
document.addEventListener("DOMContentLoaded", () => {
    [].forEach.call(document.querySelectorAll("h1,h2,h3,h4,h5,h6"), (hElement) => {
        hElement.id = hElement.textContent.toLowerCase().replace(/[^a-z0-9_]+/g, '-')
        hElement.insertAdjacentHTML("afterbegin", '<a href="#'+hElement.id+'" class="hashbang">#</a>')
    });
    [].forEach.call(document.querySelectorAll("table"), (tableElement) => tableElement.setAttribute("border", "1"));
});
        </script>
    </head>
    <body>
`
