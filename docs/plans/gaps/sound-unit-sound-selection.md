# Gaps: sound-unit-sound-selection

Spec: [Unit sound selection](../../specs/sound-unit-sound-selection.md)

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|-----|------|---------|-------------------------------|------------------|-----------------------------------|----------------|
| Dispatch and identity | behavior | USN-Q01, USN-R01, USN-R02a, USN-R02b, USN-R03, USN-R04 | Full selection entry and player conversion | Complete frame/action roots and conversion callees | 2026-10-29: client incoming-reference scan, accessor reads and import resolution locate partial caller chains and conversion import; Ghidra has no instance | Trace full roots and read conversion to outputs; corroborate encoded instructions |
| Creature table binding | behavior | USN-Q02, USN-R01 | Default/alternate record result | Loader consumer fields and effective named cells | 2026-10-29: selector and two matched imported lookup helpers read; generic lookups and sound loader mapped; alternate eligibility and compiled bindings remain unbound | Read loaders/eligibility and original effective cells, corroborate compiled correspondence |
| Action integration | integration | USN-Q03, USN-R02a, USN-R02b | Full action entry and audible playback; USN-R02b supplied selection/timing available | Selector producer, fixed cue binding, timing unit and playback dependency | 2026-10-29: action selector stage fully traced and raw-decoded; Skill1..4 loader mapping and exact vectors established; producer/playback integration remains gated | Trace callers, bind cue and document request dependency |
| Movement timing | behavior | USN-Q04, USN-R03, USN-R04 | Phase, variant and secondary cue | Complete helpers, decoded instruction runs and draw ownership | 2026-10-29: phase dispatcher and random/wrapped-distance listings read; phase/variation and material tile helpers read; raw decoding corroborates numeric helpers; room/level producer integration remains unverified | Complete raw-image decoding, producer and callee paths, then exact discriminating vectors |
| Gameplay confirmation | observation | USN-Q05, USN-O01 | Runtime confirmation only | Owned original gameplay capture | 2026-10-29: owned vanilla session allocated, narrow complete text-table read collected and session closed; no gameplay trace armed because static assertions remain unresolved | Close static gates and reproduce exact established vectors on owned runtime |
