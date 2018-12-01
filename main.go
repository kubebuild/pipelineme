package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/cenkalti/backoff"
	"github.com/shurcooL/graphql"

	"github.com/ghodss/yaml"
)

// State for db builds
type State string

// Workflow state
const (
	Scheduled State = "scheduled"
	Running   State = "running"
	Passed    State = "passed"
	Skipped   State = "skipped"
	Failed    State = "failed"
	Canceled  State = "canceled"
	Blocked   State = "blocked"
	Retrying  State = "retrying"
)

var (
	graphqlClient = graphql.NewClient("https://api.kubebuild.com/graphql", nil)
	// graphqlClient = graphql.NewClient("http://localhost:4000/graphql", nil)
)

func check(message string, e error) {
	if e != nil {
		panic(fmt.Sprintf("Message: %s Error: %s", message, e))
	}
}
func main() {
	buildID := os.Getenv("BUILD_ID")
	clusterToken := os.Getenv("CLUSTER_TOKEN")
	repoURL := os.Getenv("REPO")
	revision := os.Getenv("REVISION")
	template, err := downloadPipeline(repoURL, revision)
	if err != nil {
		errorString := err.Error()
		updateBuild(buildID, clusterToken, Failed, template, false, &errorString)
	}
	err = validateTemplate(template)
	if err != nil {
		errorString := err.Error()
		updateBuild(buildID, clusterToken, Failed, template, false, &errorString)
	} else {
		updateBuild(buildID, clusterToken, Scheduled, template, false, nil)
	}
}

func validateTemplate(template string) error {
	var wf wfv1.Workflow

	err := yaml.Unmarshal([]byte(template), &wf)
	return err
}

func downloadPipeline(repoURL string, revision string) (string, error) {
	path := fmt.Sprintf("/tmp/%s", revision)
	operation := func() error {
		gitClone := fmt.Sprintf("git clone --depth=1 -o %s %s %s", revision, repoURL, path)
		cmd := exec.Command("bash", "-c", gitClone)
		err := cmd.Run()
		return err
	}

	exponentialBackOff := backoff.NewExponentialBackOff()
	exponentialBackOff.MaxElapsedTime = 1 * time.Minute

	err := backoff.Retry(operation, exponentialBackOff)
	if err != nil {
		return "", err
	}

	dat, err := ioutil.ReadFile(fmt.Sprintf("%s/.kubebuild.yaml", path))
	if err != nil {
		return "", err
	}
	return string(dat), nil
}

func updateBuild(buildID string, clusterToken string, state State, template string, uploadPipeline graphql.Boolean, errorMessage *string) {
	var buildMutation struct {
		UpdateBuildWithPipeline struct {
			Successful graphql.Boolean
		} `graphql:"updateClusterBuild(buildId: $buildId, clusterToken: $clusterToken, template: $template, uploadPipeline: $uploadPipeline, errorMessage: $errorMessage, state: $state)"`
	}
	variables := map[string]interface{}{
		"buildId":        buildID,
		"clusterToken":   clusterToken,
		"template":       template,
		"uploadPipeline": uploadPipeline,
		"errorMessage":   errorMessage,
		"state":          string(state),
	}
	err := graphqlClient.Mutate(context.Background(), &buildMutation, variables)
	check("Failed to upload pipeline", err)
}
