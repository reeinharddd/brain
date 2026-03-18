# Skill Context: Clean Architecture

- Domain rules must not depend on transport, persistence, or frameworks.
- Use cases orchestrate behavior and depend on ports, not adapters.
- Adapters translate external concerns into domain-friendly interfaces.
- Keep dependency direction inward and verify it in reviews.
- Favor thin controllers and repositories that do one job.
