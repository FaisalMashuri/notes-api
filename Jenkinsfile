pipeline {
  agent any

  options {
    timeout(time: 20, unit: 'MINUTES')
    disableConcurrentBuilds()                     // satu build sekaligus, hemat RAM
    buildDiscarder(logRotator(numToKeepStr: '20'))
  }

  environment {
    TAG                 = "${env.GIT_COMMIT.take(8)}"
    PROJECT_NAME        = 'notes-api'
    DOCKER_REGISTRY_URL = 'ghcr.io'
    REGISTRY_PATH       = 'faisalmashuri'
    DOCKER_CREDENTIALS  = credentials('ghcr')     // otomatis tersedia sebagai _USR dan _PSW
    ENV = "production"
    NAMESPACE = "apps"
    // ENV, NAMESPACE, APP_URL sengaja tidak di sini: diisi per branch di stage Prepare
  }

  stages {

    stage('Test') {                               // jalan di semua branch dan PR
      steps {
        sh '''
          docker run --rm -v "$PWD":/src -w /src -e GOCACHE=/tmp/gocache \
            golang:1.26-alpine sh -c "go vet ./... && go test ./... -v"
        '''
      }
    }

    stage('Build') {
      steps {
        sh '''
          IMAGE=$DOCKER_REGISTRY_URL/$REGISTRY_PATH/$PROJECT_NAME
          echo "$DOCKER_CREDENTIALS_PSW" | docker login $DOCKER_REGISTRY_URL -u "$DOCKER_CREDENTIALS_USR" --password-stdin
          docker build --build-arg VERSION=$TAG -t $IMAGE:$TAG .
          docker push $IMAGE:$TAG
          docker rmi $IMAGE:$TAG
          docker logout $DOCKER_REGISTRY_URL
        '''
      }
    }

    stage('Deploy') {

      steps {
        sh '''
          helm upgrade $PROJECT_NAME ./helm/$PROJECT_NAME \
            --set-string image.repository=$DOCKER_REGISTRY_URL/$REGISTRY_PATH/$PROJECT_NAME,image.tag=$TAG \
            -f helm/$PROJECT_NAME/values.$ENV.yaml \
            --install --namespace $NAMESPACE --create-namespace \
            --wait --timeout 3m --atomic
        '''
      }
    }
  }

  post {
    always { sh 'docker image prune -f' }
  }
}