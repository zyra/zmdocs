package actions

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/urfave/cli"
	"github.com/zyra/zmdocs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var chansMtx = sync.RWMutex{}
var reloadChans = make(map[string]chan<- struct{})
var chanId = 1

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	ch := make(chan struct{})
	chId := strconv.Itoa(chanId)
	chanId++
	chansMtx.Lock()
	reloadChans[chId] = ch
	chansMtx.Unlock()

	ws.SetCloseHandler(func(code int, text string) error {
		chansMtx.Lock()
		defer chansMtx.RUnlock()
		close(ch)
		delete(reloadChans, chId)
		return nil
	})

	for {
		<-ch
		_ = ws.WriteMessage(1, []byte{})
	}
}

func Serve(ctx *cli.Context) error {
	c, cancelFn := context.WithCancel(context.Background())

	configPath := ctx.String("config")

	if !filepath.IsAbs(configPath) {
		if pwd, err := os.Getwd(); err != nil {
			return err
		} else {
			configPath = filepath.Join(pwd, configPath)
		}
	}

	var config *zmdocs.ParserConfig
	var p *zmdocs.Parser
	var e error
	var rnd *zmdocs.Renderer

	setupParser := func() {
		if config, e = zmdocs.NewConfigFromFile(configPath); e != nil {
			e = fmt.Errorf("unable to parse config: %s", e.Error())
		}

		config.BaseURL = "http://localhost:3500"

		p = zmdocs.NewParser(config)

		if e = p.LoadSourceFiles(); e != nil {
			e = fmt.Errorf("unable to load files: %s", e)
		} else if rnd, e = p.Renderer(); e != nil {
			e = fmt.Errorf("unable to create renderer: %s", e)
		}
	}

	render := func() {
		if e = rnd.Render(); e != nil {
			e = fmt.Errorf("unable to render files: %s", e)
		}
	}

	setupParser()

	if e != nil {
		return e
	}

	render()

	if e != nil {
		return e
	}

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return fmt.Errorf("unable to start watcher: %s", err.Error())
	}

	defer watcher.Close()

	if err := watcher.Add(configPath); err != nil {
		return fmt.Errorf("unable to add config to watcher: %s", err.Error())
	}

	setupFileWatchers := func() {
		for _, f := range p.Files {
			_ = watcher.Add(f.SourceFile)
		}

		for _, t := range p.Config.Templates {
			_ = watcher.Add(t.SourceFile)
		}
	}

	setupFileWatchers()

	reload := func() {
		chansMtx.RLock()
		defer chansMtx.RUnlock()
		for _, ch := range reloadChans {
			ch <- struct{}{}
		}
	}

	go func() {
		for {
			select {
			case <-c.Done():
				return

			case ev, ok := <-watcher.Events:
				if ! ok {
					return
				}

				if ev.Op&fsnotify.Remove == fsnotify.Remove {
					_ = watcher.Remove(ev.Name)
				} else if ev.Op&fsnotify.Write == fsnotify.Write {
					setupParser()

					if e != nil {
						fmt.Println(e)
						continue
					}

					render()

					if e != nil {
						fmt.Println(e)
						continue
					}

					setupFileWatchers()

					if e != nil {
						fmt.Println(e)
						e = nil
					}

					reload()
				}
			}
		}
	}()

	go func() {
		http.HandleFunc("/reload", serveWs)
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			p := filepath.Join(p.Config.OutDir, strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/"), "index.html"), "index.html")
			if fc, e := ioutil.ReadFile(p); e != nil {
				http.NotFound(w, r)
				return
			} else {
				html := strings.Replace(string(fc), "</body>", `
<script>(() => {
const ws = new WebSocket("ws://localhost:3500/reload");
ws.onopen = e => console.log("Livereload WS is open");
ws.onmessage = () => location.reload();
ws.onclose = () => console.log("Livereload WS is closed");
})()</script></body>
`, 1)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(html))
			}
		})
		if err := http.ListenAndServe(":3500", nil); err != nil {
			if context.Canceled != nil {
				return
			}

			fmt.Printf("unable to serve: %s", err.Error())
			cancelFn()
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	select {
	case <-ch:
		cancelFn()
	case <-c.Done():
	}

	return nil
}
