"""
make_report.py — generates HW10-Report.pdf using reportlab
"""

import os
from reportlab.lib.pagesizes import letter
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.lib.units import inch
from reportlab.lib import colors
from reportlab.platypus import (
    SimpleDocTemplate, Paragraph, Spacer, Image, PageBreak,
    HRFlowable, Table, TableStyle
)

OUT = os.path.join(os.path.dirname(__file__), "HW10-Report.pdf")
IMG_DIR = os.path.dirname(__file__)

doc = SimpleDocTemplate(
    OUT,
    pagesize=letter,
    leftMargin=1*inch,
    rightMargin=1*inch,
    topMargin=1*inch,
    bottomMargin=1*inch,
)

styles = getSampleStyleSheet()

title_style = ParagraphStyle(
    "ReportTitle",
    parent=styles["Title"],
    fontSize=18,
    spaceAfter=12,
    textColor=colors.HexColor("#1a237e"),
)
h1_style = ParagraphStyle(
    "H1",
    parent=styles["Heading1"],
    fontSize=14,
    spaceBefore=18,
    spaceAfter=6,
    textColor=colors.HexColor("#0d47a1"),
)
h2_style = ParagraphStyle(
    "H2",
    parent=styles["Heading2"],
    fontSize=12,
    spaceBefore=12,
    spaceAfter=4,
    textColor=colors.HexColor("#1565c0"),
)
body_style = ParagraphStyle(
    "Body",
    parent=styles["Normal"],
    fontSize=10,
    leading=15,
    spaceAfter=6,
)
caption_style = ParagraphStyle(
    "Caption",
    parent=styles["Italic"],
    fontSize=9,
    leading=12,
    spaceAfter=8,
    textColor=colors.HexColor("#555555"),
    alignment=1,  # center
)
bullet_style = ParagraphStyle(
    "Bullet",
    parent=styles["Normal"],
    fontSize=10,
    leading=15,
    leftIndent=20,
    spaceAfter=3,
    bulletIndent=10,
)

def heading1(text):
    return Paragraph(text, h1_style)

def heading2(text):
    return Paragraph(text, h2_style)

def body(text):
    return Paragraph(text, body_style)

def bullet(text):
    return Paragraph(f"\u2022  {text}", bullet_style)

def caption(text):
    return Paragraph(f"<i>{text}</i>", caption_style)

def img(filename, width=6.2*inch, max_height=3.2*inch):
    path = os.path.join(IMG_DIR, filename)
    if not os.path.exists(path):
        return body(f"[IMAGE NOT FOUND: {filename}]")
    from PIL import Image as PILImage
    with PILImage.open(path) as pil_im:
        orig_w, orig_h = pil_im.size
    aspect = orig_h / orig_w
    height = width * aspect
    if height > max_height:
        height = max_height
        width = height / aspect
    im = Image(path, width=width, height=height)
    im.hAlign = "CENTER"
    return im

def sp(n=1):
    return Spacer(1, n * 6)

story = []

# ---- Title ------------------------------------------------------------------
story.append(Spacer(1, 0.3*inch))
story.append(Paragraph("CS6650 Homework 10", title_style))
story.append(Paragraph("Distributed Databases using Replication", title_style))
story.append(HRFlowable(width="100%", thickness=2, color=colors.HexColor("#1a237e")))
story.append(sp(2))

# ---- Section 1 --------------------------------------------------------------
story.append(heading1("1. System Overview"))
story.append(body(
    "We implemented an in-memory Key-Value (KV) store distributed across 5 nodes using two "
    "architectures: <b>Leader-Follower</b> and <b>Leaderless</b>. The same Go binary serves "
    "all roles, configured at startup via environment variables (ROLE, W, R, FOLLOWERS/PEERS)."
))
story.append(sp())
story.append(body("The API exposes three endpoints:"))
story.append(bullet("POST /set {&quot;key&quot;:&quot;...&quot;,&quot;value&quot;:&quot;...&quot;} \u2192 201 Created (returns version number)"))
story.append(bullet("GET /get/{key} \u2192 200 {key, value, version} or 404"))
story.append(bullet("GET /local_read/{key} \u2192 same as /get but with no delays (test hook only)"))
story.append(sp())
story.append(body("Artificial delays simulate real-world storage latency:"))
story.append(bullet("Leader: sleeps 200 ms per goroutine before sending to each follower"))
story.append(bullet("Follower on write (/replicate): sleeps 100 ms"))
story.append(bullet("Follower on read (/get): sleeps 50 ms"))

