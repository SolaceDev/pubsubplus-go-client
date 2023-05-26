#!/bin/bash

# must have PSPGC_HOME defined

function scripts_dir()
{
    echo "$HOME/.scripts"
}

function printErr(){
    echo "$@" >&2
}

function printTrace(){
    if [ "$TRACE"=="1" ]; then
        printErr $@
    fi
}

function test_jq_found() {
    jq -e $1 $2 &> /dev/null && echo "found"
}

function get_test_env_var() {
    local CONFIG=$1
    local CONFIG_NAME=$2
    local ENV_VAR_NAME=$3
    local DEFAULT_VAL=$4
    if [[ "$(test_jq_found $CONFIG_NAME $CONFIG)" ]]; then
        local JQ_OUT="$(jq $CONFIG_NAME $CONFIG)"
    fi
    echo "${ENV_VAR_NAME}=${JQ_OUT:-"$DEFAULT_VAL"} "
}

function get_test_env(){
    local TEST_DIR="$PSPGC_HOME/test"
    local TEST_GOCONFIG_JSON="$TEST_DIR/data/config/config_testcontainers.json"
    # the docker compose files require the following env var to be set

    # for the main docker-compose.yml
    # PUBSUB_REPO_BASE
    # PUBSUB_EDITION
    # PUBSUB_TAG
    # where the above are used in image: $PUBSUB_REPO_BASE-$PUBSUB_EDITION:$PUBSUB_TAG
    # PUBSUB_HOSTNAME
    # PUBSUB_SEMP_PORT
    # PUBSUB_SECURE_SEMP_PORT
    # PUBSUB_PLAINTEXT_PORT
    # PUBSUB_COMPRESSED_PORT
    # PUBSUB_SECURE_PORT
    # PUBSUB_HEALTHCHECK_PORT
    # PUBSUB_WEB_PORT
    # PUBSUB_SECURE_WEB_PORT
    # TOXIPROXY_HOSTNAME
    # TOXIPROXY_PORT
    # TOXIPROXY_PLAINTEXT_PORT
    # TOXIPROXY_COMPRESSED_PORT
    # TOXIPROXY_SECURE_PORT

    # for the krb.yml
    # PUBSUB_KDC_HOSTNAME
    # PUBSUB_DOMAIN
    # additionally for testing there are:
    # KUSER
    # KPASSWORD
    # note there is also KBR_TEST_IMAGE but that is selected under setkrbenv

    # for the oauth.yml
    # PUBSUB_OAUTHSERVER_HOSTNAME
    # there is also $OAUTH_TEST_IMAGE but that is selected under setoauthenv

    # set all the defaults here
    local TEST_ENV=""
    #local JQ_OUT=
    #if [[ "$(test_jq_found ".testcontainers.broker_repo" $TEST_GOCONFIG_JSON )" ]]; then
    #    local JQ_OUT=$(jq ".testcontainers.broker_repo" $TEST_GOCONFIG_JSON)
    #fi
    #local TEST_ENV="PUBSUB_REPO_BASE=${JQ_OUT:-"solace/solace-pubsub"} $TEST_ENV"
    # broker image setup
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".testcontainers.broker_repo" "PUBSUB_REPO_BASE" "solace/solace-pubsub" ) $TEST_ENV"
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".testcontainers.broker_edition" "PUBSUB_EDITION" "standard" ) $TEST_ENV"
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".testcontainers.broker_tag" "PUBSUB_TAG" "10.4" ) $TEST_ENV"
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".testcontainers.broker_hostname" "PUBSUB_HOSTNAME" "solbroker" ) $TEST_ENV"
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".testcontainers.network_name" "PUBSUB_NETWORK_NAME" "solace_msg_net" ) $TEST_ENV"

    # broker ports
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".semp.port" "PUBSUB_SEMP_PORT" "8080" ) $TEST_ENV"
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".semp.secure" "PUBSUB_SECURE_SEMP_PORT" "1943" ) $TEST_ENV"
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".messaging.ports.plaintext" "PUBSUB_PLAINTEXT_PORT" "33555" ) $TEST_ENV"
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".messaging.ports.compressed" "PUBSUB_COMPRESSED_PORT" "55003" ) $TEST_ENV"
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".messaging.ports.secure" "PUBSUB_SECURE_PORT" "55443" ) $TEST_ENV"
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".messaging.ports.healthcheck" "PUBSUB_HEALTHCHECK_PORT" "5550" ) $TEST_ENV"
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".messaging.ports.web" "PUBSUB_WEB_PORT" "8008" ) $TEST_ENV"
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".messaging.ports.secure_web" "PUBSUB_SECURE_WEB_PORT" "1443" ) $TEST_ENV"

    # toxiproxy setup
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".testcontainers.toxiproxy_hostname" "TOXIPROXY_HOSTNAME" "toxiproxy" ) $TEST_ENV"
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".toxiproxy.port" "TOXIPROXY_PORT" "8474" ) $TEST_ENV"
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".toxiproxy.plaintext_port" "TOXIPROXY_PLAINTEXT_PORT" "15555" ) $TEST_ENV"
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".toxiproxy.compressed_port" "TOXIPROXY_COMPRESSED_PORT" "15003" ) $TEST_ENV"
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".toxiproxy.secure_port" "TOXIPROXY_SECURE_PORT" "15443" ) $TEST_ENV"

    # krb setup
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".kerberos.hostname" "PUBSUB_KDC_HOSTNAME" "ats-kdc-server" ) $TEST_ENV"
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".kerberos.domain" "PUBSUB_DOMAIN" "ATS.SOLACE.COM" ) $TEST_ENV"

    # oauth setup
    local TEST_ENV="$(get_test_env_var $TEST_GOCONFIG_JSON ".oauth.hostname" "PUBSUB_OAUTHSERVER_HOSTNAME" "solaceOAuth" ) $TEST_ENV"

    printTrace "Got test env: $TEST_ENV"
    for envar in $TEST_ENV ;do
        echo $envar
    done
}

