## Module: Observability

### System Health
The brain system must be observable to ensure it is operating correctly across sessions and IDEs.

### Metrics to Capture
1. **Response Time**: Track the time taken by each agent to fulfill a request.
2. **MCP Availability**: Log every failed connection attempt to an MCP server.
3. **Model Success Rate**: Track fallbacks and provider errors.
4. **Token Usage**: Log token consumption per session and project.

### Logging
- All system logs must be stored in `~/.brain/logs/`.
- Use the following categories: `[INFO]`, `[WARN]`, `[ERROR]`, `[DEBUG]`.
- Failure logging: If an MCP or Provider fails, log the exact error message and timestamp.

### Dashboards and Review
- Use `scripts/dashboard.sh` to visualize system performance.
- Review system health at the start of each week.
- If an MCP is down > 10% of the time, investigate the cause (version conflict, resource limit).

### Alerts
- Trigger an alert if a `CRITICAL` guardian check is bypassed.
- Notify the user if a `RuleCandidate` is ready for promotion but has been ignored for > 7 days.
