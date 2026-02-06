package module

import (
	"os"
	"strings"
)

func RegX() *CanalizeDB {
	var configPath = os.Getenv("DOMUS_CONFIGFILE")
	var keyPath = os.Getenv("DOMUS_KEYFILE")
	var certPath = os.Getenv("DOMUS_CERTFILE")
	var hideBannerV = os.Getenv("DOMUS_HIDEBANNER")

	return &CanalizeDB{
		configPath: configPath,
		keyPath:    keyPath,
		certPath:   certPath,
		hideBanner: (strings.ToLower(hideBannerV) == "true" ||
			strings.ToLower(hideBannerV) == "1" ||
			strings.ToLower(hideBannerV) == "yes" ||
			strings.ToLower(hideBannerV) == "y"),
	}
}
