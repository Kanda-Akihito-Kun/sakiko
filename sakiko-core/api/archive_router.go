package api

import (
	"strings"

	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/storage"
)

type archiveRouter struct {
	defaultStore *storage.ResultStore
	knightStore  *storage.ResultStore
}

func newArchiveRouter(defaultStore *storage.ResultStore, knightStore *storage.ResultStore) interfaces.ResultArchiveWriter {
	return &archiveRouter{
		defaultStore: defaultStore,
		knightStore:  knightStore,
	}
}

func (r *archiveRouter) SaveTaskArchive(snapshot interfaces.TaskArchiveSnapshot) error {
	if r == nil {
		return errServiceNotInitialized
	}
	store := r.defaultStore
	if isRemoteKnightSnapshot(snapshot) && r.knightStore != nil {
		store = r.knightStore
	}
	if store == nil {
		return errServiceNotInitialized
	}
	return store.SaveTaskArchive(snapshot)
}

func isRemoteKnightSnapshot(snapshot interfaces.TaskArchiveSnapshot) bool {
	if snapshot.Task.Environment == nil || snapshot.Task.Environment.Remote == nil {
		return false
	}
	return strings.EqualFold(string(snapshot.Task.Environment.Remote.Mode), string(interfaces.RemoteExecutionModeRemoteKnight))
}
