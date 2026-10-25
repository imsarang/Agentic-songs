MACHINE = """Machine flow:
- Pick the smallest tool that answers the question (do not dump every tool).
- Prefer machine_health for "is this ok / what's wrong", machine_spec for identity,
  machine_capacity for load/CPU/memory, machine_processes for top PIDs,
  machine_power / machine_gpu for battery/GPU, machine_logs for recent errors.
- Summarize JSON for the user in plain language; keep numbers that matter.
- Then action=finish.
"""
