name: pytest

on:
  push:
    branches:
      - main
  workflow_dispatch:

jobs:
  pytest:
    runs-on: monkci-ubuntu-24.04-4
    timeout-minutes: 30

    env:
      GB_VERSION: "6.7.1"

    steps:
      - name: Check vCPU
        run: |
          echo "vCPUs: $(nproc)"
          lscpu | grep -E 'Architecture|Model name|^CPU\(s\)|Thread\(s\) per core|Core\(s\) per socket|Socket\(s\)|Hypervisor vendor' || true
      - name: Check RAM
        run: |
          free -h
          TOTAL_KB=$(awk '/MemTotal/ {print $2}' /proc/meminfo)
          awk -v kb="$TOTAL_KB" 'BEGIN {
            printf "Total RAM: %.2f GiB\n", kb/1024/1024
          }'
      - name: Download Geekbench
        shell: bash
        run: |
          set -euo pipefail
          ARCH=$(uname -m)
          if [ "$ARCH" = "aarch64" ]; then
            PKG="Geekbench-${GB_VERSION}-LinuxARMPreview"
          else
            PKG="Geekbench-${GB_VERSION}-Linux"
          fi
          wget -q "https://cdn.geekbench.com/${PKG}.tar.gz"
          tar -xzf "${PKG}.tar.gz"
          GB_BIN=$(find "$PWD" -maxdepth 3 -type f -name geekbench6 -print -quit)
          if [ -z "$GB_BIN" ]; then
            echo "Error: Geekbench binary not found"
            exit 1
          fi
          chmod +x "$GB_BIN"
          echo "GB_BIN=$GB_BIN" >> "$GITHUB_ENV"
      - name: Run Geekbench CPU Benchmark
        shell: bash
        run: |
          set -o pipefail
          "$GB_BIN" --cpu 2>&1 | tee results.log
      - name: Upload results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: results
          path: results.log
          if-no-files-found: warn
