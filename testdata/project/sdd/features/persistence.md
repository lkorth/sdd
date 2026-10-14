---
feature: PER
---

# Persistence

Durable storage on the device.

## L2

### L2_PER_004: Acknowledgement is durable before deletion

```yaml
parent: L1_LNK_001
```

The system SHALL write an upload acknowledgement to durable storage before deleting the
acknowledged record.

Rationale: If the system deletes a record before its acknowledgement is durable, a power loss
can lose the record.

Scenario: Power is removed between acknowledgement and deletion
- WHEN power is removed after an acknowledgement is received and before the record is deleted
- THEN the record is present on restart and is not uploaded a second time
