pipeline {
    agent {
        docker {
            image 'docker:26.1.3'
            args '-v /var/run/docker.sock:/var/run/docker.sock'
        }
    }

    environment {
        CONTAINER_NAME  = credentials('GOEN_API_CONTAINER_NAME')
        PLATFORM        = credentials('PLATFORM')
        SUNFLOWER_TZ    = credentials('SUNFLOWER_TZ')

        GOEN_V2_DOMAIN  = credentials('GOEN_V2_DOMAIN')
        TRAEFIK_NETWORK = credentials('TRAEFIK_NETWORK')
        DATA_NETWORK    = credentials('DATA_NETWORK')

        JWT_SECRET      = credentials('GOEN_API_JWT_SECRET')

        POSTGRES_CONTAINER_NAME = credentials('POSTGRES_CONTAINER_NAME')
        POSTGRES_PORT           = credentials('POSTGRES_PORT')
        POSTGRES_DEFAULT        = credentials('POSTGRES_DEFAULT')
        POSTGRES_PASSWORD       = credentials('POSTGRES_PASSWORD')

        REDIS_CONTAINER_NAME    = credentials('REDIS_CONTAINER_NAME')
        REDIS_PORT              = credentials('REDIS_PORT')
        REDIS_USERNAME          = credentials('REDIS_USERNAME')
        REDIS_PASSWORD          = credentials('REDIS_PASSWORD')

        GOEN_MARKET_DATA_CONTAINER_NAME = credentials('GOEN_MARKET_DATA_CONTAINER_NAME')
        GOEN_MARKET_DATA_STATUS_PORT    = credentials('GOEN_MARKET_DATA_STATUS_PORT')
    }

    stages {
        stage('Deploy') {
            steps {
                echo 'Building and deploying goen-api...'
                sh '''#!/bin/sh
                    export GOEN_DOMAIN="${GOEN_V2_DOMAIN}"
                    export DATABASE_URL="postgres://${POSTGRES_DEFAULT}:${POSTGRES_PASSWORD}@${POSTGRES_CONTAINER_NAME}:${POSTGRES_PORT}/goen?sslmode=disable"
                    export REDIS_URL="redis://${REDIS_USERNAME}:${REDIS_PASSWORD}@${REDIS_CONTAINER_NAME}:${REDIS_PORT}/0"
                    export MARKET_DATA_STATUS_URL="http://${GOEN_MARKET_DATA_CONTAINER_NAME}:${GOEN_MARKET_DATA_STATUS_PORT}/status"

                    echo "Using GOEN_DOMAIN=$GOEN_DOMAIN"
                    docker compose -f docker-compose.prod.yml config | grep -E 'traefik.http.routers.*rule' || true
                    docker compose -f docker-compose.prod.yml up -d --build --force-recreate --remove-orphans
                '''
            }
        }

        stage('Health Check') {
            steps {
                echo 'Waiting for goen-api container to become ready...'
                sh '''#!/bin/sh
                    MAX_ATTEMPTS=24
                    ATTEMPT=0

                    while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
                        STATUS=$(docker ps -a --format '{{.Status}}' --filter "name=${CONTAINER_NAME}" | grep -i 'up')

                        if [ -n "$STATUS" ]; then
                            echo "✓ goen-api container is running"
                            exit 0
                        fi

                        ATTEMPT=$((ATTEMPT + 1))
                        echo "Waiting... (attempt: $ATTEMPT/$MAX_ATTEMPTS)"
                        sleep 5
                    done

                    echo "✗ Readiness check timed out after 2 minutes"
                    docker logs --tail 50 ${CONTAINER_NAME} || true
                    exit 1
                '''

                echo 'Verifying Traefik labels/networks and in-network API reachability...'
                sh '''#!/bin/sh
                    echo "Container labels (Traefik-related):"
                    docker inspect "${CONTAINER_NAME}" --format '{{range $k, $v := .Config.Labels}}{{println $k "=" $v}}{{end}}' | grep '^traefik[.]' || true

                    echo "Container networks:"
                    docker inspect "${CONTAINER_NAME}" --format '{{range $k, $v := .NetworkSettings.Networks}}{{println $k}}{{end}}'

                    if ! docker inspect "${CONTAINER_NAME}" --format '{{range $k, $v := .NetworkSettings.Networks}}{{println $k}}{{end}}' | grep -Fx "${TRAEFIK_NETWORK}" >/dev/null; then
                        echo "✗ ${CONTAINER_NAME} is NOT attached to TRAEFIK_NETWORK=${TRAEFIK_NETWORK}"
                        exit 1
                    fi

                    echo "Members in network ${TRAEFIK_NETWORK}:"
                    docker network inspect "${TRAEFIK_NETWORK}" --format '{{range .Containers}}{{println .Name}}{{end}}' || true

                    echo "API direct check on ${CONTAINER_NAME}:8080/api/v1/ping"
                    docker run --rm --network "${TRAEFIK_NETWORK}" curlimages/curl:8.8.0 -fsS "http://${CONTAINER_NAME}:8080/api/v1/ping" >/dev/null
                    echo "✓ Direct API ping succeeded"
                '''
            }
        }
    }

    post {
        success {
            echo '✓ goen-api deployment completed successfully'
        }
        failure {
            echo '✗ goen-api deployment failed'
        }
        always {
            echo 'Pipeline finished'
            sh 'docker image prune -f || true'
        }
    }
}