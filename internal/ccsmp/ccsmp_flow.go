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

package ccsmp

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"solace.dev/go/messaging/internal/impl/logging"
)

/*
#cgo CFLAGS: -DSOLCLIENT_PSPLUS_GO
#include <stdlib.h>
#include <stdio.h>

#include <string.h>

#include "solclient/solClient.h"
#include "solclient/solClientMsg.h"
#include "./ccsmp_helper.h"
*/
import "C"

// SolClientFlowPt is assigned a value
type SolClientFlowPt = C.solClient_opaqueFlow_pt

// SolClientFlowEventInfoPt is assigned a value
type SolClientFlowEventInfoPt = C.solClient_flow_eventCallbackInfo_pt

// SolClientFlowRxMsgDispatchFuncInfo is assigned a value
type SolClientFlowRxMsgDispatchFuncInfo = C.solClient_flow_rxMsgDispatchFuncInfo_t

// SolClientMessageSettlementOutcome is assigned a value
type SolClientMessageSettlementOutcome = C.solClient_msgOutcome_t

// SolClientSettlementOutcomeAccepted - the message was successfully processed.
const SolClientSettlementOutcomeAccepted = C.SOLCLIENT_OUTCOME_ACCEPTED

// SolClientSettlementOutcomeFailed - message processing failed temporarily, attempt redelivery if configured.
const SolClientSettlementOutcomeFailed = C.SOLCLIENT_OUTCOME_FAILED

// SolClientSettlementOutcomeRejected - message was processed and rejected, removed from the queue.
const SolClientSettlementOutcomeRejected = C.SOLCLIENT_OUTCOME_REJECTED

// Callbacks

// SolClientFlowMessageCallback is assigned a function
type SolClientFlowMessageCallback = func(msgP SolClientMessagePt) bool

// SolClientFlowEventCallback is assigned a function
type SolClientFlowEventCallback = func(flowEvent SolClientFlowEvent, responseCode SolClientResponseCode, info string)

var flowID uintptr
var flowToEventCallbackMap sync.Map
var flowToRXCallbackMap sync.Map

//export goFlowMessageReceiveCallback
func goFlowMessageReceiveCallback(flowP SolClientFlowPt, msgP SolClientMessagePt, userP unsafe.Pointer) C.solClient_rxMsgCallback_returnCode_t {
	if callback, ok := flowToRXCallbackMap.Load(uintptr(userP)); ok {
		if callback.(SolClientFlowMessageCallback)(msgP) {
			return C.SOLCLIENT_CALLBACK_TAKE_MSG
		}
		return C.SOLCLIENT_CALLBACK_OK
	}
	logging.Default.Error("Received message from core API without an associated session callback")
	return C.SOLCLIENT_CALLBACK_OK
}

//export goFlowEventCallback
func goFlowEventCallback(flowP SolClientFlowPt, eventInfoP SolClientFlowEventInfoPt, userP unsafe.Pointer) {
	if callback, ok := flowToEventCallbackMap.Load(uintptr(userP)); ok {
		callback.(SolClientFlowEventCallback)(SolClientFlowEvent(eventInfoP.flowEvent), eventInfoP.responseCode, C.GoString(eventInfoP.info_p))
	} else {
		logging.Default.Debug("Received event callback from core API without an associated session callback")
	}
}

// Type definitions

// SolClientFlow structure
type SolClientFlow struct {
	pointer SolClientFlowPt
	userP   uintptr
}

// Functionality

// SolClientSessionCreateFlow function
func (session *SolClientSession) SolClientSessionCreateFlow(properties []string, msgCallback SolClientFlowMessageCallback, eventCallback SolClientFlowEventCallback) (*SolClientFlow, *SolClientErrorInfoWrapper) {
	flowPropsP, sessionPropertiesFreeFunction := ToCArray(properties, true)
	defer sessionPropertiesFreeFunction()

	flowID := atomic.AddUintptr(&flowID, 1)

	flowToRXCallbackMap.Store(flowID, msgCallback)
	flowToEventCallbackMap.Store(flowID, eventCallback)

	flow := &SolClientFlow{}
	flow.userP = flowID
	err := handleCcsmpError(func() SolClientReturnCode {
		// this will register the goFlowMessageReceiveCallback and goFlowEventCallback callbacks with the flowID
		return C.SessionFlowCreate(session.pointer,
			flowPropsP,
			&flow.pointer,
			C.solClient_uint64_t(flowID))
	})
	if err != nil {
		return nil, err
	}
	return flow, nil
}

