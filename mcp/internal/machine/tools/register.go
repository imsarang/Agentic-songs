package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func Register(s *server.MCPServer) {
	s.AddTool(mcp.NewTool("machine_spec",
		mcp.WithDescription("Read this machine's identity: vendor, product, CPU, memory, filesystems, uptime, and whether it is physical or virtual. Returns one JSON object."),
	), specHandler())

	s.AddTool(mcp.NewTool("machine_capacity",
		mcp.WithDescription("Read this machine's load, CPU percent over 200ms, PSI, memory, file count, PID count, and thread count. Returns one JSON object."),
	), capacityHandler())

	s.AddTool(mcp.NewTool("machine_processes",
		mcp.WithDescription("Read process state counts, zombies, disk-sleep tasks, kernel-thread count, and the top processes by CPU, RSS, and elapsed run time (etime_sec). Returns one JSON object."),
	), processesHandler())

	s.AddTool(mcp.NewTool("machine_cgroups",
		mcp.WithDescription("Read cgroup v2 controllers, the cgroup count, this process's cgroup, and the top cgroups by memory. Returns one JSON object."),
	), cgroupsHandler())

	s.AddTool(mcp.NewTool("machine_namespaces",
		mcp.WithDescription("Count distinct pid, net, mnt, user, uts, ipc, cgroup, and time namespaces. Returns one JSON object."),
	), namespacesHandler())

	s.AddTool(mcp.NewTool("machine_net",
		mcp.WithDescription("Read interfaces, the default route, socket counts, TCP states, listen ports, retransmits, and listen overflows. Returns one JSON object."),
	), netHandler())

	s.AddTool(mcp.NewTool("machine_blockers",
		mcp.WithDescription("Read disk-sleep tasks, file-lock counts, and recent kernel lockdep or OOM lines. Returns one JSON object."),
	), blockersHandler())

	s.AddTool(mcp.NewTool("machine_limits",
		mcp.WithDescription("Read pid_max, threads-max, file-max, somaxconn, and this process's open-file and process ulimits. Returns one JSON object."),
	), limitsHandler())

	s.AddTool(mcp.NewTool("machine_power",
		mcp.WithDescription("Read AC status, battery charge, full energy versus design energy, cycle count, and current watts. Returns one JSON object."),
	), powerHandler())

	s.AddTool(mcp.NewTool("machine_gpu",
		mcp.WithDescription("Read DRM GPU busy percent, VRAM, temperature, and power, and whether nvidia-smi sees a device. Returns one JSON object."),
	), gpuHandler())

	s.AddTool(mcp.NewTool("machine_logs",
		mcp.WithDescription("Read up to 40 systemd journal or kernel lines. Optional source (journal or kernel), since, priority, unit, and match."),
		mcp.WithString("source", mcp.Description("journal or kernel")),
		mcp.WithString("since", mcp.Description("journalctl --since value, such as -1h")),
		mcp.WithString("priority", mcp.Description("emerg, alert, crit, err, warning, notice, info, debug, or 0-7")),
		mcp.WithString("unit", mcp.Description("systemd unit name")),
		mcp.WithString("match", mcp.Description("journalctl -g pattern, one line")),
	), logsHandler())

	s.AddTool(mcp.NewTool("machine_health",
		mcp.WithDescription("Judge disk, memory, CPU, IO, network, and battery as ok, warn, or crit using fixed thresholds. Returns one JSON object."),
	), healthHandler())

	s.AddTool(mcp.NewTool("machine_grep",
		mcp.WithString("query", mcp.Required()),
		mcp.WithString("path"),
	), grepHandler())
	
	s.AddTool(mcp.NewTool("machine_cat",
		mcp.WithString("path", mcp.Required()),
	), catHandler())
}
