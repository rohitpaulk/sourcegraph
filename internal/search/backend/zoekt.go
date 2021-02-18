package backend

import (
	"context"
	"strings"

	"github.com/google/zoekt"
	"github.com/google/zoekt/query"
	"github.com/google/zoekt/rpc"
	zoektstream "github.com/google/zoekt/stream"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

// ZoektStreamFunc is a convenience function to create a stream receiver from a
// function.
type ZoektStreamFunc func(*zoekt.SearchResult)

func (f ZoektStreamFunc) Send(event *zoekt.SearchResult) {
	f(event)
}

type Streamer interface {
	// StreamSearch returns a channel which needs to be read until closed.
	StreamSearch(ctx context.Context, q query.Q, opts *zoekt.SearchOptions, c zoektstream.Streamer) error
}

// StreamSearcher is an optional interface which sends results over a channel
// as they are found.
//
// This is a Sourcegraph extension.
type StreamSearcher interface {
	zoekt.Searcher
	Streamer
}

// StreamSearchEvent has fields optionally set representing events that happen
// during a search.
//
// This is a Sourcegraph extension.
type StreamSearchEvent struct {
	// SearchResult is non-nil if this event is a search result. These should be
	// combined with previous and later SearchResults.
	SearchResult *zoekt.SearchResult
}

// StreamSearchAdapter adapts a zoekt.Searcher to conform to the StreamSearch
// interface by calling zoekt.Searcher.Search.
type StreamSearchAdapter struct {
	zoekt.Searcher
}

func (s *StreamSearchAdapter) StreamSearch(ctx context.Context, q query.Q, opts *zoekt.SearchOptions, c zoektstream.Streamer) error {
	sr, err := s.Search(ctx, q, opts)
	if err != nil {
		return err
	}
	c.Send(sr)
	return nil
}

func (s *StreamSearchAdapter) String() string {
	return "streamSearchAdapter{" + s.Searcher.String() + "}"
}

// NewZoektStream returns a StreamSearcher. For cf == nil, we call
// httpcli.NewExternalHTTPClientFactory to create a http.Client.
func NewZoektStream(address string, cf *httpcli.Factory) StreamSearcher {
	if cf == nil {
		cf = httpcli.NewExternalHTTPClientFactory()
	}
	c, err := cf.Client()
	if err != nil {
		// We cannot recover if we cannot get a client.
		panic(err)
	}

	// TODO (stefan): this should go in the client constructor.
	addressWithScheme := address
	if !strings.HasPrefix(addressWithScheme, "http") {
		addressWithScheme = "http://" + addressWithScheme
	}

	return &zoektStream{
		rpc.Client(address),
		zoektstream.NewClient(addressWithScheme, c),
	}
}

type zoektStream struct {
	zoekt.Searcher
	Streamer
}
