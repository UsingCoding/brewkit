package api

import (
	"github.com/ispringtech/brewkit/internal/common/either"
	"github.com/ispringtech/brewkit/internal/common/maybe"
)

type Vertex struct {
	Name string // Unique Vertex name

	Stage maybe.Maybe[Stage]
	Image maybe.Maybe[Image]

	From      maybe.Maybe[*Vertex] // Parent vertex. Not available for vertexes with Image
	DependsOn []Vertex
}

type Stage struct {
	From     string               // From for stage
	Platform maybe.Maybe[string]  // Platform for image
	WorkDir  maybe.Maybe[string]  // Working directory for stage
	Env      map[string]string    // Stage env
	Cache    []Cache              // Pluggable cache for build systems
	Copy     []Copy               // Copy local or build stages artifacts
	Network  maybe.Maybe[Network] // Network options
	SSH      maybe.Maybe[SSH]     // SSH access options
	Secrets  []Secret
	Command  maybe.Maybe[Command] // Command for stage
	Output   maybe.Maybe[Output]  // Output artifacts from builder
}

type Image struct {
	Context    maybe.Maybe[string]
	Dockerfile maybe.Maybe[string]

	// result image reference
	Ref maybe.Maybe[string]

	Network maybe.Maybe[Network]
	SSH     maybe.Maybe[SSH]
	Secrets []Secret
}

type Var struct {
	Name     string // Unique Var name
	From     string
	Platform maybe.Maybe[string]
	WorkDir  maybe.Maybe[string]
	Env      map[string]string
	Cache    []Cache
	Copy     []CopyVar
	Secrets  []Secret
	Network  maybe.Maybe[Network]
	SSH      maybe.Maybe[SSH]
	Command  Command
}

type Copy struct {
	From maybe.Maybe[either.Either[*Vertex, string]]
	Src  string
	Dst  string
}

// CopyVar is Copy instruction for var
type CopyVar struct {
	From maybe.Maybe[string]
	Src  string
	Dst  string
}

type Cache struct {
	ID   string
	Path string
}

type Secret struct {
	ID        string
	MountPath string
}

type SecretSrc struct {
	ID         string
	SourcePath string
}

type SSH struct{}

type Network struct {
	Network string // It may be Host and other docker networks
}

type Command struct {
	Cmd   []string
	Shell bool
}

type Output struct {
	Artifact string
	Local    string
}
