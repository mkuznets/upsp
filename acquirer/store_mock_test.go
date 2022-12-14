// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package acquirer

import (
	"sync"
)

// Ensure, that StoreMock does implement Store.
// If this is not the case, regenerate this file with moq.
var _ Store = &StoreMock{}

// StoreMock is a mock implementation of Store.
//
//	func TestSomethingThatUsesStore(t *testing.T) {
//
//		// make and configure a mocked Store
//		mockedStore := &StoreMock{
//			CreateOrGetFunc: func(payment *Payment) (*Payment, error) {
//				panic("mock out the CreateOrGet method")
//			},
//			GetFunc: func(id PaymentId) (*Payment, error) {
//				panic("mock out the Get method")
//			},
//			ListFunc: func(state PaymentState) ([]*Payment, error) {
//				panic("mock out the List method")
//			},
//			UpdateFunc: func(id PaymentId, version string, fn func(*Payment) error) (*Payment, error) {
//				panic("mock out the Update method")
//			},
//		}
//
//		// use mockedStore in code that requires Store
//		// and then make assertions.
//
//	}
type StoreMock struct {
	// CreateOrGetFunc mocks the CreateOrGet method.
	CreateOrGetFunc func(payment *Payment) (*Payment, error)

	// GetFunc mocks the Get method.
	GetFunc func(id PaymentId) (*Payment, error)

	// ListFunc mocks the List method.
	ListFunc func(state PaymentState) ([]*Payment, error)

	// UpdateFunc mocks the Update method.
	UpdateFunc func(id PaymentId, version string, fn func(*Payment) error) (*Payment, error)

	// calls tracks calls to the methods.
	calls struct {
		// CreateOrGet holds details about calls to the CreateOrGet method.
		CreateOrGet []struct {
			// Payment is the payment argument value.
			Payment *Payment
		}
		// Get holds details about calls to the Get method.
		Get []struct {
			// ID is the id argument value.
			ID PaymentId
		}
		// List holds details about calls to the List method.
		List []struct {
			// State is the state argument value.
			State PaymentState
		}
		// Update holds details about calls to the Update method.
		Update []struct {
			// ID is the id argument value.
			ID PaymentId
			// Version is the version argument value.
			Version string
			// Fn is the fn argument value.
			Fn func(*Payment) error
		}
	}
	lockCreateOrGet sync.RWMutex
	lockGet         sync.RWMutex
	lockList        sync.RWMutex
	lockUpdate      sync.RWMutex
}

// CreateOrGet calls CreateOrGetFunc.
func (mock *StoreMock) CreateOrGet(payment *Payment) (*Payment, error) {
	if mock.CreateOrGetFunc == nil {
		panic("StoreMock.CreateOrGetFunc: method is nil but Store.CreateOrGet was just called")
	}
	callInfo := struct {
		Payment *Payment
	}{
		Payment: payment,
	}
	mock.lockCreateOrGet.Lock()
	mock.calls.CreateOrGet = append(mock.calls.CreateOrGet, callInfo)
	mock.lockCreateOrGet.Unlock()
	return mock.CreateOrGetFunc(payment)
}

// CreateOrGetCalls gets all the calls that were made to CreateOrGet.
// Check the length with:
//
//	len(mockedStore.CreateOrGetCalls())
func (mock *StoreMock) CreateOrGetCalls() []struct {
	Payment *Payment
} {
	var calls []struct {
		Payment *Payment
	}
	mock.lockCreateOrGet.RLock()
	calls = mock.calls.CreateOrGet
	mock.lockCreateOrGet.RUnlock()
	return calls
}

// Get calls GetFunc.
func (mock *StoreMock) Get(id PaymentId) (*Payment, error) {
	if mock.GetFunc == nil {
		panic("StoreMock.GetFunc: method is nil but Store.Get was just called")
	}
	callInfo := struct {
		ID PaymentId
	}{
		ID: id,
	}
	mock.lockGet.Lock()
	mock.calls.Get = append(mock.calls.Get, callInfo)
	mock.lockGet.Unlock()
	return mock.GetFunc(id)
}

// GetCalls gets all the calls that were made to Get.
// Check the length with:
//
//	len(mockedStore.GetCalls())
func (mock *StoreMock) GetCalls() []struct {
	ID PaymentId
} {
	var calls []struct {
		ID PaymentId
	}
	mock.lockGet.RLock()
	calls = mock.calls.Get
	mock.lockGet.RUnlock()
	return calls
}

// List calls ListFunc.
func (mock *StoreMock) List(state PaymentState) ([]*Payment, error) {
	if mock.ListFunc == nil {
		panic("StoreMock.ListFunc: method is nil but Store.List was just called")
	}
	callInfo := struct {
		State PaymentState
	}{
		State: state,
	}
	mock.lockList.Lock()
	mock.calls.List = append(mock.calls.List, callInfo)
	mock.lockList.Unlock()
	return mock.ListFunc(state)
}

// ListCalls gets all the calls that were made to List.
// Check the length with:
//
//	len(mockedStore.ListCalls())
func (mock *StoreMock) ListCalls() []struct {
	State PaymentState
} {
	var calls []struct {
		State PaymentState
	}
	mock.lockList.RLock()
	calls = mock.calls.List
	mock.lockList.RUnlock()
	return calls
}

// Update calls UpdateFunc.
func (mock *StoreMock) Update(id PaymentId, version string, fn func(*Payment) error) (*Payment, error) {
	if mock.UpdateFunc == nil {
		panic("StoreMock.UpdateFunc: method is nil but Store.Update was just called")
	}
	callInfo := struct {
		ID      PaymentId
		Version string
		Fn      func(*Payment) error
	}{
		ID:      id,
		Version: version,
		Fn:      fn,
	}
	mock.lockUpdate.Lock()
	mock.calls.Update = append(mock.calls.Update, callInfo)
	mock.lockUpdate.Unlock()
	return mock.UpdateFunc(id, version, fn)
}

// UpdateCalls gets all the calls that were made to Update.
// Check the length with:
//
//	len(mockedStore.UpdateCalls())
func (mock *StoreMock) UpdateCalls() []struct {
	ID      PaymentId
	Version string
	Fn      func(*Payment) error
} {
	var calls []struct {
		ID      PaymentId
		Version string
		Fn      func(*Payment) error
	}
	mock.lockUpdate.RLock()
	calls = mock.calls.Update
	mock.lockUpdate.RUnlock()
	return calls
}
