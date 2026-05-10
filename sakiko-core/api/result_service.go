package api

import (
	"sakiko.local/sakiko-core/interfaces"

	"go.uber.org/zap"
)

func (s *Service) ListResultArchives() (interfaces.ResultArchiveListResponse, error) {
	rs, err := s.requireResultStore("list result archives")
	if err != nil {
		return interfaces.ResultArchiveListResponse{}, err
	}

	items, err := rs.List()
	if err != nil {
		apiLogger().Warn("list result archives failed", zap.Error(err))
		return interfaces.ResultArchiveListResponse{}, err
	}
	return interfaces.ResultArchiveListResponse{Archives: items}, nil
}

func (s *Service) GetResultArchive(taskID string) (interfaces.ResultArchiveGetResponse, error) {
	rs, err := s.requireResultStore("get result archive")
	if err != nil {
		return interfaces.ResultArchiveGetResponse{}, err
	}

	archive, err := rs.Load(taskID)
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
	rs, err := s.requireResultStore("delete result archive")
	if err != nil {
		return interfaces.ResultArchiveDeleteResponse{}, err
	}

	if err := rs.Delete(req.TaskID); err != nil {
		apiLogger().Warn("delete result archive failed",
			zap.String("task_id", req.TaskID),
			zap.Error(err),
		)
		return interfaces.ResultArchiveDeleteResponse{}, err
	}

	apiLogger().Debug("result archive deleted", zap.String("task_id", req.TaskID))
	return interfaces.ResultArchiveDeleteResponse{TaskID: req.TaskID}, nil
}

func (s *Service) ListRemoteMasterResultArchives() (interfaces.ResultArchiveListResponse, error) {
	rs, err := s.requireRemoteMasterStore("list remote master result archives")
	if err != nil {
		return interfaces.ResultArchiveListResponse{}, err
	}
	return listResultArchives(rs)
}

func (s *Service) GetRemoteMasterResultArchive(taskID string) (interfaces.ResultArchiveGetResponse, error) {
	rs, err := s.requireRemoteMasterStore("get remote master result archive")
	if err != nil {
		return interfaces.ResultArchiveGetResponse{}, err
	}
	return getResultArchive(rs, taskID)
}

func (s *Service) DeleteRemoteMasterResultArchive(req interfaces.ResultArchiveDeleteRequest) (interfaces.ResultArchiveDeleteResponse, error) {
	rs, err := s.requireRemoteMasterStore("delete remote master result archive")
	if err != nil {
		return interfaces.ResultArchiveDeleteResponse{}, err
	}
	return deleteResultArchive(rs, req)
}

func (s *Service) ListRemoteKnightResultArchives() (interfaces.ResultArchiveListResponse, error) {
	rs, err := s.requireRemoteKnightStore("list remote knight result archives")
	if err != nil {
		return interfaces.ResultArchiveListResponse{}, err
	}
	return listResultArchives(rs)
}

func (s *Service) GetRemoteKnightResultArchive(taskID string) (interfaces.ResultArchiveGetResponse, error) {
	rs, err := s.requireRemoteKnightStore("get remote knight result archive")
	if err != nil {
		return interfaces.ResultArchiveGetResponse{}, err
	}
	return getResultArchive(rs, taskID)
}

func (s *Service) DeleteRemoteKnightResultArchive(req interfaces.ResultArchiveDeleteRequest) (interfaces.ResultArchiveDeleteResponse, error) {
	rs, err := s.requireRemoteKnightStore("delete remote knight result archive")
	if err != nil {
		return interfaces.ResultArchiveDeleteResponse{}, err
	}
	return deleteResultArchive(rs, req)
}

func listResultArchives(store interfacesResultStore) (interfaces.ResultArchiveListResponse, error) {
	items, err := store.List()
	if err != nil {
		apiLogger().Warn("list result archives failed", zap.Error(err))
		return interfaces.ResultArchiveListResponse{}, err
	}
	return interfaces.ResultArchiveListResponse{Archives: items}, nil
}

func getResultArchive(store interfacesResultStore, taskID string) (interfaces.ResultArchiveGetResponse, error) {
	archive, err := store.Load(taskID)
	if err != nil {
		apiLogger().Warn("get result archive failed",
			zap.String("task_id", taskID),
			zap.Error(err),
		)
		return interfaces.ResultArchiveGetResponse{}, err
	}
	return interfaces.ResultArchiveGetResponse{Archive: archive}, nil
}

func deleteResultArchive(store interfacesResultStore, req interfaces.ResultArchiveDeleteRequest) (interfaces.ResultArchiveDeleteResponse, error) {
	if err := store.Delete(req.TaskID); err != nil {
		apiLogger().Warn("delete result archive failed",
			zap.String("task_id", req.TaskID),
			zap.Error(err),
		)
		return interfaces.ResultArchiveDeleteResponse{}, err
	}
	apiLogger().Debug("result archive deleted", zap.String("task_id", req.TaskID))
	return interfaces.ResultArchiveDeleteResponse{TaskID: req.TaskID}, nil
}

type interfacesResultStore interface {
	List() ([]interfaces.ResultArchiveListItem, error)
	Load(taskID string) (interfaces.ResultArchive, error)
	Delete(taskID string) error
}
