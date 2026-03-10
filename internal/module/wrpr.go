package module

import (
	"os"
	"strings"
)

func RegX() *Domus {
	var configPath = os.Getenv("KUBEX_DOMUS_CONFIGFILE")
	var keyPath = os.Getenv("KUBEX_DOMUS_KEYFILE")
	var certPath = os.Getenv("KUBEX_DOMUS_CERTFILE")
	var hideBannerV = os.Getenv("KUBEX_DOMUS_HIDEBANNER")

	return &Domus{
		configPath: configPath,
		keyPath:    keyPath,
		certPath:   certPath,
		hideBanner: (strings.ToLower(hideBannerV) == "true" ||
			strings.ToLower(hideBannerV) == "1" ||
			strings.ToLower(hideBannerV) == "yes" ||
			strings.ToLower(hideBannerV) == "y"),
	}
}
