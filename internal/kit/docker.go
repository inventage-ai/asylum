package kit

func init() {
	Register(&Kit{
		Name:        "docker",
		Description: "Docker-in-Docker support",
		ConfigSnippet: `  docker:               # Docker-in-Docker support
`,
		ConfigNodes:   configNodes("docker", "Docker-in-Docker support", nil),
		ConfigComment: "docker:               # Docker-in-Docker support",
		DockerSnippet: `# Install Docker engine (repo already configured by core)
USER root
RUN apt-get update && \
    apt-get install -y --no-install-recommends docker-ce docker-buildx-plugin docker-compose-plugin && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    usermod -aG docker ${USERNAME}
USER ${USERNAME}
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
		ContainerFunc: dockerContainerFunc,
	})
}

func dockerContainerFunc(opts ContainerOpts) ([]RunArg, error) {
	return []RunArg{
		{Flag: "--privileged", Value: "", Source: "docker kit", Priority: PriorityKit},
		{Flag: "-e", Value: "ASYLUM_DOCKER=1", Source: "docker kit", Priority: PriorityKit},
	}, nil
}
