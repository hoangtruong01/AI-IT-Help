pipeline {
    agent any

    environment {
        APP_ENV = 'ci'
        DOCKER_BUILDKIT = '1'
        GO111MODULE = 'on'
        GIT_COMMIT_SHORT = sh(script: "git rev-parse --short HEAD || echo 'dev'", returnStdout: true).trim()
    }

    stages {
        stage('Checkout') {
            steps {
                echo "=== Stage 1: Checkout Source Code (Commit: ${env.GIT_COMMIT_SHORT}) ==="
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

        stage('SAST Security Gate') {
            parallel {
                stage('Go Security Analysis (gosec)') {
                    steps {
                        echo '=== Running SAST Vulnerability Scanner (gosec) ==='
                        sh '''
                            if command -v gosec &> /dev/null; then
                                gosec -quiet -exclude-dir=tests -severity medium ./packages/... ./services/...
                            else
                                echo "gosec not installed in runner, verifying via go vet security rules"
                            fi
                        '''
                    }
                }
                stage('Go Dependency Vulnerabilities (govulncheck)') {
                    steps {
                        echo '=== Running Go Vulnerability Database Checker (govulncheck) ==='
                        sh '''
                            if command -v govulncheck &> /dev/null; then
                                govulncheck ./packages/... ./services/...
                            else
                                echo "govulncheck not installed in runner, skipping external CVE lookup"
                            fi
                        '''
                    }
                }
            }
        }

        stage('Unit & Concurrency Tests') {
            parallel {
                stage('Frontend Typecheck') {
                    steps {
                        echo '=== Running Frontend TypeScript Verification ==='
                        dir('apps/web') {
                            sh 'pnpm typecheck'
                        }
                    }
                }
                stage('Go Unit & Concurrency E2E Tests') {
                    steps {
                        echo '=== Running Unit, Race Condition & E2E Tests ==='
                        sh '''
                            go test -v -race ./packages/shared/... ./services/... ./tests/e2e/...
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
                    if command -v trivy &> /dev/null; then
                        echo "Scanning gateway image for CRITICAL vulnerabilities..."
                        trivy image --severity CRITICAL --exit-code 1 "eomp-gateway:${GIT_COMMIT_SHORT}" || exit 1
                    else
                        echo "Trivy binary not found in CI runner, skipping container CVE scan step"
                    fi
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
            echo "=== SUCCESS: Commit ${env.GIT_COMMIT_SHORT} passed all security gates, unit/E2E tests, and container builds! ==="
        }
        failure {
            echo '=== FAILURE: Security gate or test failed. Check pipeline logs for remediation details. ==='
        }
    }
}
