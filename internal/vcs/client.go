package vcs

type RepositoryClient interface {
	GetFileContent(repo, path, ref string) (bool, []byte, error)
	CreatePull(repo, sourceBranch string) (int, string, error)
	CommentOnPull(repo string, pull int, driftedProjects []string) error
	VcsType() string
}

func GetFileContent(client RepositoryClient, repo, path, ref string) (bool, []byte, error) {
	return client.GetFileContent(repo, path, ref)
}

func CreatePull(client RepositoryClient, repo, sourceBranch, targetBranch string) (int, string, error) {
	return client.CreatePull(repo, sourceBranch)
}

func CommentOnPull(client RepositoryClient, repo string, pull int, driftedProjects []string) error {
	return client.CommentOnPull(repo, pull, driftedProjects)
}
