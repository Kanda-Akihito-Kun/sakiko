package main

import (
	"context"
	"fmt"
	"sync"

	coreapp "sakiko.local/sakiko-core/app"
	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/logx"

	"github.com/wailsapp/wails/v3/pkg/application"
	"go.uber.org/zap"
)

type SakikoService struct {
	appSvc  *coreapp.Service
	once    sync.Once
	initErr error
	app     *application.App
}

type ProfileTaskSubmitRequest = coreapp.ProfileTaskRequest
type RemoteProfileTaskSubmitRequest = coreapp.RemoteProfileTaskRequest
type DesktopStatus = coreapp.Status
type ProfileSummary = coreapp.ProfileSummary

func (s *SakikoService) ServiceStartup(_ context.Context, _ application.ServiceOptions) error {
	s.app = application.Get()
	setDesktopNotificationApp(s.app)
	return nil
}

func (s *SakikoService) ServiceShutdown() error {
	if s.appSvc != nil {
		s.appSvc.Stop()
	}
	clearDesktopNotificationApp()
	s.app = nil
	return nil
}

func (s *SakikoService) GetAppSettings() (AppSettings, error) {
	if err := s.ensureReady(); err != nil {
		return AppSettings{}, err
	}
	return s.appSvc.GetSettings()
}

func (s *SakikoService) UpdateAppSettings(patch AppSettingsPatch) (AppSettings, error) {
	if err := s.ensureReady(); err != nil {
		return AppSettings{}, err
	}

	settings, err := s.appSvc.UpdateSettings(patch)
	if err != nil {
		wailsServiceLogger().Warn("update app settings failed", zap.Error(err))
		return AppSettings{}, err
	}
	wailsServiceLogger().Info("app settings updated", zap.String("language", settings.Language))
	return settings, nil
}

func (s *SakikoService) DesktopStatus() (DesktopStatus, error) {
	if err := s.ensureReady(); err != nil {
		return DesktopStatus{}, err
	}
	return s.appSvc.Status(), nil
}

func (s *SakikoService) ListProfileSummaries() ([]ProfileSummary, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	return s.appSvc.ListProfileSummaries(), nil
}

func (s *SakikoService) ListProfiles() ([]interfaces.Profile, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	return s.appSvc.ListProfiles(), nil
}

func (s *SakikoService) GetProfile(profileID string) (interfaces.Profile, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.Profile{}, err
	}
	return s.appSvc.GetProfile(profileID)
}

func (s *SakikoService) SetProfileNodeEnabled(profileID string, nodeIndex int, enabled bool) (interfaces.Profile, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.Profile{}, err
	}
	return s.appSvc.SetProfileNodeEnabled(profileID, nodeIndex, enabled)
}

func (s *SakikoService) MoveProfileNode(profileID string, nodeIndex int, targetIndex int) (interfaces.Profile, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.Profile{}, err
	}
	return s.appSvc.MoveProfileNode(profileID, nodeIndex, targetIndex)
}

func (s *SakikoService) ListDownloadTargets() ([]interfaces.DownloadTarget, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	return s.appSvc.ListDownloadTargets()
}

func (s *SakikoService) SearchDownloadTargets(search string) ([]interfaces.DownloadTarget, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	return s.appSvc.SearchDownloadTargets(search)
}

func (s *SakikoService) ImportProfile(req interfaces.ProfileImportRequest) (interfaces.Profile, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.Profile{}, err
	}
	profile, err := s.appSvc.ImportProfile(req)
	if err != nil {
		wailsServiceLogger().Warn("import profile failed", zap.Error(err))
		return interfaces.Profile{}, err
	}
	wailsServiceLogger().Info("profile imported", zap.String("profile_id", profile.ID), zap.Int("node_count", len(profile.Nodes)))
	return profile, nil
}

