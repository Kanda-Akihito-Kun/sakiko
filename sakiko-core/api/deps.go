package api

import (
	"sakiko.local/sakiko-core/cluster"
	"sakiko.local/sakiko-core/downloadtargets"
	"sakiko.local/sakiko-core/kernel"
	"sakiko.local/sakiko-core/profiles"
	"sakiko.local/sakiko-core/storage"
)

func (s *Service) requireKernel(operation string) (*kernel.Service, error) {
	if s == nil || s.kernel == nil {
		return nil, rejectUnavailable(operation)
	}
	return s.kernel, nil
}

func (s *Service) requireCluster(operation string) (*cluster.Service, error) {
	if s == nil || s.cluster == nil {
		return nil, rejectUnavailable(operation)
	}
	return s.cluster, nil
}

func (s *Service) requireProfiles(operation string) (*profiles.Manager, error) {
	if s == nil || s.profiles == nil {
		return nil, rejectUnavailable(operation)
	}
	return s.profiles, nil
}

func (s *Service) requireDownloadTargets(operation string) (*downloadtargets.Manager, error) {
	if s == nil || s.downloadTargets == nil {
		return nil, rejectUnavailable(operation)
	}
	return s.downloadTargets, nil
}

func (s *Service) requireResultStore(operation string) (*storage.ResultStore, error) {
	if s == nil || s.resultStore == nil {
		return nil, rejectUnavailable(operation)
	}
	return s.resultStore, nil
}

func (s *Service) requireRemoteMasterStore(operation string) (*storage.ResultStore, error) {
	if s == nil || s.remoteMasterStore == nil {
		return nil, rejectUnavailable(operation)
	}
	return s.remoteMasterStore, nil
}

func (s *Service) requireRemoteKnightStore(operation string) (*storage.ResultStore, error) {
	if s == nil || s.remoteKnightStore == nil {
		return nil, rejectUnavailable(operation)
	}
	return s.remoteKnightStore, nil
}

func rejectUnavailable(operation string) error {
	if operation != "" {
		apiLogger().Warn(operation + " rejected: service not initialized")
	}
	return errServiceNotInitialized
}
