"""
generate_graphs.py
Reads all 16 load-test CSV files and produces the graphs required for the HW-10 report.

Graphs produced:
  1. read_latency_by_ratio.png   — read latency box plots per write ratio, all 4 modes
  2. write_latency_by_ratio.png  — write latency box plots per write ratio, all 4 modes
  3. stale_reads.png             — stale read % bar chart, all modes × ratios
  4. rw_interval.png             — distribution of time between write and subsequent read on same key
  5. latency_cdf_reads.png       — CDF of read latency per mode (shows long tail)
  6. latency_cdf_writes.png      — CDF of write latency per mode (shows long tail)
"""

import os
import glob
import pandas as pd
import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt
import matplotlib.patches as mpatches
import numpy as np

# ---- paths ------------------------------------------------------------------
RESULTS_DIR = os.path.join(os.path.dirname(__file__),
                           "../load-tester/results-try2")
OUT_DIR = os.path.dirname(__file__)

MODES    = ["w5r1", "w1r5", "w3r3", "leaderless"]
RATIOS   = [1, 10, 50, 90]          # write percentages
MODE_LABELS = {
    "w5r1":       "W=5 R=1",
    "w1r5":       "W=1 R=5",
    "w3r3":       "W=3 R=3",
    "leaderless": "Leaderless",
}
COLORS = {
    "w5r1":       "#2196F3",
    "w1r5":       "#F44336",
    "w3r3":       "#4CAF50",
    "leaderless": "#FF9800",
}

# ---- load data --------------------------------------------------------------
def load_all():
    frames = []
    for mode in MODES:
        for pct in RATIOS:
            path = os.path.join(RESULTS_DIR, f"{mode}_{pct}w.csv")
            if not os.path.exists(path):
                print(f"  MISSING: {path}")
                continue
            df = pd.read_csv(path)
            df["write_pct"] = pct
            frames.append(df)
    return pd.concat(frames, ignore_index=True)

df = load_all()
print(f"Loaded {len(df):,} records across {df['mode'].nunique()} modes")

# ---- helper -----------------------------------------------------------------
def save(fig, name):
    path = os.path.join(OUT_DIR, name)
    fig.savefig(path, dpi=150, bbox_inches="tight")
    plt.close(fig)
    print(f"  saved {name}")

# ---- 1. Read latency box plots per write ratio ------------------------------
fig, axes = plt.subplots(1, 4, figsize=(18, 5), sharey=True)
fig.suptitle("Read Latency Distribution by Write Ratio (ms)", fontsize=14, fontweight="bold")

for ax, pct in zip(axes, RATIOS):
    data   = [df[(df["write_pct"] == pct) & (df["mode"] == m) & (df["type"] == "read")]["latency_ms"].dropna()
              for m in MODES]
    labels = [MODE_LABELS[m] for m in MODES]
    bp = ax.boxplot(data, labels=labels, patch_artist=True, showfliers=True,
                    flierprops=dict(marker=".", markersize=2, alpha=0.3))
    for patch, m in zip(bp["boxes"], MODES):
        patch.set_facecolor(COLORS[m])
    ax.set_title(f"Writes {pct}% / Reads {100-pct}%")
    ax.set_xlabel("Mode")
    ax.tick_params(axis="x", labelsize=8)

axes[0].set_ylabel("Latency (ms)")
fig.tight_layout()
save(fig, "read_latency_by_ratio.png")

# ---- 2. Write latency box plots per write ratio -----------------------------
fig, axes = plt.subplots(1, 4, figsize=(18, 5), sharey=True)
fig.suptitle("Write Latency Distribution by Write Ratio (ms)", fontsize=14, fontweight="bold")

for ax, pct in zip(axes, RATIOS):
    data   = [df[(df["write_pct"] == pct) & (df["mode"] == m) & (df["type"] == "write")]["latency_ms"].dropna()
              for m in MODES]
    labels = [MODE_LABELS[m] for m in MODES]
    bp = ax.boxplot(data, labels=labels, patch_artist=True, showfliers=True,
                    flierprops=dict(marker=".", markersize=2, alpha=0.3))
    for patch, m in zip(bp["boxes"], MODES):
        patch.set_facecolor(COLORS[m])
    ax.set_title(f"Writes {pct}% / Reads {100-pct}%")
    ax.set_xlabel("Mode")
    ax.tick_params(axis="x", labelsize=8)

