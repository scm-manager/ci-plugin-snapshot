node('docker') {

  properties([
    // Keep only the last 10 build to preserve space
    buildDiscarder(logRotator(daysToKeepStr: '180')),
    parameters([
        string(name: 'version', defaultValue: 'SNAPSHOT', trim: true, description: 'version used for plugin tar')
    ])
  ])

  stage("Clean") {
    sh "rm -rf plugins-dev-*.tar.gz plugins"
  }

  stage('Collect Plugins') {
    docker.image("scmmanager/ci-plugin-snapshot:1.1.7").inside("--entrypoint=''") {
      sh "/ci-plugin-snapshot plugins"
    }
  }

  stage('Archive Plugins') {
    docker.image("alpine:3.11.3").inside {
      sh "tar cvfz plugins-dev-${params.version}.tar.gz plugins"
    }
    archiveArtifacts "plugins/plugin-center.json"
    archiveArtifacts "plugins-dev-${params.version}.tar.gz"
  }

}
