package version

// Version program version variable set by go build -ldflags
var Version = "undefined"

// GetVersion returns the current version
func GetVersion() string {
	if Version == "" || Version == "undefined" {
		return "dev"
	}
	return Version
}
