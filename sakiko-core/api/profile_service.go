package api

import (
	"fmt"

	"sakiko.local/sakiko-core/interfaces"

	"go.uber.org/zap"
)

func (s *Service) ImportProfile(req interfaces.ProfileImportRequest) (interfaces.ProfileImportResponse, error) {
	pm, err := s.requireProfiles("import profile")
	if err != nil {
		return interfaces.ProfileImportResponse{}, err
	}

	profile, err := pm.Import(req)
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
	pm, err := s.requireProfiles("refresh profile")
	if err != nil {
		return interfaces.ProfileRefreshResponse{}, err
	}

	profile, err := pm.Refresh(req.ProfileID)
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
	pm, err := s.requireProfiles("delete profile")
	if err != nil {
		return interfaces.ProfileDeleteResponse{}, err
	}

	if err := pm.Delete(req.ProfileID); err != nil {
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
	pm, err := s.requireProfiles("")
	if err != nil {
		return interfaces.ProfileListResponse{Profiles: []interfaces.Profile{}}
	}
	return interfaces.ProfileListResponse{Profiles: pm.List()}
}

func (s *Service) GetProfile(profileID string) (interfaces.ProfileGetResponse, error) {
	pm, err := s.requireProfiles("get profile")
	if err != nil {
		return interfaces.ProfileGetResponse{}, err
	}

	profile, ok := pm.Get(profileID)
	if !ok {
		apiLogger().Debug("profile not found", zap.String("profile_id", profileID))
		return interfaces.ProfileGetResponse{}, fmt.Errorf("profile not found")
	}
	return interfaces.ProfileGetResponse{Profile: profile}, nil
}

func (s *Service) UpdateProfileNodeSelection(req interfaces.ProfileNodeSelectionUpdateRequest) (interfaces.ProfileNodeSelectionUpdateResponse, error) {
	pm, err := s.requireProfiles("update profile node selection")
	if err != nil {
		return interfaces.ProfileNodeSelectionUpdateResponse{}, err
	}

	profile, err := pm.SetNodeEnabled(req.ProfileID, req.NodeIndex, req.Enabled)
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
	pm, err := s.requireProfiles("update profile node order")
	if err != nil {
		return interfaces.ProfileNodeOrderUpdateResponse{}, err
	}

	profile, err := pm.MoveNode(req.ProfileID, req.NodeIndex, req.TargetIndex)
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
