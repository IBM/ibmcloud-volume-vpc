// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/vpcvolume"
	"go.uber.org/zap"
)

type SnapshotManager struct {
	CheckSnapshotTagStub        func(string, string, string, *zap.Logger) error
	checkSnapshotTagMutex       sync.RWMutex
	checkSnapshotTagArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 string
		arg4 *zap.Logger
	}
	checkSnapshotTagReturns struct {
		result1 error
	}
	checkSnapshotTagReturnsOnCall map[int]struct {
		result1 error
	}
	CreateSnapshotStub        func(*models.Snapshot, *zap.Logger) (*models.Snapshot, error)
	createSnapshotMutex       sync.RWMutex
	createSnapshotArgsForCall []struct {
		arg1 *models.Snapshot
		arg2 *zap.Logger
	}
	createSnapshotReturns struct {
		result1 *models.Snapshot
		result2 error
	}
	createSnapshotReturnsOnCall map[int]struct {
		result1 *models.Snapshot
		result2 error
	}
	DeleteSnapshotStub        func(string, *zap.Logger) error
	deleteSnapshotMutex       sync.RWMutex
	deleteSnapshotArgsForCall []struct {
		arg1 string
		arg2 *zap.Logger
	}
	deleteSnapshotReturns struct {
		result1 error
	}
	deleteSnapshotReturnsOnCall map[int]struct {
		result1 error
	}
	DeleteSnapshotTagStub        func(string, string, string, *zap.Logger) error
	deleteSnapshotTagMutex       sync.RWMutex
	deleteSnapshotTagArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 string
		arg4 *zap.Logger
	}
	deleteSnapshotTagReturns struct {
		result1 error
	}
	deleteSnapshotTagReturnsOnCall map[int]struct {
		result1 error
	}
	GetSnapshotStub        func(string, *zap.Logger) (*models.Snapshot, error)
	getSnapshotMutex       sync.RWMutex
	getSnapshotArgsForCall []struct {
		arg1 string
		arg2 *zap.Logger
	}
	getSnapshotReturns struct {
		result1 *models.Snapshot
		result2 error
	}
	getSnapshotReturnsOnCall map[int]struct {
		result1 *models.Snapshot
		result2 error
	}
	GetSnapshotByNameStub        func(string, *zap.Logger) (*models.Snapshot, error)
	getSnapshotByNameMutex       sync.RWMutex
	getSnapshotByNameArgsForCall []struct {
		arg1 string
		arg2 *zap.Logger
	}
	getSnapshotByNameReturns struct {
		result1 *models.Snapshot
		result2 error
	}
	getSnapshotByNameReturnsOnCall map[int]struct {
		result1 *models.Snapshot
		result2 error
	}
	ListSnapshotTagsStub        func(string, string, *zap.Logger) (*[]string, error)
	listSnapshotTagsMutex       sync.RWMutex
	listSnapshotTagsArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 *zap.Logger
	}
	listSnapshotTagsReturns struct {
		result1 *[]string
		result2 error
	}
	listSnapshotTagsReturnsOnCall map[int]struct {
		result1 *[]string
		result2 error
	}
	ListSnapshotsStub        func(int, string, *models.LisSnapshotFilters, *zap.Logger) (*models.SnapshotList, error)
	listSnapshotsMutex       sync.RWMutex
	listSnapshotsArgsForCall []struct {
		arg1 int
		arg2 string
		arg3 *models.LisSnapshotFilters
		arg4 *zap.Logger
	}
	listSnapshotsReturns struct {
		result1 *models.SnapshotList
		result2 error
	}
	listSnapshotsReturnsOnCall map[int]struct {
		result1 *models.SnapshotList
		result2 error
	}
	SetSnapshotTagStub        func(string, string, string, *zap.Logger) error
	setSnapshotTagMutex       sync.RWMutex
	setSnapshotTagArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 string
		arg4 *zap.Logger
	}
	setSnapshotTagReturns struct {
		result1 error
	}
	setSnapshotTagReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *SnapshotManager) CheckSnapshotTag(arg1 string, arg2 string, arg3 string, arg4 *zap.Logger) error {
	fake.checkSnapshotTagMutex.Lock()
	ret, specificReturn := fake.checkSnapshotTagReturnsOnCall[len(fake.checkSnapshotTagArgsForCall)]
	fake.checkSnapshotTagArgsForCall = append(fake.checkSnapshotTagArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 string
		arg4 *zap.Logger
	}{arg1, arg2, arg3, arg4})
	stub := fake.CheckSnapshotTagStub
	fakeReturns := fake.checkSnapshotTagReturns
	fake.recordInvocation("CheckSnapshotTag", []interface{}{arg1, arg2, arg3, arg4})
	fake.checkSnapshotTagMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *SnapshotManager) CheckSnapshotTagCallCount() int {
	fake.checkSnapshotTagMutex.RLock()
	defer fake.checkSnapshotTagMutex.RUnlock()
	return len(fake.checkSnapshotTagArgsForCall)
}

