package store

import (
	"context"
	"net/http"
	"net/url"
	"os"

	sqlite_vec "github.com/lukasmwerner/sqlite-vec-go/cgo"
	ollama "github.com/ollama/ollama/api"
)

func ollama_embedding(model, s string) ([]byte, error) {
	ollama_host := os.Getenv("OLLAMA_HOST")
	if ollama_host == "" {
		ollama_host = "http://localhost:11434"
	}
	host, err := url.Parse(ollama_host)
	if err != nil {
		return nil, err
	}
	cli := ollama.NewClient(host, http.DefaultClient)
	req := &ollama.EmbedRequest{
		Model: model,
		Input: s,
	}

	resp, err := cli.Embed(context.Background(), req)
	if err != nil {
		return nil, err
	}

	dat, err := sqlite_vec.SerializeFloat32(resp.Embeddings[0])
	if err != nil {
		return nil, err
	}
	return dat, nil
}
