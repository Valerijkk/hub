# Case intent boundary

## Belongs here

Repository-owned `TC-*` intent, preconditions, and expected observable results
belong here independently of executable suite code. Every case participates in
the risk-linked `task -> requirement -> TC/case` chain.

## Forbidden

Executable tests, implementation details, generated runtime evidence, secrets,
untraceable case identifiers, and cases missing task, requirement, or risk
linkage are forbidden.

## First consumer

The first accepted consumer must add one reviewed `TC-*` case against the
closed case schema, including its task, requirement, and risk links, before an
executable test consumes it.
