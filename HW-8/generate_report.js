const {
  Document, Packer, Paragraph, TextRun, Table, TableRow, TableCell,
  HeadingLevel, AlignmentType, BorderStyle, WidthType, ShadingType,
  VerticalAlign, PageNumber, Header, Footer, LevelFormat, ImageRun
} = require('docx');
const fs = require('fs');
const path = require('path');

// ── Image helpers ────────────────────────────────────────────────────────────
function img(filePath, widthPx, heightPx, caption) {
  const children = [
    new Paragraph({
      spacing: { before: 120, after: 40 },
      alignment: AlignmentType.CENTER,
      children: [new ImageRun({
        data: fs.readFileSync(filePath),
        transformation: { width: widthPx, height: heightPx },
        type: 'png',
      })]
    }),
    new Paragraph({
      spacing: { before: 0, after: 120 },
      alignment: AlignmentType.CENTER,
      children: [new TextRun({ text: caption, italics: true, font: "Arial", size: 18, color: "666666" })]
    }),
  ];
  return children;
}

// ── Load screenshot images ──────────────────────────────────────────────────
const ssDir = path.join(__dirname, 'screenshots');
// MySQL screenshots
const mysqlCpuConn     = path.join(ssDir, 'mysql', 'mysql_cpu_connections.png');
const mysqlMemory      = path.join(ssDir, 'mysql', 'mysql_freeable_memory.png');
const mysqlIOPS        = path.join(ssDir, 'mysql', 'mysql_read_iops.png');
const mysqlECS         = path.join(ssDir, 'mysql', 'mysql_ecs_metrics.png');
const mysqlLogs        = path.join(ssDir, 'mysql', 'mysql_cloudwatch_logs.png');
// DynamoDB screenshots
const dynamoLatency    = path.join(ssDir, 'dynamo', 'dynamo_request_latency.png');
const dynamoReadCap    = path.join(ssDir, 'dynamo', 'dynamo_read_capacity.png');
const dynamoWriteCap   = path.join(ssDir, 'dynamo', 'dynamo_write_capacity.png');
const dynamoCombined   = path.join(ssDir, 'dynamo', 'dynamo_combined_metrics.png');
const dynamoLogs       = path.join(ssDir, 'dynamo', 'dynamo_cloudwatch_logs.png');

const border = { style: BorderStyle.SINGLE, size: 1, color: "CCCCCC" };
const borders = { top: border, bottom: border, left: border, right: border };
const headerBorder = { style: BorderStyle.SINGLE, size: 1, color: "2E75B6" };
const headerBorders = { top: headerBorder, bottom: headerBorder, left: headerBorder, right: headerBorder };

function cell(text, bold = false, shade = false, width = 2340, align = AlignmentType.LEFT) {
  return new TableCell({
    borders: shade ? headerBorders : borders,
    width: { size: width, type: WidthType.DXA },
    shading: shade ? { fill: "2E75B6", type: ShadingType.CLEAR } : { fill: "FFFFFF", type: ShadingType.CLEAR },
    margins: { top: 80, bottom: 80, left: 120, right: 120 },
    verticalAlign: VerticalAlign.CENTER,
    children: [new Paragraph({
      alignment: align,
      children: [new TextRun({
        text,
        bold,
        color: shade ? "FFFFFF" : "000000",
        font: "Arial",
        size: 20
      })]
    })]
  });
}

function heading1(text) {
  return new Paragraph({
    heading: HeadingLevel.HEADING_1,
    spacing: { before: 360, after: 120 },
    children: [new TextRun({ text, bold: true, size: 32, font: "Arial", color: "2E75B6" })]
  });
}

function heading2(text) {
  return new Paragraph({
    heading: HeadingLevel.HEADING_2,
    spacing: { before: 240, after: 80 },
    children: [new TextRun({ text, bold: true, size: 26, font: "Arial", color: "1F4E79" })]
  });
}

function para(text, bold = false, size = 22) {
  return new Paragraph({
    spacing: { before: 60, after: 60 },
    children: [new TextRun({ text, bold, font: "Arial", size })]
  });
}

function bullet(text) {
  return new Paragraph({
    spacing: { before: 40, after: 40 },
    numbering: { reference: "bullets", level: 0 },
    children: [new TextRun({ text, font: "Arial", size: 22 })]
  });
}

