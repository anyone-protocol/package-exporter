# Downloads Exporter

Prometheus exporter returning download stats for packages. Currently supporting Docker Hub, GitHub Releases and Nginx.

NPM packages support to be added.

## Configuration

Exporter could be configured to monitor multiple packages from multiple places. Below you can see example configuration.

```yaml
packages:
  - type: github-releases
    owner: openssl
    repo: openssl
  - type: docker-hub
    owner: library
    repo: ubuntu
```