# ---- Section 2 --------------------------------------------------------------
story.append(heading1("2. How the Code Works"))

story.append(heading2("2.1 Leader-Follower (Phase 2)"))
story.append(body(
    "All writes go to the leader. The leader assigns a version number using an atomic counter "
    "(<b>atomic.AddInt64</b>), writes locally, then fans out to all 4 followers via "
    "POST /replicate. Every follower gets a separate goroutine that sleeps 200 ms before "
    "sending (simulating network latency per the spec)."
))
story.append(sp())
story.append(body("<b>W</b> controls how many followers must confirm before returning 201:"))
story.append(bullet("W=5: block until all 4 followers respond (~300 ms total)"))
story.append(bullet("W=3: block until 2 followers respond, other 2 update in background"))
story.append(bullet("W=1: return immediately, all 4 followers update asynchronously"))
story.append(sp())
story.append(body("<b>R</b> controls how many nodes are consulted on reads:"))
story.append(bullet("R=1: leader reads its own store only"))
story.append(bullet("R=3: leader reads itself + 2 followers, returns the entry with the highest version"))
story.append(bullet("R=5: leader reads all 5 nodes, returns highest version"))
story.append(sp())
story.append(body(
    "Version numbers are critical for R&gt;1: each follower may have a different version of the "
    "same key during the replication window. The leader collects all responses and picks the one "
    "with the highest version, ensuring the most recent write is returned."
))
story.append(sp())
story.append(body(
    "The tricky part: <b>SetWithVersion</b> on followers only updates if the incoming version is "
    "strictly greater than what they already have. This prevents out-of-order replication messages "
    "from overwriting newer data."
))

story.append(heading2("2.2 Leaderless (Phase 3)"))
story.append(body(
    "Any node can receive a write. The receiving node becomes the <b>Write Coordinator</b>. "
    "It writes locally (using time.Now().UnixNano() as the version), then fans out to all "
    "4 peers via /replicate and waits for all to confirm (W=N=5)."
))
story.append(sp())
story.append(body("Reads are R=1: just return the local value with no coordination."))
story.append(sp())
story.append(body(
    "<b>Why timestamps instead of a counter?</b> In leader-follower, one leader assigns all "
    "versions so they are globally unique. In leaderless, two different coordinators could "
    "independently assign version 1 to different values. When the second coordinator tries to "
    "replicate its version-1 write to a node that already has version-1 data, SetWithVersion "
    "rejects it (1 &gt; 1 is false) and the write is silently lost. Using Unix nanosecond "
    "timestamps ensures each write gets a unique, globally comparable version. This is the "
    "<b>last-write-wins</b> approach used by Cassandra."
))

story.append(heading2("2.3 Error Handling"))
story.append(body(
    "The replication layer logs errors but does not fail the write if a follower is unreachable. "
    "For W&lt;5, the client still gets 201 once enough nodes confirm. The HTTP client has a "
    "5-second timeout per follower request. This is a simplification suitable for the "
    "assignment \u2014 production systems would use more sophisticated failure handling."
))

# ---- Section 3 --------------------------------------------------------------
story.append(heading1("3. Load Test Generator Design"))
story.append(body(
    "The load tester spawns N concurrent goroutines, each running for a configurable duration. "
    "Each iteration:"
))
story.append(bullet("Picks a random key from a small key space (100 keys by default)"))
story.append(bullet("With probability write_pct%, sends a write to the leader (or random node for leaderless)"))
story.append(bullet("Otherwise sends a read to a random node from the full node list"))
story.append(sp())
story.append(body(
    "<b>Local-in-time clustering:</b> by using only 100 keys with 10 workers, the same key is "
    "written and read repeatedly within milliseconds. This creates the temporal clustering needed "
    "to hit the replication window and observe stale reads."
))
story.append(sp())
story.append(body(
    "<b>Stale read detection:</b> the client maintains a map[key \u2192 latestWrittenVersion]. "
    "When a write completes, the returned version is stored. When a read returns, if the returned "
    "version is less than the client\u2019s known version for that key, it is counted as stale."
))
story.append(sp())
story.append(body(
    "The <b>rw_interval_ms</b> field in the CSV records how many milliseconds have passed since "
    "the last write to that key when a read occurs \u2014 directly showing the temporal clustering."
))

