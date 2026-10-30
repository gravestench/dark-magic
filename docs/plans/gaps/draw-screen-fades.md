# Gaps: draw-screen-fades

Spec: [Screen fades](../../specs/draw-screen-fades.md)

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|---|---|---|---|---|---|---|
| Trigger and cadence | integration | SFD-Q01, SFD-R01, SFD-R02 | Gameplay eligibility and scheduling; supplied events remain available | Complete root-to-start/cancel and frame path | 2026-10-29: read client 0xA7E20, packet 0x15 handler 0xAF640, two zero-argument callers and room/skill-pointer helpers; server 0xCBB10/0x788B0 flag-to-packet leads recorded in companion, complete producers remain unresolved | Read complete dispatch and eligibility helpers |
| Final pixels | integration | SFD-Q02, SFD-R03 | Final rendered image; request generation remains available | Backend dispatch and blend semantics | 2026-10-29: resolved client import to D2Gfx ordinal 10045 RVA 0xB9F0; read signed guards and interface call, stopped at backend boundary | Trace the fingerprinted graphics artifact through its backend consumer |
| Gameplay confirmation | observation | SFD-Q03, SFD-O01, SFD-R01, SFD-R02, SFD-R03 | Runtime-confirmation claim only | Valid matching gameplay capture | 2026-10-29: initial incomplete attempts retained in companion. Later capture yielded complete 16- and 512-record inactive windows with zero drops, deadline zero, dimensions 800 by 600, selector 0 and return zero. Normal town exit and portal return did not establish an active fade. Other windows lost input snapshots when a scratch register changed; independent clock-call overlap remained unestablished, so V10's no-clock-sampling assertion is unobserved. No promotion | Established eligible trigger and stable complete input capture with correlated output and clock-sampling observations |
