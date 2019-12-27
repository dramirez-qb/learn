node('master') {
  stage ('build') {
    openshiftBuild(buildConfig: 'learn', showBuildLogs: 'true')
  }
  stage ('deploy') {
    openshiftDeploy(deploymentConfig: 'learn')
  }
}