axes[0].set_ylabel("Latency (ms)")
fig.tight_layout()
save(fig, "write_latency_by_ratio.png")

# ---- 3. Stale read % bar chart ----------------------------------------------
fig, ax = plt.subplots(figsize=(12, 5))
x      = np.arange(len(RATIOS))
width  = 0.2

for i, mode in enumerate(MODES):
    stale_pcts = []
    for pct in RATIOS:
        sub = df[(df["write_pct"] == pct) & (df["mode"] == mode) & (df["type"] == "read")]
        if len(sub) == 0:
            stale_pcts.append(0)
        else:
            stale_pcts.append((sub["stale"] == True).sum() / len(sub) * 100)
    ax.bar(x + i * width, stale_pcts, width, label=MODE_LABELS[mode],
           color=COLORS[mode], alpha=0.85)

ax.set_xlabel("Write Percentage")
ax.set_ylabel("Stale Reads (%)")
ax.set_title("Stale Read Rate by Mode and Write Ratio", fontsize=13, fontweight="bold")
ax.set_xticks(x + width * 1.5)
ax.set_xticklabels([f"{p}% writes" for p in RATIOS])
ax.legend()
ax.grid(axis="y", alpha=0.3)
fig.tight_layout()
save(fig, "stale_reads.png")

# ---- 4. R/W interval distribution ------------------------------------------
# rw_interval_ms = time since last write to this key when a read occurred.
# Shows that reads and writes cluster closely in time (local-in-time property).
fig, axes = plt.subplots(1, 4, figsize=(18, 5), sharey=False)
fig.suptitle("Time Between Last Write and Read of Same Key (ms)\n"
             "— proves local-in-time clustering of the load generator",
             fontsize=12, fontweight="bold")

for ax, mode in zip(axes, MODES):
    intervals = df[(df["mode"] == mode) & (df["type"] == "read") &
                   (df["rw_interval_ms"] >= 0)]["rw_interval_ms"]
    if len(intervals) == 0:
        ax.text(0.5, 0.5, "no data", ha="center", va="center", transform=ax.transAxes)
    else:
        ax.hist(intervals, bins=50, color=COLORS[mode], alpha=0.8, edgecolor="none")
        ax.axvline(intervals.median(), color="black", linewidth=1.2,
                   linestyle="--", label=f"median={intervals.median():.0f}ms")
        ax.legend(fontsize=8)
    ax.set_title(MODE_LABELS[mode])
    ax.set_xlabel("Interval (ms)")
    ax.set_ylabel("Count")

fig.tight_layout()
save(fig, "rw_interval.png")

# ---- 5. CDF of read latency per mode ----------------------------------------
fig, ax = plt.subplots(figsize=(9, 5))
for mode in MODES:
    lat = df[(df["mode"] == mode) & (df["type"] == "read")]["latency_ms"].sort_values()
    cdf = np.arange(1, len(lat) + 1) / len(lat)
    ax.plot(lat.values, cdf, label=MODE_LABELS[mode], color=COLORS[mode], linewidth=1.8)

ax.set_xlabel("Latency (ms)")
ax.set_ylabel("CDF")
ax.set_title("Read Latency CDF — all write ratios combined", fontsize=12, fontweight="bold")
ax.legend()
ax.grid(alpha=0.3)
ax.set_xlim(left=0)
fig.tight_layout()
save(fig, "latency_cdf_reads.png")

# ---- 6. CDF of write latency per mode ---------------------------------------
fig, ax = plt.subplots(figsize=(9, 5))
for mode in MODES:
    lat = df[(df["mode"] == mode) & (df["type"] == "write")]["latency_ms"].sort_values()
    if len(lat) == 0:
        continue
    cdf = np.arange(1, len(lat) + 1) / len(lat)
    ax.plot(lat.values, cdf, label=MODE_LABELS[mode], color=COLORS[mode], linewidth=1.8)

ax.set_xlabel("Latency (ms)")
ax.set_ylabel("CDF")
ax.set_title("Write Latency CDF — all write ratios combined", fontsize=12, fontweight="bold")
ax.legend()
ax.grid(alpha=0.3)
ax.set_xlim(left=0)
fig.tight_layout()
save(fig, "latency_cdf_writes.png")

print("\nAll graphs generated.")
