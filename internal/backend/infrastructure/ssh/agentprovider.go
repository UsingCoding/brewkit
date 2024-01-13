package ssh

import (
	"os"

	"github.com/pkg/errors"

	"github.com/ispringtech/brewkit/internal/backend/app/ssh"
)

const (
	sshAuthSock = "SSH_AUTH_SOCK"
)

func NewAgentProvider() ssh.AgentProvider {
	return &agentProvider{}
}

type agentProvider struct{}

func (provider agentProvider) Default() (string, error) {
	socket, found := os.LookupEnv(sshAuthSock)
	if !found {
		return "", errors.Errorf("ssh auth socket via env %s not found", sshAuthSock)
	}

	return socket, nil
}
