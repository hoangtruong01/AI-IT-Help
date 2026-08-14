pipeline {
    agent any

    environment {
        APP_ENV = 'ci'
        DOCKER_BUILDKIT = '1'
        GO111MODULE = 'on'
    }

    stages {
        stage('Checkout') {
            steps {
                echo '=== Stage: Checkout Source Code ==='
                checkout scm
            }
        }

        stage('Install Dependencies') {
            parallel {
                stage('Frontend Dependencies') {
                    steps {
                        echo '=== Installing Frontend Dependencies (pnpm) ==='
                        dir('apps/web') {
                            sh 'pnpm install --frozen-lockfile'
                        }
                    }
                }
                stage('Go Modules') {
                    steps {
                        echo '=== Verifying Go Workspace Modules ==='
                        sh 'go work sync'
                    }
                }
            }
        }

        stage('Lint & Static Analysis') {
            parallel {
                stage('Frontend Lint') {
                    steps {
                        echo '=== Running Frontend ESLint ==='
                        dir('apps/web') {
                            sh 'pnpm lint'
                        }
                    }
                }
                stage('Go Vet & Formatting') {
                    steps {
                        echo '=== Running Go Vet & Format Verification ==='
                        sh '''
                            test -z "$(gofmt -l packages/ services/)" || { echo "Unformatted Go files detected"; exit 1; }
                            for dir in packages/shared services/*; do
                                if [ -f "$dir/go.mod" ]; then
                                    (cd "$dir" && go vet ./...)
                                fi
                            done
                        '''
                    }
                }
            }
        }

        stage('Run Tests') {
            parallel {
                stage('Frontend Typecheck') {
                    steps {
                        echo '=== Running Frontend TypeScript Verification ==='
                        dir('apps/web') {
                            sh 'pnpm typecheck'
                        }
                    }
                }
                stage('Go Unit Tests') {
                    steps {
                        echo '=== Running Go Unit Tests ==='
                        sh '''
                            for dir in packages/shared services/*; do
                                if [ -f "$dir/go.mod" ]; then
                                    (cd "$dir" && go test -v -race -cover ./...)
                                fi
                            done
                        '''
                    }
                }
            }
        }

        stage('Build Artifacts') {
            parallel {
                stage('Frontend Build') {
                    steps {
                        echo '=== Building Nuxt Frontend Production Bundle ==='
                        dir('apps/web') {
                            sh 'pnpm build'
                        }
                    }
                }
                stage('Go Services Build') {
                    steps {
                        echo '=== Compiling Go Microservices Binaries ==='
                        sh '''
                            mkdir -p bin
                            for dir in services/*; do
                                if [ -f "$dir/go.mod" ]; then
                                    svc=$(basename "$dir")
                                    echo "Building $svc..."
                                    (cd "$dir" && CGO_ENABLED=0 go build -ldflags="-w -s" -o "../../bin/$svc" ./cmd/server)
                                fi
                            done
                        '''
                    }
                }
            }
        }

        stage('Docker Image Build') {
            steps {
                echo '=== Building Microservice Docker Images ==='
                sh '''
                    for dir in services/*; do
                        if [ -f "$dir/Dockerfile" ]; then
                            svc=$(basename "$dir")
                            echo "Building Docker image: eomp-$svc:latest"
                            docker build -t "eomp-$svc:latest" -f "$dir/Dockerfile" .
                        fi
                    done
                '''
            }
        }
    }

    post {
        always {
            echo '=== CI Pipeline Finished ==='
            cleanWs deleteDirs: true, notFailBuild: true
        }
        success {
            echo '=== SUCCESS: All builds, tests, and images verified successfully! ==='
        }
        failure {
            echo '=== FAILURE: Pipeline encountered an error. Check logs for details. ==='
        }
    }
}
