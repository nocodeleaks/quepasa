# Metrics Instruction

## Scope
- Module: `src/metrics`
- Responsibility: generic metrics abstractions, factory creation, backend/no-op behavior.

## Ownership Rules
- Central metrics module must not contain module-specific business logic.
- Domain modules own their metric variables and usage points.
- Metrics module exposes generic factories only.

## Factory Rules
- Keep factory interfaces stable:
  - `CreateCounterRecorder(...)`
  - `CreateCounterVecRecorder(...)`
  - `CreateHistogramVecRecorder(...)`
- Factories must resolve enabled backend vs no-op behavior internally.
- Calling modules should not need conditional logic for enabled/disabled metrics.

## Backend Rules
- Respect environment-driven enable/disable behavior.
- Preserve Prometheus compatibility for enabled mode.
- Preserve no-op safety for disabled mode.
- Keep metric operations non-breaking for call sites.

## Integration Rules
- Keep API/Whatsmeow/models/rabbitmq metric usage via factories.
- Preserve existing metric names and labels unless explicitly migrated.
- Maintain compatibility with dashboards and monitoring queries.