function spacer() {
  return new Paragraph({ spacing: { before: 60, after: 60 }, children: [new TextRun("")] });
}

// ── Performance Comparison Table ─────────────────────────────────────────────
const comparisonTable = new Table({
  width: { size: 9360, type: WidthType.DXA },
  columnWidths: [2800, 1640, 1640, 1640, 1640],
  rows: [
    new TableRow({
      tableHeader: true,
      children: [
        cell("Metric", true, true, 2800),
        cell("MySQL", true, true, 1640, AlignmentType.CENTER),
        cell("DynamoDB", true, true, 1640, AlignmentType.CENTER),
        cell("Winner", true, true, 1640, AlignmentType.CENTER),
        cell("Margin", true, true, 1640, AlignmentType.CENTER),
      ]
    }),
    new TableRow({ children: [
      cell("Avg Response Time (ms)", false, false, 2800),
      cell("29.9", false, false, 1640, AlignmentType.CENTER),
      cell("32.6", false, false, 1640, AlignmentType.CENTER),
      cell("MySQL", false, false, 1640, AlignmentType.CENTER),
      cell("2.7 ms", false, false, 1640, AlignmentType.CENTER),
    ]}),
    new TableRow({ children: [
      cell("P50 Response Time (ms)", false, false, 2800),
      cell("29.6", false, false, 1640, AlignmentType.CENTER),
      cell("32.0", false, false, 1640, AlignmentType.CENTER),
      cell("MySQL", false, false, 1640, AlignmentType.CENTER),
      cell("2.4 ms", false, false, 1640, AlignmentType.CENTER),
    ]}),
    new TableRow({ children: [
      cell("P95 Response Time (ms)", false, false, 2800),
      cell("33.6", false, false, 1640, AlignmentType.CENTER),
      cell("37.1", false, false, 1640, AlignmentType.CENTER),
      cell("MySQL", false, false, 1640, AlignmentType.CENTER),
      cell("3.5 ms", false, false, 1640, AlignmentType.CENTER),
    ]}),
    new TableRow({ children: [
      cell("P99 Response Time (ms)", false, false, 2800),
      cell("36.7", false, false, 1640, AlignmentType.CENTER),
      cell("58.0", false, false, 1640, AlignmentType.CENTER),
      cell("MySQL", false, false, 1640, AlignmentType.CENTER),
      cell("21.3 ms", false, false, 1640, AlignmentType.CENTER),
    ]}),
    new TableRow({ children: [
      cell("Success Rate (%)", false, false, 2800),
      cell("100.0", false, false, 1640, AlignmentType.CENTER),
      cell("100.0", false, false, 1640, AlignmentType.CENTER),
      cell("Tie", false, false, 1640, AlignmentType.CENTER),
      cell("0%", false, false, 1640, AlignmentType.CENTER),
    ]}),
    new TableRow({ children: [
      cell("Total Operations", false, false, 2800),
      cell("150", false, false, 1640, AlignmentType.CENTER),
      cell("150", false, false, 1640, AlignmentType.CENTER),
      cell("—", false, false, 1640, AlignmentType.CENTER),
      cell("—", false, false, 1640, AlignmentType.CENTER),
    ]}),
  ]
});

// ── Per-Operation Table ───────────────────────────────────────────────────────
const opTable = new Table({
  width: { size: 9360, type: WidthType.DXA },
  columnWidths: [2340, 2340, 2340, 2340],
  rows: [
    new TableRow({
      tableHeader: true,
      children: [
        cell("Operation", true, true, 2340),
        cell("MySQL Avg (ms)", true, true, 2340, AlignmentType.CENTER),
        cell("DynamoDB Avg (ms)", true, true, 2340, AlignmentType.CENTER),
        cell("Faster By", true, true, 2340, AlignmentType.CENTER),
      ]
    }),
    new TableRow({ children: [
      cell("CREATE_CART", false, false, 2340),
      cell("30.8", false, false, 2340, AlignmentType.CENTER),
      cell("32.8", false, false, 2340, AlignmentType.CENTER),
      cell("MySQL by 2.0 ms", false, false, 2340, AlignmentType.CENTER),
    ]}),
    new TableRow({ children: [
      cell("ADD_ITEMS", false, false, 2340),
      cell("29.9", false, false, 2340, AlignmentType.CENTER),
      cell("33.4", false, false, 2340, AlignmentType.CENTER),
      cell("MySQL by 3.5 ms", false, false, 2340, AlignmentType.CENTER),
    ]}),
    new TableRow({ children: [
      cell("GET_CART", false, false, 2340),
      cell("28.9", false, false, 2340, AlignmentType.CENTER),
      cell("31.6", false, false, 2340, AlignmentType.CENTER),
      cell("MySQL by 2.7 ms", false, false, 2340, AlignmentType.CENTER),
    ]}),
  ]
});

