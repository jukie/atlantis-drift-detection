# Atlantis Drift Detection

Atlantis Drift Detection is a utility designed to detect drift in infrastructure managed by [Atlantis](https://www.runatlantis.io/). It works by comparing the infrastructure state in your version control system (GitHub, GitLab) with the actual state in your cloud provider.

## Configuration

Atlantis Drift Detection requires the following environment variables to be set:

- `ATLANTIS_URL`: The URL of your Atlantis instance.
- `ATLANTIS_TOKEN`: The API token used to authenticate with your Atlantis instance.
- `CONFIG_PATH`: The path to your VCS configuration file (in YAML format).

An API token for your Git server is also required:
-  `--gitlab-token` or `GITLAB_TOKEN`
-  `--github-token` or `GITHUB_TOKEN`

### VCS Configuration File

The VCS configuration file should have the following format:

```yaml
github:
  apiEndpoint: https://api.mygithubserver.com
  token: github_token
  repos:
    - ref: main
      name: user/repo1
    - ref: master
      name: user/repo2
gitlab:
  apiEndpoint: https://gitlab.com/api/v4
  token: gitlab_token
  repos:
    - ref: main
      name: user/repo3
```

### Usage

1. Clone the repository:
```
git clone https://github.com/yourusername/atlantis-drift-detection.git
cd atlantis-drift-detection
```

2. Build the project:
```
go build -o atlantis-drift-detection
```
3. Set the required environment variables:
```
export ATLANTIS_URL=https://your-atlantis-url.com
export ATLANTIS_TOKEN=your-atlantis-token
export CONFIG_PATH=/path/to/your/config.yaml
```

4. Run the program:
```
./atlantis-drift-detection --github-token $SOME_TOKEN --gitlab-token $SOME_TOKEN
```
