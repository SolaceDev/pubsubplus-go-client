properties([
    buildDiscarder(logRotator(daysToKeepStr: '30', numToKeepStr: '10')),
])
currentBuild.rawBuild.getParent().setQuietPeriod(0)


library 'jenkins-pipeline-library@main'

node('linux_docker') {
  notify(slackChannel: '#re-build') {
    stage('Build') {
      builder.goapi()
    }
  }
}
