package llb

import (
	`github.com/moby/buildkit/client/llb`

	`github.com/ispringtech/brewkit/internal/backend/api`
	`github.com/ispringtech/brewkit/internal/common/slices`
)

func (conv *CommonConverter) proceedSecrets(secrets []api.Secret) []llb.RunOption {
	return slices.Map(secrets, func(s api.Secret) llb.RunOption {
		return llb.AddSecret(
			s.MountPath,
			llb.SecretID(s.ID),
		)
	})
}
