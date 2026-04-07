name: /plan
version: 1.0.0
status: stable
invoked_by: user-request
delegates_to:

- planner-agent

<!-- markdownlint-disable-file -->

# Command: /plan

## Summary

Transforms a vague goal into detailed, executable breakdown using SDD phases.

## Usage

```
User: /plan Add JWT authentication to API
       ├─ Constraints: "Must support refresh tokens" "Zero downtime"
       └─ Context: "Current: session-based, 10 endpoints"

System: Delegates to planner-agent
        Returns: SDD spec + task breakdown
        Status: Ready for PROPOSE phase
```

## Output Example

```json
{
  "phase": "SPEC",
  "artifact": "docs/sdd/jwt-auth-spec.md",
  "tasks": [
    {
      "id": 1,
      "name": "Generate RSA key pair",
      "estimated_hours": 0.5,
      "dependencies": []
    },
    {
      "id": 2,
      "name": "Update login endpoint",
      "estimated_hours": 2,
      "dependencies": [1]
    }
  ],
  "total_hours": 9.5,
  "next_phase": "DESIGN"
}
```

Specification complete in 45 min.
