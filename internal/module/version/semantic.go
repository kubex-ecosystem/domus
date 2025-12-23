// Package version provides functionality to manage and check the version of the Kubex Horizon CLI tool.
// It includes methods to retrieve the current version, check for the latest version,
package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	info "github.com/kubex-ecosystem/domus/internal/module/info"

	gl "github.com/kubex-ecosystem/logz"
	"github.com/spf13/cobra"
)

var (
	manifest *info.Info
	vrs      Service
	err      error
)

type Service interface {
	// GetLatestVersion retrieves the latest version from the Git repository.
	GetLatestVersion() (string, error)
	// GetCurrentVersion returns the current version of the service.
	GetCurrentVersion() string
	// IsLatestVersion checks if the current version is the latest version.
	IsLatestVersion() (bool, error)
	// GetName returns the name of the service.
	GetName() string
	// GetVersion returns the current version of the service.
	GetVersion() string
	// GetRepository returns the Git repository URL of the service.
	GetRepository() string
	// setLastCheckedAt sets the last checked time for the version.
	setLastCheckedAt(time.Time)
	// updateLatestVersion updates the latest version from the Git repository.
	updateLatestVersion() error
}
type ServiceImpl struct {
	*info.Info

	gitModelURL    string
	latestVersion  string
	lastCheckedAt  time.Time
	currentVersion string
}

func init() {
	if vrs == nil {
		vrs = NewVersionService()
	}
}

func getLatestTag(repoURL string) (string, error) {
	defer func() {
		if rec := recover(); rec != nil {
			gl.Log("error", "Recovered from panic in getLatestTag: %v", rec)
			err = gl.Errorf("panic occurred while fetching latest tag: %v", rec)
		}
	}()

	defer func() {
		if vrs == nil {
			vrs = NewVersionService()
		}
		vrs.setLastCheckedAt(time.Now())
	}()

	if manifest == nil {
		manifest = info.NewInfo()
		if err := manifest.LoadManifest(); err != nil {
			return "", gl.Errorf("failed to load manifest: %v", err)
		}
	}
	if manifest.IsPrivate() {
		return "", gl.Errorf("cannot fetch latest tag for private repositories")
	}

	if repoURL == "" {
		repoURL = manifest.GetRepository()
		if repoURL == "" {
			return "", gl.Errorf("repository URL is not set")
		}
	}

	apiURL := fmt.Sprintf("%s/tags", repoURL)
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", gl.Errorf("failed to fetch tags: %s", resp.Status)
	}
	type Tag struct {
		Name string `json:"name"`
	}

	// Decode the JSON response into a slice of Tag structs
	// This assumes the API returns a JSON array of tags.
	// Adjust the decoding logic based on the actual API response structure.
	if resp.Header.Get("Content-Type") != "application/json" {
		return "", gl.Errorf("expected application/json, got %s", resp.Header.Get("Content-Type"))
	}

	var tags []Tag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return "", err
	}

	if len(tags) == 0 {
		return "", gl.Errorf("no tags found")
	}
	return tags[0].Name, nil
}
func (v *ServiceImpl) updateLatestVersion() error {
	if manifest.IsPrivate() {
		return gl.Errorf("cannot fetch latest version for private repositories")
	}
	repoURL := strings.TrimSuffix(v.gitModelURL, ".git")
	tag, err := getLatestTag(repoURL)
	if err != nil {
		return err
	}
	v.latestVersion = tag
	return nil
}
func (v *ServiceImpl) vrsCompare(v1, v2 []int) (int, error) {
	compare := 0
	for i := 0; i < len(v1) && i < len(v2); i++ {
		if v1[i] < v2[i] {
			compare = -1
			break
		}
		if v1[i] > v2[i] {
			compare = 1
			break
		}
	}
	return compare, nil
}
func (v *ServiceImpl) versionAtMost(versionAtMostArg, max []int) (bool, error) {
	if comp, err := v.vrsCompare(versionAtMostArg, max); err != nil {
		return false, err
	} else if comp == 1 {
		return false, nil
	}
	return true, nil
}
func (v *ServiceImpl) parseVersion(versionToParse string) []int {
	if versionToParse == "" {
		return nil
	}
	if strings.Contains(versionToParse, "-") {
		versionToParse = strings.Split(versionToParse, "-")[0]
	}
	if strings.Contains(versionToParse, "v") {
		versionToParse = strings.TrimPrefix(versionToParse, "v")
	}
	parts := strings.Split(versionToParse, ".")
	parsedVersion := make([]int, len(parts))
	for i, part := range parts {
		if num, err := strconv.Atoi(part); err != nil {
			return nil
		} else {
			parsedVersion[i] = num
		}
	}
	return parsedVersion
}
func (v *ServiceImpl) IsLatestVersion() (bool, error) {
	if manifest.IsPrivate() {
		return false, gl.Errorf("cannot check version for private repositories")
	}
	if v.latestVersion == "" {
		if err := v.updateLatestVersion(); err != nil {
			return false, err
		}
	}

	currentVersionParts := v.parseVersion(v.currentVersion)
	latestVersionParts := v.parseVersion(v.latestVersion)

	if len(currentVersionParts) == 0 || len(latestVersionParts) == 0 {
		return false, gl.Errorf("invalid version format")
	}

	if len(currentVersionParts) != len(latestVersionParts) {
		return false, gl.Errorf("version parts length mismatch")
	}

	return v.versionAtMost(currentVersionParts, latestVersionParts)
}
func (v *ServiceImpl) GetLatestVersion() (string, error) {
	if manifest.IsPrivate() {
		return "", gl.Errorf("cannot fetch latest version for private repositories")
	}
	if v.latestVersion == "" {
		if err := v.updateLatestVersion(); err != nil {
			return "", err
		}
	}
	return v.latestVersion, nil
}
func (v *ServiceImpl) GetCurrentVersion() string {
	if v.currentVersion == "" {
		v.currentVersion = manifest.GetVersion()
	}
	return v.currentVersion
}
func (v *ServiceImpl) GetName() string {
	if manifest == nil {
		return "Unknown Service"
	}
	return manifest.GetName()
}
func (v *ServiceImpl) GetVersion() string {
	if manifest == nil {
		return "Unknown version"
	}
	return manifest.GetVersion()
}
func (v *ServiceImpl) GetRepository() string {
	if manifest == nil {
		return "No repository URL set in the manifest."
	}
	return manifest.GetRepository()
}
func (v *ServiceImpl) setLastCheckedAt(t time.Time) {
	v.lastCheckedAt = t
	gl.Log("debug", "Last checked at: "+t.Format(time.RFC3339))
}

