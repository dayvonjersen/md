package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday"
)

func render(path string) []byte {
	markdown, err := ioutil.ReadFile(path)
	checkErr(err)
	return blackfriday.MarkdownCommon(markdown)
}

func main() {
	var (
		addr    string
		port    int
		preview bool
	)
	flag.StringVar(
		&addr,
		"addr",
		"",
		"(preview only) leave blank for 0.0.0.0",
	)
	flag.IntVar(
		&port,
		"port",
		8080,
		"(preview only)",
	)
	flag.BoolVar(
		&preview,
		"preview",
		false,
		"start an HTTP server to preview live changes to md files",
	)
	flag.Parse()

	if preview {
		if len(flag.Args()) != 1 {
			log.Fatalln("usage: md -preview [MARKDOWN FILE]")
		}
		serveHTTP(addr, port, flag.Args()[0])
	}

	for _, path := range flag.Args() {
		filename := strings.TrimSuffix(path, filepath.Ext(path)) + ".html"
		f, err := os.Create(filename)
		checkErr(err)
		_, err = io.WriteString(f, boilerplateHTML)
		checkErr(err)
		_, err = io.WriteString(f, boilerplateCSS)
		checkErr(err)
		_, err = io.WriteString(f, renderJS)
		checkErr(err)
		_, err = io.WriteString(f, "</head><body>")
		checkErr(err)
		_, err = f.Write(render(path))
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

const boilerplateHTML = `<!DOCTYPE html>
<html>
    <head>
        <meta charset='utf-8'>
`

const boilerplateCSS = `
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
`

const renderJS = `
        <script>
var contentloadedCallback = () => {
    [].forEach.call(document.querySelectorAll("h1,h2,h3,h4,h5,h6"), (hElement) => {
        hElement.id = hElement.textContent.toLowerCase().replace(/[^a-z0-9_]+/g, '-')
        hElement.insertAdjacentHTML("afterbegin", '<a href="#'+hElement.id+'" class="hashbang">#</a>')
    });
    [].forEach.call(document.querySelectorAll("table"), (tableElement) => tableElement.setAttribute("border", "1"));
};
document.addEventListener("DOMContentLoaded", contentloadedCallback);
        </script>
`

const previewJS = `
        <script>
var Sorbet = (function(S){

    /**
     * eventSource
     *
     * Create an EventSource from url and attach event listeners in format:
     * listeners = {
     *   "eventname": function(event) { ... }
     * }
     */
    S.eventSource = function(url, listeners) {
        var es = new EventSource(url);
        es.retryCount = es.retryCount || 0;
        if(es.retryCount > 5) {
            console.log("EventSource error! Connecting to "+url
                    +" FAILED after "+es.retryCount+" retries :(");
        }

        window.addEventListener("popstate",     function(){es.close();}, false);
        window.addEventListener("beforeunload", function(){es.close();}, false);
        window.addEventListener("unload",       function(){es.close();}, false);

        es.addEventListener('error', function(event){
            console.log("EventSource error!");

            //  getting an error event and a readyState of closed
            //  means that there was a connection error and the 
            //  eventsource must be manually re-opened
            //
            //  NOTE(tso): this just spams the console if the server is down,
            //  disabling for now until we can find a better solution
            //
            // if(es.readyState === EventSource.CLOSED) {
            //     es.close();
            //     es.retryCount++
            //     es = S.eventSource(url, listeners);
            // }
        });

        Object.keys(listeners).forEach(function(event){
            es.addEventListener(event, listeners[event]);
        });
        return es;
    };

    return S;
}(Sorbet || {}));
    Sorbet.eventSource("/es", {
        "update": function(evt) {
            var scrollY = window.scrollY;
            fetch("/update")
            .then((response) => response.text())
            .then((html) => {
                document.body.innerHTML = html;
                contentloadedCallback();
                window.scrollTo(0, scrollY);
            })
        }
    });
        </script>
`
