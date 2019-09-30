package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

type Tree struct {
	ID      int     `json:"-"`
	Comment string  `json:"title"`
	Childs  []*Tree `json:"-"`
}

func (t Tree) print(lvl int) {
	fmt.Printf("%s%d=%s\n", strings.Repeat(" ", lvl), t.ID, t.Comment)
	for _, c := range t.Childs {
		c.print(lvl + 1)
	}
}

func (t *Tree) Print() { t.print(0) }

func (t *Tree) Traverse(ctx context.Context, trees chan<- *Tree) {
	select {
	case <-ctx.Done():
		return
	case trees <- t:
		for _, tt := range t.Childs {
			tt.Traverse(ctx, trees)
		}
	}
}

func (t *Tree) URL() string {
	return fmt.Sprintf("http://jsonplaceholder.typicode.com/todos/%d", t.ID)
}

func (t *Tree) Fetch(ctx context.Context, client httpClient) error {
	req, err := http.NewRequest(http.MethodGet, t.URL(), nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(t); err != nil {
		return err
	}
	return nil
}

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Fetcher struct {
	Concurrency int
	Client      httpClient
}

func (f *Fetcher) Fetch(ctx context.Context, t *Tree) error {
	g, gCtx := errgroup.WithContext(ctx)
	trees := make(chan *Tree)
	go func() {
		t.Traverse(gCtx, trees)
		close(trees)
	}()
	for i := 0; i < f.Concurrency; i++ {
		g.Go(func() error {
			for tt := range trees {
				if err := tt.Fetch(gCtx, f.Client); err != nil {
					return err
				}
			}
			return nil
		})
	}
	return g.Wait()
}

var (
	timeout     = flag.Duration("timeout", time.Minute, "timeout, zero means no timeout")
	concurrency = flag.Int("j", runtime.NumCPU(), "concurrent requests")
)

func main() {
	flag.Parse()

	t := &Tree{
		ID: 1,
		Childs: []*Tree{
			{
				ID: 2,
			},
			{
				ID: 3,
				Childs: []*Tree{
					{ID: 4},
					{ID: 5},
				},
			},
		},
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	if *timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, *timeout)
	}

	// Cancelling on ^C.
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			cancel()
		}
	}()

	f := Fetcher{
		Concurrency: *concurrency,
		Client:      http.DefaultClient,
	}
	if err := f.Fetch(ctx, t); err != nil {
		fmt.Println("Failed:", err)
	}
	t.Print()
}
