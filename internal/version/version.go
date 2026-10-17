package version

// AppVersion represents the canonical semantic version of the Jerion desktop application.
// It can be injected at build time using:
// -ldflags "-X 'github.com/EnoughForYou/Jerion/internal/version.AppVersion=v1.2.1'"
var AppVersion = "1.2.1"

// ContractVersion represents the canonical semantic version of the backend-frontend data contract.
const ContractVersion = "1.1.0"
