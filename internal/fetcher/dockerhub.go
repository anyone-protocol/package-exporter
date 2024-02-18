package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
func (f *dockerhubPullsFetcher) FetchCount(ctx context.Context) (int, error) {
	type dockerhubRepo struct {
		PullCount int `json:"pull_count"`
	}

	url := fmt.Sprintf("https://hub.docker.com/v2/namespaces/%s/repositories/%s", f.owner, f.repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var dockerhubRepoResp dockerhubRepo
	err = json.Unmarshal(respData, &dockerhubRepoResp)
	if err != nil {
		return 0, err
	}

	return dockerhubRepoResp.PullCount, nil
}

func (f *dockerhubPullsFetcher) Name() string {
	return f.name
}
