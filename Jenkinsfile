node('docker') {

  properties([
    // Keep only the last 10 build to preserve space
    buildDiscarder(logRotator(numToKeepStr: '10')),
    parameters([
        string(name: 'version', trim: true, description: 'version used for plugin tar')
    ])
  ])

  stage("Clean") {
    sh "rm -rf plugins-dev-*.tar.gz plugins"
  }

  stage('Collect Plugins') {
    docker.image("cloudogu/scm-plugin-snapshot:1.1.0").inside("--entrypoint=''") {
      sh "/scm-plugin-snapshot plugins"
    }
  }

  stage('Archive Plugins') {
    docker.image("alpine:3.10.1").inside {
      sh "tar cvfz plugins-dev-${params.version}.tar.gz plugins"
    }
    archiveArtifacts "plugins-dev-${params.version}.tar.gz"
  }

}