function get_additional_compose_file_list(){
    local TEST_DIR=$1
    local FILES=$(find $TEST_DIR/data/compose -iname "docker-compose*.yml")
    local COMPOSE_ARG_LIST=""
    for f in $FILES; do
        local COMPOSE_ARG_LIST="${COMPOSE_ARG_LIST} -f $f"
    done
    echo $COMPOSE_ARG_LIST
}

function set_krb_env(){
    local TEST_DIR=${1:-$PSPGC_HOME/test}
    if [ "$(find $TEST_DIR/data/compose -iname "docker-compose.krb.yml")" ]; then
        export KRB_TEST_IMAGE=${KRB_TEST_IMAGE:-"apps-jenkins:18888/ats-krbserver:latest"}
        export KRB5_CONFIG=${KRB5_CONFIG:-$TEST_DIR/krb/krb5.conf}
        #export KUSER
        #export KPASSWORD
    fi
}

function set_cache_env(){
    local TEST_DIR=${1:-$PSPGC_HOME/test}
    if [ "$SKIP_CACHE" == "" ]; then
        if [ "$(find $TEST_DIR/data/compose -iname "docker-compose.cache.yml")" ]; then
            export SOLCACHE_TEST_IMAGE=${SOLCACHE_TEST_IMAGE:-"apps-jenkins:18888/ats-solcache:latest"}
        fi
        if [ "$(find $TEST_DIR/data/compose -iname "docker-compose.cacheproxy.yml")" ]; then
            export SOLCACHEPROXY_TEST_IMAGE=${SOLCACHEPROXY_TEST_IMAGE:-"apps-jenkins:18888/ats-solcacheproxy:latest"}
        fi
    fi
}

function set_oauth_env(){
    local TEST_DIR=${1:-$PSPGC_HOME/test}
    if [ "$(find $TEST_DIR/data/compose -iname "docker-compose.oauth.yml")" ]; then
        export OAUTH_TEST_IMAGE=${OAUTH_TEST_IMAGE:-"apps-jenkins:18888/ats-oauthserver:1.1.0"}
    fi
}

function krb_command() {
    if [[ -d $PSPGC_HOME ]]; then
        local TEST_DIR=$PSPGC_HOME/test
        set_krb_env $TEST_DIR
        if [ "$(find $TEST_DIR/data/compose -iname "docker-compose.krb.yml")" ]; then
            if [ "$1" ]; then
                # run keberbos command in env
                $@
            else
                printErr "Missing command to execute in $SUBCMD env"
            fi
        else
            printErr "$SUB_CMD not supported with pysolace at $PSPGC_HOME"
        fi
    fi
}

