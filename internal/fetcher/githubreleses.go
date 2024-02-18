package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

// NewGithubReleasesFetcher creates new GitHub releases downloads fetcher from config
func NewGithubReleasesFetcher(name, owner, repo string, assetsRegexp *regexp.Regexp) Fetcher {
	return &gitHubReleasesFetcher{
		name:         name,
		owner:        owner,
		repo:         repo,
		assetsRegexp: assetsRegexp,
	}
}

type gitHubReleasesFetcher struct {
	name         string
	owner        string
	repo         string
	assetsRegexp *regexp.Regexp
}

// FetchCount fetches download count from github repo API and returns it
func (f *gitHubReleasesFetcher) FetchCount(ctx context.Context) (int, error) {
	type githubRelease struct {
		Assets []struct {
			Name          string `json:"name"`
			DownloadCount int    `json:"download_count"`
		} `json:"assets"`
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", f.owner, f.repo)
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

	var githubReleasesResp []githubRelease
	err = json.Unmarshal(respData, &githubReleasesResp)
	if err != nil {
		return 0, err
	}

	// calculate total download count from all assets of all releases
	var downloadCount int
	for _, release := range githubReleasesResp {
		for _, asset := range release.Assets {
			if f.assetsRegexp == nil {
				downloadCount += asset.DownloadCount
			}

			if f.assetsRegexp.MatchString(asset.Name) {
				downloadCount += asset.DownloadCount
			}
		}
	}

	return downloadCount, nil
}

func (f *gitHubReleasesFetcher) Name() string {
	return f.name
}
