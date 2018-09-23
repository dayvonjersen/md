package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func serveHTTP(addr string, port int, path string) {
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
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, boilerplateHTML)
		fmt.Fprintf(w, "<title>[live preview] %s</title>\n", path)
		fmt.Fprintf(w, boilerplateCSS)
		fmt.Fprintf(w, renderJS)
		fmt.Fprintf(w, previewJS)
		fmt.Fprintf(w, "</head><body>")
		w.Write(render(path))
		fmt.Fprintf(w, "</body></html>")
	})
	http.HandleFunc("/es", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/event-stream")
		w.Header().Add("Access-Control-Allow-Origin", "*")
		id := time.Now().Unix()
		log.Println("CONNECT", id)
		for {
			<-ch
			fmt.Fprintf(w, "id: %d\r\nevent: update\r\ndata: asdf\r\n\r\n", id)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
				log.Println("sent event")
			}
		}
	})
	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
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
