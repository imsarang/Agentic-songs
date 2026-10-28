MACHINE = """Machine flow:
- Pick the smallest tool that answers the question (do not dump every tool).
- Prefer machine_health for "is this ok / what's wrong", machine_spec for identity,
  machine_capacity for load/CPU/memory, machine_processes for top PIDs,
  machine_power / machine_gpu for battery/GPU, machine_logs for recent errors.
- machine_grep finds files by a keyword. machine_cat reads one path from that result. The lines under --- are already shown. Do not repeat the file text.
- Summarize JSON for the user in plain language; keep numbers that matter.
- Then action=finish.
"""
