package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

type funcClient func(*http.Request) (*http.Response, error)

func (f funcClient) Do(r *http.Request) (*http.Response, error) { return f(r) }

func TestFetcher_Fetch(t *testing.T) {
	tree := &Tree{
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
	f := Fetcher{
		Concurrency: 4,
		Client: funcClient(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodGet {
				t.Error("bad method")
			}
			return &http.Response{
				Body: ioutil.NopCloser(strings.NewReader(`{"title": "foo"}`)),
			}, nil
		}),
	}
	ctx := context.Background()
	if err := f.Fetch(ctx, tree); err != nil {
		t.Fatal(err)
	}
	if tree.Comment != "foo" {
		t.Error(tree.Comment)
	}
}
