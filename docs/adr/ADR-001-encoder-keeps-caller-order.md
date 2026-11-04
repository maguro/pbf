# ADR-001: The Encoder Keeps the Caller's Entity Order

## Status

Accepted

## Context

The encoder gets entities from the caller through `Encode` and `EncodeBatch`.
It puts the entities into primitive blocks. A primitive block holds one or more
primitive groups.

The encoder sends each entity to three streams, one stream for each entity
type (`Coalesce` in `internal/encoder/batch.go`). Each stream makes its own
blocks. A merge step takes the blocks from the three streams in the order that
they become ready. Thus the order of the node, way, and relation blocks
changes from one run to the next. Input that the caller sorted by type can
give a file that is not sorted by type.

### The format

The PBF format page on the OSM wiki
(<https://wiki.openstreetmap.org/w/index.php?title=PBF_Format&oldid=3078147>)
says these things:

- Each data block holds a series of primitive groups.
- A primitive group never holds more than one type of object. If a group
  held more than one type, a reader could not get the objects back in the
  order that the writer wrote them.
- The current serializer, Osmosis, keeps the order of the OSM entities.
- `Sort.Type_then_ID` is a proposed optional header feature. It states that
  the entities are sorted by type and then by ID.

The `PrimitiveBlock` message has no field for an entity type
(`internal/pb/osm.pb.go`). The one-type rule applies only to a group.

### Consumers

Many tools need a file that is sorted by type, and some also need ID order:

- The osm2pgsql manual says that OSM data files almost always have the nodes
  first, then the ways, then the relations, each in ID order. It also says:
  "Osm2pgsql can only read OSM files ordered in this way."
  (<https://osm2pgsql.org/doc/manual.html>)
- The osmium manual says that OSM files are usually sorted in this order, but
  not always. It says that many commands, in osmium and in other programs,
  work correctly only if the file is sorted.
  (<https://osmcode.org/osmium-tool/manual.html>)
- Planetiler stops with an error if the nodes do not come first, then the
  ways, then the relations (`OsmPhaser.java`,
  <https://github.com/onthegomap/planetiler>).

These consumers check the order while they read. None of them uses
`Sort.Type_then_ID` to decide what to do. osmconvert writes
`Sort.Type_then_ID` in each header that it writes, also when the input is not
sorted (`osmconvert.c`, <https://gitlab.com/osm-c-tools/osmctools>).

Each of these readers reads every group in a block, and decodes each group by
its own type:

- libosmium (`decode_primitive_block_data` in
  `include/osmium/io/detail/pbf_decoder.hpp`,
  <https://github.com/osmcode/libosmium>)
- Planetiler (`PbfDecoder.java`, <https://github.com/onthegomap/planetiler>)
- osmconvert (`osmconvert.c`, <https://gitlab.com/osm-c-tools/osmctools>)
- GDAL (`ReadPrimitiveGroup` in `osm_parser.cpp`,
  <https://github.com/OSGeo/gdal>)
- pbf (`internal/decoder/primitive.go`)

### Other writers

Osmosis and libosmium keep the order of the caller. osm4j does not:

- Osmosis starts a new group when the entity type changes. The new group goes
  into the same block. Osmosis closes the block when the batch is full
  (`switchTypes` in
  `osmosis-pbf/src/main/java/crosby/binary/osmosis/OsmosisSerializer.java`,
  <https://github.com/openstreetmap/osmosis>). It writes no optional
  features.
- osm4j keeps the nodes, the ways, and the relations of each batch in three
  lists. It writes the three lists as groups of one block (`writeBatch` in
  `PbfWriter.java`, <https://github.com/topobyte/osm4j-pbf>). Within a batch,
  osm4j does not keep the caller's order across types.
- libosmium starts a new block when the entity type changes
  (`PrimitiveBlock::can_add` in
  `include/osmium/io/detail/pbf_output_format.hpp`,
  <https://github.com/osmcode/libosmium>). It writes `Sort.Type_then_ID` only
  when the caller sets the header option `sorting` to `Type_then_ID`.

## Decision

The encoder writes the entities in the order that the caller gives them.
When the entity type changes, the encoder starts a new primitive group in the
same block. A change of the entity type does not close the block. The encoder
closes a block when the block holds the maximum number of entities.

The encoder does not sort the entities. The encoder does not add
`Sort.Type_then_ID` to the header. A caller that gives sorted entities can add
that feature with `WithOptionalFeatures`.

## Rationale

### Sorted input gives a sorted file

When the encoder keeps the order, a caller that gives entities sorted by type
and then by ID gets a file in that order. osm2pgsql can read that file, and so
can the osmium commands that need a sorted file. With the merge step, sorted
input does not always give a sorted file.

### The order agrees with the format page, Osmosis, and libosmium

The PBF format page says that the Osmosis serializer keeps the order of the
entities. Osmosis and libosmium also keep the order of the caller. A caller
that gives the same entities to pbf, to libosmium, or to Osmosis gets the
entities in the same order.

### A type change does not need a new block

The format puts the one-type rule on a group, not on a block. Each reader in
the Context section reads every group in a block. Thus a new group is
sufficient when the type changes. A new block adds a compressed blob, a blob
header, and a string table at each type change.

## Consequences

### Positive

- The same input gives a file with the same bytes on each run. The block
  order is fixed. The encoder sorts the string table of each block, and it
  sorts the tags of each entity by key.
- Input that changes the entity type often gives full blocks, not one small
  block for each run of one type.

### Negative

- The block layout is different from the layout that libosmium writes. A
  reader that expects one type in each block cannot read some files that pbf
  writes. No reader in the Context section has this limit.
- The encoder does not check a `Sort.Type_then_ID` that the caller adds. A
  caller can declare an order that the file does not have.

### Accepted Trade-Off

The encoder does not correct the order of the input. A caller that needs a
sorted file must sort the entities before it gives them to the encoder.

## Rejected Alternatives

### Document that the block order changes

Keep the merge step, and document that the order of the node, way, and
relation blocks changes from one run to the next. The documentation does not
change the file. Sorted input could still give a file that osm2pgsql cannot
read.

### One type in each block

Start a new block when the entity type changes, as libosmium does. This keeps
the caller's order, but it writes a new block at each type change. No reader
in the Context section needs one type in each block.

### Sort the entities in the encoder

Sort the entities by type and then by ID, and declare `Sort.Type_then_ID`. A
sort of a planet-size input needs disk space and memory outside the encoder.
None of the writers in the Context section sorts. libosmium leaves the sort to
the `osmium sort` command.

### Group each batch by type

Put the nodes, the ways, and the relations of each batch into separate
groups, as osm4j does. The output is the same on each run, but it does not
keep the caller's order. Mixed input does not give a sorted file, and the
order depends on where the batches start and end.

### Declare `Sort.Type_then_ID` when the input is sorted

Check each entity against the entity before it, and add `Sort.Type_then_ID`
when the whole input is sorted. The encoder writes the header after the body.
Thus the check is possible. The consumers in the Context section check the
order themselves, and none of them uses the feature. Also, the format page
does not define the ID order of negative IDs.

### Return an error for input that is not sorted

Valid files are not always sorted. The osmium manual says that OSM files are
usually sorted, but not always. An error would stop the caller from writing
such files.

## Non-Goals

- The order in which the decoder returns entities.
- Writing history files.
- A limit on the size of a block in bytes.
