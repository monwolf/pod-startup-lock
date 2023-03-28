package watcher

import (
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
)

type callback func() error

var watcher *fsnotify.Watcher

func CloseWatcher() {
	if watcher != nil {
		watcher.Close()
	}
}

func Watch(filename string, fn callback) {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	err = watcher.Add(filename)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("watcher Watch: Modified file", event.Name)
					time.Sleep(1 * time.Second)
					err := fn()
					if err != nil {
						log.Fatal(err)
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

}
