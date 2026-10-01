#!/usr/bin/env python3
from __future__ import annotations

import argparse
import csv
import datetime as dt
import json
import statistics
import subprocess
import threading
import time
import urllib.error
import urllib.request
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Callable


@dataclass(frozen=True)
class Endpoint:
    name: str
    method: str
    path: str
    payload_factory: Callable[[int], dict[str, Any]] | None = None


@dataclass
class RequestResult:
    stack: str
    scenario: str
    endpoint: str
    method: str
    iteration: int
    status_code: int | None
    latency_ms: float
    timestamp: str
    error: str


def utc_now() -> str:
    return dt.datetime.now(dt.timezone.utc).isoformat()


def percentile(values: list[float], pct: float) -> float:
    if not values:
        return 0.0
    ordered = sorted(values)
    if len(ordered) == 1:
        return ordered[0]
    rank = (len(ordered) - 1) * (pct / 100.0)
    lower = int(rank)
    upper = min(lower + 1, len(ordered) - 1)
    weight = rank - lower
    return ordered[lower] * (1 - weight) + ordered[upper] * weight


def request_json(base_url: str, method: str, path: str, payload: dict[str, Any] | None = None) -> tuple[int, float, str]:
    url = base_url.rstrip("/") + path
    data = None
    headers = {"Content-Type": "application/json"}
    if payload is not None:
        data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(url, data=data, method=method, headers=headers)
    started = time.perf_counter()
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            body = resp.read().decode("utf-8", errors="replace")
            elapsed = (time.perf_counter() - started) * 1000.0
            return resp.status, elapsed, body.strip()
    except urllib.error.HTTPError as exc:
        elapsed = (time.perf_counter() - started) * 1000.0
        body = exc.read().decode("utf-8", errors="replace") if exc.fp else ""
        return exc.code, elapsed, body.strip()


def wait_for_service(base_url: str, timeout_seconds: int = 120) -> None:
    deadline = time.time() + timeout_seconds
    health_paths = ["/api/users", "/api/books", "/api/loans"]
    last_error = ""
    while time.time() < deadline:
        for path in health_paths:
            try:
                status, _, _ = request_json(base_url, "GET", path)
                if status == 200:
                    return
            except Exception as exc:  # noqa: BLE001
                last_error = str(exc)
        time.sleep(1)
    raise RuntimeError(f"service did not become ready: {last_error}")


def compose_container_ids(compose_dir: Path) -> list[str]:
    result = subprocess.run(
        ["docker", "compose", "ps", "-q"],
        cwd=compose_dir,
        capture_output=True,
        text=True,
        check=True,
    )
    return [line.strip() for line in result.stdout.splitlines() if line.strip()]


def collect_docker_stats(
    stack: str,
    compose_dir: Path,
    output_file: Path,
    stop_event: threading.Event,
    interval_seconds: float,
) -> None:
    container_ids = compose_container_ids(compose_dir)
    if not container_ids:
        return

    with output_file.open("w", newline="", encoding="utf-8") as handle:
        writer = csv.DictWriter(
            handle,
            fieldnames=[
                "stack",
                "timestamp",
                "container_id",
                "name",
                "cpu_percent",
                "mem_percent",
                "mem_usage",
                "net_io",
                "block_io",
            ],
        )
        writer.writeheader()
        while not stop_event.is_set():
            result = subprocess.run(
                [
                    "docker",
                    "stats",
                    "--no-stream",
                    "--format",
                    "{{.ID}}|{{.Name}}|{{.CPUPerc}}|{{.MemPerc}}|{{.MemUsage}}|{{.NetIO}}|{{.BlockIO}}",
                    *container_ids,
                ],
                cwd=compose_dir,
                capture_output=True,
                text=True,
                check=False,
            )
            timestamp = utc_now()
            for line in result.stdout.splitlines():
                parts = line.split("|")
                if len(parts) != 7:
                    continue
                writer.writerow(
                    {
                        "stack": stack,
                        "timestamp": timestamp,
                        "container_id": parts[0],
                        "name": parts[1],
                        "cpu_percent": parts[2],
                        "mem_percent": parts[3],
                        "mem_usage": parts[4],
                        "net_io": parts[5],
                        "block_io": parts[6],
                    }
                )
            handle.flush()
            stop_event.wait(interval_seconds)


def create_user(base_url: str, run_id: str, index: int) -> dict[str, Any]:
    status, _, body = request_json(
        base_url,
        "POST",
        "/api/users",
        {
            "name": f"Benchmark User {index}",
            "email": f"benchmark-user-{run_id}-{index}@example.com",
        },
    )
    if status not in (200, 201):
        raise RuntimeError(f"failed to create user {index}: {status} {body}")
    response = json.loads(body or "{}")
    return response.get("data", response)