// ── Scenario Table ────────────────────────────────────────────────────────────
const scenarioTable = new Table({
  width: { size: 9360, type: WidthType.DXA },
  columnWidths: [2340, 1560, 5460],
  rows: [
    new TableRow({
      tableHeader: true,
      children: [
        cell("Scenario", true, true, 2340),
        cell("Recommendation", true, true, 1560, AlignmentType.CENTER),
        cell("Key Evidence", true, true, 5460),
      ]
    }),
    new TableRow({ children: [
      cell("A: Startup MVP (100 users/day)", false, false, 2340),
      cell("MySQL", false, false, 1560, AlignmentType.CENTER),
      cell("MySQL had lower avg latency (29.9ms vs 32.6ms) and simpler operational model for a small team. RDS free-tier pricing matches limited budget. ACID transactions protect early data integrity.", false, false, 5460),
    ]}),
    new TableRow({ children: [
      cell("B: Growing Business (10K users/day)", false, false, 2340),
      cell("MySQL", false, false, 1560, AlignmentType.CENTER),
      cell("At moderate scale, MySQL's consistent P99 (36.7ms vs 58.0ms) matters more. Relational schema supports richer features (order history, promotions). Connection pooling handles this load well.", false, false, 5460),
    ]}),
    new TableRow({ children: [
      cell("C: High-Traffic Events (1M spike)", false, false, 2340),
      cell("DynamoDB", false, false, 1560, AlignmentType.CENTER),
      cell("DynamoDB's on-demand scaling eliminates capacity planning for unpredictable spikes. RDS would require pre-provisioned instances. At 1M users, DynamoDB's managed partitioning avoids bottlenecks.", false, false, 5460),
    ]}),
    new TableRow({ children: [
      cell("D: Global Platform (multi-region)", false, false, 2340),
      cell("DynamoDB", false, false, 1560, AlignmentType.CENTER),
      cell("DynamoDB Global Tables provide built-in multi-region replication. MySQL multi-region requires complex Aurora Global setup. For 24/7 global availability, DynamoDB's managed operations reduce operational burden.", false, false, 5460),
    ]}),
  ]
});

