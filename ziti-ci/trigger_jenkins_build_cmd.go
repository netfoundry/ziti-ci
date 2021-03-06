package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

type triggerJenkinsSmokeBuildCmd struct {
	baseCommand
	jenkinsUser      string
	jenkinsUserToken string
	jenkinsJobToken  string
}

func (cmd *triggerJenkinsSmokeBuildCmd) execute() {
	cmd.evalCurrentAndNextVersion()

	if cmd.jenkinsUser == "" {
		found := false
		cmd.jenkinsUser, found = os.LookupEnv("jenkins_user")
		if !found {
			cmd.failf("no jenkins user provided. Unable to trigger builds\n")
		}
	}

	if cmd.jenkinsUserToken == "" {
		found := false
		cmd.jenkinsUserToken, found = os.LookupEnv("jenkins_user_token")
		if !found {
			cmd.failf("no jenkins user token provided. Unable to trigger builds\n")
		}
	}

	if cmd.jenkinsJobToken == "" {
		found := false
		cmd.jenkinsJobToken, found = os.LookupEnv("jenkins_job_token")
		if !found {
			cmd.failf("no jenkins job token provided. Unable to trigger builds\n")
		}
	}

	client := resty.New()

	version := cmd.getPublishVersion().String()
	if cmd.getCurrentBranch() != "master" {
		version = fmt.Sprintf("%v-%v", version, cmd.getBuildNumber())
	}

	resp, err := client.R().
		EnableTrace().
		SetQueryParam("token", cmd.jenkinsJobToken).
		SetQueryParam("branch", cmd.getCurrentBranch()).
		SetQueryParam("version", version).
		SetQueryParam("committer", cmd.getCommitterEmail()).
		SetQueryParam("cause", fmt.Sprintf("triggered by Travis ziti-cmd build #%v", cmd.getBuildNumber())).
		SetBasicAuth(cmd.jenkinsUser, cmd.jenkinsUserToken).
		Post("https://jenkinstest.tools.netfoundry.io/job/ziti-smoke-test/buildWithParameters")

	if err != nil {
		cmd.failf("Error triggering build s\n")
		panic(err)
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusAccepted {
		cmd.logJson(resp.Body())
		cmd.failf("Error triggering build. REST call returned %v", resp.StatusCode())
	}

	cmd.infof("successfully triggered build of ziti-smoke-test for branch: %v, version: %v\n", cmd.getCurrentBranch(), version)
}

func newTriggerJenkinsBuildCmd(root *rootCommand) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "trigger-jenkins-smoke-build",
		Short: "Trigger a Jenkins CI smoke test build",
		Args:  cobra.ExactArgs(0),
	}

	result := &triggerJenkinsSmokeBuildCmd{
		baseCommand: baseCommand{
			rootCommand: root,
			cmd:         cobraCmd,
		},
	}

	cobraCmd.PersistentFlags().StringVar(&result.jenkinsUser, "user", "", "Jenkins user to use to trigger the build")
	cobraCmd.PersistentFlags().StringVar(&result.jenkinsUserToken, "user-token", "", "Jenkins user API token to use to trigger the build")
	cobraCmd.PersistentFlags().StringVar(&result.jenkinsJobToken, "job-token", "", "Jenkins job token to use to trigger the build")

	return finalize(result)
}