# ---- Section 4 --------------------------------------------------------------
story.append(heading1("4. Results and Analysis"))

story.append(heading2("4.1 Write Latency"))
story.append(img("write_latency_by_ratio.png"))
story.append(caption("Figure 1 \u2014 Write latency distribution by mode and write ratio"))
story.append(sp())
story.append(body(
    "W=5 R=1, W=3 R=3, and Leaderless all show ~295\u2013300 ms write latency regardless of "
    "write ratio. This is because all three modes wait for follower confirmation, and the "
    "dominant cost is the 200 ms sleep + 100 ms follower processing. The write ratio does not "
    "affect per-write latency since each write is independent."
))
story.append(sp())
story.append(body(
    "W=1 R=5 shows dramatically lower write latency (~1\u20135 ms) because the leader returns "
    "immediately after writing locally. At 90% writes with high concurrency, P99 rises to "
    "~46 ms due to queuing at the leader, but is still 60x faster than synchronous modes."
))
story.append(sp())
story.append(img("latency_cdf_writes.png"))
story.append(caption("Figure 2 \u2014 Write latency CDF across all write ratios"))
story.append(sp())
story.append(body(
    "The CDF clearly shows two clusters: W=1 R=5 finishes nearly all writes under 10 ms, "
    "while all other modes cluster tightly around 295\u2013310 ms."
))

story.append(heading2("4.2 Read Latency"))
story.append(img("read_latency_by_ratio.png"))
story.append(caption("Figure 3 \u2014 Read latency distribution by mode and write ratio"))
story.append(sp())
story.append(body(
    "All modes show similar read latency (~50\u201355 ms avg). This makes sense: reads always "
    "hit a single node which sleeps 50 ms (follower) or returns immediately from local store "
    "(leader R=1). The slight spread in W=5 R=1 at low write ratios is because reads mostly "
    "go to followers (which sleep 50 ms) but occasionally hit the leader (which has no sleep "
    "for R=1)."
))
story.append(sp())
story.append(body(
    "Under heavy write load (90% writes, W=1 R=5), read latency increases slightly "
    "(P99 ~93 ms) because the follower nodes are busy processing replication messages, "
    "adding queuing delay."
))
story.append(sp())
story.append(img("latency_cdf_reads.png"))
story.append(caption("Figure 4 \u2014 Read latency CDF across all write ratios"))
story.append(sp())
story.append(body(
    "The CDF shows all modes have nearly identical read latency profiles. There is a slight "
    "long tail in W=1 R=5 at high write ratios, visible as the CDF reaching 1.0 at higher "
    "latency values \u2014 this is the queuing effect from replication traffic competing "
    "with client reads."
))

story.append(heading2("4.3 Stale Reads"))
story.append(img("stale_reads.png"))
story.append(caption("Figure 5 \u2014 Stale read percentage by mode and write ratio"))
story.append(sp())
story.append(body(
    "Only W=1 R=5 produces stale reads. The stale rate increases dramatically with "
    "write percentage:"
))
story.append(bullet("1% writes: 0% stale (very few writes, most nodes already up to date)"))
story.append(bullet("10% writes: 0.79% stale"))
story.append(bullet("50% writes: 9.84% stale"))
story.append(bullet("90% writes: 55.47% stale"))
story.append(sp())
story.append(body(
    "This is the expected behavior. With more writes, there are more keys in the replication "
    "window at any moment. At 90% writes, the cluster is continuously replicating \u2014 a "
    "read to a random follower has a high probability of hitting a key that was just written "
    "and not yet propagated."
))
story.append(sp())
story.append(body("W=5 R=1, W=3 R=3, and Leaderless all show <b>0% stale reads</b>:"))
story.append(bullet("W=5 R=1: every node confirmed before 201, so all nodes have the latest value"))
story.append(bullet(
    "W=3 R=3: the W+R &gt; N quorum guarantee ensures the read set always overlaps with the "
    "write set by at least 1 node, and max-version selection returns the latest value"
))
story.append(bullet("Leaderless (W=N): same as W=5 \u2014 all nodes confirmed before 201"))

