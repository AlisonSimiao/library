#!/usr/bin/env python3
from __future__ import annotations

import argparse
import csv
from collections import defaultdict
from pathlib import Path

import matplotlib.pyplot as plt


def load_csv(path: Path) -> list[dict[str, str]]:
    with path.open(newline="", encoding="utf-8") as handle:
        return list(csv.DictReader(handle))


def to_float(value: str) -> float:
    try:
        return float(value)
    except (TypeError, ValueError):
        return 0.0


def plot_latency(summary_rows: list[dict[str, str]], output_dir: Path) -> None:
    grouped = defaultdict(dict)
    for row in summary_rows:
        grouped[row["stack"]][row["endpoint"]] = to_float(row["avg_latency_ms"])

    endpoints = sorted({row["endpoint"] for row in summary_rows})
    stacks = sorted(grouped)
    x = range(len(endpoints))
    width = 0.35 if len(stacks) == 2 else 0.8 / max(len(stacks), 1)

    fig, ax = plt.subplots(figsize=(12, 6))
    for idx, stack in enumerate(stacks):
        values = [grouped[stack].get(endpoint, 0.0) for endpoint in endpoints]
        offset = (idx - (len(stacks) - 1) / 2) * width
        ax.bar([pos + offset for pos in x], values, width=width, label=stack)

    ax.set_title("Latência média por endpoint")
    ax.set_ylabel("Milissegundos")
    ax.set_xticks(list(x))
    ax.set_xticklabels(endpoints, rotation=20, ha="right")
    ax.legend()
    ax.grid(axis="y", alpha=0.2)
    fig.tight_layout()
    fig.savefig(output_dir / "latency_average.png", dpi=180)
    plt.close(fig)


def plot_p95(summary_rows: list[dict[str, str]], output_dir: Path) -> None:
    grouped = defaultdict(dict)
    for row in summary_rows:
        grouped[row["stack"]][row["endpoint"]] = to_float(row["p95_latency_ms"])

    endpoints = sorted({row["endpoint"] for row in summary_rows})
    stacks = sorted(grouped)
    x = range(len(endpoints))
    width = 0.35 if len(stacks) == 2 else 0.8 / max(len(stacks), 1)

    fig, ax = plt.subplots(figsize=(12, 6))
    for idx, stack in enumerate(stacks):
        values = [grouped[stack].get(endpoint, 0.0) for endpoint in endpoints]
        offset = (idx - (len(stacks) - 1) / 2) * width
        ax.bar([pos + offset for pos in x], values, width=width, label=stack)

    ax.set_title("Latência p95 por endpoint")
    ax.set_ylabel("Milissegundos")
    ax.set_xticks(list(x))
    ax.set_xticklabels(endpoints, rotation=20, ha="right")
    ax.legend()
    ax.grid(axis="y", alpha=0.2)
    fig.tight_layout()
    fig.savefig(output_dir / "latency_p95.png", dpi=180)
    plt.close(fig)


def plot_rps(summary_rows: list[dict[str, str]], output_dir: Path) -> None:
    grouped = defaultdict(dict)
    for row in summary_rows:
        grouped[row["stack"]][row["endpoint"]] = to_float(row["rps"])

    endpoints = sorted({row["endpoint"] for row in summary_rows})
    stacks = sorted(grouped)
    x = range(len(endpoints))
    width = 0.35 if len(stacks) == 2 else 0.8 / max(len(stacks), 1)

    fig, ax = plt.subplots(figsize=(12, 6))
    for idx, stack in enumerate(stacks):
        values = [grouped[stack].get(endpoint, 0.0) for endpoint in endpoints]
        offset = (idx - (len(stacks) - 1) / 2) * width
        ax.bar([pos + offset for pos in x], values, width=width, label=stack)

    ax.set_title("Throughput estimado por endpoint")
    ax.set_ylabel("Requisições por segundo")
    ax.set_xticks(list(x))
    ax.set_xticklabels(endpoints, rotation=20, ha="right")
    ax.legend()
    ax.grid(axis="y", alpha=0.2)
    fig.tight_layout()
    fig.savefig(output_dir / "throughput_rps.png", dpi=180)
    plt.close(fig)


def plot_resources(container_rows: list[dict[str, str]], output_dir: Path) -> None:
    by_stack: dict[str, list[tuple[str, float, float]]] = defaultdict(list)
    for row in container_rows:
        stack = row.get("stack", "unknown")
        by_stack[stack].append((row["timestamp"], to_float(row["cpu_percent"].replace("%", "")), to_float(row["mem_percent"].replace("%", ""))))

    fig, axes = plt.subplots(2, 1, figsize=(12, 8), sharex=True)
    for stack, rows in sorted(by_stack.items()):
        if not rows:
            continue
        xs = list(range(len(rows)))
        cpu = [item[1] for item in rows]
        mem = [item[2] for item in rows]
        axes[0].plot(xs, cpu, marker="o", linewidth=1.2, label=stack)
        axes[1].plot(xs, mem, marker="o", linewidth=1.2, label=stack)

    axes[0].set_title("CPU por amostra")
    axes[0].set_ylabel("% CPU")
    axes[1].set_title("Memória por amostra")
    axes[1].set_ylabel("% memória")
    axes[1].set_xlabel("Amostras")
    for ax in axes:
        ax.grid(alpha=0.2)
        ax.legend()
    fig.tight_layout()
    fig.savefig(output_dir / "resource_usage.png", dpi=180)
    plt.close(fig)


def main() -> int:
    parser = argparse.ArgumentParser(description="Generate charts from benchmark CSV files.")
    parser.add_argument("--input-dir", required=True, help="Directory containing request_metrics.csv and container_metrics.csv.")
    args = parser.parse_args()

    input_dir = Path(args.input_dir)
    output_dir = input_dir / "plots"
    output_dir.mkdir(parents=True, exist_ok=True)

    summary_rows = load_csv(input_dir / "summary.csv")
    container_rows = load_csv(input_dir / "container_metrics.csv") if (input_dir / "container_metrics.csv").exists() else []

    plot_latency(summary_rows, output_dir)
    plot_p95(summary_rows, output_dir)
    plot_rps(summary_rows, output_dir)
    if container_rows:
        plot_resources(container_rows, output_dir)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
