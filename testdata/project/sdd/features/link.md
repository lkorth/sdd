---
feature: LNK
---

# Link

Transport and session resumption.

## L1

### L1_LNK_001: An interrupted upload completes without duplicate delivery

```yaml
parent: L0_001
```

If the link drops during an upload, the link SHALL complete the upload after it reconnects,
with no sample delivered twice.

Rationale: Links drop often in the field.

Scenario: The link drops in the middle of an upload
- WHEN the link drops after half of the samples are sent
- AND the link reconnects
- THEN the archive holds each sample exactly once

## L2

### L2_LNK_001: An interrupted upload resumes after the last acknowledgement

```yaml
parent: L1_LNK_001
```

The link SHALL resume an interrupted upload at the first sample after the last acknowledged
sample.

Rationale: The acknowledgement is the only fact both ends agree on.

Scenario: Resume after acknowledgement 40
- WHEN the last acknowledgement names sample 40
- THEN the next upload starts at sample 41, and the archive
  holds samples 1 to 40 once

### L2_LNK_002: The reconnect delay is bounded

```yaml
parent: L1_LNK_001
status: tbr
```

The link SHALL try to reconnect within TBR_1 seconds of a drop.

Rationale: The operations team chooses TBR_1 by 2026-12-01.

Scenario: The link drops
- WHEN the link drops
- THEN the link tries to reconnect within TBR_1 seconds

## Notes

This section is commentary. It names L2_LNK_001 but does not bind.
