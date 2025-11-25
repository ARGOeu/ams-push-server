pipeline {
    agent { 
        docker { 
            image 'argo.registry:5000/rocky9-go1.25:latest'
            args '-u jenkins:jenkins'
        }
    }
    options {
        checkoutToSubdirectory('ams-push-server')
        newContainerPerStage()
    }
    environment {
        PROJECT_DIR='ams-push-server'
        GOPATH="${WORKSPACE}/go"
        GOCACHE = '/tmp/go-cache'
        GOMODCACHE = '/tmp/go-mod-cache'
        GIT_COMMIT=sh(script: "cd ${WORKSPACE}/$PROJECT_DIR && git log -1 --format=\"%H\"",returnStdout: true).trim()
        GIT_COMMIT_HASH=sh(script: "cd ${WORKSPACE}/$PROJECT_DIR && git log -1 --format=\"%H\" | cut -c1-7",returnStdout: true).trim()
        GIT_COMMIT_DATE=sh(script: "date -d \"\$(cd ${WORKSPACE}/$PROJECT_DIR && git show -s --format=%ci ${GIT_COMMIT_HASH})\" \"+%Y%m%d%H%M%S\"",returnStdout: true).trim()
   }
    stages {
        stage('Build') {
            steps {
                echo 'Build...'
                sh """
                go version
                mkdir -p ${WORKSPACE}/go/src/github.com/ARGOeu
                ln -sf ${WORKSPACE}/${PROJECT_DIR} ${WORKSPACE}/go/src/github.com/ARGOeu/${PROJECT_DIR}
                rm -rf ${WORKSPACE}/go/src/github.com/ARGOeu/${PROJECT_DIR}/${PROJECT_DIR}
                cd ${WORKSPACE}/go/src/github.com/ARGOeu/${PROJECT_DIR}
                go build
                """
            }
        }
        stage('Test') {
            steps {
                echo 'Test & Coverage...'
                sh """
                cd ${WORKSPACE}/go/src/github.com/ARGOeu/${PROJECT_DIR}
                gotestsum --junitfile ${WORKSPACE}/junit.xml -- -p 1 -v -coverprofile=coverage.out ./...
                gocover-cobertura < coverage.out > ${WORKSPACE}/coverage.xml
                """
                junit '**/junit.xml'
                publishCoverage adapters: [
                    coberturaAdapter('**/coverage.xml')
                ]
            }
        }
        stage('Package') {
            steps {
                echo 'Building Rpm...'
                withCredentials(bindings: [sshUserPrivateKey(credentialsId: 'jenkins-rpm-repo', usernameVariable: 'REPOUSER', \
                                                             keyFileVariable: 'REPOKEY')]) {
                    sh "/home/jenkins/build-rpm.sh -w ${WORKSPACE} -b ${BRANCH_NAME} -d rocky9 -p ${PROJECT_DIR} -s ${REPOKEY}"
                }
                archiveArtifacts artifacts: '**/*.rpm', fingerprint: true
            }
        } 
    }
    post{
        always {
            cleanWs()
        }
        success {
            script{
                if ( env.BRANCH_NAME == 'master' || env.BRANCH_NAME == 'devel' ) {
                    slackSend( message: ":rocket: New version for <$BUILD_URL|$PROJECT_DIR>:$BRANCH_NAME Job: $JOB_NAME !")
                }
            }
        }
        failure {
            script{
                if ( env.BRANCH_NAME == 'master' || env.BRANCH_NAME == 'devel' ) {
                    slackSend( message: ":rain_cloud: Build Failed for <$BUILD_URL|$PROJECT_DIR>:$BRANCH_NAME Job: $JOB_NAME")
                }
            }   
        }
    }
}