func NewVersionService() Service {
	if manifest == nil {
		manifest = info.NewInfo()
		if err := manifest.LoadManifest(); err != nil {
			gl.Fatalf("Failed to load manifest: %v", err)
		}
	}
	return &ServiceImpl{
		Info:           manifest,
		gitModelURL:    manifest.GetRepository(),
		currentVersion: manifest.GetVersion(),
		latestVersion:  "",
	}
}

var (
	versionCmd   *cobra.Command
	subLatestCmd *cobra.Command
	subCmdCheck  *cobra.Command
	updCmd       *cobra.Command
	getCmd       *cobra.Command
	restartCmd   *cobra.Command
)

func init() {
	if versionCmd == nil {
		versionCmd = &cobra.Command{
			Use:   "version",
			Short: "Print the version number of " + manifest.GetName(),
			Long:  "Print the version number of " + manifest.GetName() + " and other related information.",
			Run: func(cmd *cobra.Command, args []string) {
				if manifest.IsPrivate() {
					gl.Log("warn", "The information shown may not be accurate for private repositories.")
					gl.Log("info", "Current version: "+GetVersion())
					gl.Log("info", "Git repository: "+GetGitRepositoryModelURL())
					return
				}
				GetVersionInfo()
			},
		}
	}
	if subLatestCmd == nil {
		subLatestCmd = &cobra.Command{
			Use:   "latest",
			Short: "Print the latest version number of " + manifest.GetName(),
			Long:  "Print the latest version number of " + manifest.GetName() + " from the Git repository.",
			Run: func(cmd *cobra.Command, args []string) {
				if manifest.IsPrivate() {
					gl.Log("error", "Cannot fetch latest version for private repositories.")
					return
				}
				GetLatestVersionInfo()
			},
		}
	}
	if subCmdCheck == nil {
		subCmdCheck = &cobra.Command{
			Use:   "check",
			Short: "Check if the current version is the latest version of " + manifest.GetName(),
			Long:  "Check if the current version is the latest version of " + manifest.GetName() + " and print the version information.",
			Run: func(cmd *cobra.Command, args []string) {
				if manifest.IsPrivate() {
					gl.Log("error", "Cannot check version for private repositories.")
					return
				}
				GetVersionInfoWithLatestAndCheck()
			},
		}
	}
	if updCmd == nil {
		updCmd = &cobra.Command{
			Use:   "update",
			Short: "Update the version information of " + manifest.GetName(),
			Long:  "Update the version information of " + manifest.GetName() + " by fetching the latest version from the Git repository.",
			Run: func(cmd *cobra.Command, args []string) {
				if manifest.IsPrivate() {
					gl.Log("error", "Cannot update version for private repositories.")
					return
				}
				if err := vrs.updateLatestVersion(); err != nil {
					gl.Log("error", "Failed to update version: "+err.Error())
				} else {
					latestVersion, err := vrs.GetLatestVersion()
					if err != nil {
						gl.Log("error", "Failed to get latest version: "+err.Error())
					} else {
						gl.Log("info", "Current version: "+vrs.GetCurrentVersion())
						gl.Log("info", "Latest version: "+latestVersion)
					}
					vrs.setLastCheckedAt(time.Now())
				}
			},
		}
	}
	if getCmd == nil {
		getCmd = &cobra.Command{
			Use:   "get",
			Short: "Get the current version of " + manifest.GetName(),
			Long:  "Get the current version of " + manifest.GetName() + " from the manifest.",
			Run: func(cmd *cobra.Command, args []string) {
				gl.Log("info", "Current version: "+vrs.GetCurrentVersion())
			},
		}
	}
	if restartCmd == nil {
		restartCmd = &cobra.Command{
			Use:   "restart",
			Short: "Restart the " + manifest.GetName() + " service",
			Long:  "Restart the " + manifest.GetName() + " service to apply any changes made.",
			Run: func(cmd *cobra.Command, args []string) {
				gl.Log("info", "Restarting the service...")
				// Logic to restart the service can be added here
				gl.Log("success", "Service restarted successfully")
			},
		}
	}

}
func GetVersion() string {
	if manifest == nil {
		manifest = info.NewInfo()
		if err := manifest.LoadManifest(); err != nil {
			gl.Log("error", "Failed to load manifest: "+err.Error())
			return "Unknown version"
		}
	}
	return manifest.GetVersion()
}
func GetGitRepositoryModelURL() string {
	if manifest.GetRepository() == "" {
		return "No repository URL set in the manifest."
	}
	return manifest.GetRepository()
}
func GetVersionInfo() string {
	gl.Log("info", "Version: "+GetVersion())
	gl.Log("info", "Git repository: "+GetGitRepositoryModelURL())
	return fmt.Sprintf("Version: %s\nGit repository: %s", GetVersion(), GetGitRepositoryModelURL())
}
func GetLatestVersionFromGit() string {
	if manifest.IsPrivate() {
		gl.Log("error", "Cannot fetch latest version for private repositories.")
		return "Cannot fetch latest version for private repositories."
	}

	netClient := &http.Client{
		Timeout: time.Second * 10,
	}

	gitURLWithoutGit := strings.TrimSuffix(GetGitRepositoryModelURL(), ".git")
	if gitURLWithoutGit == "" {
		gl.Log("error", "No repository URL set in the manifest.")
		return "No repository URL set in the manifest."
	}

	response, err := netClient.Get(gitURLWithoutGit + "/releases/latest")
	if err != nil {
		gl.Log("error", "Error fetching latest version: "+err.Error())
		gl.Log("error", gitURLWithoutGit+"/releases/latest")
		return err.Error()
	}

	if response.StatusCode != 200 {
		gl.Log("error", "Error fetching latest version: "+response.Status)
		gl.Log("error", "Url: "+gitURLWithoutGit+"/releases/latest")
		body, _ := io.ReadAll(response.Body)
		return fmt.Sprintf("Error: %s\nResponse: %s", response.Status, string(body))
	}

	tag := strings.Split(response.Request.URL.Path, "/")

	return tag[len(tag)-1]
}
func GetLatestVersionInfo() string {
	if manifest.IsPrivate() {
		gl.Log("error", "Cannot fetch latest version for private repositories.")
		return "Cannot fetch latest version for private repositories."
	}
	gl.Log("info", "Latest version: "+GetLatestVersionFromGit())
	return "Latest version: " + GetLatestVersionFromGit()
}
func GetVersionInfoWithLatestAndCheck() string {
	if manifest.IsPrivate() {
		gl.Log("error", "Cannot check version for private repositories.")
		return "Cannot check version for private repositories."
	}
	if GetVersion() == GetLatestVersionFromGit() {
		gl.Log("info", "You are using the latest version.")
		return fmt.Sprintf("You are using the latest version.\n%s\n%s", GetVersionInfo(), GetLatestVersionInfo())
	} else {
		gl.Log("warn", "You are using an outdated version.")
		return fmt.Sprintf("You are using an outdated version.\n%s\n%s", GetVersionInfo(), GetLatestVersionInfo())
	}
}
func CliCommand() *cobra.Command {
	versionCmd.AddCommand(subLatestCmd)
	versionCmd.AddCommand(subCmdCheck)
	versionCmd.AddCommand(updCmd)
	versionCmd.AddCommand(getCmd)
	versionCmd.AddCommand(restartCmd)
	return versionCmd
}
