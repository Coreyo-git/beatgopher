pipeline {
    agent any
    
    environment {
        // Define environment variables
        DOCKER_IMAGE = "beatgopher"
        CONTAINER_NAME = "beatgopher-prod"
        DOCKER_TAG = "${BUILD_NUMBER}"
        DISCORD_TOKEN = credentials('DISCORD_TOKEN')
    }
    
    // Only trigger on prod branch
    triggers {
        githubPush()
    }
    
    stages {
        stage('Checkout') {
            steps {
                echo 'Checking out code from prod branch...'
                checkout scm
            }
        }

        stage('Run Tests') {
            steps {
                echo 'Running tests...'
                script {
                    sh "docker build --target test -t ${DOCKER_IMAGE}-test ."
                    sh "docker run --rm ${DOCKER_IMAGE}-test"
                }
            }
        }

        stage('Build Docker Image') {
			steps {
				echo 'Building Production Docker image...'
				script {
					// --target release to get production image
					// --pull to ensure latest base images
					sh "docker build --target release --pull -t ${DOCKER_IMAGE}:${DOCKER_TAG} ."
					
					// Also tag as latest
					sh "docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:latest"
				}
			}
		}
		
        stage('Stop Existing Container') {
            steps {
                echo 'Stopping existing container...'
                script {
                    // Stop and remove existing container if it exists
                    sh '''
                        if docker ps -a | grep -q ${CONTAINER_NAME}; then
                            echo "Stopping existing container..."
                            docker stop ${CONTAINER_NAME} || true
                            docker rm ${CONTAINER_NAME} || true
                        else
                            echo "No existing container found"
                        fi
                    '''
                }
            }
        }
        
        stage('Deploy') {
            steps {
                echo 'Deploying new container...'
                script {
                    // Run the new container with environment variables from Jenkins credentials
                    sh '''
                        docker run -d \
                            --name ${CONTAINER_NAME} \
                            --restart unless-stopped \
                            -e DISCORD_TOKEN="${DISCORD_TOKEN}" \
                            ${DOCKER_IMAGE}:latest
                    '''
                }
            }
        }
        
        stage('Health Check') {
            steps {
                echo 'Performing health check...'
                script {
                    // Wait a moment for container to start
                    sh 'sleep 10'
                    
                    // Check if container is running
                    sh '''
                        if docker ps | grep -q ${CONTAINER_NAME}; then
                            echo "✅ Container is running successfully"
                            docker logs --tail 20 ${CONTAINER_NAME}
                        else
                            echo "❌ Container failed to start"
                            docker logs ${CONTAINER_NAME}
                            exit 1
                        fi
                    '''
                }
            }
        }
    }
    
    post {
        success {
            echo '🎉 Deployment successful!'
            // Optional: Send notification (Slack, email, etc.)
        }
        
        failure {
            echo '❌ Deployment failed!'
            // Cleanup on failure
            script {
                sh '''
                    if docker ps -a | grep -q ${CONTAINER_NAME}; then
                        docker stop ${CONTAINER_NAME} || true
                        docker rm ${CONTAINER_NAME} || true
                    fi
                '''
            }
        }
        
        always {
            // Clean up test image and old images to save space
            script {
                sh '''
                    # Remove test image
                    docker rmi ${DOCKER_IMAGE}-test || true

                    # Remove old images (keep last 2)
                    docker images ${DOCKER_IMAGE} --format "table {{.Tag}}" | grep -E "^[0-9]+$" | sort -nr | tail -n +3 | xargs -I {} docker rmi ${DOCKER_IMAGE}:{} || true
                '''
            }
        }
    }
}
