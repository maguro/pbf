# Test data

| File | Source |
|---|---|
| `sample.osm.pbf` | OSM |
| `bremen.osm.pbf` | OSM extract |
| `greater-london.osm.pbf` | OSM extract |
| `san-francisco.osm.pbf` | OSM extract |
| `edge-cases.osm.pbf` | Written by this repository |

The round trip integration test reads every file.

## edge-cases.osm.pbf

This file holds the cases that the OSM extracts do not have. Package
`internal/edgecases` defines each case and the code path that it covers.
`TestEdgeCasesFixture` decodes the file and compares each entity with its
case. A failure names the case ID.

| ID | Case |
|---|---|
| N1 | Node with zero Info values and an empty user |
| N2 | Node with many tags: UTF-8, spaces, `=`, empty value |
| N3 | Node with an empty tag key |
| N4, N5 | Nodes at the maximum and minimum latitude and longitude |
| N6a, N6b | Nodes with the minimum and maximum ID. The ID delta wraps. |
| N7a, N7b | Nodes with the maximum and minimum UID. The UID delta wraps. |
| N8 | Node with negative Info values and a timestamp before 1970 |
| N9 | Node with a long UTF-8 user name |
| W1 | Way without refs and tags |
| W2 | Way with one ref and many tags |
| W3 | Closed way with a repeated ref |
| W4 | Way with refs that make the ref delta wrap |
| W5 | Way with the minimum ID and refs to missing nodes |
| W6 | Way with negative Info values and a timestamp before 1970 |
| R1 | Relation without members |
| R2 | Relation with members of each type, an empty role, and itself as a member |
| R3 | Relation with member IDs that make the member delta wrap |
| R4 | Relation with the maximum ID and a missing member |
| R5a, R5b | Two relations with the same ID |
| M1 | Node after the relations: a second dense group in the same block |
| H1 | Header with each settable field not at its default |
| H2 | Header bounding box equal to the node extent |

Not in the file:

| Case | Reason |
|---|---|
| Entity without Info | The encoder panics. #13 |
| Dense nodes without metadata | The decoder panics. #14 |
| Deleted entities (`Visible: false`) | The encoder does not write history files. #24 |
| Non-dense nodes | The encoder does not write them. |
| Compressions other than ZLIB | The encoder writes ZLIB by default. |

To write the file again after a change to the cases:

```sh
go run ./internal/tools/edgecases -o testdata/edge-cases.osm.pbf
```

Each run writes a file with the same bytes.
