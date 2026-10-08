# History

Adoption logs. Each one records what a platform port changed, what it found, and what it
deliberately left for later, as of the day it was written. They are kept because the findings
are the record of decisions that are otherwise invisible in the code; they are not kept current,
so a path or a version number in one is the port's, not the repository's today.

| Log                                                                | Port                                                     |
|--------------------------------------------------------------------|----------------------------------------------------------|
| [platform-go-v13-adoption.md](platform-go-v13-adoption.md)         | v12 → v13: what changed, and what it left for later      |
| [platform-go-v14-test-port.md](platform-go-v14-test-port.md)       | v14 at `main`: whether anything in it would force a v15  |
| [v14-port-completion-findings.md](v14-port-completion-findings.md) | Finishing the v14 port: what landed, and what stopped it |

For what the code does now, read `../writing_go.md` and the domain docs beside it.
