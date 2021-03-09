package main

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/fsnotify/fsnotify"

	// TODO - deduplicate me
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

func watch(ctx context.Context, args []string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				fmt.Printf("> %v\n", event)

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				// TODO - return instead
				panic(err.Error())
			}
		}
	}()

	root, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	i := 0
	if err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		// if err != nil {
		// 	return err
		// }

		// if info.Name() == ".git" || info.Name() == "node_modules" {
		// 	return filepath.SkipDir
		// }

		if !info.Mode().IsDir() {
			return nil
		}

		i++
		if err := watcher.Add(path); err != nil {
			fmt.Printf("> %s (%d) -> %+v\n", path, i, err)
			return err
		}

		return nil
	}); err != nil {
		fmt.Printf("WOT\n")
		return err
	}

	<-done
	return nil
}
