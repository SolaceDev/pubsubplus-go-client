// pubsubplus-go-client
//
// Copyright 2021-2025 Solace Corporation. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package receiver

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
	"unsafe"

	"solace.dev/go/messaging/internal/ccsmp"
	"solace.dev/go/messaging/internal/impl/core"
	"solace.dev/go/messaging/pkg/solace"
	"solace.dev/go/messaging/pkg/solace/message"
	"solace.dev/go/messaging/pkg/solace/resource"
	"solace.dev/go/messaging/pkg/solace/subcode"
)

func TestMessageReceiverStartStateChecks(t *testing.T) {
	receiver := &basicMessageReceiver{}
	receiver.construct(&mockInternalReceiver{})

	// Not started
	if receiver.IsRunning() {
		t.Error("expected IsRunning to return false when not started")
	}
	if receiver.IsTerminating() {
		t.Error("expected IsTerminating to return false when not terminating")
	}
	if receiver.IsTerminated() {
		t.Error("expected IsTerminated to return false when not terminating")
	}

	// Starting
	receiver.starting()
	if receiver.IsRunning() {
		t.Error("expected IsRunning to return false when starting")
	}
	if receiver.IsTerminating() {
		t.Error("expected IsTerminating to return false when not terminating")
	}
	if receiver.IsTerminated() {
		t.Error("expected IsTerminated to return false when not terminating")
	}
	if proceed, err := receiver.terminate(); proceed || err == nil {
		t.Error("expected terminate to return an error when not in running state, got nil")
	}

	// Started
	receiver.started(nil)
	if !receiver.IsRunning() {
		t.Error("expected IsRunning to return true when started")
	}
	if receiver.IsTerminating() {
		t.Error("expected IsTerminating to return false when not terminating")
	}
	if receiver.IsTerminated() {
		t.Error("expected IsTerminated to return false when not terminating")
	}

	// Termianting
	receiver.terminate()
	if proceed, err := receiver.starting(); proceed || err == nil {
		t.Error("expected starting to return an error when in terminating state, got nil")
	}
	if receiver.IsRunning() {
		t.Error("expected IsRunning to return false when terminating")
	}
	if !receiver.IsTerminating() {
		t.Error("expected IsTerminating to return true when terminating")
	}
	if receiver.IsTerminated() {
		t.Error("expected IsTerminated to return false when not terminating")
	}

	// Terminated
	receiver.terminated(nil)
	if receiver.IsRunning() {
		t.Error("expected IsRunning to return false when terminated")
	}
	if receiver.IsTerminating() {
		t.Error("expected IsTerminating to return false when terminated")
	}
	if !receiver.IsTerminated() {
		t.Error("expected IsTerminated to return true when terminated")
	}
	if proceed, err := receiver.starting(); proceed || err == nil {
		t.Error("expected starting to return an error when in terminated state, got nil")
	}
}

func TestMessageReceiverTerminateWhenNotStarted(t *testing.T) {
	receiver := &basicMessageReceiver{}
	receiver.construct(&mockInternalReceiver{})

	proceed, err := receiver.terminate()
	if proceed {
		t.Error("expected proceed to be false")
	}
	if err != nil {
		t.Error("expected error to be nil")
	}
	if !receiver.IsTerminated() {
		t.Error("expected IsTerminated to return true when terminated")
	}
	proceed, err = receiver.starting()
	if proceed {
		t.Error("expected proceed to be false")
	}
	if err == nil {
		t.Error("expected illegal state error")
	}
	// try calling terminate again
	proceed, err = receiver.terminate()
	if proceed {
		t.Error("expected proceed to be false")
	}
	if err != nil {
		t.Error("expected error to be nil")
	}
}

