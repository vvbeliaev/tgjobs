# Agent Instructions for JobHunter (Hexagonal Architecture)

This document provides context and rules for AI agents working on the JobHunter backend.

## Architecture Principles

1. **Hexagonal Architecture**:
   - `core/`: Domain models and interfaces (Ports). NO dependencies on other packages or external libraries (except basic ones).
   - `usecases/`: Business logic implementations. Depends on `core` interfaces.
   - `adapters/in/`: Driving adapters (HTTP handlers, PB hooks, TG listeners).
   - `adapters/out/`: Driven adapters (LLM implementations, external APIs).
2. **Dependency Injection**: All dependencies are wired in `pb/main.go`. Avoid using `init()` functions for logic.
3. **PocketBase Integration**:
   - We use `app.Dao()` or `app.FindRecordById` directly in services/usecases for simplicity (no repository abstraction unless it becomes a bottleneck).
   - Domain objects (Aggregates) wrap `*core.Record` to provide business methods.

## Modules

### Collector Module (`pb/pkg/collector`)
Responsible for ingesting data from external sources (currently Telegram).
- Depends on `JobService` interface to submit raw findings.
- Performs fast pre-filtering before calling the Job module.

### Job Module (`pb/pkg/job`)
The core domain of the application.
- **Aggregate**: `pkg/job/core/job.go` - handles state transitions (`Raw` -> `Processing` -> `Processed`).
- **Extraction**: Uses LLM to convert unstructured text into `ParsedData`.
- **Offer Generation**: Uses LLM to create cold messages based on User's CV.

## Development Rules

1. **Go Context**: Always pass `context.Context` through services and adapters.
2. **LLM Prompts**: Keep prompts inside `adapters/out/` of the respective module.
3. **State Machine**: Only the `Job` aggregate should change its own status. Use methods like `job.Complete(data)`.
4. **Error Handling**: Use structured logging with `zap`.
5. **Config**: Use `pb/config/config.go` for all environment-based settings.

## Common Tasks

- **New Source**: Add a new folder in `pkg/collector/adapters/in` (e.g., `discord.go`).
- **New LLM Model**: Update the `Extractor` or `OfferGenerator` in `pkg/job/adapters/out`.
- **New Job Field**: 
  1. Add to `ParsedData` in `pkg/job/core/ports.go`.
  2. Update `job.Complete()` in `pkg/job/core/job.go`.
  3. Update prompt/mapping in `pkg/job/adapters/out/extractor.go`.