func (s *SakikoService) RefreshProfile(profileID string) (interfaces.Profile, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.Profile{}, err
	}
	profile, err := s.appSvc.RefreshProfile(profileID)
	if err != nil {
		wailsServiceLogger().Warn("refresh profile failed", zap.String("profile_id", profileID), zap.Error(err))
		return interfaces.Profile{}, err
	}
	wailsServiceLogger().Info("profile refreshed", zap.String("profile_id", profile.ID), zap.Int("node_count", len(profile.Nodes)))
	return profile, nil
}

func (s *SakikoService) DeleteProfile(profileID string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := s.appSvc.DeleteProfile(profileID); err != nil {
		wailsServiceLogger().Warn("delete profile failed", zap.String("profile_id", profileID), zap.Error(err))
		return err
	}
	wailsServiceLogger().Info("profile deleted", zap.String("profile_id", profileID))
	return nil
}

func (s *SakikoService) ListTasks() ([]interfaces.TaskState, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	return s.appSvc.ListTasks(), nil
}

func (s *SakikoService) GetTask(taskID string) (interfaces.TaskStatusResponse, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.TaskStatusResponse{}, err
	}
	return s.appSvc.GetTask(taskID)
}

func (s *SakikoService) CancelTask(taskID string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := s.appSvc.CancelTask(taskID); err != nil {
		wailsServiceLogger().Warn("cancel task failed", zap.String("task_id", taskID), zap.Error(err))
		return err
	}
	wailsServiceLogger().Info("task cancel requested", zap.String("task_id", taskID))
	return nil
}

func (s *SakikoService) DeleteTask(taskID string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := s.appSvc.DeleteTask(taskID); err != nil {
		wailsServiceLogger().Warn("delete task failed", zap.String("task_id", taskID), zap.Error(err))
		return err
	}
	wailsServiceLogger().Info("task deleted", zap.String("task_id", taskID))
	return nil
}

func (s *SakikoService) ListResultArchives() ([]interfaces.ResultArchiveListItem, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	return s.appSvc.ListResultArchives()
}

func (s *SakikoService) GetResultArchive(taskID string) (interfaces.ResultArchive, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.ResultArchive{}, err
	}
	return s.appSvc.GetResultArchive(taskID)
}

func (s *SakikoService) DeleteResultArchive(taskID string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := s.appSvc.DeleteResultArchive(taskID); err != nil {
		wailsServiceLogger().Warn("delete result archive failed", zap.String("task_id", taskID), zap.Error(err))
		return err
	}
	wailsServiceLogger().Info("result archive deleted", zap.String("task_id", taskID))
	return nil
}

func (s *SakikoService) GetRemoteStatus() (interfaces.ClusterStatus, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.ClusterStatus{}, err
	}
	return s.appSvc.GetRemoteStatus()
}

func (s *SakikoService) ProbeRemoteMasterEligibility() (interfaces.MasterEligibility, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.MasterEligibility{}, err
	}
	return s.appSvc.ProbeRemoteMasterEligibility()
}

func (s *SakikoService) EnableRemoteMaster(listenHost string, listenPort int) (interfaces.ClusterStatus, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.ClusterStatus{}, err
	}
	return s.appSvc.EnableRemoteMaster(listenHost, listenPort)
}

func (s *SakikoService) CreateRemotePairingCode(knightName string, ttlSeconds int) (interfaces.ClusterPairingCode, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.ClusterPairingCode{}, err
	}
	return s.appSvc.CreateRemotePairingCode(knightName, ttlSeconds)
}

func (s *SakikoService) EnableRemoteKnight(masterHost string, masterPort int, oneTimeCode string) (interfaces.ClusterStatus, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.ClusterStatus{}, err
	}
	return s.appSvc.EnableRemoteKnight(masterHost, masterPort, oneTimeCode)
}

func (s *SakikoService) DisableRemoteMode() (interfaces.ClusterStatus, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.ClusterStatus{}, err
	}
	return s.appSvc.DisableRemoteMode()
}

func (s *SakikoService) ListRemoteKnights() ([]interfaces.ClusterConnectedKnight, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	return s.appSvc.ListRemoteKnights()
}

