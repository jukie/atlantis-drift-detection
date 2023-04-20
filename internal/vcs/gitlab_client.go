package vcs

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/xanzy/go-gitlab"
)

type GitlabClient struct {
	Client *gitlab.Client
}

func NewGitlabClient(hostname, token string) (*GitlabClient, error) {
	glClient, err := gitlab.NewClient(token, gitlab.WithBaseURL(hostname))
	if err != nil {
		return nil, err
	}
	return &GitlabClient{glClient}, err
}

func (g *GitlabClient) GetFileContent(repo, path, ref string) (bool, []byte, error) {
	opt := gitlab.GetRawFileOptions{Ref: gitlab.String(ref)}

	bytes, resp, err := g.Client.RepositoryFiles.GetRawFile(repo, path, &opt)
	if resp.StatusCode == http.StatusNotFound {
		return false, []byte{}, nil
	}

	if err != nil {
		return false, []byte{}, err
	}

	return true, bytes, nil
}

func (c *GitlabClient) CreatePull(repo, sourceBranch string) (int, string, error) {
	// TODO
	// mrReviewers := c.reviewerIDs()

	head, _, err := c.Client.Commits.GetCommit(repo, sourceBranch)
	if err != nil {
		return 0, "", err
	}
	targetBranch := "atlantis-drift-" + head.ShortID

	err = c.CommitFileChange(repo, sourceBranch, targetBranch)
	if err != nil {
		return 0, "", err
	}
	mr, _, err := c.Client.MergeRequests.CreateMergeRequest(repo, &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.String("Atlantis drift detector"),
		SourceBranch: gitlab.String(sourceBranch),
		TargetBranch: gitlab.String(targetBranch),
		//ReviewerIDs:        mrReviewers,
		RemoveSourceBranch: gitlab.Bool(true),
		Squash:             gitlab.Bool(true),
	})

	return mr.IID, mr.WebURL, err
}

func (g *GitlabClient) driftCommitFileAction(repo, branch string) (gitlab.FileActionValue, error) {
	driftFileExists, _, err := g.GetFileContent(repo, "drift/date.txt", branch)
	if err != nil {
		return "", err
	}
	var action gitlab.FileActionValue = "create"
	if driftFileExists {
		action = "update"
	}
	return action, nil
}

func (g *GitlabClient) CommitFileChange(repo, sourceBranch, targetBranch string) error {
	action, err := g.driftCommitFileAction(repo, sourceBranch)
	if err != nil {
		return err
	}

	_, _, err = g.Client.Commits.CreateCommit(repo, &gitlab.CreateCommitOptions{
		Branch:        gitlab.String(sourceBranch),
		CommitMessage: gitlab.String("Update date.txt"),
		StartBranch:   gitlab.String(targetBranch),
		Actions: []*gitlab.CommitActionOptions{{
			Action:   gitlab.FileAction(action),
			FilePath: gitlab.String("drift-date.txt"),
			Content:  gitlab.String(time.Now().String()),
		}},
		AuthorName: gitlab.String("group_1301_bot2"),
		Force:      gitlab.Bool(true),
	})
	return err
}

func (c *GitlabClient) VcsType() string {
	return "Gitlab"
}

func (c *GitlabClient) CommentOnPull(repo string, pull int, driftedProjects []string) error {
	projectRegexp := strings.Join(driftedProjects, "|")
	commentBody := fmt.Sprintf("atlantis plan -p %s", projectRegexp)
	_, _, err := c.Client.Notes.CreateMergeRequestNote(repo, pull, &gitlab.CreateMergeRequestNoteOptions{
		Body: gitlab.String(commentBody),
	})
	return err
}
