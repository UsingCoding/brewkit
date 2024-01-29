package llb

import (
	"context"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/docker/docker/api/types/strslice"
	"github.com/google/shlex"
	"github.com/moby/buildkit/client/llb"

	"github.com/ispringtech/brewkit/internal/backend/api"
	"github.com/ispringtech/brewkit/internal/backend/app/progress/progresslabel"
	"github.com/ispringtech/brewkit/internal/common/maybe"
	"github.com/ispringtech/brewkit/internal/common/slices"
)

type cmdDTO struct {
	name         string
	labelPayload map[string]string

	command maybe.Maybe[api.Command]
	cache   []api.Cache
	ssh     maybe.Maybe[api.SSH]
	secrets []api.Secret
	network maybe.Maybe[api.Network]

	ignoreCache   bool
	progressGroup maybe.Maybe[llb.ConstraintsOpt]

	params map[string]string
}

func (conv *CommonConverter) proceedCommand(ctx context.Context, c cmdDTO, st llb.State) (llb.State, error) {
	command, ok := maybe.JustValid(c.command)
	if !ok {
		return st, nil
	}

	args := command.Cmd
	args = proceedParams(args, c.params)

	customName := customCmdName(c.name, args)

	if command.Shell {
		value, err := st.Value(ctx, shellKey)
		if err != nil {
			return llb.State{}, err
		}
		shell, ok := value.(strslice.StrSlice)
		if !ok {
			shell = defaultShell()
		}

		//nolint:gocritic
		args = append(shell, strings.Join(args, " "))
	}

	options := []llb.RunOption{
		llb.Args(args),
	}

	if n, ok := maybe.JustValid(c.network); ok {
		options = append(options, conv.network(n))
	}

	if g, ok := maybe.JustValid(c.progressGroup); ok {
		options = append(options, g)
	}

	if s, ok := maybe.JustValid(conv.proceedSSH(c.ssh)); ok {
		options = append(options, s)
	}

	options = append(options, conv.proceedSecrets(c.secrets)...)
	options = append(options, conv.proceedCache(c.cache)...)

	if c.ignoreCache {
		options = append(options, llb.IgnoreCache)
	}

	if len(c.labelPayload) != 0 {
		var err error
		customName, err = progresslabel.MakePayloadLabel(c.labelPayload, customName)
		if err != nil {
			return llb.State{}, err
		}
	}
	options = append(options, llb.WithCustomName(customName))

	execState := st.Run(
		options...,
	)

	return execState.Root(), nil
}

func proceedParams(args []string, params map[string]string) []string {
	if len(params) == 0 {
		return args
	}

	return slices.Map(args, func(a string) string {
		return os.Expand(a, func(s string) string {
			return params[s]
		})
	})
}

func defaultShell() []string {
	return []string{"/bin/sh", "-c"}
}

func customCmdName(name string, args []string) string {
	a := shortenArgs(args)

	return name + ": " + a
}

func shortenArgs(args []string) string {
	const (
		maxLen = 15
	)

	if len(args) == 1 {
		split, err := shlex.Split(args[0])
		if err == nil {
			args = split
		}
	}

	//nolint:prealloc
	var (
		res   []string
		currN int
	)
	for _, arg := range args {
		n := utf8.RuneCountInString(arg)

		// add arg to res if:
		// it`s first arg
		// maxLen for args is not reached
		if currN == 0 || currN+n < maxLen {
			currN += n

			// clear unnecessary characters
			arg = strings.ReplaceAll(arg, "\n", "")

			res = append(res, arg)
			continue
		}

		res = append(res, "...")

		break
	}

	return strings.Join(res, " ")
}
