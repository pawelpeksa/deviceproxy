// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks_test

import mock "github.com/stretchr/testify/mock"

// Listener is an autogenerated mock type for the Listener type
type Listener struct {
	mock.Mock
}

// OnClientDisconnected provides a mock function with given fields: connectionID, deviceTag
func (_m *Listener) OnClientDisconnected(connectionID string, deviceTag *string) error {
	ret := _m.Called(connectionID, deviceTag)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *string) error); ok {
		r0 = rf(connectionID, deviceTag)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OnConnectionEstabilishedFromClient provides a mock function with given fields: connectionID, deviceTag
func (_m *Listener) OnConnectionEstabilishedFromClient(connectionID string, deviceTag *string) error {
	ret := _m.Called(connectionID, deviceTag)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *string) error); ok {
		r0 = rf(connectionID, deviceTag)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OnMessageReceivedFromClient provides a mock function with given fields: connectionID, msg, deviceTag
func (_m *Listener) OnMessageReceivedFromClient(connectionID string, msg *string, deviceTag *string) error {
	ret := _m.Called(connectionID, msg, deviceTag)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *string, *string) error); ok {
		r0 = rf(connectionID, msg, deviceTag)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OnServerStopped provides a mock function with given fields:
func (_m *Listener) OnServerStopped() {
	_m.Called()
}