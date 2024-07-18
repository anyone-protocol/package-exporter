package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/ATOR-Development/downloads-exporter/internal/counter"
)

// NewGithubReleasesFetcher creates new GitHub releases downloads fetcher from config
func NewGithubReleasesFetcher(name, owner, repo string, assetsRegexp *regexp.Regexp, labels map[string]*regexp.Regexp, counter *counter.Counter) Fetcher {
	return &gitHubReleasesFetcher{
		name:         name,
		owner:        owner,
		repo:         repo,
		assetsRegexp: assetsRegexp,
		labels:       labels,
		counter:      counter,
	}
}

type gitHubReleasesFetcher struct {
	name         string
	owner        string
	repo         string
	assetsRegexp *regexp.Regexp
	labels       map[string]*regexp.Regexp
	counter      *counter.Counter
}

// FetchCount fetches download count from github repo API and returns it
func (f *gitHubReleasesFetcher) FetchCount(ctx context.Context) ([]*counter.Result, error) {
	type githubRelease struct {
		Assets []struct {
			Name          string `json:"name"`
			DownloadCount int    `json:"download_count"`
		} `json:"assets"`
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", f.owner, f.repo)
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

	var githubReleasesResp []githubRelease
	err = json.Unmarshal(respData, &githubReleasesResp)
	if err != nil {
		return nil, err
	}

	for _, release := range githubReleasesResp {
		for _, asset := range release.Assets {
			if f.assetsRegexp == nil || f.assetsRegexp.MatchString(asset.Name) {
				labels := make(map[string]string)
				for labelName, labelRegexp := range f.labels {
					submatch := labelRegexp.FindStringSubmatch(asset.Name)
					if len(submatch) >= 2 {
						labels[labelName] = submatch[1]
					}
				}

				f.counter.Set(labels, asset.DownloadCount)
			}
		}
	}

	return f.counter.Results(), nil
}

func (f *gitHubReleasesFetcher) Name() string {
	return f.name
}