func (fake *SnapshotManager) CheckSnapshotTagCalls(stub func(string, string, string, *zap.Logger) error) {
	fake.checkSnapshotTagMutex.Lock()
	defer fake.checkSnapshotTagMutex.Unlock()
	fake.CheckSnapshotTagStub = stub
}

func (fake *SnapshotManager) CheckSnapshotTagArgsForCall(i int) (string, string, string, *zap.Logger) {
	fake.checkSnapshotTagMutex.RLock()
	defer fake.checkSnapshotTagMutex.RUnlock()
	argsForCall := fake.checkSnapshotTagArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *SnapshotManager) CheckSnapshotTagReturns(result1 error) {
	fake.checkSnapshotTagMutex.Lock()
	defer fake.checkSnapshotTagMutex.Unlock()
	fake.CheckSnapshotTagStub = nil
	fake.checkSnapshotTagReturns = struct {
		result1 error
	}{result1}
}

func (fake *SnapshotManager) CheckSnapshotTagReturnsOnCall(i int, result1 error) {
	fake.checkSnapshotTagMutex.Lock()
	defer fake.checkSnapshotTagMutex.Unlock()
	fake.CheckSnapshotTagStub = nil
	if fake.checkSnapshotTagReturnsOnCall == nil {
		fake.checkSnapshotTagReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.checkSnapshotTagReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *SnapshotManager) CreateSnapshot(arg1 *models.Snapshot, arg2 *zap.Logger) (*models.Snapshot, error) {
	fake.createSnapshotMutex.Lock()
	ret, specificReturn := fake.createSnapshotReturnsOnCall[len(fake.createSnapshotArgsForCall)]
	fake.createSnapshotArgsForCall = append(fake.createSnapshotArgsForCall, struct {
		arg1 *models.Snapshot
		arg2 *zap.Logger
	}{arg1, arg2})
	stub := fake.CreateSnapshotStub
	fakeReturns := fake.createSnapshotReturns
	fake.recordInvocation("CreateSnapshot", []interface{}{arg1, arg2})
	fake.createSnapshotMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *SnapshotManager) CreateSnapshotCallCount() int {
	fake.createSnapshotMutex.RLock()
	defer fake.createSnapshotMutex.RUnlock()
	return len(fake.createSnapshotArgsForCall)
}

func (fake *SnapshotManager) CreateSnapshotCalls(stub func(*models.Snapshot, *zap.Logger) (*models.Snapshot, error)) {
	fake.createSnapshotMutex.Lock()
	defer fake.createSnapshotMutex.Unlock()
	fake.CreateSnapshotStub = stub
}

func (fake *SnapshotManager) CreateSnapshotArgsForCall(i int) (*models.Snapshot, *zap.Logger) {
	fake.createSnapshotMutex.RLock()
	defer fake.createSnapshotMutex.RUnlock()
	argsForCall := fake.createSnapshotArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *SnapshotManager) CreateSnapshotReturns(result1 *models.Snapshot, result2 error) {
	fake.createSnapshotMutex.Lock()
	defer fake.createSnapshotMutex.Unlock()
	fake.CreateSnapshotStub = nil
	fake.createSnapshotReturns = struct {
		result1 *models.Snapshot
		result2 error
	}{result1, result2}
}

func (fake *SnapshotManager) CreateSnapshotReturnsOnCall(i int, result1 *models.Snapshot, result2 error) {
	fake.createSnapshotMutex.Lock()
	defer fake.createSnapshotMutex.Unlock()
	fake.CreateSnapshotStub = nil
	if fake.createSnapshotReturnsOnCall == nil {
		fake.createSnapshotReturnsOnCall = make(map[int]struct {
			result1 *models.Snapshot
			result2 error
		})
	}
	fake.createSnapshotReturnsOnCall[i] = struct {
		result1 *models.Snapshot
		result2 error
	}{result1, result2}
}

func (fake *SnapshotManager) DeleteSnapshot(arg1 string, arg2 *zap.Logger) error {
	fake.deleteSnapshotMutex.Lock()
	ret, specificReturn := fake.deleteSnapshotReturnsOnCall[len(fake.deleteSnapshotArgsForCall)]
	fake.deleteSnapshotArgsForCall = append(fake.deleteSnapshotArgsForCall, struct {
		arg1 string
		arg2 *zap.Logger
	}{arg1, arg2})
	stub := fake.DeleteSnapshotStub
	fakeReturns := fake.deleteSnapshotReturns
	fake.recordInvocation("DeleteSnapshot", []interface{}{arg1, arg2})
	fake.deleteSnapshotMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *SnapshotManager) DeleteSnapshotCallCount() int {
	fake.deleteSnapshotMutex.RLock()
	defer fake.deleteSnapshotMutex.RUnlock()
	return len(fake.deleteSnapshotArgsForCall)
}

func (fake *SnapshotManager) DeleteSnapshotCalls(stub func(string, *zap.Logger) error) {
	fake.deleteSnapshotMutex.Lock()
	defer fake.deleteSnapshotMutex.Unlock()
	fake.DeleteSnapshotStub = stub
}

func (fake *SnapshotManager) DeleteSnapshotArgsForCall(i int) (string, *zap.Logger) {
	fake.deleteSnapshotMutex.RLock()
	defer fake.deleteSnapshotMutex.RUnlock()
	argsForCall := fake.deleteSnapshotArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *SnapshotManager) DeleteSnapshotReturns(result1 error) {
	fake.deleteSnapshotMutex.Lock()
	defer fake.deleteSnapshotMutex.Unlock()
	fake.DeleteSnapshotStub = nil
	fake.deleteSnapshotReturns = struct {
		result1 error
	}{result1}
}

func (fake *SnapshotManager) DeleteSnapshotReturnsOnCall(i int, result1 error) {
	fake.deleteSnapshotMutex.Lock()
	defer fake.deleteSnapshotMutex.Unlock()
	fake.DeleteSnapshotStub = nil
	if fake.deleteSnapshotReturnsOnCall == nil {
		fake.deleteSnapshotReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.deleteSnapshotReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *SnapshotManager) DeleteSnapshotTag(arg1 string, arg2 string, arg3 string, arg4 *zap.Logger) error {
	fake.deleteSnapshotTagMutex.Lock()
	ret, specificReturn := fake.deleteSnapshotTagReturnsOnCall[len(fake.deleteSnapshotTagArgsForCall)]
	fake.deleteSnapshotTagArgsForCall = append(fake.deleteSnapshotTagArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 string
		arg4 *zap.Logger
	}{arg1, arg2, arg3, arg4})
	stub := fake.DeleteSnapshotTagStub
	fakeReturns := fake.deleteSnapshotTagReturns
	fake.recordInvocation("DeleteSnapshotTag", []interface{}{arg1, arg2, arg3, arg4})
	fake.deleteSnapshotTagMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *SnapshotManager) DeleteSnapshotTagCallCount() int {
	fake.deleteSnapshotTagMutex.RLock()
	defer fake.deleteSnapshotTagMutex.RUnlock()
	return len(fake.deleteSnapshotTagArgsForCall)
}

func (fake *SnapshotManager) DeleteSnapshotTagCalls(stub func(string, string, string, *zap.Logger) error) {
	fake.deleteSnapshotTagMutex.Lock()
	defer fake.deleteSnapshotTagMutex.Unlock()
	fake.DeleteSnapshotTagStub = stub
}

func (fake *SnapshotManager) DeleteSnapshotTagArgsForCall(i int) (string, string, string, *zap.Logger) {
	fake.deleteSnapshotTagMutex.RLock()
	defer fake.deleteSnapshotTagMutex.RUnlock()
	argsForCall := fake.deleteSnapshotTagArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *SnapshotManager) DeleteSnapshotTagReturns(result1 error) {
	fake.deleteSnapshotTagMutex.Lock()
	defer fake.deleteSnapshotTagMutex.Unlock()
	fake.DeleteSnapshotTagStub = nil
	fake.deleteSnapshotTagReturns = struct {
		result1 error
	}{result1}
}

func (fake *SnapshotManager) DeleteSnapshotTagReturnsOnCall(i int, result1 error) {
	fake.deleteSnapshotTagMutex.Lock()
	defer fake.deleteSnapshotTagMutex.Unlock()
	fake.DeleteSnapshotTagStub = nil
	if fake.deleteSnapshotTagReturnsOnCall == nil {
		fake.deleteSnapshotTagReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.deleteSnapshotTagReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *SnapshotManager) GetSnapshot(arg1 string, arg2 *zap.Logger) (*models.Snapshot, error) {
	fake.getSnapshotMutex.Lock()
	ret, specificReturn := fake.getSnapshotReturnsOnCall[len(fake.getSnapshotArgsForCall)]
	fake.getSnapshotArgsForCall = append(fake.getSnapshotArgsForCall, struct {
		arg1 string
		arg2 *zap.Logger
	}{arg1, arg2})
	stub := fake.GetSnapshotStub
	fakeReturns := fake.getSnapshotReturns
	fake.recordInvocation("GetSnapshot", []interface{}{arg1, arg2})
	fake.getSnapshotMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *SnapshotManager) GetSnapshotCallCount() int {
	fake.getSnapshotMutex.RLock()
	defer fake.getSnapshotMutex.RUnlock()
	return len(fake.getSnapshotArgsForCall)
}

func (fake *SnapshotManager) GetSnapshotCalls(stub func(string, *zap.Logger) (*models.Snapshot, error)) {
	fake.getSnapshotMutex.Lock()
	defer fake.getSnapshotMutex.Unlock()
	fake.GetSnapshotStub = stub
}

func (fake *SnapshotManager) GetSnapshotArgsForCall(i int) (string, *zap.Logger) {
	fake.getSnapshotMutex.RLock()
	defer fake.getSnapshotMutex.RUnlock()
	argsForCall := fake.getSnapshotArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *SnapshotManager) GetSnapshotReturns(result1 *models.Snapshot, result2 error) {
	fake.getSnapshotMutex.Lock()
	defer fake.getSnapshotMutex.Unlock()
	fake.GetSnapshotStub = nil
	fake.getSnapshotReturns = struct {
		result1 *models.Snapshot
		result2 error
	}{result1, result2}
}

func (fake *SnapshotManager) GetSnapshotReturnsOnCall(i int, result1 *models.Snapshot, result2 error) {
	fake.getSnapshotMutex.Lock()
	defer fake.getSnapshotMutex.Unlock()
	fake.GetSnapshotStub = nil
	if fake.getSnapshotReturnsOnCall == nil {
		fake.getSnapshotReturnsOnCall = make(map[int]struct {
			result1 *models.Snapshot
			result2 error
		})
	}
	fake.getSnapshotReturnsOnCall[i] = struct {
		result1 *models.Snapshot
		result2 error
	}{result1, result2}
}

func (fake *SnapshotManager) GetSnapshotByName(arg1 string, arg2 *zap.Logger) (*models.Snapshot, error) {
	fake.getSnapshotByNameMutex.Lock()
	ret, specificReturn := fake.getSnapshotByNameReturnsOnCall[len(fake.getSnapshotByNameArgsForCall)]
	fake.getSnapshotByNameArgsForCall = append(fake.getSnapshotByNameArgsForCall, struct {
		arg1 string
		arg2 *zap.Logger
	}{arg1, arg2})
	stub := fake.GetSnapshotByNameStub
	fakeReturns := fake.getSnapshotByNameReturns
	fake.recordInvocation("GetSnapshotByName", []interface{}{arg1, arg2})
	fake.getSnapshotByNameMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *SnapshotManager) GetSnapshotByNameCallCount() int {
	fake.getSnapshotByNameMutex.RLock()
	defer fake.getSnapshotByNameMutex.RUnlock()
	return len(fake.getSnapshotByNameArgsForCall)
}

func (fake *SnapshotManager) GetSnapshotByNameCalls(stub func(string, *zap.Logger) (*models.Snapshot, error)) {
	fake.getSnapshotByNameMutex.Lock()
	defer fake.getSnapshotByNameMutex.Unlock()
	fake.GetSnapshotByNameStub = stub
}

func (fake *SnapshotManager) GetSnapshotByNameArgsForCall(i int) (string, *zap.Logger) {
	fake.getSnapshotByNameMutex.RLock()
	defer fake.getSnapshotByNameMutex.RUnlock()
	argsForCall := fake.getSnapshotByNameArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *SnapshotManager) GetSnapshotByNameReturns(result1 *models.Snapshot, result2 error) {
	fake.getSnapshotByNameMutex.Lock()
	defer fake.getSnapshotByNameMutex.Unlock()
	fake.GetSnapshotByNameStub = nil
	fake.getSnapshotByNameReturns = struct {
		result1 *models.Snapshot
		result2 error
	}{result1, result2}
}

func (fake *SnapshotManager) GetSnapshotByNameReturnsOnCall(i int, result1 *models.Snapshot, result2 error) {
	fake.getSnapshotByNameMutex.Lock()
	defer fake.getSnapshotByNameMutex.Unlock()
	fake.GetSnapshotByNameStub = nil
	if fake.getSnapshotByNameReturnsOnCall == nil {
		fake.getSnapshotByNameReturnsOnCall = make(map[int]struct {
			result1 *models.Snapshot
			result2 error
		})
	}
	fake.getSnapshotByNameReturnsOnCall[i] = struct {
		result1 *models.Snapshot
		result2 error
	}{result1, result2}
}

func (fake *SnapshotManager) ListSnapshotTags(arg1 string, arg2 string, arg3 *zap.Logger) (*[]string, error) {
	fake.listSnapshotTagsMutex.Lock()
	ret, specificReturn := fake.listSnapshotTagsReturnsOnCall[len(fake.listSnapshotTagsArgsForCall)]
	fake.listSnapshotTagsArgsForCall = append(fake.listSnapshotTagsArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 *zap.Logger
	}{arg1, arg2, arg3})
	stub := fake.ListSnapshotTagsStub
	fakeReturns := fake.listSnapshotTagsReturns
	fake.recordInvocation("ListSnapshotTags", []interface{}{arg1, arg2, arg3})
	fake.listSnapshotTagsMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *SnapshotManager) ListSnapshotTagsCallCount() int {
	fake.listSnapshotTagsMutex.RLock()
	defer fake.listSnapshotTagsMutex.RUnlock()
	return len(fake.listSnapshotTagsArgsForCall)
}

func (fake *SnapshotManager) ListSnapshotTagsCalls(stub func(string, string, *zap.Logger) (*[]string, error)) {
	fake.listSnapshotTagsMutex.Lock()
	defer fake.listSnapshotTagsMutex.Unlock()
	fake.ListSnapshotTagsStub = stub
}

func (fake *SnapshotManager) ListSnapshotTagsArgsForCall(i int) (string, string, *zap.Logger) {
	fake.listSnapshotTagsMutex.RLock()
	defer fake.listSnapshotTagsMutex.RUnlock()
	argsForCall := fake.listSnapshotTagsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *SnapshotManager) ListSnapshotTagsReturns(result1 *[]string, result2 error) {
	fake.listSnapshotTagsMutex.Lock()
	defer fake.listSnapshotTagsMutex.Unlock()
	fake.ListSnapshotTagsStub = nil
	fake.listSnapshotTagsReturns = struct {
		result1 *[]string
		result2 error
	}{result1, result2}
}

func (fake *SnapshotManager) ListSnapshotTagsReturnsOnCall(i int, result1 *[]string, result2 error) {
	fake.listSnapshotTagsMutex.Lock()
	defer fake.listSnapshotTagsMutex.Unlock()
	fake.ListSnapshotTagsStub = nil
	if fake.listSnapshotTagsReturnsOnCall == nil {
		fake.listSnapshotTagsReturnsOnCall = make(map[int]struct {
			result1 *[]string
			result2 error
		})
	}
	fake.listSnapshotTagsReturnsOnCall[i] = struct {
		result1 *[]string
		result2 error
	}{result1, result2}
}

func (fake *SnapshotManager) ListSnapshots(arg1 int, arg2 string, arg3 *models.LisSnapshotFilters, arg4 *zap.Logger) (*models.SnapshotList, error) {
	fake.listSnapshotsMutex.Lock()
	ret, specificReturn := fake.listSnapshotsReturnsOnCall[len(fake.listSnapshotsArgsForCall)]
	fake.listSnapshotsArgsForCall = append(fake.listSnapshotsArgsForCall, struct {
		arg1 int
		arg2 string
		arg3 *models.LisSnapshotFilters
		arg4 *zap.Logger
	}{arg1, arg2, arg3, arg4})
	stub := fake.ListSnapshotsStub
	fakeReturns := fake.listSnapshotsReturns
	fake.recordInvocation("ListSnapshots", []interface{}{arg1, arg2, arg3, arg4})
	fake.listSnapshotsMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *SnapshotManager) ListSnapshotsCallCount() int {
	fake.listSnapshotsMutex.RLock()
	defer fake.listSnapshotsMutex.RUnlock()
	return len(fake.listSnapshotsArgsForCall)
}

func (fake *SnapshotManager) ListSnapshotsCalls(stub func(int, string, *models.LisSnapshotFilters, *zap.Logger) (*models.SnapshotList, error)) {
	fake.listSnapshotsMutex.Lock()
	defer fake.listSnapshotsMutex.Unlock()
	fake.ListSnapshotsStub = stub
}

func (fake *SnapshotManager) ListSnapshotsArgsForCall(i int) (int, string, *models.LisSnapshotFilters, *zap.Logger) {
	fake.listSnapshotsMutex.RLock()
	defer fake.listSnapshotsMutex.RUnlock()
	argsForCall := fake.listSnapshotsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *SnapshotManager) ListSnapshotsReturns(result1 *models.SnapshotList, result2 error) {
	fake.listSnapshotsMutex.Lock()
	defer fake.listSnapshotsMutex.Unlock()
	fake.ListSnapshotsStub = nil
	fake.listSnapshotsReturns = struct {
		result1 *models.SnapshotList
		result2 error
	}{result1, result2}
}

func (fake *SnapshotManager) ListSnapshotsReturnsOnCall(i int, result1 *models.SnapshotList, result2 error) {
	fake.listSnapshotsMutex.Lock()
	defer fake.listSnapshotsMutex.Unlock()
	fake.ListSnapshotsStub = nil
	if fake.listSnapshotsReturnsOnCall == nil {
		fake.listSnapshotsReturnsOnCall = make(map[int]struct {
			result1 *models.SnapshotList
			result2 error
		})
	}
	fake.listSnapshotsReturnsOnCall[i] = struct {
		result1 *models.SnapshotList
		result2 error
	}{result1, result2}
}

func (fake *SnapshotManager) SetSnapshotTag(arg1 string, arg2 string, arg3 string, arg4 *zap.Logger) error {
	fake.setSnapshotTagMutex.Lock()
	ret, specificReturn := fake.setSnapshotTagReturnsOnCall[len(fake.setSnapshotTagArgsForCall)]
	fake.setSnapshotTagArgsForCall = append(fake.setSnapshotTagArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 string
		arg4 *zap.Logger
	}{arg1, arg2, arg3, arg4})
	stub := fake.SetSnapshotTagStub
	fakeReturns := fake.setSnapshotTagReturns
	fake.recordInvocation("SetSnapshotTag", []interface{}{arg1, arg2, arg3, arg4})
	fake.setSnapshotTagMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *SnapshotManager) SetSnapshotTagCallCount() int {
	fake.setSnapshotTagMutex.RLock()
	defer fake.setSnapshotTagMutex.RUnlock()
	return len(fake.setSnapshotTagArgsForCall)
}

