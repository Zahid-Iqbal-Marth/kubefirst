/*
Copyright (C) 2021-2023, Kubefirst

This program is licensed under MIT.
See the LICENSE file for more details.
*/
package cmd

import (
	"log"

	"github.com/konstructio/kubefirst-api/pkg/certificates"
	"github.com/konstructio/kubefirst/internal/progress"
	"github.com/spf13/cobra"
)

// Certificate check
var domainNameFlag string

func LetsEncryptCommand() *cobra.Command {
	letsEncryptCommand := &cobra.Command{
		Use:   "letsencrypt",
		Short: "interact with LetsEncrypt certificates for a domain",
		Long:  "interact with LetsEncrypt certificates for a domain",
	}

	// wire up new commands
	letsEncryptCommand.AddCommand(status())

	return letsEncryptCommand
}

func status() *cobra.Command {
	statusCmd := &cobra.Command{
		Use:              "status",
		Short:            "check the usage statistics for a LetsEncrypt certificate",
		TraverseChildren: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := certificates.CheckCertificateUsage(domainNameFlag); err != nil {
				log.Printf("failed to check certificate usage for domain %q: %w", domainNameFlag, err)
			}
			progress.Progress.Quit()
		},
	}

	// todo review defaults and update descriptions
	statusCmd.Flags().StringVar(&domainNameFlag, "domain-name", "", "the domain to check certificates for (i.e. your-domain.com|subdomain.your-domain.com) (required)")
	statusCmd.MarkFlagRequired("domain-name")

	return statusCmd
}
