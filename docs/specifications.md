# Specifications and Canonical References

This project implements OpenStreetMap PBF encoding/decoding. For format behavior and compatibility decisions, use the canonical OSM PBF specification as the source of truth.

## OpenStreetMap PBF

- Canonical format spec: https://wiki.openstreetmap.org/wiki/PBF_Format

## Guidance

- When implementation behavior is ambiguous, prefer spec-compliant behavior.
- When adding tests for format semantics, include a short note referencing the relevant spec section.
- Keep implementation comments concise; put broad reference links here instead of repeating them across files.