func (s *SakikoService) KickRemoteKnight(knightID string) (interfaces.ClusterStatus, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.ClusterStatus{}, err
	}
	return s.appSvc.KickRemoteKnight(knightID)
}

func (s *SakikoService) ListRemoteTasks() ([]interfaces.ClusterRemoteTask, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	return s.appSvc.ListRemoteTasks()
}

func (s *SakikoService) ListRemoteMasterResultArchives() ([]interfaces.ResultArchiveListItem, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	return s.appSvc.ListRemoteMasterResultArchives()
}

func (s *SakikoService) GetRemoteMasterResultArchive(taskID string) (interfaces.ResultArchive, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.ResultArchive{}, err
	}
	return s.appSvc.GetRemoteMasterResultArchive(taskID)
}

func (s *SakikoService) DeleteRemoteMasterResultArchive(taskID string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	return s.appSvc.DeleteRemoteMasterResultArchive(taskID)
}

func (s *SakikoService) ListRemoteKnightResultArchives() ([]interfaces.ResultArchiveListItem, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	return s.appSvc.ListRemoteKnightResultArchives()
}

func (s *SakikoService) GetRemoteKnightResultArchive(taskID string) (interfaces.ResultArchive, error) {
	if err := s.ensureReady(); err != nil {
		return interfaces.ResultArchive{}, err
	}
	return s.appSvc.GetRemoteKnightResultArchive(taskID)
}

func (s *SakikoService) DeleteRemoteKnightResultArchive(taskID string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	return s.appSvc.DeleteRemoteKnightResultArchive(taskID)
}

func (s *SakikoService) SubmitProfileTask(req ProfileTaskSubmitRequest) (string, error) {
	if err := s.ensureReady(); err != nil {
		return "", err
	}

	wailsServiceLogger().Info("submit profile task requested",
		zap.String("profile_id", req.ProfileID),
		zap.String("preset", req.Preset),
		zap.Strings("presets", req.Presets),
		zap.String("task_name", req.Name),
	)
	taskID, err := s.appSvc.SubmitProfileTask(req)
	if err != nil {
		wailsServiceLogger().Warn("submit profile task failed",
			zap.String("profile_id", req.ProfileID),
			zap.String("preset", req.Preset),
			zap.Strings("presets", req.Presets),
			zap.Error(err),
		)
		return "", err
	}
	wailsServiceLogger().Info("profile task submitted", zap.String("profile_id", req.ProfileID), zap.String("task_id", taskID))
	return taskID, nil
}

func (s *SakikoService) SubmitRemoteProfileTask(req RemoteProfileTaskSubmitRequest) ([]interfaces.ClusterRemoteTask, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	tasks, err := s.appSvc.SubmitRemoteProfileTask(req)
	if err != nil {
		wailsServiceLogger().Warn("submit remote profile task failed",
			zap.String("profile_id", req.ProfileID),
			zap.Int("knight_count", len(req.KnightIDs)),
			zap.Error(err),
		)
		return nil, err
	}
	wailsServiceLogger().Info("remote profile task dispatched",
		zap.String("profile_id", req.ProfileID),
		zap.Int("knight_count", len(req.KnightIDs)),
		zap.Int("remote_task_count", len(tasks)),
	)
	return tasks, nil
}

func (s *SakikoService) ensureReady() error {
	if s == nil {
		return fmt.Errorf("sakiko service is nil")
	}

	s.once.Do(func() {
		wailsServiceLogger().Info("initializing sakiko service")
		appSvc, err := coreapp.New(coreapp.Config{})
		if err != nil {
			wailsServiceLogger().Error("initialize core app failed", zap.Error(err))
			s.initErr = err
			return
		}

		s.appSvc = appSvc
		wailsServiceLogger().Info("sakiko service ready", zap.String("profiles_path", appSvc.Paths().ProfilesPath))
	})

	return s.initErr
}

func wailsServiceLogger() *zap.Logger {
	return logx.Named("service")
}
