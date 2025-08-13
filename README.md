# Distributed Video Platform

## Table of Contents
1. [System Architecture](#system-architecture)
2. [RAFT Consensus Protocol](#raft-consensus-protocol)
3. [Byzantine Fault Tolerance](#byzantine-fault-tolerance)
4. [End-to-End Workflow](#end-to-end-workflow)
5. [Component Analysis](#component-analysis)
6. [Mathematical Foundations](#mathematical-foundations)
7. [Deployment and Testing](#deployment-and-testing)

## System Architecture

The distributed video platform implements a microservices architecture with the following components:

```
┌─────────────┐    ┌─────────────┐    ┌─────────────────┐
│   Frontend  │───▶│  Gateway    │───▶│  RAFT Cluster   │
│  (Next.js)  │    │ (Load Bal.) │    │ (Consensus)     │
└─────────────┘    └─────────────┘    └─────────────────┘
                                       │
                                       ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────────┐
│   Worker    │◀───│  RabbitMQ   │◀───│  MinIO Storage  │
│ (Thumbnail) │    │ (Message Q) │    │ (Object Store)  │
└─────────────┘    └─────────────┘    └─────────────────┘
```

### Component Roles

- **Frontend**: React-based dashboard for video upload and cluster monitoring
- **Gateway**: API gateway implementing service discovery and load balancing
- **RAFT Cluster**: Three-node consensus cluster for metadata replication
- **RabbitMQ**: Message broker for asynchronous video processing
- **MinIO**: S3-compatible object storage for video files and thumbnails
- **Worker**: Background service for thumbnail generation

## RAFT Consensus Protocol

### Theoretical Foundation

RAFT solves the distributed consensus problem by implementing a leader-based approach with the following properties:

**Safety Properties:**
- Election Safety: At most one leader can be elected in a given term
- Leader Append-Only: Leaders never overwrite or delete entries in their log
- Log Matching: If two logs contain an entry with the same index and term, then the logs are identical in all entries up through the given index
- Leader Completeness: If a log entry is committed in a given term, then that entry will be present in the logs of the leaders for all higher-numbered terms
- State Machine Safety: If a server has applied a log entry at a given index to its state machine, no other server will ever apply a different log entry for the same index

### Mathematical Model

**Node States:**
- FOLLOWER: Passive state, responds to RPCs from candidates and leaders
- CANDIDATE: Actively seeking votes to become leader
- LEADER: Handles all client requests and log replication

**Timing Constraints:**
```
broadcastTime << electionTimeout << MTBF (Mean Time Between Failures)
```

Where:
- broadcastTime: Time for a server to send RPCs to every server in the cluster
- electionTimeout: Time a follower waits before becoming a candidate
- MTBF: Average time between failures of a single server

**Majority Quorum:**
For a cluster of N nodes, consensus requires agreement from at least ⌊N/2⌋ + 1 nodes.

For our 3-node cluster:
- Total nodes: 3
- Required for majority: ⌊3/2⌋ + 1 = 2 nodes
- Fault tolerance: Can survive 1 node failure

### Leader Election Algorithm

```
1. All nodes start as FOLLOWERS
2. If election timeout elapses without receiving heartbeat:
   a. Increment currentTerm
   b. Transition to CANDIDATE
   c. Vote for self
   d. Send RequestVote RPCs to all other servers
3. If receives majority votes: become LEADER
4. If receives RPC from valid leader: become FOLLOWER
5. If election timeout elapses: start new election
```

**RequestVote RPC Parameters:**
- term: Candidate's term
- candidateId: Candidate requesting vote
- lastLogIndex: Index of candidate's last log entry
- lastLogTerm: Term of candidate's last log entry

### Log Replication

**AppendEntries RPC:**
- term: Leader's term
- leaderId: Leader's identifier
- prevLogIndex: Index of log entry immediately preceding new ones
- prevLogTerm: Term of prevLogIndex entry
- entries[]: Log entries to store (empty for heartbeat)
- leaderCommit: Leader's commitIndex


## Byzantine Fault Tolerance

### Problem Definition

The Byzantine Generals Problem models the challenge of achieving consensus in a distributed system where components may fail in arbitrary ways, including:
- Crash failures (nodes stop responding)
- Omission failures (messages are lost)
- Byzantine failures (nodes send conflicting information)

### RAFT vs Byzantine Fault Tolerance

**RAFT Assumptions:**
- Non-Byzantine failures only (crash-stop model)
- Network may partition but does not corrupt messages
- Messages may be lost, duplicated, or reordered

**Byzantine Tolerance Requirements:**
For Byzantine fault tolerance with f malicious nodes:
- Total nodes required: N ≥ 3f + 1
- For 1 Byzantine failure: Need 4 nodes minimum

The system uses RAFT (not Byzantine-tolerant) because:
1. Simpler implementation and reasoning
2. Better performance (fewer message rounds)
3. Suitable for trusted environments (controlled data center)
4. Network partitions more common than Byzantine failures in practice

### Fault Tolerance Analysis

**3-Node RAFT Cluster:**
- Tolerates: 1 crash failure
- Cannot tolerate: Network partition with 2 isolated nodes
- Recovery: Automatic when failed node rejoins

**Failure Scenarios:**

1. **Single Node Failure:**
   - 2 remaining nodes maintain majority
   - Leader election proceeds if leader fails
   - Read/write operations continue

2. **Network Partition (2+1):**
   - Majority partition (2 nodes) continues operation
   - Minority partition (1 node) becomes read-only
   - Automatic healing when partition resolves

3. **Two Node Failures:**
   - Cluster loses majority quorum
   - System becomes read-only
   - Requires manual intervention

## End-to-End Workflow

### Video Upload Workflow

```
1. User selects video file in frontend
2. Frontend sends multipart/form-data to Gateway
3. Gateway discovers current RAFT leader
4. Gateway proxies request to leader node
5. Leader node processes upload:
   a. Validates request
   b. Uploads file to MinIO storage
   c. Generates video metadata
   d. Stores metadata in RAFT log
   e. Replicates to follower nodes
   f. Publishes message to RabbitMQ
6. Worker consumes message from RabbitMQ
7. Worker downloads video from MinIO
8. Worker generates thumbnail
9. Worker uploads thumbnail to MinIO
10. Frontend displays updated video library
```