story.append(heading2("4.4 Read-Write Interval Distribution"))
story.append(img("rw_interval.png"))
story.append(caption("Figure 6 \u2014 Time between last write and subsequent read on the same key"))
story.append(sp())
story.append(body(
    "The histograms show that reads and writes on the same key are clustered within "
    "milliseconds of each other. The median interval is well under 1000 ms for all modes, "
    "confirming that the small key space (100 keys) successfully creates local-in-time "
    "clustering. This is what makes the stale read results meaningful \u2014 reads are "
    "happening while the replication window is still open."
))

# ---- Section 5 --------------------------------------------------------------
story.append(PageBreak())
story.append(heading1("5. Discussion"))

story.append(heading2("5.1 Which mode performs best for each workload?"))

story.append(body("<b>1% writes / 99% reads (read-heavy):</b>"))
story.append(body(
    "W=1 R=5 is the best choice. Writes are rare so the stale read rate is effectively 0%, "
    "and write latency of ~2 ms means the occasional write has no impact on throughput. The "
    "system delivers high read throughput with fast (non-blocking) writes."
))
story.append(sp())

story.append(body("<b>10% writes / 90% reads:</b>"))
story.append(body(
    "W=1 R=5 still performs best for raw throughput \u2014 writes are 150x faster than "
    "synchronous modes. The 0.79% stale rate is acceptable for many applications (social media "
    "feeds, analytics dashboards, recommendation systems)."
))
story.append(sp())

story.append(body("<b>50% writes / 50% reads:</b>"))
story.append(body(
    "This is the difficult case. W=1 R=5 has 9.84% stale reads \u2014 unacceptable for many "
    "applications. W=3 R=3 provides the best balance: zero stale reads with quorum guarantees, "
    "while tolerating one node failure. W=5 R=1 also works but provides no fault tolerance "
    "benefit over W=3 R=3."
))
story.append(sp())

story.append(body("<b>90% writes / 10% reads:</b>"))
story.append(body(
    "W=1 R=5 is impractical (55.47% stale reads). W=3 R=3 or W=5 R=1 should be used. For "
    "write-heavy workloads, the synchronous write cost (~300 ms) is the bottleneck regardless "
    "of mode, so W=3 R=3 is preferred for its fault tolerance."
))

story.append(heading2("5.2 Which database for which application?"))

data = [
    ["Mode", "Best Use Case"],
    ["W=5 R=1",
     "When reads must always be served quickly from a single authoritative node and writes\n"
     "can be slow. Example: configuration management systems where the leader serves all reads."],
    ["W=1 R=5",
     "When writes must be fast and occasional stale reads are acceptable. Example: social\n"
     "media like counts, page view counters, analytics events."],
    ["W=3 R=3\n(Quorum)",
     "The general-purpose choice. Strong consistency without requiring all nodes to be\n"
     "available. Example: shopping carts, user sessions, inventory management."],
    ["Leaderless\n(W=N, R=1)",
     "No single point of failure; any node can accept writes. Example: IoT sensor data\n"
     "collection where reads after confirmation are always consistent."],
]

t = Table(data, colWidths=[1.4*inch, 5.0*inch])
t.setStyle(TableStyle([
    ("BACKGROUND",   (0, 0), (-1, 0), colors.HexColor("#1565c0")),
    ("TEXTCOLOR",    (0, 0), (-1, 0), colors.white),
    ("FONTNAME",     (0, 0), (-1, 0), "Helvetica-Bold"),
    ("FONTSIZE",     (0, 0), (-1, 0), 10),
    ("BOTTOMPADDING",(0, 0), (-1, 0), 8),
    ("TOPPADDING",   (0, 0), (-1, 0), 8),
    ("BACKGROUND",   (0, 1), (-1, -1), colors.HexColor("#f5f5f5")),
    ("ROWBACKGROUNDS",(0, 1), (-1, -1), [colors.white, colors.HexColor("#e3f2fd")]),
    ("FONTSIZE",     (0, 1), (-1, -1), 9),
    ("LEADING",      (0, 1), (-1, -1), 13),
    ("TOPPADDING",   (0, 1), (-1, -1), 6),
    ("BOTTOMPADDING",(0, 1), (-1, -1), 6),
    ("GRID",         (0, 0), (-1, -1), 0.5, colors.HexColor("#bbbbbb")),
    ("VALIGN",       (0, 0), (-1, -1), "TOP"),
    ("FONTNAME",     (0, 1), (0, -1), "Helvetica-Bold"),
]))
story.append(sp(2))
story.append(t)

doc.build(story)
print(f"Report written to: {OUT}")
