package app

import (
	"context"
	"sync"
	"time"

	coreapi "sakiko.local/sakiko-core/api"
	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/netenv"
	mihomovendor "sakiko.local/sakiko-core/vendors/mihomo"
)

const (
	defaultConnConcurrency     = 24
	defaultSpeedConcurrency    = 1
	defaultSpeedInterval       = 300 * time.Millisecond
	defaultProfileFetchTimeout = 20 * time.Second
)

type Service struct {
	api           *coreapi.Service
	paths         Paths
	mihomoVersion string

	networkEnv       interfaces.BackendInfo
	networkEnvMu     sync.RWMutex
	networkProbeBusy bool
	lastNetworkProbe string
	now              func() time.Time
}

func New(cfg Config) (*Service, error) {
	paths, err := ResolvePaths(cfg)
	if err != nil {
		return nil, err
	}

	settings, err := LoadSettings(paths.SettingsPath)
	if err != nil {
		return nil, err
	}

	apiCfg := coreapi.Config{
		Mode:                defaultMode(cfg.Mode),
		ConnConcurrency:     defaultUint(cfg.ConnConcurrency, defaultConnConcurrency),
		SpeedConcurrency:    defaultUint(cfg.SpeedConcurrency, defaultSpeedConcurrency),
		SpeedInterval:       defaultDuration(cfg.SpeedInterval, defaultSpeedInterval),
		ProfilesPath:        paths.ProfilesPath,
		ProfileFetchTimeout: defaultDuration(cfg.ProfileFetchTimeout, defaultProfileFetchTimeout),
		DNS:                 selectDNS(cfg.DNS, settings.DNS),
	}

	apiService, err := coreapi.New(apiCfg)
	if err != nil {
		return nil, err
	}

	service := &Service{
		api:           apiService,
		paths:         paths,
		mihomoVersion: mihomovendor.LibraryVersion(),
		now:           time.Now,
	}
	service.RefreshNetworkEnvAsync()
	return service, nil
}

func (s *Service) Stop() {
	if s == nil || s.api == nil {
		return
	}
	s.api.Stop()
}

func (s *Service) API() *coreapi.Service {
	if s == nil {
		return nil
	}
	return s.api
}

func (s *Service) Paths() Paths {
	if s == nil {
		return Paths{}
	}
	return s.paths
}

func (s *Service) GetSettings() (Settings, error) {
	if s == nil {
		return Settings{}, errNilService
	}
	return LoadSettings(s.paths.SettingsPath)
}

func (s *Service) UpdateSettings(patch SettingsPatch) (Settings, error) {
	if s == nil {
		return Settings{}, errNilService
	}

	settings, err := LoadSettings(s.paths.SettingsPath)
	if err != nil {
		return Settings{}, err
	}
	settings = ApplySettingsPatch(settings, patch)
	if err := s.api.UpdateDNSConfig(settings.DNS); err != nil {
		return Settings{}, err
	}
	if err := SaveSettings(s.paths.SettingsPath, settings); err != nil {
		return Settings{}, err
	}
	return settings, nil
}

func (s *Service) Status() Status {
	if s == nil {
		return Status{}
	}

	status := Status{
		ProfilesPath:  s.paths.ProfilesPath,
		SettingsPath:  s.paths.SettingsPath,
		MihomoVersion: s.mihomoVersion,
	}
	status.NetworkEnv, status.LastNetworkProbe, status.NetworkProbeBusy = s.networkProbeSnapshot()
	if s.api != nil {
		status.Runtime = s.api.RuntimeStatus().Status
		if remote, err := s.api.ClusterStatus(); err == nil {
			status.Remote = remote.Status
		}
	}
	return status
}

func (s *Service) RefreshNetworkEnvAsync() {
	if s == nil {
		return
	}
	go func() {
		_ = s.RefreshNetworkEnv(context.Background())
	}()
}

func (s *Service) RefreshNetworkEnv(ctx context.Context) interfaces.BackendInfo {
	if s == nil {
		return interfaces.BackendInfo{}
	}

	s.networkEnvMu.Lock()
	s.networkProbeBusy = true
	s.networkEnvMu.Unlock()

	info := netenv.Probe(ctx)

	s.networkEnvMu.Lock()
	s.networkEnv = info
	s.networkProbeBusy = false
	if s.now != nil {
		s.lastNetworkProbe = s.now().UTC().Format(time.RFC3339)
	}
	s.networkEnvMu.Unlock()
	return info
}

func (s *Service) CurrentNetworkEnv() interfaces.BackendInfo {
	if s == nil {
		return interfaces.BackendInfo{}
	}
	s.networkEnvMu.RLock()
	defer s.networkEnvMu.RUnlock()
	return s.networkEnv
}

func (s *Service) networkProbeSnapshot() (interfaces.BackendInfo, string, bool) {
	s.networkEnvMu.RLock()
	defer s.networkEnvMu.RUnlock()
	return s.networkEnv, s.lastNetworkProbe, s.networkProbeBusy
}

func defaultMode(value interfaces.Mode) interfaces.Mode {
	if value == "" {
		return interfaces.ModeParallel
	}
	return value
}

func defaultUint(value uint, fallback uint) uint {
	if value == 0 {
		return fallback
	}
	return value
}

func defaultDuration(value time.Duration, fallback time.Duration) time.Duration {
	if value == 0 {
		return fallback
	}
	return value
}

func selectDNS(configured interfaces.DNSConfig, settings interfaces.DNSConfig) interfaces.DNSConfig {
	if len(configured.BootstrapServers) > 0 || len(configured.ResolverServers) > 0 {
		return configured.Normalize()
	}
	return settings.Normalize()
}
