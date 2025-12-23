package docker

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"

	"github.com/docker/go-connections/nat"

	c "github.com/docker/docker/api/types/container"
	i "github.com/docker/docker/api/types/image"
	v "github.com/docker/docker/api/types/volume"
	k "github.com/docker/docker/client"

	evs "github.com/kubex-ecosystem/domus/internal/events"
	ci "github.com/kubex-ecosystem/domus/internal/interfaces"
	"github.com/kubex-ecosystem/domus/internal/module/kbx"
	"github.com/kubex-ecosystem/domus/internal/types"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
	logz "github.com/kubex-ecosystem/logz"

	_ "embed"
)

func NewServices(name, image string, env map[string]string, ports []nat.PortMap, volumes map[string]struct{}, cmd []string) *ci.Services {
	if containersCache == nil {
		containersCache = make(map[string]*ci.Services)
	}
	service := &ci.Services{
		Name:     name,
		Image:    image,
		Env:      kbxGet.ValOrType(env, map[string]string{}),
		Ports:    ports,
		Volumes:  kbxGet.ValOrType(volumes, map[string]struct{}{}),
		Cmd:      kbxGet.ValOrType(cmd, []string{}),
		StateMap: make(map[string]any),
	}
	if _, ok := containersCache[name]; !ok {
		containersCache[name] = service
	} else {
		containersCache[name].Name = name
		containersCache[name].Image = image
		containersCache[name].Env = kbxGet.ValOrType(env, map[string]string{})
		containersCache[name].Ports = ports
		containersCache[name].Volumes = kbxGet.ValOrType(volumes, map[string]struct{}{})
		containersCache[name].Cmd = kbxGet.ValOrType(cmd, []string{})
	}
	return service
}

type DockerService struct {
	*ContainerNameReport
	*ContainerImageReport
	*ContainerVolumeReport
	*DockerUtils

	Logger    *logz.LoggerZ
	reference kbx.Reference
	mutexes   *types.Mutexes

	services map[string]any

	Cli  ci.IDockerClient
	pool *sync.Pool

	properties map[string]any
	eventBus   *evs.EventBus
}

func newDockerServiceBus(logger *logz.LoggerZ) (ci.IDockerService, error) {
	EnsureDockerIsRunning()

	if logger == nil {
		logger = logz.GetLoggerZ("DockerService")
	}

	cli, err := k.NewClientWithOpts(k.FromEnv, k.WithAPIVersionNegotiation())
	if err != nil {
		return nil, logz.Errorf("Error creating Docker client: %v", err)
	}
	dockerService := &DockerService{
		Logger:     logger,
		reference:  kbx.NewReference("DockerService"),
		mutexes:    types.NewMutexesType(),
		pool:       &sync.Pool{},
		Cli:        cli,
		properties: make(map[string]any),

		DockerUtils:           NewDockerUtils(),
		ContainerNameReport:   NewContainerNameReport(),
		ContainerImageReport:  NewContainerImageReport(),
		ContainerVolumeReport: NewContainerVolumeReport(),
	}
	if dockerService.eventBus == nil {
		dockerService.eventBus = evs.NewEventBus()
	}
	return dockerService, nil
}
func newDockerService(logger *logz.LoggerZ) (ci.IDockerService, error) {
	EnsureDockerIsRunning()
	return newDockerServiceBus(logger)
}
func NewDockerService(logger *logz.LoggerZ) (ci.IDockerService, error) {
	return newDockerService(logger)
}

func (d *DockerService) Initialize() error {
	if d.properties != nil {
		dbServiceConfigT, exists := d.properties["dbConfig"]
		if exists {
			if dbServiceConfig, ok := dbServiceConfigT.(*types.Property[*kbx.RootConfig]); !ok {
				return logz.Errorf("Error converting database configuration")
			} else {
				dbSrvCfg := dbServiceConfig.GetValue()
				if err := SetupDatabaseServices(context.Background(), d, dbSrvCfg); err != nil {
					return logz.Errorf("Error setting up database services: %v", err)
				}
				d.properties["dbConfig"] = dbServiceConfig
			}
		} else {
			logz.Log("warn", "Database configuration not found in DockerService properties... skipping database services setup")
		}
	}
	logz.Debug("Database settings stored in DockerService properties")
	d.properties["volumes"] = make(map[string]map[string]struct{})
	d.properties["services"] = make(map[string]string)

	logz.Success("DockerService initialized successfully")
	return nil
}

