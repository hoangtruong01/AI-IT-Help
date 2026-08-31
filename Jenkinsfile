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
                checkout scm
                script {
                    env.GIT_COMMIT_SHORT = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                }
                echo "=== Stage 1: Checkout Source Code (Commit: ${env.GIT_COMMIT_SHORT}) ==="
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

        stage('Code Quality & Linting') {
            parallel {
                stage('Frontend Lint') {
                    steps {
                        echo '=== Running Frontend ESLint ==='
                        dir('apps/web') {
                            sh 'pnpm lint'
                        }
                    }
                }
                stage('Go Vet & Format Verification') {
                    steps {
                        echo '=== Running Go Vet & Format Verification ==='
                        sh '''
                            test -z "$(gofmt -l packages/ services/ scripts/)" || { echo "Unformatted Go files detected"; exit 1; }
                            for dir in packages/shared services/*; do
                                if [ -f "$dir/go.mod" ]; then
                                    (cd "$dir" && go vet ./...)
                                fi
                            done
                        '''
                    }
                }
                stage('Runtime Route & OpenAPI Contract') {
                    steps {
                        echo '=== Verifying Route Compatibility and OpenAPI Coverage ==='
                        sh 'go run scripts/check_openapi_coverage.go'
                    }
                }
            }
        }

        stage('SAST Security Gate') {
            parallel {
                stage('Go Security Analysis (gosec)') {
                    steps {
                        echo '=== Running SAST Vulnerability Scanner (gosec) ==='
                        sh '''
                            command -v gosec >/dev/null 2>&1 || { echo "gosec is required on the CI runner"; exit 1; }
                            for dir in packages/shared services/*; do
                                if [ -f "$dir/go.mod" ]; then
                                    (cd "$dir" && gosec -quiet -severity medium ./...)
                                fi
                            done
                        '''
                    }
                }
                stage('Go Dependency Vulnerabilities (govulncheck)') {
                    steps {
                        echo '=== Running Go Vulnerability Database Checker (govulncheck) ==='
                        sh '''
                            command -v govulncheck >/dev/null 2>&1 || { echo "govulncheck is required on the CI runner"; exit 1; }
                            for dir in packages/shared services/*; do
                                if [ -f "$dir/go.mod" ]; then
                                    (cd "$dir" && govulncheck ./...)
                                fi
                            done
                        '''
                    }
                }
            }
        }

        stage('Unit & Concurrency Tests') {
            parallel {
                stage('Frontend Unit Tests & Typecheck') {
                    steps {
                        echo '=== Running Frontend TypeScript Verification ==='
                        dir('apps/web') {
                            sh 'pnpm test && pnpm typecheck'
                        }
                    }
                }
                stage('Go Unit & In-Memory Simulation Tests') {
                    steps {
                        echo '=== Running Unit, Race Condition & In-Memory Simulation Tests ==='
                        sh '''
                            for dir in packages/shared services/* tests/e2e; do
                                if [ -f "$dir/go.mod" ]; then
                                    (cd "$dir" && go test -race ./...)
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
                                    echo "Compiling static binary for $svc..."
                                    (cd "$dir" && CGO_ENABLED=0 go build -ldflags="-w -s" -o "../../bin/$svc" ./cmd/server)
                                fi
                            done
                        '''
                    }
                }
            }
        }

        stage('Container Image Packaging & Pinning') {
            steps {
                echo "=== Building Docker Images with Git SHA Tag (${env.GIT_COMMIT_SHORT}) ==="
                sh '''
                    # Build Web Frontend Image
                    docker build -t "eomp-web:${GIT_COMMIT_SHORT}" -f deploy/docker/Dockerfile.web .

                    # Build 11 Go Microservice Images
                    for dir in services/*; do
                        if [ -f "$dir/go.mod" ]; then
                            svc=$(basename "$dir")
                            echo "Building Docker image: eomp-$svc:${GIT_COMMIT_SHORT}"
                            docker build \
                                --build-arg SERVICE_NAME="$svc" \
                                -t "eomp-$svc:${GIT_COMMIT_SHORT}" \
                                -f deploy/docker/Dockerfile.go-service .
                        fi
                    done
                '''
            }
        }

        stage('Container Vulnerability Scanning (Trivy)') {
            steps {
                echo '=== Scanning Docker Images for CVE Vulnerabilities (Trivy) ==='
                sh '''
                    command -v trivy >/dev/null 2>&1 || { echo "Trivy is required on the CI runner"; exit 1; }
                    trivy image --severity HIGH,CRITICAL --exit-code 1 "eomp-web:${GIT_COMMIT_SHORT}"
                    for dir in services/*; do
                        if [ -f "$dir/go.mod" ]; then
                            svc=$(basename "$dir")
                            trivy image --severity HIGH,CRITICAL --exit-code 1 "eomp-$svc:${GIT_COMMIT_SHORT}"
                        fi
                    done
                '''
            }
        }
    }

    post {
        always {
            echo '=== CI/CD DevSecOps Pipeline Execution Finished ==='
            cleanWs deleteDirs: true, notFailBuild: true
        }
        success {
            echo "=== SUCCESS: Commit ${env.GIT_COMMIT_SHORT} passed configured security gates, unit/simulation tests, and container builds; this is not deployed E2E or production acceptance. ==="
        }
        failure {
            echo '=== FAILURE: Security gate or test failed. Check pipeline logs for remediation details. ==='
        }
    }
}
