package api

import (
	"fmt"
	"strings"
	"time"

	"sakiko.local/sakiko-core/downloadtargets"
	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/kernel"
	"sakiko.local/sakiko-core/logx"
	"sakiko.local/sakiko-core/profiles"
	"sakiko.local/sakiko-core/storage"

	"go.uber.org/zap"
)

type Config struct {
	Mode                interfaces.Mode
	ConnConcurrency     uint
	SpeedConcurrency    uint
	SpeedInterval       time.Duration
	ProfilesPath        string
	ProfileFetchTimeout time.Duration
}

type Service struct {
	kernel          *kernel.Service
	profiles        *profiles.Manager
	downloadTargets *downloadtargets.Manager
	resultStore     *storage.ResultStore
}

func New(cfg Config) (*Service, error) {
	resultStore := storage.NewResultStore(cfg.ProfilesPath)
	k, err := kernel.New(kernel.Config{
		Mode:             cfg.Mode,
		ConnConcurrency:  cfg.ConnConcurrency,
		SpeedConcurrency: cfg.SpeedConcurrency,
		SpeedInterval:    cfg.SpeedInterval,
		ArchiveWriter:    resultStore,
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
		kernel:          k,
		profiles:        pm,
		downloadTargets: downloadtargets.NewManager(downloadtargets.Config{}),
		resultStore:     resultStore,
	}
	apiLogger().Info("service initialized",
		zap.String("mode", string(cfg.Mode)),
		zap.Uint("conn_concurrency", cfg.ConnConcurrency),
		zap.Uint("speed_concurrency", cfg.SpeedConcurrency),
		zap.Duration("speed_interval", cfg.SpeedInterval),
		zap.String("profiles_path", cfg.ProfilesPath),
		zap.Duration("profile_fetch_timeout", cfg.ProfileFetchTimeout),
	)
	return service, nil
}

func (s *Service) Stop() {
	if s != nil && s.kernel != nil {
		apiLogger().Info("service stopping")
		s.kernel.Stop()
	}
}

func (s *Service) SubmitTask(req interfaces.TaskSubmitRequest, onEvent func(interfaces.Event)) (interfaces.TaskSubmitResponse, error) {
	if s == nil || s.kernel == nil {
		apiLogger().Warn("submit task rejected: service not initialized")
		return interfaces.TaskSubmitResponse{}, fmt.Errorf("service not initialized")
	}
	task := req.Task
	if identity := strings.TrimSpace(task.Config.BackendIdentity); identity != "" {
		task.Environment = &interfaces.TaskEnvironment{
			Identity: identity,
		}
	}
	taskID, err := s.kernel.Submit(task, onEvent)
	if err != nil {
		apiLogger().Warn("submit task failed",
			zap.String("task_name", task.Name),
			zap.Int("node_count", len(task.Nodes)),
			zap.Error(err),
		)
		return interfaces.TaskSubmitResponse{}, err
	}
	apiLogger().Debug("task submitted",
		zap.String("task_id", taskID),
		zap.String("task_name", task.Name),
		zap.Int("node_count", len(task.Nodes)),
	)
	return interfaces.TaskSubmitResponse{TaskID: taskID}, nil
}

func (s *Service) ListTasks() interfaces.TaskListResponse {
	if s == nil || s.kernel == nil {
		apiLogger().Warn("list tasks rejected: service not initialized")
		return interfaces.TaskListResponse{Tasks: []interfaces.TaskState{}}
	}
	return interfaces.TaskListResponse{Tasks: s.kernel.ListTasks()}
}

func (s *Service) GetTask(taskID string) (interfaces.TaskStatusResponse, error) {
	if s == nil || s.kernel == nil {
		apiLogger().Warn("get task rejected: service not initialized")
		return interfaces.TaskStatusResponse{}, fmt.Errorf("service not initialized")
	}
	task, ok := s.kernel.GetTask(taskID)
	if !ok {
		apiLogger().Debug("task not found", zap.String("task_id", taskID))
		return interfaces.TaskStatusResponse{}, fmt.Errorf("task not found")
	}
	return task, nil
}

func (s *Service) RuntimeStatus() interfaces.RuntimeStatusResponse {
	if s == nil || s.kernel == nil {
		apiLogger().Warn("runtime status rejected: service not initialized")
		return interfaces.RuntimeStatusResponse{}
	}
	return interfaces.RuntimeStatusResponse{Status: s.kernel.RuntimeStatus()}
}

func (s *Service) ImportProfile(req interfaces.ProfileImportRequest) (interfaces.ProfileImportResponse, error) {
	if s == nil || s.profiles == nil {
		apiLogger().Warn("import profile rejected: service not initialized")
		return interfaces.ProfileImportResponse{}, fmt.Errorf("service not initialized")
	}
	profile, err := s.profiles.Import(req)
	if err != nil {
		apiLogger().Warn("import profile failed", zap.Error(err))
		return interfaces.ProfileImportResponse{}, err
	}
	apiLogger().Debug("profile imported",
		zap.String("profile_id", profile.ID),
		zap.Int("node_count", len(profile.Nodes)),
	)
	return interfaces.ProfileImportResponse{Profile: profile}, nil
}

func (s *Service) RefreshProfile(req interfaces.ProfileRefreshRequest) (interfaces.ProfileRefreshResponse, error) {
	if s == nil || s.profiles == nil {
		apiLogger().Warn("refresh profile rejected: service not initialized")
		return interfaces.ProfileRefreshResponse{}, fmt.Errorf("service not initialized")
	}
	profile, err := s.profiles.Refresh(req.ProfileID)
	if err != nil {
		apiLogger().Warn("refresh profile failed",
			zap.String("profile_id", req.ProfileID),
			zap.Error(err),
		)
		return interfaces.ProfileRefreshResponse{}, err
	}
	apiLogger().Debug("profile refreshed",
		zap.String("profile_id", profile.ID),
		zap.Int("node_count", len(profile.Nodes)),
	)
	return interfaces.ProfileRefreshResponse{Profile: profile}, nil
}

func (s *Service) DeleteProfile(req interfaces.ProfileDeleteRequest) (interfaces.ProfileDeleteResponse, error) {
	if s == nil || s.profiles == nil {
		apiLogger().Warn("delete profile rejected: service not initialized")
		return interfaces.ProfileDeleteResponse{}, fmt.Errorf("service not initialized")
	}
	if err := s.profiles.Delete(req.ProfileID); err != nil {
		apiLogger().Warn("delete profile failed",
			zap.String("profile_id", req.ProfileID),
			zap.Error(err),
		)
		return interfaces.ProfileDeleteResponse{}, err
	}
	apiLogger().Debug("profile deleted", zap.String("profile_id", req.ProfileID))
	return interfaces.ProfileDeleteResponse{ProfileID: req.ProfileID}, nil
}

func (s *Service) ListProfiles() interfaces.ProfileListResponse {
	if s == nil || s.profiles == nil {
		return interfaces.ProfileListResponse{Profiles: []interfaces.Profile{}}
	}
	return interfaces.ProfileListResponse{Profiles: s.profiles.List()}
}

func (s *Service) GetProfile(profileID string) (interfaces.ProfileGetResponse, error) {
	if s == nil || s.profiles == nil {
		apiLogger().Warn("get profile rejected: service not initialized")
		return interfaces.ProfileGetResponse{}, fmt.Errorf("service not initialized")
	}
	profile, ok := s.profiles.Get(profileID)
	if !ok {
		apiLogger().Debug("profile not found", zap.String("profile_id", profileID))
		return interfaces.ProfileGetResponse{}, fmt.Errorf("profile not found")
	}
	return interfaces.ProfileGetResponse{Profile: profile}, nil
}

func (s *Service) UpdateProfileNodeSelection(req interfaces.ProfileNodeSelectionUpdateRequest) (interfaces.ProfileNodeSelectionUpdateResponse, error) {
	if s == nil || s.profiles == nil {
		apiLogger().Warn("update profile node selection rejected: service not initialized")
		return interfaces.ProfileNodeSelectionUpdateResponse{}, fmt.Errorf("service not initialized")
	}

	profile, err := s.profiles.SetNodeEnabled(req.ProfileID, req.NodeIndex, req.Enabled)
	if err != nil {
		apiLogger().Warn("update profile node selection failed",
			zap.String("profile_id", req.ProfileID),
			zap.Int("node_index", req.NodeIndex),
			zap.Bool("enabled", req.Enabled),
			zap.Error(err),
		)
		return interfaces.ProfileNodeSelectionUpdateResponse{}, err
	}

	return interfaces.ProfileNodeSelectionUpdateResponse{Profile: profile}, nil
}

func (s *Service) UpdateProfileNodeOrder(req interfaces.ProfileNodeOrderUpdateRequest) (interfaces.ProfileNodeOrderUpdateResponse, error) {
	if s == nil || s.profiles == nil {
		apiLogger().Warn("update profile node order rejected: service not initialized")
		return interfaces.ProfileNodeOrderUpdateResponse{}, fmt.Errorf("service not initialized")
	}

	profile, err := s.profiles.MoveNode(req.ProfileID, req.NodeIndex, req.TargetIndex)
	if err != nil {
		apiLogger().Warn("update profile node order failed",
			zap.String("profile_id", req.ProfileID),
			zap.Int("node_index", req.NodeIndex),
			zap.Int("target_index", req.TargetIndex),
			zap.Error(err),
		)
		return interfaces.ProfileNodeOrderUpdateResponse{}, err
	}

	return interfaces.ProfileNodeOrderUpdateResponse{Profile: profile}, nil
}

func (s *Service) ListDownloadTargets() (interfaces.DownloadTargetListResponse, error) {
	if s == nil || s.downloadTargets == nil {
		apiLogger().Warn("list download targets rejected: service not initialized")
		return interfaces.DownloadTargetListResponse{}, fmt.Errorf("service not initialized")
	}

	targets, err := s.downloadTargets.List()
	if err != nil {
		apiLogger().Warn("list download targets failed", zap.Error(err))
		return interfaces.DownloadTargetListResponse{}, err
	}

	return interfaces.DownloadTargetListResponse{Targets: targets}, nil
}

func (s *Service) SearchDownloadTargets(search string) (interfaces.DownloadTargetListResponse, error) {
	if s == nil || s.downloadTargets == nil {
		apiLogger().Warn("search download targets rejected: service not initialized")
		return interfaces.DownloadTargetListResponse{}, fmt.Errorf("service not initialized")
	}

	targets, err := s.downloadTargets.ListBySearch(search)
	if err != nil {
		apiLogger().Warn("search download targets failed",
			zap.String("search", search),
			zap.Error(err),
		)
		return interfaces.DownloadTargetListResponse{}, err
	}

	return interfaces.DownloadTargetListResponse{Targets: targets}, nil
}

func (s *Service) ListResultArchives() (interfaces.ResultArchiveListResponse, error) {
	if s == nil || s.resultStore == nil {
		apiLogger().Warn("list result archives rejected: service not initialized")
		return interfaces.ResultArchiveListResponse{}, fmt.Errorf("service not initialized")
	}

	items, err := s.resultStore.List()
	if err != nil {
		apiLogger().Warn("list result archives failed", zap.Error(err))
		return interfaces.ResultArchiveListResponse{}, err
	}
	return interfaces.ResultArchiveListResponse{Archives: items}, nil
}

func (s *Service) GetResultArchive(taskID string) (interfaces.ResultArchiveGetResponse, error) {
	if s == nil || s.resultStore == nil {
		apiLogger().Warn("get result archive rejected: service not initialized")
		return interfaces.ResultArchiveGetResponse{}, fmt.Errorf("service not initialized")
	}

	archive, err := s.resultStore.Load(taskID)
	if err != nil {
		apiLogger().Warn("get result archive failed",
			zap.String("task_id", taskID),
			zap.Error(err),
		)
		return interfaces.ResultArchiveGetResponse{}, err
	}
	return interfaces.ResultArchiveGetResponse{Archive: archive}, nil
}

func (s *Service) DeleteResultArchive(req interfaces.ResultArchiveDeleteRequest) (interfaces.ResultArchiveDeleteResponse, error) {
	if s == nil || s.resultStore == nil {
		apiLogger().Warn("delete result archive rejected: service not initialized")
		return interfaces.ResultArchiveDeleteResponse{}, fmt.Errorf("service not initialized")
	}

	if err := s.resultStore.Delete(req.TaskID); err != nil {
		apiLogger().Warn("delete result archive failed",
			zap.String("task_id", req.TaskID),
			zap.Error(err),
		)
		return interfaces.ResultArchiveDeleteResponse{}, err
	}

	apiLogger().Debug("result archive deleted", zap.String("task_id", req.TaskID))
	return interfaces.ResultArchiveDeleteResponse{TaskID: req.TaskID}, nil
}

func apiLogger() *zap.Logger {
	return logx.Named("core.api")
}
