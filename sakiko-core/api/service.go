package api

import (
	"errors"
	"path/filepath"
	"sync"
	"time"

	"sakiko.local/sakiko-core/cluster"
	"sakiko.local/sakiko-core/downloadtargets"
	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/kernel"
	"sakiko.local/sakiko-core/logx"
	"sakiko.local/sakiko-core/profiles"
	"sakiko.local/sakiko-core/storage"
	mihomovendor "sakiko.local/sakiko-core/vendors/mihomo"

	"go.uber.org/zap"
)

var errServiceNotInitialized = errors.New("service not initialized")

type Config struct {
	Mode                interfaces.Mode
	ConnConcurrency     uint
	SpeedConcurrency    uint
	SpeedInterval       time.Duration
	ProfilesPath        string
	ProfileFetchTimeout time.Duration
	DNS                 interfaces.DNSConfig
}

type Service struct {
	kernel            *kernel.Service
	cluster           *cluster.Service
	profiles          *profiles.Manager
	downloadTargets   *downloadtargets.Manager
	resultStore       *storage.ResultStore
	remoteMasterStore *storage.ResultStore
	remoteKnightStore *storage.ResultStore
	remoteKnightTasks map[string]string
	remoteTaskLock    sync.RWMutex
}

func New(cfg Config) (*Service, error) {
	cfg.DNS = cfg.DNS.Normalize()
	if err := mihomovendor.ConfigureDNSConfig(cfg.DNS); err != nil {
		return nil, err
	}

	resultStore := storage.NewResultStore(cfg.ProfilesPath)
	remoteMasterStore := storage.NewResultStoreForSubdir(cfg.ProfilesPath, "results/remote-master")
	remoteKnightStore := storage.NewResultStoreForSubdir(cfg.ProfilesPath, "results/remote-knight")
	archiveWriter := newArchiveRouter(resultStore, remoteKnightStore)
	k, err := kernel.New(kernel.Config{
		Mode:             cfg.Mode,
		ConnConcurrency:  cfg.ConnConcurrency,
		SpeedConcurrency: cfg.SpeedConcurrency,
		SpeedInterval:    cfg.SpeedInterval,
		ArchiveWriter:    archiveWriter,
	})
	if err != nil {
		return nil, err
	}

	pm, err := profiles.NewManager(profiles.Config{
		StorePath:    cfg.ProfilesPath,
		FetchTimeout: cfg.ProfileFetchTimeout,
	})
	if err != nil {
		return nil, err
	}

	service := &Service{
		kernel:            k,
		profiles:          pm,
		downloadTargets:   downloadtargets.NewManager(downloadtargets.Config{}),
		resultStore:       resultStore,
		remoteMasterStore: remoteMasterStore,
		remoteKnightStore: remoteKnightStore,
		remoteKnightTasks: map[string]string{},
	}
	service.cluster = cluster.New(cluster.Config{
		KnightAssignmentRunner:    service.runKnightAssignment,
		MasterArchiveWriter:       service.saveRemoteMasterArchive,
		KnightProgressSnapshotter: service.snapshotRemoteKnightTaskState,
		StateDir:                  filepath.Join(filepath.Dir(cfg.ProfilesPath), "remote"),
	})
	if err := service.cluster.Restore(); err != nil {
		apiLogger().Warn("restore remote cluster state failed", zap.Error(err))
	}

	apiLogger().Info("service initialized",
		zap.String("mode", string(cfg.Mode)),
		zap.Uint("conn_concurrency", cfg.ConnConcurrency),
		zap.Uint("speed_concurrency", cfg.SpeedConcurrency),
		zap.Duration("speed_interval", cfg.SpeedInterval),
		zap.String("profiles_path", cfg.ProfilesPath),
		zap.Duration("profile_fetch_timeout", cfg.ProfileFetchTimeout),
		zap.Int("dns_bootstrap_count", len(cfg.DNS.BootstrapServers)),
		zap.Int("dns_resolver_count", len(cfg.DNS.ResolverServers)),
	)
	return service, nil
}

func (s *Service) Stop() {
	if s == nil {
		return
	}

	apiLogger().Info("service stopping")
	if s.cluster != nil {
		s.cluster.Stop()
	}
	if s.kernel != nil {
		s.kernel.Stop()
	}
}

func (s *Service) UpdateDNSConfig(cfg interfaces.DNSConfig) error {
	cfg = cfg.Normalize()
	if err := mihomovendor.ConfigureDNSConfig(cfg); err != nil {
		return err
	}

	apiLogger().Info("dns config updated",
		zap.Int("bootstrap_count", len(cfg.BootstrapServers)),
		zap.Int("resolver_count", len(cfg.ResolverServers)),
	)
	return nil
}

func apiLogger() *zap.Logger {
	return logx.Named("core.api")
}
