package plugins

type Plugin interface {
	// Name
	Name() string

	// Requires
	Requires() []string

	// Init
	Init() error

	// UnInit
	UnInit() error
}
