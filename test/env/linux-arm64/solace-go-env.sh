#!/bin/bash
if [ -z "$PATH_SOL_COMMON" ]; then
    export PATH_SOL_COMMON=$PATH
fi

# envs for PSPGC
# export PATH=$PATH_SOL_COMMON:$WORKSPACE/
export PSPGC_TMP_DIR=$WORKSPACE/tmp
export PSPGC_HOME=$WORKSPACE

# envs for running local tests
export KRB_TEST_IMAGE=docker.solacedev.net/ats-krbserver:1.2.0
export OAUTH_TEST_IMAGE=docker.solacedev.net/ats-oauthserver:1.2.0
export PUBSUB_PLAINTEXT_PORT=33555

# Toxiproxy envs
export TOXIPROXY_HOST=localhost
export TOXIPROXY_HOSTNAME=toxiproxy
export TOXIPROXY_UPSTREAM=solbroker
export TOXIPROXY_PORT=8474
export TOXIPROXY_PLAINTEXT_PORT=15555
export TOXIPROXY_COMPRESSED_PORT=15003
export TOXIPROXY_SECURE_PORT=15443

# Pubsub envs
export PUBSUB_REPO_BASE=solace/solace-pubsub
export PUBSUB_EDITION=standard
export PUBSUB_TAG=10.4
export PUBSUB_HOSTNAME=solbroker
export PUBSUB_NETWORK_NAME=solace_msg_net

# Kerberos envs
export PUBSUB_KDC_HOSTNAME=ats-kdc-server
export PUBSUB_DOMAIN=ATS.SOLACE.COM

# Oauthserver envs
export PUBSUB_OAUTHSERVER_HOSTNAME=solaceOAuth
export PUBSUB_OAUTHSERVER_JWKS_ENDPOINT=https://solaceOAuth:30000/
export PUBSUB_OAUTHSERVER_USERINFO_ENDPOINT=https://solaceOAuth:30001/

# Alias
alias gosolctrl="gosoltestctrl"


