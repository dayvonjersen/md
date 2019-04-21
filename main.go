package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday"
)

func render(path string) string {
	markdown, err := ioutil.ReadFile(path)
	checkErr(err)
	return string(blackfriday.MarkdownCommon(markdown))
}

func main() {
	var (
		addr    string
		port    int
		preview bool
	)
	flag.StringVar(&addr, "addr", "", "(preview only) leave blank for 0.0.0.0")
	flag.IntVar(&port, "port", 8080, "(preview only)")
	flag.BoolVar(&preview, "preview", false, "start an HTTP server to preview live changes to md files")
	flag.Parse()
	args := flag.Args()

	if preview {
		if len(args) != 1 {
			log.Fatal("usage: md -preview [MARKDOWN FILE]")
		}
		serveHTTP(addr, port, args[0])
	}

	for _, path := range args {
		filename := strings.TrimSuffix(path, filepath.Ext(path)) + ".html"

		f, err := os.Create(filename)
		checkErr(err)
		_, err = fmt.Fprint(f, boilerplateHTML, boilerplateCSS, renderJS,
			"</head><body>", render(path), "</body></html>",
		)
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

const boilerplateHTML = `<!DOCTYPE html>
<html>
    <head>
        <meta charset='utf-8'>
`

const boilerplateCSS = `
        <style>
*, *::after, *::before {
    box-sizing: border-box;
}

:root {
    font-size: 14px;
    line-height: 1.5;
}

body {
    font-family: "Noto Sans", sans-serif;
    background: #fefefe;
    color: #444;
    padding: 1rem;
    max-width: 42rem;
    margin: auto;
}

::selection, a::selection {
    background: rgba(255,255,0,0.3);
    color: #000;
}

a::selection {
    color: #0645ad;
}

a { 
    color: #0645ad; 
    text-decoration: underline;
}

a:hover { 
    color: #06e; 
    text-decoration: none;
}

a:active {
    color:#faa700;
}

p {
    margin: 1rem 0;
}

img {
    max-width: 100%;
    max-height: 100%;
    vertical-align: middle;
    z-index: 1;
}

h1, h2, h3, h4, h5, h6 {
    font-weight: normal;
    line-height: 1;
    color: #111;
}

h4, h5, h6 {
    font-weight: bold;
}

h1 {
    font-size: 2.5rem;
}

h2 {
    font-size: 2rem;
}

h3 {
    font-size: 1.5rem;
}

h4 {
    font-size: 1.2rem;
}

h5 {
    font-size: 1rem;
}

h6 {
    font-size: 0.9rem;
}

blockquote {
    color: #666;
    margin: 0;
    padding-left: 3rem;
    border-left: 0.5rem solid #eee;
}

hr {
    display: block;
    height: 2px;
    border: 0;
    border-top: 1px solid #aaa;
    border-bottom: 1px solid #eee;
    margin: 1rem 0;
    padding: 0;
}

pre, code, kbd, samp {
    background: #444;
    color: #fefefe;
    font-family: "Hack", monospace;
    font-size: 0.98rem;
    border-radius: 2px;
}

pre {
    white-space: pre-wrap;
    word-wrap: break-word;
}

:not(pre) > code {
    padding: 0 0.3333rem;
}

dfn {
    font-style: italic;
}

ins {
    background: #ff9;
    color: #000;
    text-decoration: none;
}

mark {
    background: #ff0;
    color: #000;
    font-style: italic;
    font-weight: bold;
}

sub, sup {
    font-size: 75%;
    line-height: 0;
    position: relative;
    vertical-align: baseline;
}

sup {
    top: -0.5em;
}

sub {
    bottom: -0.25em;
}

ul, ol {
    margin: 1rem 0;
    padding: 0 0 0 2em;
}

li p:last-child {
    margin: 0;
}

dd {
    margin: 0 0 0 2rem;
}

table {
    border-radius: 5px;
    border-collapse: collapse;
    border-spacing: 0;
    width: 100%;
}

th {
    border-bottom: 1px solid black;
}

td {
    vertical-align: top;
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
    h1[id],h2[id],h3[id],h4[id],h5[id],h6[id] {
        left: 3rem;
        width: calc(100% - 3rem);
    }
    a.hashbang {
        opacity: 1;
        width: calc(100% - 3rem);
    }
}

a.active {
    display: block;
}
        </style>
`

const renderJS = `
        <script>
let lightbox = {
    imgElement: null,
    active: false,
    toggle: () => {
        document.documentElement.style.overflow = lightbox.active ? "hidden" : "visible";

        if(!lightbox.imgElement) return;
        
        let rect = lightbox.imgElement.getBoundingClientRect();
        if(lightbox.active) {
            lightbox.imgElement.parentElement.classList.add("active");
            lightbox.imgElement.parentElement.style.backgroundImage = "url("+lightbox.imgElement.src+")";
            lightbox.imgElement.parentElement.style.width  = rect.width + "px";
            lightbox.imgElement.parentElement.style.height = rect.height + "px";
        } else {
            lightbox.imgElement.parentElement.classList.remove("active");
        }

        lightbox.imgElement.style.position  = lightbox.active ? "fixed" : "static";
        lightbox.imgElement.style.boxShadow = lightbox.active ? "rgba(254,254,254,0.53) 0 0 50vw -6vw, rgba(0,0,0,0.9) 0 0 0 100vh" : "none";
        rect = lightbox.imgElement.getBoundingClientRect();
        lightbox.imgElement.style.left = lightbox.active ? ((window.innerWidth  - rect.width )/2)+"px" : 0;
        lightbox.imgElement.style.top  = lightbox.active ? ((window.innerHeight - rect.height)/2)+"px" : 0;
    },
    close: () => {
        lightbox.active = false;
        lightbox.toggle();
    }
}

let contentloadedCallback = () => {
    [].forEach.call(document.querySelectorAll("h1,h2,h3,h4,h5,h6"), (hElement) => {
        hElement.id = hElement.textContent.toLowerCase().replace(/[^a-z0-9_]+/g, '-')
        hElement.insertAdjacentHTML("afterbegin", '<a href="#'+hElement.id+'" class="hashbang">#</a>')
    });

    [].forEach.call(document.querySelectorAll("table"), (tableElement) => tableElement.setAttribute("border", "1"));
    
    document.documentElement.addEventListener("click", lightbox.close); 

    window.addEventListener("keyup", (e) => {
        if(!lightbox.active) return;
        if(e.key === "Escape") {
            lightbox.close();
            return;
        }

        let imgs = document.querySelectorAll("img");
        if(imgs.length < 2) return;

        let idx  = [].findIndex.call(imgs, (img) => img === lightbox.imgElement);

        switch(e.key) {
        default: return;
        case "ArrowUp":  case "ArrowLeft":  idx--; break;
        case "ArrowDown":case "ArrowRight": idx++; break;
        }

        if(idx < 0) idx = imgs.length - 1;
        if(idx >= imgs.length) idx = 0;

        lightbox.close();
        lightbox.imgElement = imgs[idx];
        lightbox.active = true;
        lightbox.toggle();
    });

    [].forEach.call(document.querySelectorAll("img"), (imgElement) => {
        let anchorElement = document.createElement("a");
        anchorElement.setAttribute("href", imgElement.src);
        anchorElement.setAttribute("target", "_blank");
        imgElement.parentElement.appendChild(anchorElement)
        anchorElement.appendChild(imgElement)
        anchorElement.addEventListener("click", (e) => {
            e.preventDefault();
            e.stopPropagation();
            lightbox.imgElement = imgElement;
            lightbox.active = !lightbox.active;
            lightbox.toggle();
        });
    });
};
document.addEventListener("DOMContentLoaded", contentloadedCallback);
        </script>
`

const previewJS = `
        <script>
let es = new EventSource("/es");

window.addEventListener("popstate",     es.close, false);
window.addEventListener("beforeunload", es.close, false);
window.addEventListener("unload",       es.close, false);

es.addEventListener('error', (event) => {
    console.error("EventSource error!");
});

es.addEventListener("update", (evt) => {
    let scrollY = window.scrollY;
    fetch("/update")
    .then((response) => response.text())
    .then((html) => {
        document.body.innerHTML = html;
        contentloadedCallback();
        window.scrollTo(0, scrollY);
    });
});
        </script>
`
