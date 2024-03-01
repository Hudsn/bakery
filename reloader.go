package bakery

import (
	"context"
	"io/fs"
	"strings"
	"time"
)

func (b *Bakery) watch(ctx context.Context, localFS fs.FS, reloaderChan chan bool) {
	ticker := time.NewTicker(250 * time.Millisecond)
outer:
	for {
		select {
		case <-ctx.Done():
			break outer
		case <-ticker.C:
			b.checkFiles(ctx, localFS, reloaderChan)
		}
	}

}

func (b Bakery) checkFiles(ctx context.Context, localFS fs.FS, reloaderChan chan bool) {
	timeAgo := time.Now().Add(-250 * time.Millisecond)
	fs.WalkDir(localFS, ".", func(path string, d fs.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			return fs.SkipAll
		default:
		}
		if d.IsDir() {
			return nil
		}

		for _, ext := range b.watchExtensions {
			if strings.HasSuffix(path, ext) {

				fileInfo, err := d.Info()
				if err != nil {
					continue
				}

				if fileInfo.ModTime().After(timeAgo) {
					select {
					case reloaderChan <- true:
					default:
						continue
					}
				}

			}
		}
		return nil
	})

}
