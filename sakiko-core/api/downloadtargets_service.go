package api

import (
	"sakiko.local/sakiko-core/interfaces"

	"go.uber.org/zap"
)

func (s *Service) ListDownloadTargets() (interfaces.DownloadTargetListResponse, error) {
	dt, err := s.requireDownloadTargets("list download targets")
	if err != nil {
		return interfaces.DownloadTargetListResponse{}, err
	}

	targets, err := dt.List()
	if err != nil {
		apiLogger().Warn("list download targets failed", zap.Error(err))
		return interfaces.DownloadTargetListResponse{}, err
	}

	return interfaces.DownloadTargetListResponse{Targets: targets}, nil
}

func (s *Service) SearchDownloadTargets(search string) (interfaces.DownloadTargetListResponse, error) {
	dt, err := s.requireDownloadTargets("search download targets")
	if err != nil {
		return interfaces.DownloadTargetListResponse{}, err
	}

	targets, err := dt.ListBySearch(search)
	if err != nil {
		apiLogger().Warn("search download targets failed",
			zap.String("search", search),
			zap.Error(err),
		)
		return interfaces.DownloadTargetListResponse{}, err
	}

	return interfaces.DownloadTargetListResponse{Targets: targets}, nil
}
