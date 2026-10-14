# System

What the system is for, and what is always true of it.

### L0_001: Recorded data reaches the archive

```yaml
parent: self
```

The system SHALL deliver each recorded sample to the archive at most once.

Rationale: The archive is the only copy that outlives the device. A lost sample cannot be
measured again.

Scenario: A sample is recorded
- WHEN the device records a sample
- THEN the archive holds the sample exactly once after the next upload

### L0_002: No feature reads the wall clock

```yaml
parent: self
```

A monotonic clock is a clock that never goes back. Each feature SHALL measure time with a
monotonic clock.

Rationale: The wall clock jumps when the device synchronizes it.

That jump breaks each timeout that spans it.

Scenario: The wall clock jumps back
- WHEN the wall clock moves back one hour during a timeout
- THEN the timeout ends at its configured duration
