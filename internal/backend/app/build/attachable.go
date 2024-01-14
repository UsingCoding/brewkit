package build

import (
	"os"

	dockerconfig "github.com/docker/cli/cli/config"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/session/secrets/secretsprovider"
	"github.com/moby/buildkit/session/sshforward/sshprovider"

	"github.com/ispringtech/brewkit/internal/backend/api"
	"github.com/ispringtech/brewkit/internal/common/maybe"
	"github.com/ispringtech/brewkit/internal/common/slices"
)

func (s *service) makeVertexAttachable(v api.Vertex, secrets []api.SecretSrc) ([]session.Attachable, error) {
	res := []session.Attachable{s.dockerAuth()}

	if vertexUseSSH(v) {
		sshAttachable, err := s.sshAttachable()
		if err != nil {
			return nil, err
		}

		res = append(res, sshAttachable)
	}

	secretsAttachable, err := s.secretsAttachable(secrets)
	if err != nil {
		return nil, err
	}

	if secretsAttachable != nil {
		res = append(res, secretsAttachable)
	}

	return res, nil
}

func (s *service) makeVarsAttachable(vars []api.Var, secrets []api.SecretSrc) ([]session.Attachable, error) {
	res := []session.Attachable{s.dockerAuth()}

	if varsUseSSH(vars) {
		sshAttachable, err := s.sshAttachable()
		if err != nil {
			return nil, err
		}

		res = append(res, sshAttachable)
	}

	secretsAttachable, err := s.secretsAttachable(secrets)
	if err != nil {
		return nil, err
	}

	if secretsAttachable != nil {
		res = append(res, secretsAttachable)
	}

	return res, nil
}

// attach docker authorization
func (s *service) dockerAuth() session.Attachable {
	// Set up docker config auth.
	dockerConfig := dockerconfig.LoadDefaultConfigFile(os.Stderr)
	return authprovider.NewDockerAuthProvider(dockerConfig)
}

// creates secretprovider for llbsolver
func (s *service) secretsAttachable(secrets []api.SecretSrc) (session.Attachable, error) {
	if len(secrets) == 0 {
		return nil, nil
	}

	fs := slices.Map(secrets, func(s api.SecretSrc) secretsprovider.Source {
		return secretsprovider.Source{
			ID:       s.ID,
			FilePath: s.SourcePath,
		}
	})
	store, err := secretsprovider.NewStore(fs)
	if err != nil {
		return nil, err
	}
	return secretsprovider.NewSecretProvider(store), nil
}

// creates sshprovider for llbsolver or returns nil if vertex don't use ssh
func (s *service) sshAttachable() (session.Attachable, error) {
	agentPath, err := s.sshAgentProvider.Default()
	if err != nil {
		return nil, err
	}

	const defaultID = "default"
	return sshprovider.NewSSHAgentProvider([]sshprovider.AgentConfig{{
		ID:    defaultID,
		Paths: []string{agentPath},
	}})
}

func varsUseSSH(vars []api.Var) bool {
	for _, v := range vars {
		if maybe.Valid(v.SSH) {
			return true
		}
	}
	return false
}

func vertexUseSSH(v api.Vertex) (use bool) {
	if from, ok := maybe.JustValid(v.From); ok {
		use = vertexUseSSH(*from)
		if use {
			return use
		}
	}

	if s, ok := maybe.JustValid(v.Stage); ok {
		if maybe.Valid(s.SSH) {
			return true
		}

		for _, c := range s.Copy {
			copyFrom, ok := maybe.JustValid(c.From)
			if !ok {
				continue
			}

			copyFrom.MapLeft(func(l *api.Vertex) {
				use = vertexUseSSH(*l)
			})
			if use {
				return use
			}
		}
	}

	for _, dep := range v.DependsOn {
		use = vertexUseSSH(dep)
		if use {
			return use
		}
	}

	return use
}
