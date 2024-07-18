package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ATOR-Development/downloads-exporter/internal/counter"
)

// NewDockerhubPullsFetcher creates new Docker Hub pulls fetcher from config
func NewDockerhubPullsFetcher(name, owner, repo string) Fetcher {
	return &dockerhubPullsFetcher{
		name:  name,
		owner: owner,
		repo:  repo,
	}
}

type dockerhubPullsFetcher struct {
	name  string
	owner string
	repo  string
}

// FetchCount fetches pull count from docker hub repo API and returns it
func (f *dockerhubPullsFetcher) FetchCount(ctx context.Context) ([]*counter.Result, error) {
	type dockerhubRepo struct {
		PullCount int `json:"pull_count"`
	}

	url := fmt.Sprintf("https://hub.docker.com/v2/namespaces/%s/repositories/%s", f.owner, f.repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var dockerhubRepoResp dockerhubRepo
	err = json.Unmarshal(respData, &dockerhubRepoResp)
	if err != nil {
		return nil, err
	}

	var results []*counter.Result

	results = append(results, &counter.Result{
		Count:  dockerhubRepoResp.PullCount,
		Labels: make(map[string]string),
	})

	return results, nil
}

func (f *dockerhubPullsFetcher) Name() string {
	return f.name
}
