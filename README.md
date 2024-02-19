# Package Exporter

Prometheus exporter returning download stats for packages. Currently supporting Docker Hub, GitHub Releases and Nginx.

NPM packages support to be added.

## Configuration

Exporter could be configured to monitor multiple packages from multiple places. Default configuration filename is `config.yml`, feel free to use different with `--config` option. Below you can see example configuration.

```yaml
fetchers:
  dockerhub_pulls:
    - name: anon_dev_dockerhub
      owner: svforte
      repo: anon-dev
    - name: anon_stage_dockerhub
      owner: svforte
      repo: anon-stage
  github_releases:
    - name: anon_dev_github_releases
      owner: ATOR-Development
      repo: ator-protocol
      assets_regexp: ^anon.+-dev-.+\.deb
    - name: anon_stage_github_releases
      owner: ATOR-Development
      repo: ator-protocol
      assets_regexp: ^anon.+-stage-.+\.deb
  nginx_access_log:
    - name: anon_dev_debian_repo
      access_log_path: /var/log/nginx/access.log
      access_log_regexp: '"GET /pool/.+anon_[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+-dev.+\.deb HTTP\/1\.1" 200'
    - name: anon_stage_debian_repo
      access_log_path: /var/log/nginx/access.log
      access_log_regexp: '"GET /pool/.+anon_[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+-stage.+\.deb HTTP\/1\.1" 200'
```

### Available command line options:

```
Usage of ./bin/exporter:
  -config string
    	Config file. (default "config.yml")
  -listen-address string
    	Exporter HTTP listen address. (default ":8080")
  -metrics-path string
    	URL path for metrics endpoint. (default "/metrics")
```

## Build

Make sure you have Go installed and it is in your `PATH`.

```
make build
```

## Run

Make sure you created `config.yml` before running.

```
make run
```

## Access Metrics

Let's test our metrics. Default port could be changed using `--listen-address` option and default metrics endpoint path with `--metrics-path`.

```
curl http://127.0.0.1:8080/metrics
```
