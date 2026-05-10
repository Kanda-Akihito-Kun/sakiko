package app

import "sakiko.local/sakiko-core/interfaces"

func (s *Service) ListProfiles() []interfaces.Profile {
	return s.api.ListProfiles().Profiles
}

func (s *Service) GetProfile(profileID string) (interfaces.Profile, error) {
	resp, err := s.api.GetProfile(profileID)
	if err != nil {
		return interfaces.Profile{}, err
	}
	return resp.Profile, nil
}

func (s *Service) SetProfileNodeEnabled(profileID string, nodeIndex int, enabled bool) (interfaces.Profile, error) {
	resp, err := s.api.UpdateProfileNodeSelection(interfaces.ProfileNodeSelectionUpdateRequest{
		ProfileID: profileID,
		NodeIndex: nodeIndex,
		Enabled:   enabled,
	})
	if err != nil {
		return interfaces.Profile{}, err
	}
	return resp.Profile, nil
}

func (s *Service) MoveProfileNode(profileID string, nodeIndex int, targetIndex int) (interfaces.Profile, error) {
	resp, err := s.api.UpdateProfileNodeOrder(interfaces.ProfileNodeOrderUpdateRequest{
		ProfileID:   profileID,
		NodeIndex:   nodeIndex,
		TargetIndex: targetIndex,
	})
	if err != nil {
		return interfaces.Profile{}, err
	}
	return resp.Profile, nil
}

func (s *Service) ImportProfile(req interfaces.ProfileImportRequest) (interfaces.Profile, error) {
	resp, err := s.api.ImportProfile(req)
	if err != nil {
		return interfaces.Profile{}, err
	}
	return resp.Profile, nil
}

func (s *Service) RefreshProfile(profileID string) (interfaces.Profile, error) {
	resp, err := s.api.RefreshProfile(interfaces.ProfileRefreshRequest{ProfileID: profileID})
	if err != nil {
		return interfaces.Profile{}, err
	}
	return resp.Profile, nil
}

func (s *Service) DeleteProfile(profileID string) error {
	_, err := s.api.DeleteProfile(interfaces.ProfileDeleteRequest{ProfileID: profileID})
	return err
}

func (s *Service) ListDownloadTargets() ([]interfaces.DownloadTarget, error) {
	resp, err := s.api.ListDownloadTargets()
	if err != nil {
		return nil, err
	}
	return resp.Targets, nil
}

func (s *Service) SearchDownloadTargets(search string) ([]interfaces.DownloadTarget, error) {
	resp, err := s.api.SearchDownloadTargets(search)
	if err != nil {
		return nil, err
	}
	return resp.Targets, nil
}

func (s *Service) ListTasks() []interfaces.TaskState {
	return s.api.ListTasks().Tasks
}

func (s *Service) GetTask(taskID string) (interfaces.TaskStatusResponse, error) {
	return s.api.GetTask(taskID)
}

func (s *Service) CancelTask(taskID string) error {
	return s.api.CancelTask(taskID)
}

func (s *Service) DeleteTask(taskID string) error {
	return s.api.DeleteTask(taskID)
}

func (s *Service) ListResultArchives() ([]interfaces.ResultArchiveListItem, error) {
	resp, err := s.api.ListResultArchives()
	if err != nil {
		return nil, err
	}
	return resp.Archives, nil
}

func (s *Service) GetResultArchive(taskID string) (interfaces.ResultArchive, error) {
	resp, err := s.api.GetResultArchive(taskID)
	if err != nil {
		return interfaces.ResultArchive{}, err
	}
	return resp.Archive, nil
}

func (s *Service) DeleteResultArchive(taskID string) error {
	_, err := s.api.DeleteResultArchive(interfaces.ResultArchiveDeleteRequest{TaskID: taskID})
	return err
}

func (s *Service) GetRemoteStatus() (interfaces.ClusterStatus, error) {
	resp, err := s.api.ClusterStatus()
	if err != nil {
		return interfaces.ClusterStatus{}, err
	}
	return resp.Status, nil
}

