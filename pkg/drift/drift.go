package drift

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"regexp"
	"time"

	"github.com/jukie/atlantis-drift-detection/config"
	"github.com/jukie/atlantis-drift-detection/pkg/vcs"
	"gopkg.in/yaml.v3"
)

const atlantisCfgFile = "atlantis.yaml"

type Path struct {
	Directory string `yaml:"dir"`
	Workspace string
}

type PlanApiResponse struct {
	Error          interface{}
	Failure        string
	ProjectResults []struct {
		RepoRelDir  string
		Error       interface{}
		Failure     string
		PlanSuccess struct {
			TerraformOutput string
		}
		ProjectName string
	}
}

type PlanApiRequest struct {
	Repository string
	Ref        string
	Type       string
	Paths      []Path
}

func BuildPlanReq(client vcs.RepositoryClient, repo, ref, vcsType string) ([]byte, error) {
	planInput := PlanApiRequest{
		Repository: repo,
		Ref:        ref,
		Type:       vcsType,
		Paths: []Path{{
			Directory: ".",
		}},
	}
	hasRepoCfg, atlantisCfgBytes, _ := client.GetFileContent(repo, atlantisCfgFile, ref)
	if hasRepoCfg {
		var projects struct {
			Paths []Path `yaml:"projects"`
		}
		_ = yaml.Unmarshal(atlantisCfgBytes, &projects)
		planInput.Paths = projects.Paths
	}

	json_data, err := json.Marshal(planInput)
	if err != nil {
		return nil, err
	}
	return json_data, nil
}

func Run(client vcs.RepositoryClient, repo config.Repo, driftCfg config.DriftCfg) error {
	resp, err := ApiPlan(client, repo, driftCfg.AtlantisUrl, driftCfg.AtlantisToken)
	if err != nil {
		return err
	}
	driftedProjects, err := DriftChecker(resp)
	if err != nil {
		return err
	}
	err = DriftHandler(client, driftedProjects, repo)
	if err != nil {
		return err
	}
	return nil
}

func ApiPlan(client vcs.RepositoryClient, r config.Repo, atlantisHost, atlantisToken string) (PlanApiResponse, error) {
	planReq, err := BuildPlanReq(client, r.Name, r.Ref, client.VcsType())
	if err != nil {
		return PlanApiResponse{}, err
	}
	return httpPost(atlantisHost+"/api/plan", atlantisToken, planReq)
}

func httpPost(url, token string, reqBody []byte) (PlanApiResponse, error) {
	var planResp PlanApiResponse

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return planResp, err
	}
	req.Header.Set("X-Atlantis-Token", token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return planResp, err
	}

	if resp.StatusCode != http.StatusOK {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return planResp, fmt.Errorf("issue during http request to Atlantis server:\nRequest body: %v\nAdditional error: %q", string(reqBody), err)
		}
		return planResp, fmt.Errorf("issue during http request to Atlantis server\nRequest body: %v\nResponse dump: %v", string(reqBody), string(dump))
	}

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&planResp)

	if err != nil {
		return planResp, fmt.Errorf("issue parsing response from Atlantis: %v", err)
	}
	return planResp, nil
}

func DriftChecker(res PlanApiResponse) ([]string, error) {
	r := regexp.MustCompile("No changes. Your infrastructure matches the configuration")
	failedProjects := []string{}
	driftedProjects := []string{}
	var err error
	for _, p := range res.ProjectResults {
		if p.Error != nil {
			fmt.Printf("Errors during plan: %v\nFailure message: %s\n", p.Error, p.Failure)
			failedProjects = append(failedProjects, p.ProjectName)
			continue
		}
		if !r.Match([]byte(p.PlanSuccess.TerraformOutput)) {
			fmt.Printf("Found drifted project %s\n", p.ProjectName)
			driftedProjects = append(driftedProjects, p.ProjectName)
		}
	}
	if len(failedProjects) > 0 {
		err = fmt.Errorf("plan execution failed for following projects: %s", failedProjects)
	}
	return driftedProjects, err
}

func DriftHandler(client vcs.RepositoryClient, driftedProjects []string, repo config.Repo) error {
	if len(driftedProjects) < 1 {
		fmt.Println("No drifted projects found, party on. ༼つ▀̿_▀̿ ༽つ")
		return nil
	}

	fmt.Printf("Drift detected for the following projects: %s\n", driftedProjects)

	pull, url, err := client.CreatePull(repo.Name, repo.Ref)
	if err != nil {
		return err
	}

	fmt.Printf("MR can be seen here: %s\n", url)

	// When testing with Gitlab EE an immediate call would confuse Atlantis
	time.Sleep(15 * time.Second)
	err = client.CommentOnPull(repo.Name, pull, driftedProjects)
	if err != nil {
		return fmt.Errorf("issue creating MR comment: %q", err)
	}
	return nil
}
