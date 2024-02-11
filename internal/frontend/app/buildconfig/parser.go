package buildconfig

type Spec struct {
	Path string
	Args map[string]string
}

type Parser interface {
	Parse(spec Spec) (Config, error)
	// CompileConfig templates config file and returns it raw without parsing
	CompileConfig(spec Spec) (string, error)
}
