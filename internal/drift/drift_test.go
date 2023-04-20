package drift_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jukie/atlantis-drift-detection/internal/config"
	"github.com/jukie/atlantis-drift-detection/internal/drift"
	"github.com/stretchr/testify/assert"
)

// MockRepositoryClient is a mock implementation of vcs.RepositoryClient.
type MockRepositoryClient struct {
}

func (m *MockRepositoryClient) GetFileContent(repo, path, ref string) (bool, []byte, error) {
	// Mock the behavior of GetFileContent here.
	return true, []byte{}, nil
}

func (m *MockRepositoryClient) VcsType() string {
	// Mock the behavior of VcsType here.
	return "github"
}

func (m *MockRepositoryClient) CreatePull(repo, ref string) (int, string, error) {
	// Mock the behavior of CreatePull here.
	return 1, "https://example.com/pull/1", nil
}

func (m *MockRepositoryClient) CommentOnPull(repo string, pullID int, driftedProjects []string) error {
	// Mock the behavior of CommentOnPull here.
	return nil
}

func TestBuildPlanReq(t *testing.T) {
	mockClient := &MockRepositoryClient{}
	repo := "test-repo"
	ref := "test-ref"
	vcsType := "github"

	req, err := drift.BuildPlanReq(mockClient, repo, ref, vcsType)
	assert.NoError(t, err)

	var planReq drift.PlanApiRequest
	err = json.Unmarshal(req, &planReq)
	assert.NoError(t, err)

	assert.Equal(t, repo, planReq.Repository)
	assert.Equal(t, ref, planReq.Ref)
	assert.Equal(t, vcsType, planReq.Type)
}

func TestDriftChecker(t *testing.T) {
	// Test case 1: no drift detected.
	res := drift.PlanApiResponse{
		ProjectResults: []struct {
			RepoRelDir  string
			Error       interface{}
			Failure     string
			PlanSuccess struct{ TerraformOutput string }
			ProjectName string
		}{
			{
				PlanSuccess: struct{ TerraformOutput string }{TerraformOutput: "No changes. Your infrastructure matches the configuration"},
				ProjectName: "project1",
			},
		},
	}
	driftedProjects, err := drift.DriftChecker(res)
	assert.NoError(t, err)
	assert.Empty(t, driftedProjects)

	// Test case 2: drift detected.
	res.ProjectResults[0].PlanSuccess.TerraformOutput = "1 to add, 0 to change, 0 to destroy"
	driftedProjects, err = drift.DriftChecker(res)
	assert.NoError(t, err)
	assert.NotEmpty(t, driftedProjects)
	assert.Equal(t, "project1", driftedProjects[0])
}

func TestRun(t *testing.T) {
	mockClient := &MockRepositoryClient{}
	repo := config.Repo{
		Name: "test-repo",
		Ref:  "test-ref",
	}
	driftCfg := config.DriftCfg{
		AtlantisUrl:   "http://localhost:4141",
		AtlantisToken: "test-token",
	}

	// Note: In a real test scenario, you should replace the httptest.NewServer with a mock implementation of the Atlantis server.
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ProjectResults": [{"PlanSuccess": {"TerraformOutput": "No changes. Your infrastructure matches the configuration"}, "ProjectName": "project1"}]}`))
	})
	testServer := httptest.NewServer(testHandler)
	defer testServer.Close()

	driftCfg.AtlantisUrl = testServer.URL

	err := drift.Run(mockClient, repo, driftCfg)
	assert.NoError(t, err)
}

func TestApiPlan(t *testing.T) {
	mockClient := &MockRepositoryClient{}
	repo := config.Repo{
		Name: "test-repo",
		Ref:  "test-ref",
	}
	atlantisHost := "http://localhost:4141"
	atlantisToken := "test-token"

	// Note: In a real test scenario, you should replace the httptest.NewServer with a mock implementation of the Atlantis server.
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ProjectResults": [{"PlanSuccess": {"TerraformOutput": "No changes. Your infrastructure matches the configuration"}, "ProjectName": "project1"}]}`))
	})
	testServer := httptest.NewServer(testHandler)
	defer testServer.Close()

	atlantisHost = testServer.URL

	planResp, err := drift.ApiPlan(mockClient, repo, atlantisHost, atlantisToken)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(planResp.ProjectResults))
	assert.Equal(t, "No changes. Your infrastructure matches the configuration", planResp.ProjectResults[0].PlanSuccess.TerraformOutput)
	assert.Equal(t, "project1", planResp.ProjectResults[0].ProjectName)
}

func TestDriftHandler(t *testing.T) {
	mockClient := &MockRepositoryClient{}
	driftedProjects := []string{"project1"}
	repo := config.Repo{
		Name: "test-repo",
		Ref:  "test-ref",
	}

	err := drift.DriftHandler(mockClient, driftedProjects, repo)
	assert.NoError(t, err)
}
