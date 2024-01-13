package api

type BuildParams struct {
	ForcePull bool
	Context   string
}

type BootstrapParams struct {
	Image string
	Wait  bool
}

type ClearParams struct {
	All bool
}
