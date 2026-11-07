# Event Architecture

Domain events are typed facts emitted after successful state transitions. Events carry an event ID, entity ID, generation, timestamp and schema version. They do not carry prompts, credentials or arbitrary raw process output.

The application event bus is in-process and bounded. Consumers must tolerate duplicate delivery and stale generations. Event handlers must not block inference or lifecycle control. Persist only selected operational events with retention limits.