func (s *Service) ProbeRemoteMasterEligibility() (interfaces.MasterEligibility, error) {
	resp, err := s.api.ProbeMasterEligibility()
	if err != nil {
		return interfaces.MasterEligibility{}, err
	}
	return resp.Eligibility, nil
}

func (s *Service) EnableRemoteMaster(listenHost string, listenPort int) (interfaces.ClusterStatus, error) {
	resp, err := s.api.EnableMaster(interfaces.ClusterEnableMasterRequest{
		ListenHost: listenHost,
		ListenPort: listenPort,
	})
	if err != nil {
		return interfaces.ClusterStatus{}, err
	}
	return resp.Status, nil
}

func (s *Service) CreateRemotePairingCode(knightName string, ttlSeconds int) (interfaces.ClusterPairingCode, error) {
	resp, err := s.api.CreatePairingCode(interfaces.ClusterCreatePairingCodeRequest{
		KnightName: knightName,
		TTLSeconds: ttlSeconds,
	})
	if err != nil {
		return interfaces.ClusterPairingCode{}, err
	}
	return resp.PairingCode, nil
}

func (s *Service) EnableRemoteKnight(masterHost string, masterPort int, oneTimeCode string) (interfaces.ClusterStatus, error) {
	resp, err := s.api.EnableKnight(interfaces.ClusterEnableKnightRequest{
		MasterHost:  masterHost,
		MasterPort:  masterPort,
		OneTimeCode: oneTimeCode,
	})
	if err != nil {
		return interfaces.ClusterStatus{}, err
	}
	return resp.Status, nil
}

func (s *Service) DisableRemoteMode() (interfaces.ClusterStatus, error) {
	resp, err := s.api.DisableRemote()
	if err != nil {
		return interfaces.ClusterStatus{}, err
	}
	return resp.Status, nil
}

func (s *Service) ListRemoteKnights() ([]interfaces.ClusterConnectedKnight, error) {
	resp, err := s.api.ListClusterKnights()
	if err != nil {
		return nil, err
	}
	return resp.Knights, nil
}

func (s *Service) KickRemoteKnight(knightID string) (interfaces.ClusterStatus, error) {
	resp, err := s.api.KickKnight(interfaces.ClusterKickKnightRequest{KnightID: knightID})
	if err != nil {
		return interfaces.ClusterStatus{}, err
	}
	return resp.Status, nil
}

func (s *Service) ListRemoteTasks() ([]interfaces.ClusterRemoteTask, error) {
	resp, err := s.api.ListRemoteTasks()
	if err != nil {
		return nil, err
	}
	return resp.Tasks, nil
}

func (s *Service) ListRemoteMasterResultArchives() ([]interfaces.ResultArchiveListItem, error) {
	resp, err := s.api.ListRemoteMasterResultArchives()
	if err != nil {
		return nil, err
	}
	return resp.Archives, nil
}

func (s *Service) GetRemoteMasterResultArchive(taskID string) (interfaces.ResultArchive, error) {
	resp, err := s.api.GetRemoteMasterResultArchive(taskID)
	if err != nil {
		return interfaces.ResultArchive{}, err
	}
	return resp.Archive, nil
}

func (s *Service) DeleteRemoteMasterResultArchive(taskID string) error {
	_, err := s.api.DeleteRemoteMasterResultArchive(interfaces.ResultArchiveDeleteRequest{TaskID: taskID})
	return err
}

func (s *Service) ListRemoteKnightResultArchives() ([]interfaces.ResultArchiveListItem, error) {
	resp, err := s.api.ListRemoteKnightResultArchives()
	if err != nil {
		return nil, err
	}
	return resp.Archives, nil
}

func (s *Service) GetRemoteKnightResultArchive(taskID string) (interfaces.ResultArchive, error) {
	resp, err := s.api.GetRemoteKnightResultArchive(taskID)
	if err != nil {
		return interfaces.ResultArchive{}, err
	}
	return resp.Archive, nil
}

func (s *Service) DeleteRemoteKnightResultArchive(taskID string) error {
	_, err := s.api.DeleteRemoteKnightResultArchive(interfaces.ResultArchiveDeleteRequest{TaskID: taskID})
	return err
}