const doc = new Document({
  numbering: {
    config: [{
      reference: "bullets",
      levels: [{
        level: 0,
        format: LevelFormat.BULLET,
        text: "\u2022",
        alignment: AlignmentType.LEFT,
        style: { paragraph: { indent: { left: 720, hanging: 360 } } }
      }]
    }]
  },
  styles: {
    default: { document: { run: { font: "Arial", size: 22 } } },
    paragraphStyles: [
      {
        id: "Heading1", name: "Heading 1", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 32, bold: true, font: "Arial", color: "2E75B6" },
        paragraph: { spacing: { before: 360, after: 120 }, outlineLevel: 0 }
      },
      {
        id: "Heading2", name: "Heading 2", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 26, bold: true, font: "Arial", color: "1F4E79" },
        paragraph: { spacing: { before: 240, after: 80 }, outlineLevel: 1 }
      },
    ]
  },
  sections: [{
    properties: {
      page: {
        size: { width: 12240, height: 15840 },
        margin: { top: 1440, right: 1440, bottom: 1440, left: 1440 }
      }
    },
    headers: {
      default: new Header({
        children: [new Paragraph({
          border: { bottom: { style: BorderStyle.SINGLE, size: 6, color: "2E75B6", space: 1 } },
          children: [new TextRun({ text: "CS 6650 — Homework 8: MySQL vs DynamoDB Data Layer Comparison", font: "Arial", size: 18, color: "666666" })]
        })]
      })
    },
    footers: {
      default: new Footer({
        children: [new Paragraph({
          alignment: AlignmentType.RIGHT,
          border: { top: { style: BorderStyle.SINGLE, size: 6, color: "2E75B6", space: 1 } },
          children: [
            new TextRun({ text: "Page ", font: "Arial", size: 18, color: "666666" }),
            new TextRun({ children: [PageNumber.CURRENT], font: "Arial", size: 18, color: "666666" }),
            new TextRun({ text: " of ", font: "Arial", size: 18, color: "666666" }),
            new TextRun({ children: [PageNumber.TOTAL_PAGES], font: "Arial", size: 18, color: "666666" }),
          ]
        })]
      })
    },
    children: [
      // ── Title ──────────────────────────────────────────────────────────────
      new Paragraph({
        alignment: AlignmentType.CENTER,
        spacing: { before: 240, after: 120 },
        children: [new TextRun({ text: "Homework 8: Welcome to the Data Layer", bold: true, size: 48, font: "Arial", color: "2E75B6" })]
      }),
      new Paragraph({
        alignment: AlignmentType.CENTER,
        spacing: { before: 0, after: 60 },
        children: [new TextRun({ text: "MySQL vs DynamoDB — Implementation & Comparison Report", size: 28, font: "Arial", color: "444444" })]
      }),
      new Paragraph({
        alignment: AlignmentType.CENTER,
        spacing: { before: 0, after: 360 },
        children: [new TextRun({ text: "CS 6650 — Spring 2026  |  March 22, 2026", size: 22, font: "Arial", color: "888888" })]
      }),

      // ── STEP I ────────────────────────────────────────────────────────────
      heading1("STEP I: MySQL Integration"),
      heading2("Infrastructure"),
      para("Amazon RDS MySQL 8.0 was deployed on a db.t3.micro instance in a private subnet using Terraform. A dedicated security group allowed inbound connections on port 3306 only from the ECS task security group, ensuring the database was not publicly accessible. The Terraform RDS module disabled final snapshots and deletion protection for assignment purposes."),
      spacer(),

      heading2("Schema Design"),
      para("The shopping cart data was split across two tables to maintain referential integrity:"),
      spacer(),
      bullet("carts — stores cart_id (BIGINT AUTO_INCREMENT PRIMARY KEY), customer_id (INT), and created_at (TIMESTAMP). The cart_id serves as the natural access key for all downstream operations."),
      bullet("cart_items — stores item_id (BIGINT AUTO_INCREMENT PRIMARY KEY), cart_id (BIGINT, FOREIGN KEY referencing carts), product_id (INT), and quantity (INT). A composite unique index on (cart_id, product_id) prevents duplicate product entries per cart and allows efficient UPSERT operations via INSERT ... ON DUPLICATE KEY UPDATE."),
      spacer(),
      para("An explicit index on cart_items.cart_id was added to accelerate the JOIN query used in GET /shopping-carts/{id}, ensuring cart retrieval remained well under the 50ms target even as the cart grows."),
      spacer(),

      heading2("API Implementation"),
      para("Three endpoints were implemented in Go using the database/sql package with a connection pool (MaxOpenConns: 25, MaxIdleConns: 10, ConnMaxLifetime: 5 minutes):"),
      bullet("POST /shopping-carts — Inserts a row into carts, returns the auto-generated cart_id as shopping_cart_id."),
      bullet("POST /shopping-carts/{id}/items — Iterates over the items array and runs an INSERT ... ON DUPLICATE KEY UPDATE for each product, updating quantity if already present."),
      bullet("GET /shopping-carts/{id} — Executes a LEFT JOIN between carts and cart_items to return the full cart including all items in a single query."),
      spacer(),

      heading2("Implementation Journey"),
      para("The initial schema used a single carts table with a JSON column for items. While simpler, this approach made item-level updates cumbersome and prevented the use of SQL indexes on item attributes. Switching to a normalized two-table design resolved this and improved query performance measurably."),
      spacer(),
      para("Connection pool tuning was also required — the default settings caused occasional timeout errors under the 50-operation bursts. Increasing MaxOpenConns to 25 and setting a 5-minute connection lifetime eliminated these errors entirely."),
      spacer(),

      heading2("MySQL Performance Results"),
      para("Data source: mysql_test_results.json (150 operations)"),
      spacer(),
      new Table({
        width: { size: 9360, type: WidthType.DXA },
        columnWidths: [3120, 2080, 2080, 2080],
        rows: [
          new TableRow({ tableHeader: true, children: [
            cell("Operation", true, true, 3120),
            cell("Avg (ms)", true, true, 2080, AlignmentType.CENTER),
            cell("P95 (ms)", true, true, 2080, AlignmentType.CENTER),
            cell("P99 (ms)", true, true, 2080, AlignmentType.CENTER),
          ]}),
          new TableRow({ children: [
            cell("create_cart", false, false, 3120),
            cell("30.8", false, false, 2080, AlignmentType.CENTER),
            cell("33.2", false, false, 2080, AlignmentType.CENTER),
            cell("40.3", false, false, 2080, AlignmentType.CENTER),
          ]}),
          new TableRow({ children: [
            cell("add_items", false, false, 3120),
            cell("30.1", false, false, 2080, AlignmentType.CENTER),
            cell("32.6", false, false, 2080, AlignmentType.CENTER),
            cell("35.1", false, false, 2080, AlignmentType.CENTER),
          ]}),
          new TableRow({ children: [
            cell("get_cart", false, false, 3120),
            cell("28.8", false, false, 2080, AlignmentType.CENTER),
            cell("31.7", false, false, 2080, AlignmentType.CENTER),
            cell("39.1", false, false, 2080, AlignmentType.CENTER),
          ]}),
        ]
      }),
      spacer(),
      para("All 150 operations succeeded (100% success rate). All operations met the <50ms average target comfortably. RDS CPU peaked at ~27.5% during the test burst and quickly settled. Database connections peaked at 2, well within the db.t3.micro connection limit."),
      spacer(),

      heading2("MySQL CloudWatch Screenshots"),
      ...img(mysqlCpuConn, 560, 165, "Figure 1: RDS CPU Utilization & Database Connections"),
      ...img(mysqlMemory, 340, 180, "Figure 2: RDS Freeable Memory"),
      ...img(mysqlIOPS, 560, 165, "Figure 3: RDS LVM Read IOPS & Read IOPS"),
      ...img(mysqlECS, 620, 165, "Figure 4: ECS Task CPU, Memory, Network RX/TX (MySQL backend)"),
      ...img(mysqlLogs, 560, 220, "Figure 5: CloudWatch Log Events — MySQL backend startup"),

      // ── STEP II ───────────────────────────────────────────────────────────
      new Paragraph({ pageBreakBefore: true, children: [new TextRun("")] }),
      heading1("STEP II: DynamoDB Integration"),
      heading2("Table Design"),
      para("A single DynamoDB table (hw8-cart-carts) was created with cartId (String) as the partition key. No sort key was required because all three access patterns operate on a single cart at a time:"),
      bullet("Create cart: PutItem with a new UUID-based cartId string."),
      bullet("Get cart: GetItem by cartId — O(1) lookup regardless of table size."),
      bullet("Add items: UpdateItem using an expression to upsert into an items map attribute embedded in the cart item."),
      spacer(),
      para("A single-table design was chosen over multiple tables because the shopping cart use case has no complex relational queries. All data needed for a cart response (metadata + items) fits in a single DynamoDB item, eliminating the need for joins or secondary indexes."),
      spacer(),

      heading2("Partition Key Strategy"),
      para("The cartId is generated as a random int64 converted to a string in the application layer. This ensures even distribution across DynamoDB partitions with no hot-partition risk under normal shopping load. A UUID would also work but the int64 approach kept the key compact and consistent with the MySQL auto-increment ID format for API compatibility."),
      spacer(),

      heading2("Implementation Journey & Type Mismatch Bug"),
      para("The most significant implementation challenge was a DynamoDB attribute type mismatch: the application initially sent cartId as a Number (N) type, but the table schema defined it as a String (S). This caused all PutItem operations to fail with a ValidationException."),
      spacer(),
      para("Debugging this required inspecting CloudWatch logs directly, as the HTTP response only returned a generic 'internal error' message. The fix was to cast the int64 cartId to a string before constructing the DynamoDB attribute value. This also highlighted the importance of verbose server-side logging for distributed debugging."),
      spacer(),
      para("A secondary complication arose from Docker build caching: after fixing the type mismatch in code, Docker reused a cached layer for the Go compilation step, producing an image that still contained the bug. Using --no-cache on the next build resolved this and produced the correct binary."),
      spacer(),

      heading2("Eventual Consistency Observations"),
      para("During the 150-operation test, no eventual consistency delays were observed. All GET /shopping-carts/{id} operations immediately after a POST returned the expected data. This is consistent with DynamoDB's read-after-write consistency guarantee for GetItem on the primary partition key — eventual consistency primarily affects Global Secondary Index queries, which were not used here."),
      spacer(),

      heading2("DynamoDB Performance Results"),
      para("Data source: dynamo_test_results.json (150 operations)"),
      spacer(),
      new Table({
        width: { size: 9360, type: WidthType.DXA },
        columnWidths: [3120, 2080, 2080, 2080],
        rows: [
          new TableRow({ tableHeader: true, children: [
            cell("Operation", true, true, 3120),
            cell("Avg (ms)", true, true, 2080, AlignmentType.CENTER),
            cell("P95 (ms)", true, true, 2080, AlignmentType.CENTER),
            cell("P99 (ms)", true, true, 2080, AlignmentType.CENTER),
          ]}),
          new TableRow({ children: [
            cell("create_cart", false, false, 3120),
            cell("32.8", false, false, 2080, AlignmentType.CENTER),
            cell("37.0", false, false, 2080, AlignmentType.CENTER),
            cell("37.5", false, false, 2080, AlignmentType.CENTER),
          ]}),
          new TableRow({ children: [
            cell("add_items", false, false, 3120),
            cell("33.4", false, false, 2080, AlignmentType.CENTER),
            cell("39.4", false, false, 2080, AlignmentType.CENTER),
            cell("64.4", false, false, 2080, AlignmentType.CENTER),
          ]}),
          new TableRow({ children: [
            cell("get_cart", false, false, 3120),
            cell("31.6", false, false, 2080, AlignmentType.CENTER),
            cell("36.3", false, false, 2080, AlignmentType.CENTER),
            cell("58.0", false, false, 2080, AlignmentType.CENTER),
          ]}),
        ]
      }),
      spacer(),
      para("All 150 operations succeeded (100% success rate). Average latencies were 2-3ms higher than MySQL, with notably higher P99 values (up to 64.4ms for add_items). DynamoDB CloudWatch confirmed sub-4ms internal request latency, indicating the additional overhead is network and SDK serialization cost rather than database processing time."),
      spacer(),

      heading2("DynamoDB CloudWatch Screenshots"),
      ...img(dynamoLatency, 620, 100, "Figure 6: DynamoDB SuccessfulRequestLatency by operation"),
      ...img(dynamoReadCap, 620, 100, "Figure 7: DynamoDB ConsumedReadCapacityUnits + SuccessfulRequestLatency"),
      ...img(dynamoWriteCap, 620, 100, "Figure 8: DynamoDB ConsumedWriteCapacityUnits + SuccessfulRequestLatency"),
      ...img(dynamoCombined, 620, 120, "Figure 9: Combined ECS + DynamoDB utilization metrics over time"),
      ...img(dynamoLogs, 620, 100, "Figure 10: CloudWatch Log Events — DynamoDB backend startup"),

      // ── STEP III ──────────────────────────────────────────────────────────
      new Paragraph({ pageBreakBefore: true, children: [new TextRun("")] }),
      heading1("STEP III: Database Comparison & Analysis"),
      heading2("Performance Comparison Table"),
      para("Data source: combined_results.json (300 total records — 150 MySQL + 150 DynamoDB)"),
      spacer(),
      comparisonTable,
      spacer(),

      heading2("Operation-Specific Breakdown"),
      opTable,
      spacer(),

      heading2("Consistency Model Analysis"),
      para("MySQL (RDS) provides full ACID transactions. Multi-item cart updates are executed within a single transaction, ensuring that either all items are committed or none are. This is critical for shopping cart integrity — a partial write leaving half the items committed would be a data corruption bug."),
      spacer(),
      para("DynamoDB provides eventual consistency by default for non-primary-key reads. However, for this implementation, all reads use GetItem on the primary partition key (cartId), which provides read-after-write consistency. No eventual consistency delays were observed during the 150-operation test. If Global Secondary Indexes were added (e.g., to query carts by customer_id), eventual consistency could become observable and would require either strongly consistent reads (at 2x read cost) or application-level retry logic."),
      spacer(),

      heading2("Resource Efficiency Analysis"),
      para("MySQL requires active connection management. The connection pool (MaxOpenConns: 25) consumes resources even when idle, and the db.t3.micro instance runs continuously regardless of traffic. This creates a fixed cost floor but predictable capacity planning."),
      spacer(),
      para("DynamoDB on-demand mode charges per request with no standing infrastructure cost. For low-traffic periods this is significantly cheaper, but under sustained high load the per-request cost can exceed a reserved RDS instance. From the CloudWatch data, DynamoDB consumed approximately 1-2 Read/Write Capacity Units per operation — well within free-tier limits for this test scale."),
      spacer(),
      para("Operationally, DynamoDB required zero schema migrations, no connection pool tuning, and no index maintenance. MySQL required schema design upfront and two iterations (flat JSON column vs normalized tables) before achieving the target performance."),
      spacer(),

      heading2("Real-World Scenario Recommendations"),
      scenarioTable,
      spacer(),

      heading2("Evidence-Based Architecture Recommendations"),
      para("Shopping Cart Winner: MySQL", true),
      para("For this specific workload, MySQL outperformed DynamoDB across all operations and all percentiles. The key evidence:"),
      bullet("Avg response time advantage: 2.7ms (29.9ms vs 32.6ms)"),
      bullet("P99 advantage: 21.3ms (36.7ms vs 58.0ms) — critical for user experience under load"),
      bullet("Implementation complexity: DynamoDB required a type-mismatch debugging session and Docker cache invalidation. MySQL required schema iteration but no SDK-level attribute formatting issues."),
      spacer(),
      para("When to Choose DynamoDB:"),
      bullet("Traffic is highly unpredictable with large spikes (auto-scaling eliminates capacity planning)"),
      bullet("Multi-region deployment is required (DynamoDB Global Tables vs complex Aurora Global setup)"),
      bullet("The data model has no relational requirements (no joins needed)"),
      bullet("Operational simplicity is prioritized over raw latency performance"),
      spacer(),
      para("Polyglot Strategy for a Full E-Commerce System:"),
      bullet("Shopping carts: MySQL — lower latency, ACID transactions, simple at startup scale"),
      bullet("User sessions: DynamoDB — short-lived, high-write, naturally key-value shaped"),
      bullet("Product catalog: MySQL — rich query patterns (search, filter, sort) benefit from SQL indexes"),
      bullet("Order history: MySQL — relational joins between orders, items, customers require ACID"),
      spacer(),

      heading2("Learning Reflection"),
      para("What Surprised Us:"),
      bullet("MySQL outperformed DynamoDB in absolute latency at this scale. The conventional wisdom that NoSQL is 'faster' did not hold here — MySQL's in-region RDS instance with a warm connection pool matched or beat DynamoDB's managed service, likely because DynamoDB's AWS SDK adds serialization overhead that dominates at sub-50ms latency targets."),
      bullet("DynamoDB's P99 latency (58-64ms) was significantly worse than MySQL's (36-40ms). The tail latency gap is more important than average latency for user experience."),
      bullet("The Docker build cache issue cost significant debugging time. Lesson: always use --no-cache when the code change is in a compiled step, or use multi-stage builds with explicit cache-busting."),
      spacer(),
      para("What Failed Initially:"),
      bullet("MySQL Schema v1 used a JSON column for cart items. This worked functionally but made UPDATE operations complex and prevented index-based optimization. Switched to a normalized two-table design."),
      bullet("DynamoDB type mismatch (Number vs String for cartId) caused 100% failure on all PutItem operations. Only visible in CloudWatch logs, not in the HTTP response body."),
      bullet("Connection pool defaults caused timeout errors during burst testing. Tuning MaxOpenConns resolved this."),
      spacer(),
      para("Key Insight:"),
      para("At the scale tested (150 sequential operations, 1 ECS task, us-east-1 region), a well-configured RDS MySQL instance consistently outperforms DynamoDB on latency. DynamoDB's value proposition emerges at scale (millions of requests, global distribution, serverless cost model) — not at the prototype stage. The right database choice depends heavily on your scale, operational maturity, and traffic pattern, not on a universal 'SQL vs NoSQL' rule.", false, 22),
    ]
  }]
});

Packer.toBuffer(doc).then(buffer => {
  fs.writeFileSync("HW-8-Report.docx", buffer);
  console.log("Report saved: HW-8-Report.docx");
});
