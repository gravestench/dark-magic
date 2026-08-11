# Next research tranche

After the foundation baseline is reviewed, the next tightly coupled research group should be:

1. combat, damage, defense, and death;
2. skills, missiles, states, and combat actions;
3. monsters, spawning, AI, and lifecycle;
4. hirelings, mercenaries, pets, and owned units;
5. movement, collision, pathfinding, and room streaming.

Research these together because they share the same timing, stat-source, target, movement, death, and ownership vocabulary, but keep one primary document per workstream so Codex can implement them in independently reviewable slices.

The combat tranche should not reopen foundation architecture unless evidence contradicts it. Prefer to extend the existing authoritative session, ECS, item authority, game-data generation, and replay/checkpoint boundaries.
