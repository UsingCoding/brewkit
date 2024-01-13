package llb

import (
	`github.com/moby/buildkit/client/llb`

	`github.com/ispringtech/brewkit/internal/backend/api`
	`github.com/ispringtech/brewkit/internal/common/maybe`
)

func (conv *CommonConverter) proceedSSH(ssh maybe.Maybe[api.SSH]) maybe.Maybe[llb.RunOption] {
	_, ok := maybe.JustValid(ssh)
	if !ok {
		return maybe.Maybe[llb.RunOption]{}
	}

	id := "default"

	opts := []llb.SSHOption{llb.SSHID(id)}

	return maybe.NewJust(llb.AddSSHSocket(opts...))
}
