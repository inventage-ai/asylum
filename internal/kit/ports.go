package kit

import (
	"strconv"

	"github.com/inventage-ai/asylum/internal/ports"
)

func init() {
	Register(&Kit{
		Name:        "ports",
		Description: "Automatic port forwarding for web services",
		Tier:        TierAlwaysOn,
		ConfigSnippet: `  # ports:              # Automatic port forwarding (on by default)
  #   count: 5          # Number of ports to allocate
`,
		ConfigComment: "ports:                # Automatic port forwarding (on by default)\n  count: 5            # Number of ports to allocate",
		ContainerFunc: portsContainerFunc,
	})
}

func portsContainerFunc(opts ContainerOpts) ([]RunArg, error) {
	pr, err := ports.Allocate(opts.ProjectDir, opts.ContainerName, opts.Config.PortCount())
	if err != nil {
		return nil, err
	}
	var args []RunArg
	for _, p := range pr.Ports() {
		ps := strconv.Itoa(p)
		args = append(args, RunArg{
			Flag:     "-p",
			Value:    ps + ":" + ps,
			Source:   "ports kit",
			Priority: PriorityKit,
		})
	}
	return args, nil
}
