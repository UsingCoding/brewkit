package ssh

type AgentProvider interface {
	Default() (string, error)
}