def create_book(base_url: str, run_id: str, index: int, quantity: int) -> dict[str, Any]:
    status, _, body = request_json(
        base_url,
        "POST",
        "/api/books",
        {
            "title": f"Benchmark Book {run_id}-{index}",
            "author": "Codex Benchmark",
            "isbn": f"9780{run_id[-8:]}{index:04d}",
            "quantity": quantity,
        },
    )
    if status not in (200, 201):
        raise RuntimeError(f"failed to create book {index}: {status} {body}")
    response = json.loads(body or "{}")
    return response.get("data", response)


def create_loan(base_url: str, user_id: int, book_id: int) -> dict[str, Any]:
    status, _, body = request_json(base_url, "POST", "/api/loans", {"user_id": user_id, "book_id": book_id})
    if status not in (200, 201):
        raise RuntimeError(f"failed to create loan for user={user_id} book={book_id}: {status} {body}")
    response = json.loads(body or "{}")
    return response.get("data", response)


def run_scenario(
    stack: str,
    scenario: str,
    base_url: str,
    endpoints: list[Endpoint],
    iterations: int,
    concurrency: int,
) -> tuple[list[RequestResult], float]:
    results: list[RequestResult] = []
    lock = threading.Lock()
    index = 0
    started = time.perf_counter()

    def worker() -> None:
        nonlocal index
        while True:
            with lock:
                if index >= iterations:
                    return
                current = index
                index += 1
            endpoint = endpoints[current % len(endpoints)]
            payload = endpoint.payload_factory(current) if endpoint.payload_factory else None
            try:
                status, latency, error = request_json(base_url, endpoint.method, endpoint.path, payload)
            except Exception as exc:  # noqa: BLE001
                status = None
                latency = 0.0
                error = str(exc)
            else:
                if status is None:
                    error = "unknown error"
            results.append(
                RequestResult(
                    stack=stack,
                    scenario=scenario,
                    endpoint=endpoint.name,
                    method=endpoint.method,
                    iteration=current + 1,
                    status_code=status,
                    latency_ms=latency,
                    timestamp=utc_now(),
                    error=error,
                )
            )

    threads = [threading.Thread(target=worker, daemon=True) for _ in range(concurrency)]
    for thread in threads:
        thread.start()
    for thread in threads:
        thread.join()
    elapsed = time.perf_counter() - started
    return results, elapsed


def write_request_csv(path: Path, rows: list[RequestResult]) -> None:
    with path.open("w", newline="", encoding="utf-8") as handle:
        writer = csv.DictWriter(
            handle,
            fieldnames=[
                "stack",
                "scenario",
                "endpoint",
                "method",
                "iteration",
                "status_code",
                "latency_ms",
                "timestamp",
                "error",
            ],
        )
        writer.writeheader()
        for row in rows:
            writer.writerow(row.__dict__)


def write_summary_csv(path: Path, rows: list[RequestResult], durations: dict[tuple[str, str], float]) -> None:
    grouped: dict[tuple[str, str, str], list[float]] = {}
    statuses: dict[tuple[str, str, str], list[int]] = {}
    for row in rows:
        key = (row.stack, row.scenario, row.endpoint)
        grouped.setdefault(key, []).append(row.latency_ms)
        statuses.setdefault(key, []).append(row.status_code or 0)

    with path.open("w", newline="", encoding="utf-8") as handle:
        writer = csv.DictWriter(
            handle,
            fieldnames=[
                "stack",
                "scenario",
                "endpoint",
                "requests",
                "avg_latency_ms",
                "median_latency_ms",
                "p95_latency_ms",
                "min_latency_ms",
                "max_latency_ms",
                "success_rate",
                "rps",
            ],
        )
        writer.writeheader()
        for key, values in sorted(grouped.items()):
            stack, scenario, endpoint = key
            total = len(values)
            successes = sum(1 for code in statuses[key] if 200 <= code < 400)
            total_time_s = durations.get((stack, scenario), 0.0)
            writer.writerow(
                {
                    "stack": stack,
                    "scenario": scenario,
                    "endpoint": endpoint,
                    "requests": total,
                    "avg_latency_ms": round(statistics.fmean(values), 3),
                    "median_latency_ms": round(statistics.median(values), 3),
                    "p95_latency_ms": round(percentile(values, 95), 3),
                    "min_latency_ms": round(min(values), 3),
                    "max_latency_ms": round(max(values), 3),
                    "success_rate": round((successes / total) * 100.0, 2),
                    "rps": round(total / total_time_s, 3) if total_time_s else 0.0,
                }
            )


def append_note(path: Path, text: str) -> None:
    with path.open("a", encoding="utf-8") as handle:
        handle.write(text.rstrip() + "\n")


def resolve_compose_dir(raw_value: str) -> Path:
    candidate = Path(raw_value)
    if candidate.exists():
        return candidate

    stripped = raw_value.removeprefix("library/")
    candidate = Path(stripped)
    if candidate.exists():
        return candidate

    script_dir = Path(__file__).resolve().parent
    repo_library_dir = script_dir.parent
    for base in (repo_library_dir, repo_library_dir.parent):
        candidate = base / raw_value
        if candidate.exists():
            return candidate
        candidate = base / stripped
        if candidate.exists():
            return candidate

    raise FileNotFoundError(f"compose directory not found: {raw_value}")