function docker_compose_cmd() {
    local CMD=docker-compose
    echo "Running command: $CMD $@"
    $CMD $@
}

function test_broker(){
    local CMD=$1
    if [[ -d $PSPGC_HOME ]]; then
        local TEST_DIR="$PSPGC_HOME/test"
        local TEST_COMPOSE_OPTS="$(get_additional_compose_file_list $TEST_DIR)"
        local DOCKER_CMD="docker"
        #local COMPOSE_CMD="docker-compose"
        local COMPOSE_CMD="docker_compose_cmd"
        #eval $(get_test_env | awk '{if ($1 ~ /^.*D\:/) { print substr($1, 3) } else { print }}' | awk '{print("export", $1)}')

        #get_test_env
        #echo "$(get_test_env | awk '{print("export", $1)}' )"
        eval $(get_test_env | awk '{print("export", $1)}' )
        set_krb_env $TEST_DIR
        set_cache_env $TEST_DIR
        set_oauth_env $TEST_DIR
        if [ "$CMD" == "start" ]; then
            # start gosolace test broker using docker
            $COMPOSE_CMD $TEST_COMPOSE_OPTS down
            if [ "$(docker network ls -f "name=$PUBSUB_NETWORK_NAME" -q)" == "" ]; then
                $DOCKER_CMD network create $PUBSUB_NETWORK_NAME
            fi
            $COMPOSE_CMD $TEST_COMPOSE_OPTS up -d $COMPOSE_SERVICES
        elif [ "$CMD" == "status" ]; then
            # check gosolace test broker status
            $COMPOSE_CMD $TEST_COMPOSE_OPTS ps $COMPOSE_SERVICES
        elif [ "$CMD" == "stop" ]; then
            # stop gosolace test broker
            $COMPOSE_CMD $TEST_COMPOSE_OPTS down && $DOCKER_CMD network prune -f
        elif [[ "$CMD" == "help" || "$CMD" == "-h" ]]; then
            echo "${SUB_CMD:-broker} <subcmd>"
            echo "<subcmd>:"
            echo "    start :   Starts gosolace test services using docker-compose.yml from PSPGC_HOME"
            echo "    stop  :   Stops gosolace test services using docker-compose.yml from PSPGC_HOME"
            echo "    status:   Display current status of test services docker-compose.yml from PSPGC_HOME"
        else
            printErr "Unrecognized CMD: '$CMD'"
        fi
    else
        printErr "Missing PSPGC_HOME directory or '$PSPGC_HOME' does not exist"

    fi
}

function test_proxy(){
    local COMPOSE_SERVICES=toxiproxy
    # filter by
    test_broker $@
}

function test_kdc(){
    local COMPOSE_SERVICES=krbserver
    #filter by
    test_broker $@
}

