package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/shurcooL/graphql"
)

var (
	// graphqlClient    = graphql.NewClient("https://api.kubebuild.com/graphql", nil)
	graphqlClient = graphql.NewClient("http://localhost:4000/graphql", nil)
)

func check(message string, e error) {
	if e != nil {
		panic(fmt.Sprintf("Message: %s Error: %s", message, e))
	}
}
func main() {
	token := os.Getenv("TOKEN")
	repoURL := os.Getenv("REPO_URL")
	revision := os.Getenv("REVISION")
	template := downloadPipeline(repoURL, revision)
	updateBuild(token, template)
}

func downloadPipeline(repoURL string, revision string) string {
	gitArchive := fmt.Sprintf("git archive --remote=%s %s .kubebuild.yaml | tar -x", repoURL, revision)
	cmd := exec.Command("bash", "-c", gitArchive)
	err := cmd.Run()
	check("Could not download .kubebuild.yaml, make sure file exists on branch", err)
	dat, err := ioutil.ReadFile(".kubebuild.yaml")
	check("Failed to read .kubebuild.yaml", err)
	return string(dat)
}

func updateBuild(token string, template string) {
	var buildMutation struct {
		UpdateBuildWithPipeline struct {
			Successful graphql.Boolean
		} `graphql:"updateBuildWithPipeline(token: $token, template: $template, uploadPipeline: $uploadPipeline)"`
	}
	variables := map[string]interface{}{
		"token":          token,
		"template":       template,
		"uploadPipeline": false,
	}
	err := graphqlClient.Mutate(context.Background(), &buildMutation, variables)
	check("Failed to upload pipeline", err)
}