func TestMessageReceiverStartIdempotence(t *testing.T) {
	receiver := &basicMessageReceiver{}
	receiver.construct(&mockInternalReceiver{})
	proceed, err := receiver.starting()
	if !proceed || err != nil {
		t.Errorf("expected first call to start to be successful, got proceed %t and error %s", proceed, err)
	}
	numWaiters := 3
	type result struct {
		proceed bool
		err     error
	}
	results := make(chan result, numWaiters)
	for i := 0; i < numWaiters; i++ {
		go func() {
			proceed, err := receiver.starting()
			results <- result{
				proceed: proceed,
				err:     err,
			}
		}()
	}
	select {
	case <-results:
		t.Error("did not expect subsequent calls to starting to return")
	case <-time.After(100 * time.Millisecond):
		// success
	}
	myError := fmt.Errorf("hello world")
	receiver.started(myError)
	for i := 0; i < numWaiters; i++ {
		select {
		case r := <-results:
			if r.proceed || r.err != myError {
				t.Errorf("got bad response: proceed %t, error %s", r.proceed, r.err)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("timed out waiting for results to be published")
		}
	}
}

func TestMessageReceiverTerminateIdempotence(t *testing.T) {
	receiver := &basicMessageReceiver{}
	receiver.construct(&mockInternalReceiver{})
	receiver.starting()
	receiver.started(nil)
	proceed, err := receiver.terminate()
	if !proceed || err != nil {
		t.Errorf("expected first call to terminate to be successful, got proceed %t and error %s", proceed, err)
	}
	numWaiters := 3
	results := make(chan result, numWaiters)
	for i := 0; i < numWaiters; i++ {
		go func() {
			proceed, err := receiver.terminate()
			results <- result{
				proceed: proceed,
				err:     err,
			}
		}()
	}
	select {
	case <-results:
		t.Error("did not expect subsequent calls to terminate to return")
	case <-time.After(100 * time.Millisecond):
		// success
	}
	myError := fmt.Errorf("hello world")
	receiver.terminated(myError)
	for i := 0; i < numWaiters; i++ {
		select {
		case r := <-results:
			if r.proceed || r.err != myError {
				t.Errorf("got bad response: proceed %t, error %s", r.proceed, r.err)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("timed out waiting for results to be published")
		}
	}
}

func TestMessageReceiverStartAtomic(t *testing.T) {
	first := true
	notify := make(chan struct{})
	isRunningDelayed := func() bool {
		if first {
			first = false
			<-notify
		}
		return true
	}
	receiver := &basicMessageReceiver{}
	receiver.construct(&mockInternalReceiver{
		isRunning: isRunningDelayed,
	})
	firstStart := make(chan result)
	go func() {
		proceed, err := receiver.starting()
		firstStart <- result{
			proceed: proceed,
			err:     err,
		}
	}()
	// give the goroutine above a chance to start
	<-time.After(10 * time.Millisecond)
	proceed, err := receiver.starting()
	if !proceed || err != nil {
		t.Errorf("expected procceed to be true and err to be nil, got proceed: %t, err: %s", proceed, err)
	}
	close(notify)
	select {
	case <-firstStart:
		t.Error("did not expect first call to starting to return")
	case <-time.After(100 * time.Millisecond):
		// success
	}
	myError := fmt.Errorf("hello world")
	receiver.started(myError)
	select {
	case r := <-firstStart:
		if r.proceed || r.err != myError {
			t.Errorf("got bad response: proceed %t, error %s", r.proceed, r.err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timed out waiting for results to be published")
	}
}

// can't test terminate in the same way as start..

func TestStartWhenInternalReceiverIsNotRunning(t *testing.T) {
	receiver := &basicMessageReceiver{}
	receiver.construct(&mockInternalReceiver{
		isRunning: func() bool {
			return false
		},
	})
	proceed, err := receiver.starting()
	if proceed {
		t.Error("expected receiver to not start when internal receiver is not started")
	}
	if err == nil {
		t.Error("expected error to not be nil and instead indicate that internal receiver is down")
	}
}

func TestSendCacheRequestIsNotBlocking(t *testing.T) {

	var validateNativeError = func(err error, sc subcode.Code) {
		var nativeErr *solace.NativeError
		if err == nil {
			t.Errorf("Expected to get an error from sending the cache request")
		}
		ret := errors.As(err, &nativeErr)
		if ret != true {
			t.Errorf("Expected ret to be true")
		}
		if !subcode.Is(nativeErr.SubCode(), sc) {
			t.Errorf("Expected native error to have subcode %d, instead had %d", sc, nativeErr.SubCode())
		}
	}
	wouldBlockSubcodeNum := 1
	var sendCacheRequestFailsWithSolClientWouldBlockFunc = func(mock *mockInternalReceiver, _ core.CacheRequest, _ core.CoreCacheEventCallback, _ uintptr) error {
		solClientErr := ccsmp.NewInternalSolClientErrorInfoWrapper(ccsmp.SolClientReturnCode(1),
			/* NOTE: This subcode doesn't make sense, but is used so that we can assert its value
			 * later in the test. For purposes of clarity, it matches the return code value. The
			 * return code is what we'd actually like to assert on, but it is not accessible
			 * through any of the existing solace.errors interfaces.
			 */
			ccsmp.SolClientSubCode(wouldBlockSubcodeNum),
			ccsmp.SolClientResponseCode(0),
			"This is a generated error info")
		return core.ToNativeError(solClientErr)
	}
	notReadySubcodeNum := 3
	var sendCacheRequestFailsWithSolClientNotReadyFunc = func(_ *mockInternalReceiver, _ core.CacheRequest, _ core.CoreCacheEventCallback, _ uintptr) error {
		solClientErr := ccsmp.NewInternalSolClientErrorInfoWrapper(ccsmp.SolClientReturnCode(3),
			/* NOTE: This subcode doesn't make sense, but is used so that we can assert its value
			 * later in the test. For purposes of clarity, it matches the return code value. The
			 * return code is what we'd actually like to assert on, but it is not accessible
			 * through any of the existing solace.errors interfaces.
			 */
			ccsmp.SolClientSubCode(notReadySubcodeNum),
			ccsmp.SolClientResponseCode(0),
			"This is a generated error info")
		return core.ToNativeError(solClientErr)
	}
	var createCacheRequestFunc = func(mock *mockInternalReceiver, cacheRequestConfig resource.CachedMessageSubscriptionRequest, cacheRequestID message.CacheRequestID, cacheResponseProcessor core.CacheResponseProcessor, dispatchID uintptr) (core.CacheRequest, error) {
		cacheRequest := defaultMockCacheRequest()
		return cacheRequest, nil
	}
	var createCacheReceiver = func(sendCacheRequestFunc func(*mockInternalReceiver, core.CacheRequest, core.CoreCacheEventCallback, uintptr) error) *directMessageReceiverImpl {
		receiver := &directMessageReceiverImpl{}
		receiver.construct(&directMessageReceiverProps{
			internalReceiver: &mockInternalReceiver{
				isRunning: func() bool {
					return true
				},
				sendCacheRequestFunc:   sendCacheRequestFunc,
				createCacheRequestFunc: createCacheRequestFunc,
			},
			backpressureStrategy:   strategyDropOldest,
			backpressureBufferSize: 5,
		})
		return receiver
	}
	receiver := createCacheReceiver(sendCacheRequestFailsWithSolClientWouldBlockFunc)
	receiver.Start()

	receiver2 := createCacheReceiver(sendCacheRequestFailsWithSolClientNotReadyFunc)
	receiver2.Start()

	cacheRequestID := message.CacheRequestID(1)
	cacheRequestConfig := resource.NewCachedMessageSubscriptionRequest(
		resource.CacheRequestStrategyAsAvailable,
		"cache name",
		resource.TopicSubscriptionOf("/some/test/topic"),
		int32(5000),
		int32(5000),
		int32(5000))

	signal := make(chan bool)
	var requestCachedAsync = func(directReceiver *directMessageReceiverImpl, cacheRequestConfig resource.CachedMessageSubscriptionRequest, cacheRequestID message.CacheRequestID, subcodeAsNum int) {

		channel, err := directReceiver.RequestCachedAsync(cacheRequestConfig, cacheRequestID)
		validateNativeError(err, subcode.Code(subcodeAsNum))
		if channel != nil {
			t.Errorf("Expected channel for cache request to be nil since the cache request should have failed before being sent.")
		}
		signal <- true
	}
	var requestCachedAsyncWithCallback = func(directReceiver *directMessageReceiverImpl, cacheRequestConfig resource.CachedMessageSubscriptionRequest, cacheRequestID message.CacheRequestID, subcodeAsNum int) {
		err := directReceiver.RequestCachedAsyncWithCallback(cacheRequestConfig, cacheRequestID, func(_ solace.CacheResponse) {})
		validateNativeError(err, subcode.Code(subcodeAsNum))
		signal <- true
	}
	var assertNonBlocking = func() {
		select {
		case <-signal:
		case <-time.After(1 * time.Second):
			t.Errorf("Expected cache request to fail immediately and not block for at least 1 second.")
		}
	}
	go requestCachedAsync(receiver, cacheRequestConfig, cacheRequestID, wouldBlockSubcodeNum)
	assertNonBlocking()

	go requestCachedAsyncWithCallback(receiver, cacheRequestConfig, cacheRequestID, wouldBlockSubcodeNum)
	assertNonBlocking()

	go requestCachedAsync(receiver2, cacheRequestConfig, cacheRequestID, notReadySubcodeNum)
	assertNonBlocking()

	go requestCachedAsyncWithCallback(receiver2, cacheRequestConfig, cacheRequestID, notReadySubcodeNum)
	assertNonBlocking()
}

type mockCacheResponseProcessor struct {
	handlerFunc func(solace.CacheResponse)
}

func (mock *mockCacheResponseProcessor) ProcessCacheResponse(cacheResponse solace.CacheResponse) {
	if mock.handlerFunc != nil {
		mock.handlerFunc(cacheResponse)
	}
}

func newMockCacheRequest(cacheRequestConfig resource.CachedMessageSubscriptionRequest, cacheRequestID message.CacheRequestID, cacheResponseHandler core.CacheResponseProcessor, index core.CacheRequestMapIndex, usesLocalDispatch bool, messageFilter *core.ReceivedMessageFilter) core.CacheRequest {
	return mockCacheRequest{
		requestConfigFunc: func(_ *mockCacheRequest) resource.CachedMessageSubscriptionRequest {
			return cacheRequestConfig
		},
		cacheRequestIDFunc: func(_ *mockCacheRequest) message.CacheRequestID {
			return cacheRequestID
		},
		cacheResponseHandlerFunc: func(_ *mockCacheRequest) core.CacheResponseProcessor {
			return cacheResponseHandler
		},
		cacheSessionFunc: func(_ *mockCacheRequest) core.CoreCacheSession {
			return core.GetCacheSessionFromCacheRequestIndex(core.CacheRequestMapIndex(index))
		},
		indexFunc: func(_ *mockCacheRequest) core.CacheRequestMapIndex {
			return core.CacheRequestMapIndex(index)
		},
		usesLocalDispatchFunc: func(_ *mockCacheRequest) bool {
			return usesLocalDispatch
		},
		messageFilterFunc: func(_ *mockCacheRequest) *core.ReceivedMessageFilter {
			return messageFilter
		},
	}
}

func defaultMockCacheRequest() core.CacheRequest {
	cacheRequestConfig := resource.NewCachedMessageSubscriptionRequest(resource.CacheRequestStrategyAsAvailable, "cache_name", resource.TopicSubscriptionOf("/some/test/topic"), int32(5000), int32(5000), int32(5000))
	cacheRequestID := defaultCacheRequestID
	index := defaultCacheRequestMapIndex
	cacheResponseHandler := &mockCacheResponseProcessor{}
	return newMockCacheRequest(cacheRequestConfig, cacheRequestID, cacheResponseHandler, index, false, nil)
}

var defaultCacheRequestMapIndex = core.CacheRequestMapIndex(0)
var defaultCoreCacheSession = core.GetCacheSessionFromCacheRequestIndex(defaultCacheRequestMapIndex)
var defaultCacheRequestID = message.CacheRequestID(0)

type mockCacheRequest struct {
	requestConfigFunc        func(*mockCacheRequest) resource.CachedMessageSubscriptionRequest
	cacheRequestIDFunc       func(*mockCacheRequest) message.CacheRequestID
	cacheResponseHandlerFunc func(*mockCacheRequest) core.CacheResponseProcessor
	cacheSessionFunc         func(*mockCacheRequest) core.CoreCacheSession
	indexFunc                func(*mockCacheRequest) core.CacheRequestMapIndex
	usesLocalDispatchFunc    func(*mockCacheRequest) bool
	messageFilterFunc        func(*mockCacheRequest) *core.ReceivedMessageFilter
}

func (mock mockCacheRequest) RequestConfig() resource.CachedMessageSubscriptionRequest {
	if mock.requestConfigFunc != nil {
		return mock.requestConfigFunc(&mock)
	}
	return nil
}

func (mock mockCacheRequest) ID() message.CacheRequestID {
	if mock.cacheRequestIDFunc != nil {
		return mock.cacheRequestIDFunc(&mock)
	}
	return defaultCacheRequestID
}

func (mock mockCacheRequest) Processor() core.CacheResponseProcessor {
	if mock.cacheResponseHandlerFunc != nil {
		return mock.cacheResponseHandlerFunc(&mock)
	}
	return &mockCacheResponseProcessor{}
}

func (mock mockCacheRequest) CacheSession() core.CoreCacheSession {
	if mock.cacheSessionFunc != nil {
		return mock.cacheSessionFunc(&mock)
	}
	return defaultCoreCacheSession
}

func (mock mockCacheRequest) Index() core.CacheRequestMapIndex {
	if mock.indexFunc != nil {
		return mock.indexFunc(&mock)
	}
	return defaultCacheRequestMapIndex
}

func (mock mockCacheRequest) MessageFilter() *core.ReceivedMessageFilter {
	if mock.messageFilterFunc != nil {
		return mock.messageFilterFunc(&mock)
	}
	return nil
}

func (mock mockCacheRequest) UsesLocalDispatch() bool {
	if mock.usesLocalDispatchFunc != nil {
		return mock.usesLocalDispatchFunc(&mock)
	}
	return false
}

type result struct {
	proceed bool
	err     error
}

type mockInternalReceiver struct {
	events                               func() core.Events
	replier                              func() core.Replier
	isRunning                            func() bool
	registerRxCallback                   func(callback core.RxCallback) uintptr
	unregisterRxCallback                 func(ptr uintptr)
	subscribe                            func(topic string, ptr uintptr) (core.SubscriptionCorrelationID, <-chan core.SubscriptionEvent, core.ErrorInfo)
	unsubscribe                          func(topic string, ptr uintptr) (core.SubscriptionCorrelationID, <-chan core.SubscriptionEvent, core.ErrorInfo)
	incrementMetric                      func(metric core.NextGenMetric, amount uint64)
	incrementDuplicateAckCount           func()
	newPersistentReceiver                func(props []string, callback core.RxCallback, eventCallback core.PersistentEventCallback) (core.PersistentReceiver, *ccsmp.SolClientErrorInfoWrapper)
	processCacheResponseFunc             func(*mockInternalReceiver, *sync.Map, core.CoreCacheEventInfo)
	sendCacheRequestFunc                 func(*mockInternalReceiver, core.CacheRequest, core.CoreCacheEventCallback, uintptr) error
	cancelPendingCacheRequestsFunc       func(*mockInternalReceiver, core.CacheRequestMapIndex, core.CacheRequest) *core.CoreCacheEventInfo
	createCacheRequestFunc               func(*mockInternalReceiver, resource.CachedMessageSubscriptionRequest, message.CacheRequestID, core.CacheResponseProcessor, uintptr) (core.CacheRequest, error)
	destroyCacheRequestFunc              func(*mockInternalReceiver, core.CacheRequest) error
	cleanupCacheRequestSubscriptionsFunc func(*mockInternalReceiver, core.CacheRequest) error
}

func (mock *mockInternalReceiver) CacheRequestor() core.CacheRequestor {
	return mock
}

func (mock *mockInternalReceiver) CreateCacheRequest(cachedMessageSubscriptionRequest resource.CachedMessageSubscriptionRequest, cacheRequestID message.CacheRequestID, cacheResponseHandler core.CacheResponseProcessor, dispatchID uintptr) (core.CacheRequest, error) {
	if mock.createCacheRequestFunc != nil {
		return mock.createCacheRequestFunc(mock, cachedMessageSubscriptionRequest, cacheRequestID, cacheResponseHandler, dispatchID)
	}
	/* If not set, presume no-op is intended. */
	return nil, nil
}

func (mock *mockInternalReceiver) DestroyCacheRequest(cacheRequest core.CacheRequest) error {
	if mock.destroyCacheRequestFunc != nil {
		return mock.destroyCacheRequestFunc(mock, cacheRequest)
	}
	/* If not set, presume no-op is intended. */
	return nil
}

func (mock *mockInternalReceiver) SendCacheRequest(cacheRequest core.CacheRequest, cacheEventCallback core.CoreCacheEventCallback, dispatchID uintptr) error {
	if mock.sendCacheRequestFunc != nil {
		return mock.sendCacheRequestFunc(mock, cacheRequest, cacheEventCallback, dispatchID)
	}
	/* If not set, presume no-op is intended. */
	return nil
}

func (mock *mockInternalReceiver) ProcessCacheEvent(cacheRequestMap *sync.Map, eventInfo core.CoreCacheEventInfo) {
	if mock.processCacheResponseFunc != nil {
		mock.processCacheResponseFunc(mock, cacheRequestMap, eventInfo)
	}
	/* If not set, presume no-op is intended. */
}

func (mock *mockInternalReceiver) CancelPendingCacheRequests(cacheRequestIndex core.CacheRequestMapIndex, cacheRequest core.CacheRequest) *core.CoreCacheEventInfo {
	if mock.cancelPendingCacheRequestsFunc != nil {
		return mock.cancelPendingCacheRequestsFunc(mock, cacheRequestIndex, cacheRequest)
	}
	/* If not set, presume no-op is intended. */
	return nil
}

func (mock *mockInternalReceiver) CleanupCacheRequestSubscriptions(cacheRequest core.CacheRequest) error {
	if mock.cleanupCacheRequestSubscriptionsFunc != nil {
		return mock.cleanupCacheRequestSubscriptionsFunc(mock, cacheRequest)
	}
	return nil
}

func (mock *mockInternalReceiver) Events() core.Events {
	if mock.events != nil {
		return mock.events()
	}
	return &mockEvents{}
}

func (mock *mockInternalReceiver) Replier() core.Replier {
	if mock.replier != nil {
		return mock.replier()
	}
	return &mockReplier{}
}

func (mock *mockInternalReceiver) IsRunning() bool {
	if mock.isRunning != nil {
		return mock.isRunning()
	}
	return true
}

// Register an RX callback, returns a correlation pointer used when adding and removing subscriptions
func (mock *mockInternalReceiver) RegisterRXCallback(msgCallback core.RxCallback) uintptr {
	if mock.registerRxCallback != nil {
		return mock.registerRxCallback(msgCallback)
	}
	return 0
}

// Remove the callback allowing GC to cleanup the function registered
func (mock *mockInternalReceiver) UnregisterRXCallback(ptr uintptr) {
	if mock.unregisterRxCallback != nil {
		mock.unregisterRxCallback(ptr)
	}
}

// Add a subscription to the given correlation pointer
func (mock *mockInternalReceiver) Subscribe(topic string, ptr uintptr) (core.SubscriptionCorrelationID, <-chan core.SubscriptionEvent, core.ErrorInfo) {
	if mock.subscribe != nil {
		return mock.subscribe(topic, ptr)
	}
	return 0, nil, nil
}

// Remove a subscription from the given correlation pointer
func (mock *mockInternalReceiver) Unsubscribe(topic string, ptr uintptr) (core.SubscriptionCorrelationID, <-chan core.SubscriptionEvent, core.ErrorInfo) {
	if mock.unsubscribe != nil {
		return mock.unsubscribe(topic, ptr)
	}
	return 0, nil, nil
}

func (mock *mockInternalReceiver) ClearSubscriptionCorrelation(id core.SubscriptionCorrelationID) {

}

func (mock *mockInternalReceiver) ProvisionEndpoint(name string, exclusive bool) *ccsmp.SolClientErrorInfoWrapper {
	return nil
}

func (mock *mockInternalReceiver) EndpointUnsubscribe(queueName string, topic string) (core.SubscriptionCorrelationID, <-chan core.SubscriptionEvent, core.ErrorInfo) {
	return 0, nil, nil
}

func (mock *mockInternalReceiver) IncrementMetric(metric core.NextGenMetric, amount uint64) {
	if mock.incrementMetric != nil {
		mock.incrementMetric(metric, amount)
	}
}

func (mock *mockInternalReceiver) IncrementDuplicateAckCount() {
	if mock.incrementDuplicateAckCount != nil {
		mock.incrementDuplicateAckCount()
	}
}

func (mock *mockInternalReceiver) NewPersistentReceiver(props []string, callback core.RxCallback, eventCallback core.PersistentEventCallback) (core.PersistentReceiver, *ccsmp.SolClientErrorInfoWrapper) {
	if mock.newPersistentReceiver != nil {
		return mock.newPersistentReceiver(props, callback, eventCallback)
	}
	return &mockPersistentReceiver{}, nil
}

type mockPersistentReceiver struct {
	destroy     func(freeMemory bool) *ccsmp.SolClientErrorInfoWrapper
	start       func() *ccsmp.SolClientErrorInfoWrapper
	stop        func() *ccsmp.SolClientErrorInfoWrapper
	subscribe   func(topic string) (core.SubscriptionCorrelationID, <-chan core.SubscriptionEvent, core.ErrorInfo)
	unsubscribe func(topic string) (core.SubscriptionCorrelationID, <-chan core.SubscriptionEvent, core.ErrorInfo)
	ack         func(msg core.MessageID) *ccsmp.SolClientErrorInfoWrapper
	settle      func(msg core.MessageID, outcome core.MessageSettlementOutcome) *ccsmp.SolClientErrorInfoWrapper
}

// Destroy destroys the flow
func (mock *mockPersistentReceiver) Destroy(freeMemory bool) *ccsmp.SolClientErrorInfoWrapper {
	if mock.destroy != nil {
		return mock.destroy(freeMemory)
	}
	return nil
}

// Start will start the receiption of messages
// Persistent Receivers are started by default after creation
func (mock *mockPersistentReceiver) Start() *ccsmp.SolClientErrorInfoWrapper {
	if mock.start != nil {
		return mock.start()
	}
	return nil
}

// Stop will stop the reception of messages
func (mock *mockPersistentReceiver) Stop() *ccsmp.SolClientErrorInfoWrapper {
	if mock.stop != nil {
		return mock.stop()
	}
	return nil
}

// Subscribe will add a subscription to the persistent receiver
func (mock *mockPersistentReceiver) Subscribe(topic string) (core.SubscriptionCorrelationID, <-chan core.SubscriptionEvent, core.ErrorInfo) {
	if mock.subscribe != nil {
		return mock.subscribe(topic)
	}
	return 0, nil, nil
}

// Unsubscribe will remove the subscription from the persistent receiver
func (mock *mockPersistentReceiver) Unsubscribe(topic string) (core.SubscriptionCorrelationID, <-chan core.SubscriptionEvent, core.ErrorInfo) {
	if mock.unsubscribe != nil {
		return mock.unsubscribe(topic)
	}
	return 0, nil, nil
}

// Ack will acknowledge the given message
func (mock *mockPersistentReceiver) Ack(msgID core.MessageID) *ccsmp.SolClientErrorInfoWrapper {
	if mock.ack != nil {
		return mock.ack(msgID)
	}
	return nil
}

// Settle will settle the given message with the provided outcome
func (mock *mockPersistentReceiver) Settle(msgID core.MessageID, outcome core.MessageSettlementOutcome) *ccsmp.SolClientErrorInfoWrapper {
	if mock.settle != nil {
		return mock.settle(msgID, outcome)
	}
	return nil
}

func (mock *mockPersistentReceiver) Destination() (destination string, durable bool, errorInfo *ccsmp.SolClientErrorInfoWrapper) {
	return
}

type mockEvents struct {
}

func (events *mockEvents) AddEventHandler(sessionEvent core.Event, responseCode core.EventHandler) uint {
	return 0
}

func (events *mockEvents) RemoveEventHandler(id uint) {
}

type mockReplier struct {
	sendReply func(replyMsg core.ReplyPublishable) core.ErrorInfo
}

func (replier *mockReplier) SendReply(replyMsg core.ReplyPublishable) core.ErrorInfo {
	if replier.sendReply != nil {
		return replier.sendReply(replyMsg)
	}
	return nil
}

type notASubscription struct{}

func (sub *notASubscription) GetName() string {
	return "some topic"
}

func (sub *notASubscription) GetSubscriptionType() string {
	return "notASubscription"
}

type mockEvent struct {
	err error
}

func (event mockEvent) GetError() error {
	return event.err
}

func (event mockEvent) GetInfoString() string {
	return ""
}

func (event mockEvent) GetCorrelationPointer() unsafe.Pointer {
	return nil
}

func (event mockEvent) GetUserPointer() unsafe.Pointer {
	return nil
}

type mockSubscriptionEvent struct {
	err error
}

func (event mockSubscriptionEvent) GetID() core.SubscriptionCorrelationID {
	return 0
}

func (event mockSubscriptionEvent) GetError() error {
	return event.err
}
