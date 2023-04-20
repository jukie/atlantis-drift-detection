package vcs

type Client interface {
	GetFileContent(repo, path, ref string) (bool, []byte, error)
	CreatePull(repo, sourceBranch string) (int, string, error)
	CommentOnPull(repo string, pull int, driftedProjects []string) error
	VcsType() string
}

func GetFileContent(client Client, repo, path, ref string) (bool, []byte, error) {
	return client.GetFileContent(repo, path, ref)
}

func CreatePull(client Client, repo, sourceBranch, targetBranch string) (int, string, error) {
	return client.CreatePull(repo, sourceBranch)
}

func CommentOnPull(client Client, repo string, pull int, driftedProjects []string) error {
	return client.CommentOnPull(repo, pull, driftedProjects)
}
