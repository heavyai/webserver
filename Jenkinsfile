def buildStepComplete = false
// Sets individual commit (vs. PR) status. From example at https://plugins.jenkins.io/github .
void setBuildStatus(String message, String state) {
  step([
      $class: "GitHubCommitStatusSetter",
      reposSource: [$class: "ManuallyEnteredRepositorySource", url: "https://github.com/heavyai/webserver"],
      contextSource: [$class: "ManuallyEnteredCommitContextSource", context: "${JOB_NAME}"],
      errorHandlers: [[$class: "ChangingBuildStatusErrorHandler", result: "UNSTABLE"]],
      statusResultSource: [ $class: "ConditionalStatusResultSource", results: [[$class: "AnyBuildResult", message: message, state: state]] ]
  ]);
}

pipeline {
  agent { label "${label_expression}" }
  options {
    timestamps()
  }
  stages {
    stage('Setup Environment') {
      steps {
        // Set up the environment based on kind of node we are on and store the environment
        // information into a source-able file for later steps.
        sh '''
          if [ -e /etc/profile.d/modules.sh ] ; then
            source /etc/profile.d/modules.sh
          elif [ -e /etc/profile.d/xx-mapd-deps.sh ] ; then
            source /etc/profile.d/xx-mapd-deps.sh
          fi
          if command -v module 2> /dev/null; then
            module unload cuda mapd-deps
            module load cuda mapd-deps
          fi
          declare -px > environment.groovy
        '''
      }
    }
    stage('Install Golang Libraries') {
      steps {
        sh '''
          set +x // temporarily turn off Jenkins command echoing to clear up logs
          source environment.groovy
          set -x // re-enable echoing
          go get golang.org/x/crypto@v0.35.0 # needed for security updates.  Can remove after golang version is bumped to a version that is not vulnerable.
          go get golang.org/x/net@v0.33.0 # security, as above.
          go install golang.org/x/tools/cmd/goimports@latest
          go install golang.org/x/lint/golint@latest
          go mod tidy # needed to clean up the checksums after installing new module versions
        '''
      }
    }
    stage('Verify') {
      steps {
        sh '''
          set +x // temporarily turn off Jenkins command echoing to clear up logs
          source environment.groovy
          set -x // re-enable echoing
          ./scripts/verify.sh
        '''
      }
    }
    stage('Build All') {
      steps {
        sh '''
          set +x // temporarily turn off Jenkins command echoing to clear up logs
          source environment.groovy
          set -x // re-enable echoing
          ./scripts/build-all.sh
        '''
        script {
            buildStepComplete = true
        }
      }
    }
    /*
    stage('Verify licenses') {
      steps {
        sh 'go get github.com/omnisci/golicense@v0.3.0-omnisci'
        withCredentials([usernamePassword(credentialsId: 'f942705b-2827-427b-8741-fb2f1dbe9d46', usernameVariable: 'GITHUB_USER', passwordVariable: 'GITHUB_TOKEN')]) {
          sh 'GITHUB_TOKEN=${GITHUB_TOKEN} ./scripts/verify-licenses.sh'
        }
      }
    }
    */
    stage('Archive and symlink') {
      when {
        expression {
          buildStepComplete
        }
        anyOf {
          environment name: 'JOB_NAME', value: 'omnisci-webserver'
        }
      }
      steps {
        echo "Writing files to /theHoard/export/home/www/builds.mapd.com/frontend/webserver/${BUILD_NUMBER}"
        sshPublisher(
          alwaysPublishFromMaster: true,
          publishers: [
            sshPublisherDesc(
              configName: 'hoarder.mapd.com',
              transfers: [
                sshTransfer(
                  cleanRemote: false,
                  excludes: '',
                  execCommand: 'ln -sfvnr /theHoard/export/home/www/builds.mapd.com/frontend/webserver/${BUILD_NUMBER} /theHoard/export/home/www/builds.mapd.com/frontend/webserver/master',
                  execTimeout: 120000,
                  flatten: false,
                  makeEmptyDirs: false,
                  noDefaultExcludes: false,
                  patternSeparator: '[, ]+',
                  remoteDirectory: '/theHoard/export/home/www/builds.mapd.com/frontend/webserver/${BUILD_NUMBER}',
                  remoteDirectorySDF: false,
                  removePrefix: 'build',
                  sourceFiles: 'build/*.tar.gz'
                )
              ],
              usePromotionTimestamp: false,
              useWorkspaceInPromotion: false,
              verbose: false
            )
          ]
        )
      }
    }
    stage('Archive only') {
      when {
        expression {
          buildStepComplete
        }
        anyOf {
          environment name: 'JOB_NAME', value: 'omnisci-webserver-ondemand'
          environment name: 'JOB_NAME', value: 'omnisci-webserver-pr'
        }
      }
      steps {
        echo "Writing files to /theHoard/export/home/www/builds.mapd.com/frontend/${JOB_NAME}/${BUILD_NUMBER}"
        sshPublisher(
          alwaysPublishFromMaster: true,
          publishers: [
            sshPublisherDesc(
              configName: 'hoarder.mapd.com',
              transfers: [
                sshTransfer(
                  cleanRemote: false,
                  excludes: '',
                  execTimeout: 120000,
                  flatten: false,
                  makeEmptyDirs: false,
                  noDefaultExcludes: false,
                  patternSeparator: '[, ]+',
                  remoteDirectory: '/theHoard/export/home/www/builds.mapd.com/frontend/${JOB_NAME}/${BUILD_NUMBER}',
                  remoteDirectorySDF: false,
                  removePrefix: 'build',
                  sourceFiles: 'build/*.tar.gz'
                )
              ],
              usePromotionTimestamp: false,
              useWorkspaceInPromotion: false,
              verbose: false
            )
          ]
        )
      }
    }
  }
  post {
    success {
      script {
        setBuildStatus("Build succeeded", "SUCCESS")
      }
    }
    failure {
      script {
        setBuildStatus("Build failed", "FAILURE")
      }
    }
  }
}
