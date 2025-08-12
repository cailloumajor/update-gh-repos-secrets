package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"slices"

	"github.com/google/go-github/v74/github"
	"golang.org/x/crypto/nacl/box"
)

// / Represents a GitHub repository
type Repository struct {
	owner string
	name  string
}

func (r Repository) String() string {
	return fmt.Sprintf("%s/%s", r.owner, r.name)
}

// GitHubAPIClient represents a specialized GitHub API client.
type GitHubAPIClient struct {
	ic *github.Client
}

// NewGitHubApiClient returns a new API client, provided the auth token.
func NewGitHubAPIClient(token string) *GitHubAPIClient {
	ic := github.NewClient(nil).WithAuthToken(token)

	return &GitHubAPIClient{ic}
}

func (g *GitHubAPIClient) ListRepositories(ctx context.Context) ([]Repository, error) {
	o := &github.RepositoryListByAuthenticatedUserOptions{
		Affiliation: "owner",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	rl, _, err := g.ic.Repositories.ListByAuthenticatedUser(ctx, o)
	if err != nil {
		return nil, fmt.Errorf("error getting the repositories for authenticated user: %w", err)
	}

	var rn []Repository
	for _, r := range rl {
		if !r.GetArchived() && !r.GetDisabled() {
			rn = append(rn, Repository{
				owner: r.GetOwner().GetLogin(),
				name:  r.GetName(),
			})
		}
	}

	return rn, nil
}

func (g *GitHubAPIClient) RepositoryHasSecret(ctx context.Context, repo Repository, sName string) (bool, error) {
	o := &github.ListOptions{
		PerPage: 100,
	}
	sl, _, err := g.ic.Actions.ListRepoSecrets(ctx, repo.owner, repo.name, o)
	if err != nil {
		return false, fmt.Errorf("error getting the secrets: %w", err)
	}

	return slices.ContainsFunc(sl.Secrets, func(s *github.Secret) bool {
		return s.Name == sName
	}), nil
}

func (g *GitHubAPIClient) UpdateRepositorySecret(ctx context.Context, repo Repository, sName string, sValue string) error {
	k, _, err := g.ic.Actions.GetRepoPublicKey(ctx, repo.owner, repo.name)
	if err != nil {
		return fmt.Errorf("error getting the public key: %w", err)
	}

	d, err := base64.StdEncoding.DecodeString(k.GetKey())
	if err != nil {
		return fmt.Errorf("error decoding public key from base64: %w", err)
	}
	var rpk [32]byte
	if len(d) != len(rpk) {
		return fmt.Errorf("bad decoded public key: want %d, got %d", len(rpk), len(d))
	}
	copy(rpk[:], d)

	s, err := box.SealAnonymous(nil, []byte(sValue), &rpk, nil)
	if err != nil {
		return fmt.Errorf("error encrypting secret: %w", err)
	}

	e := base64.StdEncoding.EncodeToString(s)

	es := &github.EncryptedSecret{
		Name:           sName,
		KeyID:          k.GetKeyID(),
		EncryptedValue: e,
	}

	if _, err := g.ic.Actions.CreateOrUpdateRepoSecret(ctx, repo.owner, repo.name, es); err != nil {
		return fmt.Errorf("error updating secret: %w", err)
	}

	return nil
}
