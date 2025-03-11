// pubsubplus-go-client
//
// Copyright 2021-2025 Solace Corporation. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/**
 * ccsmp_callbacks.c is used to adapt C code to go code
 * Go functions exported with `//export <funcname>` can be called from c code
 * These functions can then be referenced in golang with C.messageReceiveCallback
 */

#include "solclient/solClient.h"
#include "solclient/solClientMsg.h"
#include "solclient/solCache.h"
#include "./ccsmp_helper.h"
#include <stdlib.h>

solClient_rxMsgCallback_returnCode_t
messageReceiveCallback(solClient_opaqueSession_pt opaqueSession_p, solClient_opaqueMsg_pt msg_p, void *user_p)
{
        solClient_rxMsgCallback_returnCode_t goMessageReceiveCallback(solClient_opaqueSession_pt, solClient_opaqueMsg_pt, void *);
        return goMessageReceiveCallback(opaqueSession_p, msg_p, user_p);
}

solClient_rxMsgCallback_returnCode_t
cacheFilterCallback(solClient_opaqueSession_pt opaqueSession_p, solClient_opaqueMsg_pt msg_p, void * user_p)
{
        solClientgo_msgDispatchCacheRequestIdFilterInfo_t * info_p = (solClientgo_msgDispatchCacheRequestIdFilterInfo_t *) user_p;
        if ( solClientgo_filterCachedMessageByCacheRequestId(opaqueSession_p, msg_p, user_p) != SOLCLIENT_OK) {
                /* NOTE: We were unable to find the expected cache request ID in the message, so we discard it. */
                return SOLCLIENT_CALLBACK_OK;
        }
        return info_p->callback_p(opaqueSession_p, msg_p, (void *)info_p->dispatchID);
}

solClient_rxMsgCallback_returnCode_t
requestResponseReplyMessageReceiveCallback(solClient_opaqueSession_pt opaqueSession_p, solClient_opaqueMsg_pt msg_p, void *user_p) {
    solClient_rxMsgCallback_returnCode_t goReplyMessageReceiveCallback(solClient_opaqueSession_pt, solClient_opaqueMsg_pt, void *, char *);
    char * correlationId = NULL;
    // when receiving message that is not a reply deliver to subscription dispatch
    if ( SOLCLIENT_OK != solClientgo_msg_isRequestReponseMsg(msg_p, &correlationId) ) {
        // discard any message that is not a reply message
        // note any subscription that matches the replyto topic will get an independent dispatch callback
        return SOLCLIENT_CALLBACK_OK;
    }
    return goReplyMessageReceiveCallback(opaqueSession_p, msg_p, user_p, correlationId);
}

solClient_rxMsgCallback_returnCode_t
defaultMessageReceiveCallback(solClient_opaqueSession_pt opaqueSession_p, solClient_opaqueMsg_pt msg_p, void *user_p)
{
    solClient_rxMsgCallback_returnCode_t goDefaultMessageReceiveCallback(solClient_opaqueSession_pt, solClient_opaqueMsg_pt, void *);
    return goDefaultMessageReceiveCallback(opaqueSession_p, msg_p, user_p);
}

void eventCallback(solClient_opaqueSession_pt opaqueSession_p, solClient_session_eventCallbackInfo_pt eventInfo_p, void *user_p)
{
    void goEventCallback(solClient_opaqueSession_pt, solClient_session_eventCallbackInfo_pt, void *);
    goEventCallback(opaqueSession_p, eventInfo_p, user_p);
}

void handleLogCallback(solClient_log_callbackInfo_pt logInfo_p, void *user_p)
{
    void goLogCallback(solClient_log_callbackInfo_pt, void *);
    goLogCallback(logInfo_p, user_p);
}

solClient_rxMsgCallback_returnCode_t
flowMessageReceiveCallback(solClient_opaqueFlow_pt opaqueFlow_p, solClient_opaqueMsg_pt msg_p, void *user_p)
{
    solClient_rxMsgCallback_returnCode_t goFlowMessageReceiveCallback(solClient_opaqueFlow_pt, solClient_opaqueMsg_pt, void *);
    return goFlowMessageReceiveCallback(opaqueFlow_p, msg_p, user_p);
}

void flowEventCallback(solClient_opaqueFlow_pt opaqueFlow_p, solClient_flow_eventCallbackInfo_pt eventInfo_p, void *user_p)
{
    void goFlowEventCallback(solClient_opaqueFlow_pt, solClient_flow_eventCallbackInfo_pt, void *);
    goFlowEventCallback(opaqueFlow_p, eventInfo_p, user_p);
}

void cacheEventCallback(solClient_opaqueSession_pt opaqueSession_p, solCache_eventCallbackInfo_pt eventInfo_p, void *user_p)
{
    void goCacheEventCallback(solClient_opaqueSession_pt, solCache_eventCallbackInfo_pt, void *);
    goCacheEventCallback(opaqueSession_p, eventInfo_p, user_p);
}
