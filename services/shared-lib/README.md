# Shared Library (`shared-lib`)

This module contains shared Go libraries (logging, tracing, Kafka helpers, middleware) used by multiple microservices.

**Architectural Note:**
It is **not** deployed as a runtime service. It is a library dependency included at build time to avoid tight runtime coupling and "distributed monolith" anti-patterns.
