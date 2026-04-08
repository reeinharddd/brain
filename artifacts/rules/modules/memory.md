---
module: memory
title: Memory
---

## Memory module

This module documents the repository conventions and required scripts for
persistent memory used by the Brain environment.

Required scripts and artifacts

- `memory-namespace.sh` — utility to derive namespace for project-scoped memory
- `$HOME/.brain/memory` — storage directory for vector chunks and memory artefacts

Notes

- Memory implementations should follow the Memory Protocol described in
  `memory-protocol.md` and provide a `memory-namespace.sh` helper so tools
  and CI can derive stable namespaces for project memory.

Examples

- To derive the namespace for a project root, run:

```sh
# derive namespace for current project
bash ~/.brain/scripts/memory-namespace.sh /path/to/project
```

This file exists to satisfy the Doctor checks which expect a memory module and
the presence of a memory-namespace helper referenced from the module text.
