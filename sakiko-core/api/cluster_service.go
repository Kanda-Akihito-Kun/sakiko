package api

import (
	"context"

	"sakiko.local/sakiko-core/interfaces"

	"go.uber.org/zap"
)

func (s *Service) ClusterStatus() (interfaces.ClusterStatusResponse, error) {
	c, err := s.requireCluster("cluster status")
	if err != nil {
		return interfaces.ClusterStatusResponse{}, err
	}
	return interfaces.ClusterStatusResponse{Status: c.Status()}, nil
}

func (s *Service) ProbeMasterEligibility() (interfaces.ClusterMasterEligibilityResponse, error) {
	c, err := s.requireCluster("cluster master eligibility")
	if err != nil {
		return interfaces.ClusterMasterEligibilityResponse{}, err
	}

	eligibility := c.ProbeMasterEligibility(context.Background())
	return interfaces.ClusterMasterEligibilityResponse{
		Eligibility: eligibility,
		Status:      c.Status(),
	}, nil
}

func (s *Service) EnableMaster(req interfaces.ClusterEnableMasterRequest) (interfaces.ClusterEnableMasterResponse, error) {
	c, err := s.requireCluster("enable master")
	if err != nil {
		return interfaces.ClusterEnableMasterResponse{}, err
	}

	status, err := c.EnableMaster(context.Background(), req)
	if err != nil {
		apiLogger().Warn("enable master failed",
			zap.String("listen_host", req.ListenHost),
			zap.Int("listen_port", req.ListenPort),
			zap.Error(err),
		)
		return interfaces.ClusterEnableMasterResponse{}, err
	}
	return interfaces.ClusterEnableMasterResponse{Status: status}, nil
}

func (s *Service) EnableKnight(req interfaces.ClusterEnableKnightRequest) (interfaces.ClusterEnableKnightResponse, error) {
	c, err := s.requireCluster("enable knight")
	if err != nil {
		return interfaces.ClusterEnableKnightResponse{}, err
	}

	status, err := c.EnableKnight(req)
	if err != nil {
		apiLogger().Warn("enable knight failed",
			zap.String("master_host", req.MasterHost),
			zap.Int("master_port", req.MasterPort),
			zap.Error(err),
		)
		return interfaces.ClusterEnableKnightResponse{}, err
	}
	return interfaces.ClusterEnableKnightResponse{Status: status}, nil
}

func (s *Service) DisableRemote() (interfaces.ClusterDisableRemoteResponse, error) {
	c, err := s.requireCluster("disable remote")
	if err != nil {
		return interfaces.ClusterDisableRemoteResponse{}, err
	}
	return interfaces.ClusterDisableRemoteResponse{Status: c.DisableRemote()}, nil
}

func (s *Service) ListClusterKnights() (interfaces.ClusterListKnightsResponse, error) {
	c, err := s.requireCluster("list cluster knights")
	if err != nil {
		return interfaces.ClusterListKnightsResponse{}, err
	}
	return interfaces.ClusterListKnightsResponse{Knights: c.ListKnights()}, nil
}

func (s *Service) KickKnight(req interfaces.ClusterKickKnightRequest) (interfaces.ClusterKickKnightResponse, error) {
	c, err := s.requireCluster("kick knight")
	if err != nil {
		return interfaces.ClusterKickKnightResponse{}, err
	}

	status, err := c.KickKnight(req)
	if err != nil {
		apiLogger().Warn("kick knight failed",
			zap.String("knight_id", req.KnightID),
			zap.Error(err),
		)
		return interfaces.ClusterKickKnightResponse{}, err
	}
	return interfaces.ClusterKickKnightResponse{Status: status}, nil
}

func (s *Service) DispatchRemoteTask(req interfaces.ClusterDispatchTaskRequest) (interfaces.ClusterDispatchTaskResponse, error) {
	c, err := s.requireCluster("dispatch remote task")
	if err != nil {
		return interfaces.ClusterDispatchTaskResponse{}, err
	}

	tasks, err := c.DispatchTask(req)
	if err != nil {
		apiLogger().Warn("dispatch remote task failed",
			zap.Int("knight_count", len(req.KnightIDs)),
			zap.String("task_name", req.Task.Name),
			zap.Error(err),
		)
		return interfaces.ClusterDispatchTaskResponse{}, err
	}
	return interfaces.ClusterDispatchTaskResponse{Tasks: tasks}, nil
}

func (s *Service) ListRemoteTasks() (interfaces.ClusterListRemoteTasksResponse, error) {
	c, err := s.requireCluster("list remote tasks")
	if err != nil {
		return interfaces.ClusterListRemoteTasksResponse{}, err
	}
	return interfaces.ClusterListRemoteTasksResponse{Tasks: c.ListRemoteTasks()}, nil
}

func (s *Service) CreatePairingCode(req interfaces.ClusterCreatePairingCodeRequest) (interfaces.ClusterCreatePairingCodeResponse, error) {
	c, err := s.requireCluster("create pairing code")
	if err != nil {
		return interfaces.ClusterCreatePairingCodeResponse{}, err
	}

	code, err := c.CreatePairingCode(req)
	if err != nil {
		apiLogger().Warn("create pairing code failed", zap.Error(err))
		return interfaces.ClusterCreatePairingCodeResponse{}, err
	}
	return interfaces.ClusterCreatePairingCodeResponse{
		PairingCode: code,
		Status:      c.Status(),
	}, nil
}

func (s *Service) BootstrapKnight(req interfaces.ClusterPairingBootstrapRequest) (interfaces.ClusterPairingBootstrapResponse, error) {
	c, err := s.requireCluster("bootstrap knight")
	if err != nil {
		return interfaces.ClusterPairingBootstrapResponse{}, err
	}

	resp, err := c.BootstrapKnight(req)
	if err != nil {
		apiLogger().Warn("bootstrap knight failed",
			zap.String("knight_id", req.KnightID),
			zap.String("knight_name", req.KnightName),
			zap.Error(err),
		)
		return interfaces.ClusterPairingBootstrapResponse{}, err
	}
	return resp, nil
}
