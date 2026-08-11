def vaultConfig = [
    vaultUrl: 'http://sunflower-openbao:8200',
    vaultCredentialId: 'vault-approle',
    engineVersion: 2
]

def vaultSecrets = [
    [path: 'secret/common', engineVersion: 2, secretValues: [
        [envVar: 'PLATFORM', vaultKey: 'PLATFORM'],
        [envVar: 'SUNFLOWER_TZ', vaultKey: 'SUNFLOWER_TZ'],
        [envVar: 'TRAEFIK_NETWORK', vaultKey: 'TRAEFIK_NETWORK'],
        [envVar: 'DATA_NETWORK', vaultKey: 'DATA_NETWORK']
    ]],
    [path: 'secret/postgresql', engineVersion: 2, secretValues: [
        [envVar: 'POSTGRES_CONTAINER_NAME', vaultKey: 'POSTGRES_CONTAINER_NAME'],
        [envVar: 'POSTGRES_PORT', vaultKey: 'POSTGRES_PORT'],
        [envVar: 'POSTGRES_DEFAULT', vaultKey: 'POSTGRES_DEFAULT'],
        [envVar: 'POSTGRES_PASSWORD', vaultKey: 'POSTGRES_PASSWORD']
    ]],
    [path: 'secret/redis', engineVersion: 2, secretValues: [
        [envVar: 'REDIS_CONTAINER_NAME', vaultKey: 'REDIS_CONTAINER_NAME'],
        [envVar: 'REDIS_PORT', vaultKey: 'REDIS_PORT'],
        [envVar: 'REDIS_USERNAME', vaultKey: 'REDIS_USERNAME'],
        [envVar: 'REDIS_PASSWORD', vaultKey: 'REDIS_PASSWORD']
    ]],
    [path: 'secret/seaweedfs', engineVersion: 2, secretValues: [
        [envVar: 'SEAWEEDFS_ACCESS_KEY_ID', vaultKey: 'SEAWEEDFS_ACCESS_KEY_ID'],
        [envVar: 'SEAWEEDFS_SECRET_ACCESS_KEY', vaultKey: 'SEAWEEDFS_SECRET_ACCESS_KEY']
    ]],
    [path: 'secret/goen', engineVersion: 2, secretValues: [
        [envVar: 'CONTAINER_NAME', vaultKey: 'GOEN_API_CONTAINER_NAME'],
        [envVar: 'GOEN_V2_DOMAIN', vaultKey: 'GOEN_V2_DOMAIN'],
        [envVar: 'JWT_SECRET', vaultKey: 'GOEN_API_JWT_SECRET'],
        [envVar: 'GOEN_MARKET_DATA_CONTAINER_NAME', vaultKey: 'GOEN_MARKET_DATA_CONTAINER_NAME'],
        [envVar: 'GOEN_MARKET_DATA_STATUS_PORT', vaultKey: 'GOEN_MARKET_DATA_STATUS_PORT']
    ]]
]

pipeline {
    agent {
        docker {
            image 'docker:26.1.3'
            args '-v /var/run/docker.sock:/var/run/docker.sock'
        }
    }

    stages {
        stage('Deploy') {
            steps {
                echo 'Building and deploying goen-api...'
                withVault(configuration: vaultConfig, vaultSecrets: vaultSecrets) {
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
        }

        stage('Health Check') {
            steps {
                echo 'Waiting for goen-api container to become ready...'
                withVault(configuration: vaultConfig, vaultSecrets: vaultSecrets) {
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