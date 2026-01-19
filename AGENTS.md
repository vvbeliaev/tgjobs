# Agent Instructions for JobHunter (Hexagonal Architecture)

This document provides context and rules for AI agents working on the JobHunter backend.

## Architecture Principles

### Backend (Hexagonal Architecture)
1. **Hexagonal Architecture**:
   - `core/`: Domain models and interfaces (Ports). NO dependencies on external libraries.
   - `usecases/`: Business logic implementations. Depends on `core` interfaces.
   - `adapters/in/`: Driving adapters (HTTP handlers, PB hooks, TG listeners).
   - `adapters/out/`: Driven adapters (implementations of `core` interfaces for external services like LLM).
2. **Dependency Injection**: All dependencies are wired in `pb/main.go`.
3. **PocketBase Integration**: Aggregates wrap `*core.Record` to provide business methods.

### Frontend (Svelte 5)
1. **Svelte 5 Runes**: Always use `$state`, `$derived`, `$effect`. No more `export let` or `writable` stores for component state.
2. **State Management**: Use `.svelte.ts` classes for complex state (e.g., the job feed).
3. **Styling**: Tailwind CSS 4 + DaisyUI 5. Prefer utility classes over custom CSS.
4. **PocketBase Client**: Use the shared client in `src/lib/shared/pb.ts`.
5. **Types**: Use auto-generated types from `pocketbase-typegen`.

## Modules

### Backend
- **Collector Module** (`pb/pkg/collector`): Responsible for ingesting data from external sources (Telegram).
- **Job Module** (`pb/pkg/job`): The core domain (Extraction, Offer Generation).

### Frontend
- `src/lib/shared`: Generic UI components, utils, and PB client.
- `src/lib/apps`: App-specific logic (e.g., `dashboard` for viewing vacancies).
- `src/routes`: SvelteKit routes.

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
