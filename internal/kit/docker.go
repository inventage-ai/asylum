package kit

func init() {
	Register(&Kit{
		Name:        "docker",
		Description: "Docker-in-Docker support",
		DockerSnippet: `# Install Docker engine (for Docker-in-Docker support)
RUN curl -fsSL https://download.docker.com/linux/debian/gpg | \
    gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg && \
    chmod 644 /usr/share/keyrings/docker-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" \
    > /etc/apt/sources.list.d/docker.list && \
    apt-get update && \
    apt-get install -y docker-ce docker-buildx-plugin docker-compose-plugin && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    usermod -aG docker claude
`,
		RulesSnippet: `### Docker (docker kit)
Full Docker Engine is available (not just CLI). The container runs in privileged mode with dockerd started automatically. You can build and run containers directly.
`,
		EntrypointSnippet: `# Start Docker daemon if enabled and running privileged
if [ "${ASYLUM_DOCKER:-}" = "1" ]; then
    if [ -S /var/run/docker.sock ] && docker info >/dev/null 2>&1; then
        echo "Docker socket already available"
    elif command -v dockerd >/dev/null 2>&1; then
        echo "Starting Docker daemon..."
        sudo dockerd --storage-driver=vfs --log-level=warn >/tmp/dockerd.log 2>&1 &
        DOCKERD_PID=$!
        disown $DOCKERD_PID
        for i in $(seq 1 30); do
            if docker info >/dev/null 2>&1; then
                echo "Docker daemon ready"
                break
            fi
            sleep 1
        done
        if ! docker info >/dev/null 2>&1; then
            echo "Warning: Docker daemon failed to start (check /tmp/dockerd.log)"
        fi
    fi
fi
`,
	})
}
