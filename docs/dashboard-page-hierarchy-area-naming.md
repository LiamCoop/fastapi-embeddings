# RAG Platform — Dashboard Page Hierarchy & Area Naming

This hierarchy is intended to support early wireframing and exploratory prototypes.  
It emphasizes user mental model, trust, and iteration rather than infrastructure concepts.

The structure should feel like navigating an AI reliability workspace, not a vector database admin panel.

---

# 🏢 1. Organization Layer

### Workspace Home
Primary landing page after login.

Purpose
* Quick orientation
* High-level health + activity
* Entry point into projects

Key surfaces
* Recent corpuses
* AI health summary
* ingestion activity
* evaluation regressions
* quick actions (upload, test, connect MCP)

Naming alternatives
* Workspace
* Control Center
* AI Reliability Hub

---

# 📁 2. Project / Environment Layer

Projects represent logical groupings of corpuses and integrations.

### Project Overview

Purpose
* Context switching
* Project health
* Entry into knowledge + evaluation

Key surfaces
* corpuses within project
* MCP integrations
* evaluation summaries
* query analytics

Naming alternatives
* Project
* Knowledge Space
* AI Domain

---

# 📚 3. Knowledge Layer

This is where users spend significant time.

## Corpus Library

Purpose
* Browse and manage knowledge bases

Key surfaces
* corpus list
* ingestion status
* sync schedules
* version indicators

Naming alternatives
* Knowledge Library
* Corpus Hub
* Knowledge Vault

---

## Corpus Detail

Primary deep workspace for a single knowledge base.

Tabs / sub-areas:

---

### 📦 Ingestion
Vibe: observable pipeline

* upload sources
* sync configuration
* ingestion logs
* failure diagnostics
* indexing progress

---

### 🧱 Chunk Explorer
Vibe: structural transparency

* document browser
* chunk visualization
* metadata inspection
* chunk diffing across versions

Naming alternatives:
* Chunk Explorer
* Knowledge Structure
* Document Anatomy

---

### 🧬 Embedding & Index Strategy
Vibe: controlled optimization

* embedding model selection
* hybrid retrieval toggles
* metadata indexing
* reranker configuration

Naming alternatives:
* Retrieval Strategy
* Index Configuration
* Representation

---

# 🔎 4. Investigation Layer

This area builds trust and becomes a major differentiator.

## Retrieval Inspector

Purpose
Help users understand retrieval behavior.

Capabilities
* query → retrieved chunks
* similarity scores
* ranking explanation
* filter visibility
* “why wasn’t this retrieved?”

Naming alternatives:
* Retrieval Debugger
* Query Inspector
* Grounding Inspector

---

# 🤖 5. Playground Layer

## Answer Playground

Purpose
Safe experimentation and answer inspection.

Capabilities
* ask questions
* inspect grounding
* compare configurations
* test prompt variants
* multi-corpus queries

Naming alternatives:
* Playground
* AI Lab
* Grounded Chat

---

# 🧪 6. Evaluation Layer

## Evaluation Studio

Purpose
Measure quality and prevent silent regressions.

Capabilities
* golden question sets
* retrieval recall metrics
* hallucination indicators
* regression comparison
* answer grounding scores

Naming alternatives:
* Evaluation Studio
* Quality Lab
* Reliability Tests

---

# 📊 7. Observability Layer

## AI Health Dashboard

Purpose
Production monitoring and insight.

Capabilities
* common queries
* failed retrieval
* unanswered questions
* latency + cost metrics
* drift indicators

Naming alternatives:
* AI Health
* Reliability Dashboard
* Knowledge Observability

---

# 🔌 8. Integration Layer

## MCP Connections

Purpose
Enable external LLM/agent integration.

Capabilities
* MCP server credentials
* scoped access per corpus
* SDK quickstart snippets
* usage analytics
* rate limits

Naming alternatives:
* Integrations
* MCP Access
* AI Connect

---

# 🔐 9. Organization & Admin Layer

## Access & Permissions
* roles
* corpus access control
* audit logs

---

## Usage & Billing
* token usage
* retrieval volume
* cost attribution

---

## Settings
* environments (dev / prod)
* global defaults
* organization preferences

---

# 🧭 10. Suggested Primary Navigation Structure

A clean top-level nav could look like:

* Workspace
* Projects
* Knowledge
* Playground
* Evaluation
* AI Health
* Integrations
* Settings

This keeps navigation aligned with user intent, not implementation details.

---

# 🎯 Design Guidance

The hierarchy should reinforce:

* confidence over configuration
* investigation over magic
* safe iteration
* explainability
* retrieval-first thinking

The product should feel like:

a reliability workspace for AI knowledge systems

—not a vector database console.
