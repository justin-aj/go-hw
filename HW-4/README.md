# HW-4: MapReduce Implementation & Scaling Analysis

A containerized MapReduce implementation for distributed word count processing. This assignment demonstrates big data processing patterns, Docker containerization, orchestration, and horizontal scaling analysis.

## Overview

This project implements a MapReduce pipeline to process large text files (e.g., Shakespeare's Hamlet) in a distributed manner. It explores:
- Data splitting and distributed processing
- Container orchestration and networking
- Horizontal scaling performance analysis
- Fault tolerance and retry mechanisms
- Performance metrics collection

## Project Structure

```
HW-4/
├── mapreduce/
│   ├── orchestrator.py              # Main orchestration script
│   ├── performance.py               # Performance metrics collection
│   ├── verify.py                    # Output verification tool
│   ├── shakespeare-hamlet.txt        # Input data for processing
│   ├── SCALING_README.md            # Scaling experiment results
│   ├── mapreduce_performance.png    # Performance visualization
│   │
│   ├── mapper/                      # Mapper container
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── main.go                  # Maps chunks to word counts
│   │
│   ├── reducer/                     # Reducer container
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── main.go                  # Aggregates word counts
│   │
│   └── splitter/                    # Splitter container
│       ├── Dockerfile
│       ├── go.mod
│       └── main.go                  # Splits input into chunks

```

## Prerequisites

- Docker & Docker Compose
- Python 3.x with requests library
- Go 1.16+ (for building containers)
- 2GB+ free disk space

## Pipeline Architecture

The MapReduce pipeline consists of three stages:

### 1. Splitter
- Reads input file (shakespeare-hamlet.txt)
- Divides into equal-sized chunks
- Creates intermediate files for distribution
- **Parallelism**: Single, orchestrated job

### 2. Mapper
- Processes individual chunks
- Counts word occurrences per chunk
- Outputs word → count pairs
- **Parallelism**: Multiple mappers (one per chunk)

### 3. Reducer
- Aggregates results from all mappers
- Combines counts for each word
- Produces final result
- **Parallelism**: Single, aggregation stage

## Setup & Execution

### Build Docker Images

```bash
cd HW-4/mapreduce
docker-compose build
```

### Run MapReduce Pipeline

```bash
python orchestrator.py run
```

This command:
1. Starts Docker containers for all stages
2. Runs the complete pipeline
3. Collects timing metrics
4. Displays results

### Scaling Experiment

Test performance with multiple chunk sizes:

```bash
python orchestrator.py scale
```

Measures pipeline execution time with:
- 1 chunk (no parallelism)
- 2 chunks
- 4 chunks
- 8 chunks
- 16 chunks

### Verify Results

```bash
python verify.py
```

Validates output correctness against expected word counts.

## Configuration

Key parameters in `orchestrator.py`:

```python
# Number of chunks to split input into
NUM_CHUNKS = 4

# Retry attempts for failed stages
MAX_RETRIES = 3

# Docker network timeout (seconds)
TIMEOUT = 60

# Container image names
SPLITTER_IMAGE = "mapreduce-splitter"
MAPPER_IMAGE = "mapreduce-mapper"
REDUCER_IMAGE = "mapreduce-reducer"
```

## Output & Results

### Pipeline Output
```
============================================================
MapReduce Pipeline — 4 chunks, 4 mapper(s)
============================================================

Phase 1: Splitting...
  Splitter succeeded in 0.596s (attempt 1)
  Created 4 chunks

Phase 2: Mapping...
  Mapper (chunk 0) succeeded in 0.553s (attempt 1)
  Mapper (chunk 1) succeeded in 0.542s (attempt 1)
  Mapper (chunk 2) succeeded in 0.548s (attempt 1)
  Mapper (chunk 3) succeeded in 0.541s (attempt 1)
  Map phase wall time: 0.557s

Phase 3: Reducing...
  Reducer succeeded in 0.335s (attempt 1)

============================================================
PIPELINE COMPLETE
============================================================
  Split time:    0.596s
  Map time:      0.557s
  Reduce time:   0.335s
  Total time:    1.488s
============================================================
```

### Performance Metrics

Results stored in `mapreduce_performance.png`:
- Execution time vs number of chunks
- Mapper parallelism speedup
- Diminishing returns analysis

**See** `SCALING_README.md` for detailed experiment results and analysis.

## Key Experiments

### Scaling Analysis
Tests how system scales with increasing parallelism:
- **1 chunk**: Baseline (no parallelism)
- **2-16 chunks**: Measure speedup

**Findings**:
- Linear speedup with chunk count (up to hardware limits)
- Overhead from splitting and reduction stages
- Network latency becomes significant at high parallelism

### Fault Tolerance
Demonstrates automatic retry on failure:
- Mapper failures trigger automatic retry
- Failed stage repeats up to MAX_RETRIES times
- Pipeline continues if stage succeeds on retry

### Container Orchestration
Shows Docker networking and service communication:
- Splitter outputs chunks to shared volume
- Mappers read from shared volume
- Reducer consumes mapper outputs

## Performance Characteristics

### Time Complexity
- **Splitting**: O(n) where n = file size
- **Mapping**: O(n/k) per mapper, k = number of chunks
- **Reducing**: O(w) where w = unique words
- **Overall**: O(n + w) with parallelism

### Space Complexity
- **Per chunk**: O(n/k)
- **Intermediate data**: O(n) for chunk files
- **Final output**: O(w) for word counts

### Scaling Characteristics
- **Strong scaling**: Decreases with increasing parallelism
- **Optimal chunks**: 4-8 for typical systems
- **Overhead**: Coordination, network I/O

## Common Tasks

### Process Different File
Replace `shakespeare-hamlet.txt` with your file:
```bash
cp /path/to/your/file.txt shakespeare-hamlet.txt
```

### Adjust Parallelism
Modify `NUM_CHUNKS` in `orchestrator.py`:
```python
NUM_CHUNKS = 8  # Increase parallelism
```

### Debug Failures
Enable verbose logging:
```python
DEBUG = True  # In orchestrator.py
```

### Collect Performance Data
```bash
python performance.py
```

Generates detailed metrics and charts.

## Docker Images

### Splitter
- **Input**: shakespeare-hamlet.txt
- **Output**: chunk_0.txt, chunk_1.txt, ...
- **Purpose**: Data partitioning

### Mapper
- **Input**: Single chunk file
- **Output**: word_count_mapper_X.txt
- **Purpose**: Parallel word counting

### Reducer
- **Input**: All mapper outputs
- **Output**: final_result.txt
- **Purpose**: Aggregation

### Building Images Manually
```bash
docker build -t mapreduce-splitter ./splitter
docker build -t mapreduce-mapper ./mapper
docker build -t mapreduce-reducer ./reducer
```

## Troubleshooting

### Container Won't Start
```bash
docker logs <container_name>
docker-compose logs
```

### Timeout Errors
Increase timeout in orchestrator.py:
```python
TIMEOUT = 120  # Increase from default
```

### Network Issues
Ensure containers are on same Docker network:
```bash
docker network ls
docker network inspect <network_name>
```

### File Permission Issues
```bash
chmod 755 *.py
chmod 755 */main.go
```

## Optimization Ideas

1. **Streaming Processing**: Process files without loading entirely
2. **Compression**: Compress intermediate chunk files
3. **Parallel Reduce**: Hierarchical reduce tree
4. **Caching**: Cache mapper outputs
5. **Load Balancing**: Dynamic chunk sizing based on content

## Technologies Used

- **Python 3.x**: Orchestration and scripting
- **Go 1.16+**: Mapper and reducer implementation
- **Docker & Compose**: Containerization and orchestration
- **Bash**: Scripting and execution

## Concepts Covered

- **MapReduce Programming Model**: Distributed data processing
- **Data Partitioning**: Splitting data for parallel processing
- **Container Orchestration**: Coordinating multiple services
- **Performance Measurement**: Timing and throughput analysis
- **Fault Tolerance**: Automatic retry mechanisms
- **Horizontal Scaling**: Adding resources to improve performance
- **Docker Networking**: Inter-container communication

## References

- [MapReduce Paper](https://research.google/pubs/mapreduce-simplified-data-processing-on-large-clusters/)
- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Go Standard Library](https://golang.org/pkg/)

## Notes

- Pipeline is sequential: splitter → mappers → reducer
- Mappers run in parallel on separate chunks
- Output files stored in `/tmp/mapreduce/`
- Clean up with `docker-compose down -v`
- Test with smaller files first before scaling

## Further Reading

See `SCALING_README.md` for detailed experimental results including:
- Execution time measurements
- Parallelism efficiency analysis
- Overhead breakdown
- Scaling recommendations