func (d *DockerService) InitializeWithConfig(ctx context.Context, dbConfig *kbx.RootConfig) error {
	if dbConfig != nil && kbx.DefaultTrue(dbConfig.Enabled) {
		if err := SetupDatabaseServices(ctx, d, dbConfig); err != nil {
			return logz.Errorf("Error setting up database services: %v", err)
		}
		logz.Success("Database services setup completed successfully")
	} else if dbConfig == nil {
		logz.Warn("Database configuration is nil... skipping database services setup")
		return nil
	}
	logz.Debug("Database settings stored in DockerService properties")
	d.properties["dbConfig"] = types.NewProperty("dbConfig", &dbConfig, true, nil)
	d.properties["volumes"] = make(map[string]map[string]struct{})
	d.properties["services"] = make(map[string]string)

	logz.Success("DockerService initialized successfully (Custom Config)")
	return nil
}

func (d *DockerService) GetContainerLogs(ctx context.Context, containerName string, follow bool) error {
	cli, err := k.NewClientWithOpts(k.FromEnv, k.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("error creating Docker client: %v", err)
	}

	logsReader, err := cli.ContainerLogs(ctx, containerName, c.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Follow:     follow,
	})
	if err != nil {
		return logz.Errorf("error getting logs for container %s: %v", containerName, err)
	}
	defer func(logsReader io.ReadCloser) {
		_ = logsReader.Close()
	}(logsReader)

	scanner := bufio.NewScanner(logsReader)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	return scanner.Err()
}
func (d *DockerService) StartContainer(serviceName, image string, envVars map[string]string, portBindings map[nat.Port]struct{}, volumes map[string]struct{}, cmd []string) error {
	if err := IsDockerAvailable(); err != nil {
		return logz.Errorf("Docker is not available: %v", err)
	}

	if IsServiceRunning(serviceName) {
		logz.Noticef("%s is already running!\n", serviceName)
		return nil
	}

	ctx := context.Background()

	logz.Info("Pulling image...")
	reader, err := d.Cli.ImagePull(ctx, image, i.PullOptions{})
	if err != nil {
		return logz.Errorf("Error pulling image: %v", err)
	}
	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)
	_, _ = io.Copy(io.Discard, reader)

	envSlice := []string{}
	for key, value := range envVars {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", key, value))
	}

	logz.Info("Creating container...")
	containerConfig := &c.Config{
		Image:        image,
		Env:          envSlice,
		ExposedPorts: d.ExtractPorts(portBindings),
		Cmd:          kbxGet.ValOrType(cmd, []string{}),
	}

	binds := []string{}

	for volume := range volumes {
		// Por enquanto coloquei os campos repetidos, mas depois PRECISAMOS melhorar isso
		structuredVolume, err := d.GetStructuredVolume(volume, volume)
		if err != nil {
			return logz.Errorf("error getting structured volume: %v", err)
		}

		binds = append(binds, fmt.Sprintf("%s:%s", structuredVolume.HostPath, structuredVolume.ContainerPath))
	}

	hostIP := ResolveHostIP()
	portBindingsT := make(nat.PortMap)
	for hostPort := range portBindings {
		containerPort := strings.TrimSuffix(hostPort.Port(), "/tcp")
		hostPortBinding := nat.PortBinding{
			HostIP:   hostIP,
			HostPort: hostPort.Port(),
		}
		prtPort := nat.Port(containerPort + "/tcp")
		portBindingsT[prtPort] = []nat.PortBinding{hostPortBinding}
	}

	hostConfig := &c.HostConfig{
		Binds:        binds,
		PortBindings: portBindingsT,
		RestartPolicy: c.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	resp, err := d.Cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, serviceName)
	if err != nil {
		return logz.Errorf("error creating container %s: %v", serviceName, err)
	}

	return d.Cli.ContainerStart(ctx, resp.ID, c.StartOptions{})
}
func (d *DockerService) CreateVolume(volumeName, pathsForBind string) error {
	structuredVolume, err := d.GetStructuredVolume(volumeName, pathsForBind)
	if err != nil {
		return logz.Errorf("error getting structured volume: %v", err)
	}
	ctx := context.Background()

	volumes, err := d.Cli.VolumeList(ctx, v.ListOptions{})
	if err != nil { // Check if this will break anything
		return logz.Errorf("error listing volumes: %v", err)
	}
	for _, vol := range volumes.Volumes {
		if vol.Name == volumeName {
			logz.Debugf("Volume %s already exists, skipping creation", volumeName)
			return nil
		}
	}

	if filepath.IsAbs(structuredVolume.HostPath) {
		var createOpts v.CreateOptions
		if structuredVolume.HostPath == "" {
			createOpts = v.CreateOptions{
				Name:   structuredVolume.Name,
				Labels: map[string]string{"created_by": "kubexdb"},
			}
		} else {
			// Ensure the host path exists
			// if err := ensureDirWithOwner(structuredVolume.HostPath, os.Getuid(), os.Getgid(), 0755); err != nil {
			// 	return fmt.Errorf("error ensuring host path %s exists: %v", structuredVolume.HostPath, err)
			// }
			createOpts = v.CreateOptions{
				Name:   structuredVolume.Name,
				Labels: map[string]string{"created_by": "kubexdb"},
				Driver: "local",
				DriverOpts: map[string]string{
					"type":   "none",
					"device": structuredVolume.HostPath,
					"o":      "bind,rbind,rshared",
				},
			}
		}

		// Create the volume with the bind mount option1
		vol, err := d.Cli.VolumeCreate(ctx, createOpts)
		if err != nil {
			return err
		}

		logz.Infof("Volume %s created at %s", vol.Name, structuredVolume.HostPath)
	}

	return nil
}
func (d *DockerService) GetContainersList() ([]c.Summary, error) {
	containers, err := d.Cli.ContainerList(context.Background(), c.ListOptions{All: true})
	if err != nil {
		return nil, logz.Errorf("Error listing containers: %v", err)
	}

	var containerList []c.Summary
	for _, container := range containers {
		if container.State == "running" {
			containerList = append(containerList, container)
		}
	}

	if len(containersCache) > 0 {
		logz.Debugf("Containers cache has %d entries", len(containersCache))
	} else {
		logz.Debug("Containers cache is empty")
	}

	return containerList, nil
}
func (d *DockerService) GetVolumesList() ([]*v.Volume, error) {
	volumes, err := d.Cli.VolumeList(context.Background(), v.ListOptions{})
	if err != nil {
		return nil, logz.Errorf("Error listing volumes: %v", err)
	}

	var volumeList []*v.Volume
	for _, volume := range volumes.Volumes {
		if volume.Name == "kubexdb-pg-data" /* || volume.Name != "kubexdb-redis-data" */ {
			volumeList = append(volumeList, volume)
		}
	}

	return volumeList, nil
}
func (d *DockerService) StartContainerByName(containerName string) error {
	return d.Cli.ContainerStart(context.Background(), containerName, c.StartOptions{})
}
func (d *DockerService) StopContainerByName(containerName string, stopOptions c.StopOptions) error {
	return d.Cli.ContainerStop(context.Background(), containerName, stopOptions)
}
func (d *DockerService) GetProperty(name string) any {
	if prop, ok := d.properties[name]; ok {
		return prop
	}
	return nil
}
func (d *DockerService) On(name string, event string, callback func(...any)) {
	if d.mutexes == nil {
		d.mutexes = types.NewMutexesType()
	}
	if d.pool == nil {
		d.pool = &sync.Pool{}
	}
	// d.mutexes.MuRLock()
	// defer d.mutexes.MuRUnlock()
	if callback != nil {
		d.pool.Put(callback)
	}
}
func (d *DockerService) Off(name string, event string) {
	if d.mutexes == nil {
		d.mutexes = types.NewMutexesType()
	}
	if d.pool == nil {
		d.pool = &sync.Pool{}
	}
	// d.mutexes.MuRLock()
	// defer d.mutexes.MuRUnlock()
	d.pool.Put(nil)
}
func (d *DockerService) GetContainersCache() map[string]*ci.Services {
	if containersCache == nil {
		containersCache = make(map[string]*ci.Services)
	}
	return containersCache
}
func (d *DockerService) GetEventBus() *evs.EventBus {
	if d.eventBus == nil {
		d.eventBus = evs.NewEventBus()
	}
	return d.eventBus
}
func (d *DockerService) AddService(name string, image string, env map[string]string, ports []nat.PortMap, volumes map[string]struct{}) *ci.Services {
	if containersCache == nil {
		containersCache = make(map[string]*ci.Services)
	}
	service := &ci.Services{
		Name:     name,
		Image:    image,
		Env:      env,
		Ports:    ports,
		Volumes:  volumes,
		StateMap: make(map[string]any),
	}
	if d.services == nil {
		d.services = make(map[string]any)
	}

	d.services[name] = service

	if _, ok := containersCache[name]; !ok {
		containersCache[name] = service
	} else {
		containersCache[name].Name = name
		containersCache[name].Image = image
		containersCache[name].Env = env
		containersCache[name].Ports = ports
		containersCache[name].Volumes = volumes
	}
	return service
}
func (d *DockerService) Error() string {
	return ""
}
