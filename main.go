// Main package
package main

import (
	"context"
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AuthToken   string `required:"true" split_words:"true"`
	SecretName  string `required:"true" split_words:"true"`
	SecretValue string `required:"true" split_words:"true"`
}

func main() {
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Println("error getting configuration from environment:", err)
		exitCode = 1
		return
	}

	c := NewGitHubAPIClient(cfg.AuthToken)

	ctx := context.Background()

	repos, err := c.ListRepositories(ctx)
	if err != nil {
		log.Println("error getting the repositories list:", err)
		exitCode = 1
		return
	}

	for _, r := range repos {
		has, err := c.RepositoryHasSecret(ctx, r, cfg.SecretName)
		if err != nil {
			log.Printf("error checking if repository %q has the secret: %s", r, err)
			exitCode = 1
			return
		}
		if has {
			log.Printf("updating secret on repository %q", r)
			if err := c.UpdateRepositorySecret(ctx, r, cfg.SecretName, cfg.SecretValue); err != nil {
				log.Printf("error updating secret for repository %q: %s", r, err)
			}
		}
	}

}
