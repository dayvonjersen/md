package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func serveHTTP(addr string, port int, default_path string) {
	path := default_path
	ch := make(chan struct{})
	watch, err := newWatcher(
		func(p string) bool {
			return p == path
		},
		func() {
			ch <- struct{}{}
		},
	)
	checkErr(err)
	watch.w.Add(".")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var req_path string
		log.Println("->", r.Method, r.URL.Path)
		if r.URL.Path == "/" {
			path = default_path
			goto default_route
		}
		req_path = strings.TrimPrefix(r.URL.Path, "/")
		if fileExists(req_path) {
			switch strings.ToLower(filepath.Ext(req_path)) {
			case ".md", ".mdown", ".markdown", ".mkd":
				path = req_path
				goto default_route
			}
			http.ServeFile(w, r, req_path)
			return
		}
		w.WriteHeader(404)
		fmt.Fprintln(w, "404", r.URL.Path, "was not found on this server.")
		return
	default_route:
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, boilerplateHTML)
		fmt.Fprintf(w, "<title>[live preview] %s</title>\n", path)
		fmt.Fprint(w, boilerplateCSS)
		fmt.Fprint(w, renderJS)
		fmt.Fprint(w, previewJS)
		fmt.Fprint(w, "</head><body>")
		w.Write(render(path))
		fmt.Fprint(w, "</body></html>")
	})
	http.HandleFunc("/es", func(w http.ResponseWriter, r *http.Request) {
		log.Println("->", r.Method, r.URL)
		w.Header().Add("Content-Type", "text/event-stream")
		w.Header().Add("Access-Control-Allow-Origin", "*")
		id := time.Now().Unix()
		log.Printf("[event-source:%d] connected", id)
		for {
			select {
			case <-ch:
				evt := fmt.Sprintf("id: %d\r\nevent: update\r\ndata: %d\r\n\r\n", id, time.Now().Unix())
				fmt.Fprint(w, evt)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
					log.Printf("[event-source:%d] sent event: %#v", id, evt)
				}
			case <-r.Context().Done():
				log.Printf("[event-source:%d] exited", id)
				return
			}
		}
	})
	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		log.Println("->", r.Method, r.URL)
		w.Write(render(path))
	})

	listenAddr := fmt.Sprintf("%s:%d", addr, port)
	log.Println("listening on", listenAddr, "(HTTP) ...")
	log.Fatalln(http.ListenAndServe(listenAddr, nil))
}

type event struct {
	filename string
	t        time.Time
}

type watcher struct {
	w         *fsnotify.Watcher
	paths     []string
	events    chan *event
	validator func(string) bool
	callback  func()
}

func newWatcher(validator func(string) bool, callback func()) (*watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	watch := &watcher{
		w:         w,
		paths:     []string{},
		events:    make(chan *event),
		validator: validator,
		callback:  callback,
	}

	go watch.relay()
	go watch.dispatch()

	return watch, nil
}

func (w *watcher) relay() {
	for {
		select {
		case e := <-w.w.Events:
			w.events <- &event{
				filename: normalizePathSeparators(e.Name),
				t:        time.Now(),
			}
		case err := <-w.w.Errors:
			checkErr(err)
		}
	}
}

func (w *watcher) dispatch() {
	var last time.Time
	for e := range w.events {
		diff := time.Since(last) - time.Since(e.t)
		last = e.t
		// log.Println("got:", path.Base(e.filename), diff)
		if !w.validator(e.filename) {
			// log.Println("file is not valid,          skipping...")
			continue
		}
		if diff < time.Millisecond*100 {
			// log.Println("last event was < 100ms ago, skipping...")
			continue
		}
		go w.callback()
	}
}

func normalizePathSeparators(path string) string {
	return strings.Replace(path, "\\", "/", -1)
}

func fileExists(filename string) bool {
	f, err := os.Open(filename)
	if os.IsNotExist(err) {
		return false
	}
	checkErr(err)
	checkErr(f.Close())
	return true
}
