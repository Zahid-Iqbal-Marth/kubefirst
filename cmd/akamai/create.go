/*
Copyright (C) 2021-2023, Kubefirst

This program is licensed under MIT.
See the LICENSE file for more details.
*/
package akamai

import (
	"fmt"
	"os"
	"strings"

	internalssh "github.com/konstructio/kubefirst-api/pkg/ssh"
	pkg "github.com/konstructio/kubefirst-api/pkg/utils"
	"github.com/konstructio/kubefirst/internal/catalog"
	"github.com/konstructio/kubefirst/internal/cluster"
	"github.com/konstructio/kubefirst/internal/gitShim"
	"github.com/konstructio/kubefirst/internal/launch"
	"github.com/konstructio/kubefirst/internal/progress"
	"github.com/konstructio/kubefirst/internal/provision"
	"github.com/konstructio/kubefirst/internal/utilities"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func createAkamai(cmd *cobra.Command, args []string) error {
	cliFlags, err := utilities.GetFlags(cmd, "akamai")
	if err != nil {
		progress.Error(err.Error())
		return nil
	}

	progress.DisplayLogHints(25)

	isValid, catalogApps, err := catalog.ValidateCatalogApps(cliFlags.InstallCatalogApps)
	if !isValid {
		return err
	}

	err = ValidateProvidedFlags(cliFlags.GitProvider)
	if err != nil {
		progress.Error(err.Error())
		return nil
	}

	// If cluster setup is complete, return

	utilities.CreateK1ClusterDirectory(clusterNameFlag)

	gitAuth, err := gitShim.ValidateGitCredentials(cliFlags.GitProvider, cliFlags.GithubOrg, cliFlags.GitlabGroup)
	if err != nil {
		progress.Error(err.Error())
		return nil
	}

	// Validate git
	executionControl := viper.GetBool(fmt.Sprintf("kubefirst-checks.%s-credentials", cliFlags.GitProvider))
	if !executionControl {
		newRepositoryNames := []string{cliFlags.GitopsRepoName, cliFlags.MetaphorRepoName}
		newTeamNames := []string{cliFlags.AdminTeamName, cliFlags.DeveloperTeamName}

		initGitParameters := gitShim.GitInitParameters{
			GitProvider:  cliFlags.GitProvider,
			GitToken:     gitAuth.Token,
			GitOwner:     gitAuth.Owner,
			Repositories: newRepositoryNames,
			Teams:        newTeamNames,
		}

		err = gitShim.InitializeGitProvider(&initGitParameters)
		if err != nil {
			progress.Error(err.Error())
			return nil
		}
	}
	viper.Set(fmt.Sprintf("kubefirst-checks.%s-credentials", cliFlags.GitProvider), true)
	viper.WriteConfig()

	k3dClusterCreationComplete := viper.GetBool("launch.deployed")
	isK1Debug := strings.ToLower(os.Getenv("K1_LOCAL_DEBUG")) == "true"

	if !k3dClusterCreationComplete && !isK1Debug {
		launch.Up(nil, true, cliFlags.UseTelemetry)
	}

	err = pkg.IsAppAvailable(fmt.Sprintf("%s/api/proxyHealth", cluster.GetConsoleIngresUrl()), "kubefirst api")
	if err != nil {
		progress.Error("unable to start kubefirst api")
	}

	provision.CreateMgmtCluster(gitAuth, cliFlags, catalogApps)

	return nil
}

func ValidateProvidedFlags(gitProvider string) error {
	progress.AddStep("Validate provided flags")

	if os.Getenv("LINODE_TOKEN") == "" {
		// telemetryShim.Transmit(useTelemetryFlag, segmentClient, segment.MetricCloudCredentialsCheckFailed, "LINODE_TOKEN environment variable was not set")
		return fmt.Errorf("your LINODE_TOKEN is not set - please set and re-run your last command")
	}

	// Validate required environment variables for dns provider
	if dnsProviderFlag == "cloudflare" {
		if os.Getenv("CF_API_TOKEN") == "" {
			return fmt.Errorf("your CF_API_TOKEN environment variable is not set. Please set and try again")
		}
	}

	switch gitProvider {
	case "github":
		key, err := internalssh.GetHostKey("github.com")
		if err != nil {
			return fmt.Errorf("known_hosts file does not exist - please run `ssh-keyscan github.com >> ~/.ssh/known_hosts` to remedy")
		} else {
			log.Info().Msgf("%s %s\n", "github.com", key.Type())
		}
	case "gitlab":
		key, err := internalssh.GetHostKey("gitlab.com")
		if err != nil {
			return fmt.Errorf("known_hosts file does not exist - please run `ssh-keyscan gitlab.com >> ~/.ssh/known_hosts` to remedy")
		} else {
			log.Info().Msgf("%s %s\n", "gitlab.com", key.Type())
		}
	}

	progress.CompleteStep("Validate provided flags")

	return nil
}
