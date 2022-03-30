properties([
    buildDiscarder(logRotator(daysToKeepStr: '30', numToKeepStr: '10')),
])
currentBuild.rawBuild.getParent().setQuietPeriod(0)


library 'jenkins-pipeline-library@SOL-63732/tags_versioned'

node('linux_docker') {
  notify(slackChannel: '#re-build') {
    stage ('Checkout') {
      checkout scm
    }
    stage('Build') {
      deployer.code()
    }
  }
}