// SolClientFlowRemoveCallbacks function
func (flow *SolClientFlow) SolClientFlowRemoveCallbacks() {
	flowToRXCallbackMap.Delete(flow.userP)
	flowToEventCallbackMap.Delete(flow.userP)
}

// SolClientFlowDestroy function
func (flow *SolClientFlow) SolClientFlowDestroy() *SolClientErrorInfoWrapper {
	errInfo := handleCcsmpError(func() SolClientReturnCode {
		return C.solClient_flow_destroy(&flow.pointer)
	})
	return errInfo
}

// SolClientFlowStart function
func (flow *SolClientFlow) SolClientFlowStart() *SolClientErrorInfoWrapper {
	return handleCcsmpError(func() SolClientReturnCode {
		return C.solClient_flow_start(flow.pointer)
	})
}

// SolClientFlowStop function
func (flow *SolClientFlow) SolClientFlowStop() *SolClientErrorInfoWrapper {
	return handleCcsmpError(func() SolClientReturnCode {
		return C.solClient_flow_stop(flow.pointer)
	})
}

// SolClientFlowSubscribe function
func (flow *SolClientFlow) SolClientFlowSubscribe(topic string, correlationID uintptr) *SolClientErrorInfoWrapper {
	return handleCcsmpError(func() SolClientReturnCode {
		cString := C.CString(topic)
		defer C.free(unsafe.Pointer(cString))
		return C.FlowTopicSubscribeWithDispatch(flow.pointer,
			C.SOLCLIENT_SUBSCRIBE_FLAGS_REQUEST_CONFIRM,
			cString,
			nil,
			C.solClient_uint64_t(correlationID))
	})
}

// SolClientFlowUnsubscribe function
func (flow *SolClientFlow) SolClientFlowUnsubscribe(topic string, correlationID uintptr) *SolClientErrorInfoWrapper {
	return handleCcsmpError(func() SolClientReturnCode {
		cString := C.CString(topic)
		defer C.free(unsafe.Pointer(cString))
		return C.FlowTopicUnsubscribeWithDispatch(flow.pointer,
			C.SOLCLIENT_SUBSCRIBE_FLAGS_REQUEST_CONFIRM,
			cString,
			nil,
			C.solClient_uint64_t(correlationID))
	})
}

// SolClientFlowAck function
func (flow *SolClientFlow) SolClientFlowAck(msgID SolClientMessageID) *SolClientErrorInfoWrapper {
	return handleCcsmpError(func() SolClientReturnCode {
		return C.solClient_flow_sendAck(flow.pointer, msgID)
	})
}

// SolClientFlowSettleMessage function
func (flow *SolClientFlow) SolClientFlowSettleMessage(msgID SolClientMessageID, settlementOutcome SolClientMessageSettlementOutcome) *SolClientErrorInfoWrapper {
	return handleCcsmpError(func() SolClientReturnCode {
		return C.solClient_flow_settleMsg(flow.pointer, msgID, settlementOutcome)
	})
}

// SolClientFlowGetDestination function
func (flow *SolClientFlow) SolClientFlowGetDestination() (name string, durable bool, err *SolClientErrorInfoWrapper) {
	var dest *SolClientDestination = &SolClientDestination{}
	errorInfo := handleCcsmpError(func() SolClientReturnCode {
		return C.solClient_flow_getDestination(flow.pointer, dest, (C.size_t)(unsafe.Sizeof(*dest)))
	})
	if errorInfo != nil {
		return "", false, errorInfo
	}
	name = C.GoString(dest.dest)
	durable = dest.destType != C.SOLCLIENT_QUEUE_TEMP_DESTINATION
	return name, durable, errorInfo
}
