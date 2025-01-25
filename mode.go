package app

// DefaultUser is the default user for the application, used when the application needs to set a username but the
// application is the "user"
const DefaultUser = "app"

type ApplicationMode string

const (
	ReleaseMode = ApplicationMode("release")
	DevMode     = ApplicationMode("dev")
	DebugMode   = ApplicationMode("debug")
)

var (
	// Mode is the mode the application is running in
	Mode = ReleaseMode
)

// InProductionMode returns true if the application is running in production mode
func InProductionMode() bool {
	return Mode == ReleaseMode
}