func (fake *SnapshotManager) SetSnapshotTagCalls(stub func(string, string, string, *zap.Logger) error) {
	fake.setSnapshotTagMutex.Lock()
	defer fake.setSnapshotTagMutex.Unlock()
	fake.SetSnapshotTagStub = stub
}

func (fake *SnapshotManager) SetSnapshotTagArgsForCall(i int) (string, string, string, *zap.Logger) {
	fake.setSnapshotTagMutex.RLock()
	defer fake.setSnapshotTagMutex.RUnlock()
	argsForCall := fake.setSnapshotTagArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *SnapshotManager) SetSnapshotTagReturns(result1 error) {
	fake.setSnapshotTagMutex.Lock()
	defer fake.setSnapshotTagMutex.Unlock()
	fake.SetSnapshotTagStub = nil
	fake.setSnapshotTagReturns = struct {
		result1 error
	}{result1}
}

func (fake *SnapshotManager) SetSnapshotTagReturnsOnCall(i int, result1 error) {
	fake.setSnapshotTagMutex.Lock()
	defer fake.setSnapshotTagMutex.Unlock()
	fake.SetSnapshotTagStub = nil
	if fake.setSnapshotTagReturnsOnCall == nil {
		fake.setSnapshotTagReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.setSnapshotTagReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *SnapshotManager) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.checkSnapshotTagMutex.RLock()
	defer fake.checkSnapshotTagMutex.RUnlock()
	fake.createSnapshotMutex.RLock()
	defer fake.createSnapshotMutex.RUnlock()
	fake.deleteSnapshotMutex.RLock()
	defer fake.deleteSnapshotMutex.RUnlock()
	fake.deleteSnapshotTagMutex.RLock()
	defer fake.deleteSnapshotTagMutex.RUnlock()
	fake.getSnapshotMutex.RLock()
	defer fake.getSnapshotMutex.RUnlock()
	fake.getSnapshotByNameMutex.RLock()
	defer fake.getSnapshotByNameMutex.RUnlock()
	fake.listSnapshotTagsMutex.RLock()
	defer fake.listSnapshotTagsMutex.RUnlock()
	fake.listSnapshotsMutex.RLock()
	defer fake.listSnapshotsMutex.RUnlock()
	fake.setSnapshotTagMutex.RLock()
	defer fake.setSnapshotTagMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *SnapshotManager) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ vpcvolume.SnapshotManager = new(SnapshotManager)
