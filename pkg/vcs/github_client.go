package vcs

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-github/v51/github"
	"golang.org/x/oauth2"
)

type GithubClient struct {
	Client *github.Client
	Ctx    context.Context
}

func NewGithubClient(hostname, token string) (*GithubClient, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	if hostname != "" && hostname != "https://api.github.com/" {
		var err error
		client.BaseURL, err = url.Parse(hostname)
		if err != nil {
			return nil, err
		}
	}

	return &GithubClient{Client: client, Ctx: ctx}, nil
}

func (g *GithubClient) GetFileContent(owner, repo, path, ref string) (bool, []byte, error) {
	fileContent, _, _, err := g.Client.Repositories.GetContents(g.Ctx, owner, repo, path, &github.RepositoryContentGetOptions{
		Ref: ref,
	})

	if err != nil {
		if _, ok := err.(*github.ErrorResponse); ok && err.(*github.ErrorResponse).Response.StatusCode == http.StatusNotFound {
			return false, []byte{}, nil
		}
		return false, []byte{}, err
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return false, []byte{}, err
	}

	return true, []byte(content), nil
}

func (g *GithubClient) CreatePull(owner, repo, sourceBranch string) (int, string, error) {
	head, _, err := g.Client.Repositories.GetCommit(g.Ctx, owner, repo, sourceBranch, nil)
	if err != nil {
		return 0, "", err
	}
	targetBranch := "atlantis-drift-" + *head.SHA

	err = g.CommitFileChange(owner, repo, sourceBranch, targetBranch)
	if err != nil {
		return 0, "", err
	}
	pr, _, err := g.Client.PullRequests.Create(g.Ctx, owner, repo, &github.NewPullRequest{
		Title:               github.String("Atlantis drift detector"),
		Head:                github.String(sourceBranch),
		Base:                github.String(targetBranch),
		Body:                github.String(""),
		MaintainerCanModify: github.Bool(true),
	})

	if err != nil {
		return 0, "", err
	}

	return *pr.Number, *pr.HTMLURL, err
}

func (g *GithubClient) CommitFileChange(owner, repo, sourceBranch, targetBranch string) error {
	filePath := "drift-date.txt"
	fileExists, _, err := g.GetFileContent(owner, repo, filePath, sourceBranch)

	if err != nil {
		return err
	}

	content := []byte(time.Now().String())
	opts := &github.RepositoryContentFileOptions{
		Message: github.String("Update date.txt"),
		Content: content,
		Branch:  github.String(targetBranch),
	}

	if fileExists {
		_, _, err = g.Client.Repositories.UpdateFile(g.Ctx, owner, repo, filePath, opts)
	} else {
		_, _, err = g.Client.Repositories.CreateFile(g.Ctx, owner, repo, filePath, opts)
	}

	return err
}

func (g *GithubClient) VcsType() string {
	return "GitHub"
}

func (g *GithubClient) CommentOnPull(owner, repo string, pull int, driftedProjects []string) error {
	projectRegexp := strings.Join(driftedProjects, "|")
	commentBody := fmt.Sprintf("atlantis plan -p %s", projectRegexp)
	_, _, err := g.Client.Issues.CreateComment(g.Ctx, owner, repo, pull, &github.IssueComment{
		Body: github.String(commentBody),
	})
	return err
}