function runtests(){
    local TEST_DIR=/opt/jenkins/test
    local BRANCH=/opt/jenkins
    local TEST_OUTDIR=$( mktemp -d /opt/jenkins/test/reports/test-report.XXXXX)
    #local TMP_DIR=${PSPGC_TMP_DIR:-$HOME/tmp}
    #local BRANCH_DIR=${TMP_DIR}/$BRANCH
    #if [[ -d $BRANCH_DIR ]]; then
    #    local TEST_OUTDIR=$( mktemp -d $BRANCH_DIR/test-report.XXXXX )
    #else
    #    local TEST_OUTDIR=$( mktemp -d $TMP_DIR/test-report.XXXXX )
    #fi

    echo "$*"
    #parse args
    while [ "$1" ]; do

        local ARG=$1
        shift
        if [ "$ARG" == "-b" ]; then
            local BROKER_TYPE=$1
            shift
        elif [ "$ARG" == "-n" ]; then
            local ITERATIONS=$1
            shift
        #elif [ "$ARG" == "-t" ]; then
        #    local TEST_FILTER=$1
        #    shift
        elif [[ "$ARG" == "-h" || "$ARG" == "--help" || "$ARG" == "help" ]]; then
            echo "${SUB_CMD:-run} <options>"
            echo "  executes pysolace tests from PSPGC_HOME"
            echo "<options>:"
            echo "    -n <num>        :     executes tests with <num> iterations num must be > 0"
            echo "    -b <broker_type>:     sets gloabl test configuration of test broker,"
            echo "                          valid values are: 'local', 'container', default is 'container'"
            echo "                          others may be available see source at PSPGC_HOME"
            echo "    -t <test_filter>:     sets the test filter for pytest,"
            echo "                          where <test_filter> matches the test expression passed to pytest"
            echo "                          this can be passed multiple times"
        else
            printErr "Ignoring test run argument '$ARG'"

        fi

    done
    local ITERATIONS=${ITERATIONS:-1}
    local BROKER_TYPE=${BROKER_TYPE:-"container"}
    local COUNTER=0
    local RET=0

    local GO_TEST_CMD_ARGS=""
    set_krb_env $TEST_DIR
    set_oauth_env $TEST_DIR
    set_cache_env $TEST_DIR

    if [ "$TEST_FILTER" ]; then
        #local FILTER_ARGS="$( for f in $TEST_FILTER; do echo "--focus=$f"; done  ) "
        local FILTER_ARGS="--focus="
        #local GO_TEST_CMD="ginkgo $FILTER_ARGS\"$FILTER\""
        local GO_TEST_CMD="ginkgo"
    else
        # use go test for simple testing note needs additional timeout as test can be longer then 10 mins
        local GO_TEST_CMD="go test -timeout 1200s"
        local GO_TEST_CMD="ginkgo "
    fi

    if [[ $ITERATIONS -gt 1 ]]; then
        local GO_TEST_CMD_ARGS="$GO_TEST_CMD_ARGS --repeat=$ITERATIONS"
    fi

    if [[ "$BROKER_TYPE" == "container" ]]; then
        # use defaults
        printTrace "Running defaults"
    else
        # using local
        # set PUBSUB_IT_CONFIG to the TEST_DIR/data/config/config.json which has the example defaults for localhost testing
        # unless otherwise set
        export PUBSUB_IT_CONFIG=${PUBSUB_IT_CONFIG:-$TEST_DIR/data/config/config.json}
        local GO_TEST_CMD_ARGS="$GO_TEST_CMD_ARGS -tags remote --covermode atomic -coverprofile ./reports/coverage.out -coverpkg solace.dev/go/messaging/internal/...,solace.dev/go/messaging/pkg/..."
        printTrace "Run tests using $PUBSUB_IT_CONFIG"
    fi
    local log_capture=$TEST_OUTDIR/test_output.log
    local GO_TEST_CMD_ARGS="$GO_TEST_CMD_ARGS --junit-report=./reports/junit/test_report.xml --no-color"
    if [ "$FILTER_ARGS" ]; then
        echo "Running test with cmd $GO_TEST_CMD $GO_TEST_CMD_ARGS $FILTER_ARGS\"$TEST_FILTER\" to $TEST_OUTDIR"
        (cd $TEST_DIR && pwd && $GO_TEST_CMD $GO_TEST_CMD_ARGS $FILTER_ARGS"$TEST_FILTER" 2>&1 | tee $log_capture)
    else
        echo "Running test with cmd $GO_TEST_CMD $GO_TEST_CMD_ARGS to $TEST_OUTDIR"
        (cd $TEST_DIR && pwd && $GO_TEST_CMD $GO_TEST_CMD_ARGS  2>&1 | tee $log_capture)
    fi
    if [[ "$BROKER_TYPE" == "container" ]]; then
        # copy the diagnostic.tgz into test report dir
        local DIAG_FILE=$TEST_DIR/diagnostics.tgz
        local TEST_OUT_DIAG_FILE=$TEST_OUTDIR/diagnostics.tgz
        if [[ -e $DIAG_FILE ]]; then
            cp $DIAG_FILE $TEST_OUT_DIAG_FILE
            printTrace "Inflating diagnostics file $TEST_OUT_DIAG_FILE"
            (cd $TEST_OUTDIR && tar xvf diagnostics.tgz)
        else
            printTrace "Missing container diagnostics archive $DIAG_FILE"
        fi
    fi
    echo "Finished running test output in $TEST_OUTDIR"
}