def main() -> int:
    parser = argparse.ArgumentParser(description="Run repeatable benchmarks against the library API.")
    parser.add_argument("--stack", required=True, choices=["mono", "micro"], help="Architecture label for the results.")
    parser.add_argument("--base-url", required=True, help="Base URL of the API under test, e.g. http://localhost:8080")
    parser.add_argument(
        "--compose-dir",
        required=True,
        help="Directory containing docker-compose.yml for the running stack.",
    )
    parser.add_argument("--output-dir", default="benchmarks/results", help="Directory to store CSV files.")
    parser.add_argument("--requests", type=int, default=60, help="Number of requests per scenario.")
    parser.add_argument("--concurrency", type=int, default=8, help="Worker threads per scenario.")
    parser.add_argument("--sample-interval", type=float, default=1.0, help="docker stats sampling interval in seconds.")
    parser.add_argument("--seed-count", type=int, default=120, help="Number of seed users/books for read and loan tests.")
    args = parser.parse_args()

    output_dir = Path(args.output_dir) / args.stack
    output_dir.mkdir(parents=True, exist_ok=True)

    base_url = args.base_url.rstrip("/")
    compose_dir = resolve_compose_dir(args.compose_dir)
    run_id = dt.datetime.now(dt.timezone.utc).strftime("%Y%m%d%H%M%S")

    wait_for_service(base_url)

    seeded_users = [create_user(base_url, run_id, i + 1) for i in range(args.seed_count)]
    seeded_books = [create_book(base_url, run_id, i + 1, quantity=5) for i in range(args.seed_count)]
    seeded_loans = [create_loan(base_url, seeded_users[i]["id"], seeded_books[i]["id"]) for i in range(args.seed_count)]

    scenarios: list[tuple[str, list[Endpoint]]] = [
        (
            "read_users",
            [Endpoint("GET /api/users", "GET", "/api/users")],
        ),
        (
            "read_books",
            [Endpoint("GET /api/books", "GET", "/api/books")],
        ),
        (
            "read_loans",
            [Endpoint("GET /api/loans", "GET", "/api/loans")],
        ),
        (
            "write_users",
            [
                Endpoint(
                        "POST /api/users",
                        "POST",
                        "/api/users",
                        lambda i: {
                            "name": f"Benchmark User {run_id}-{i + 1}",
                            "email": f"benchmark-user-{run_id}-{i + 1}-{time.time_ns()}@example.com",
                        },
                    )
            ],
        ),
        (
            "write_books",
            [
                Endpoint(
                        "POST /api/books",
                        "POST",
                        "/api/books",
                        lambda i: {
                        "title": f"Benchmark Book {run_id}-{i + 1}",
                        "author": "Codex Benchmark",
                        "isbn": f"9781{run_id[-6:]}{i + 1:04d}{time.time_ns() % 100:02d}",
                        "quantity": 3,
                    },
                    )
            ],
        ),
        (
            "write_loans",
            [
                Endpoint(
                    "POST /api/loans",
                    "POST",
                    "/api/loans",
                    lambda i: {
                        "user_id": seeded_users[i % len(seeded_users)]["id"],
                        "book_id": seeded_books[i % len(seeded_books)]["id"],
                    },
                )
            ],
        ),
    ]

    all_results: list[RequestResult] = []
    durations: dict[tuple[str, str], float] = {}
    stats_stop = threading.Event()
    stats_file = output_dir / "container_metrics.csv"
    stats_thread = threading.Thread(
        target=collect_docker_stats,
        args=(args.stack, compose_dir, stats_file, stats_stop, args.sample_interval),
        daemon=True,
    )
    stats_thread.start()

    started_at = utc_now()
    append_note(output_dir / "run_info.txt", f"stack={args.stack}")
    append_note(output_dir / "run_info.txt", f"base_url={base_url}")
    append_note(output_dir / "run_info.txt", f"started_at={started_at}")
    append_note(output_dir / "run_info.txt", f"seed_users={len(seeded_users)}")
    append_note(output_dir / "run_info.txt", f"seed_books={len(seeded_books)}")
    append_note(output_dir / "run_info.txt", f"seed_loans={len(seeded_loans)}")

    try:
        for scenario_name, endpoints in scenarios:
            scenario_results, elapsed = run_scenario(args.stack, scenario_name, base_url, endpoints, args.requests, args.concurrency)
            all_results.extend(scenario_results)
            durations[(args.stack, scenario_name)] = elapsed
    finally:
        stats_stop.set()
        stats_thread.join(timeout=5)

    finished_at = utc_now()
    append_note(output_dir / "run_info.txt", f"finished_at={finished_at}")

    request_csv = output_dir / "request_metrics.csv"
    summary_csv = output_dir / "summary.csv"
    write_request_csv(request_csv, all_results)
    write_summary_csv(summary_csv, all_results, durations)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
