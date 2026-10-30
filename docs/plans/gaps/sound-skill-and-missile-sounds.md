# Gaps: sound-skill-and-missile-sounds

Spec: [Skill and missile sounds](../../specs/sound-skill-and-missile-sounds.md)

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|---|---|---|---|---|---|---|
| Active handler producer | integration | SMS-Q01, SMS-R01, SMS-R02 | actual state/skill eligibility and cadence; supplied handler remains available | complete state/list producer and selector provenance | 2026-10-29: state dispatcher and original slots identified, handlers and Skills descriptor bindings read; full state producer unclosed | trace eligible state through active handler with exact skill and selector inputs |
| Other skill phases | behavior | SMS-R03b, SMS-Q02 | start/do/progressive timing and variations | complete producer, scheduling and actual overlay/request execution contracts; supplied R03a remains available | 2026-10-29: progressive candidate read and corrected from cast-sound label to prgsound/prgoverlay consumer; lookup/stat/overlay dependencies remain | resolve flags and dependencies, identify ordinary start/do producers |
| Missile and state audio lifetime | behavior | SMS-R04d, SMS-Q03 | travel/hit/progression and state on/off execution, counts and stopping | request and stop paths from actual triggers | 2026-10-29: missile descriptor search and state schema reuse; creation tail B30F0, hit B1420, progression BB470 and state stages 0x73760/0x73630 sound stages traced; exact callback/collision admission, handlerdispatch, retirement deviceeffects remain | root-to-request-to-retirement traces for each phase |
| Audible request integration | integration | SMS-R02, SMS-Q04 | actual playback/voice effects; request log remains available | admission/device lifetime contract | 2026-10-29: common sound request identified; playback contract leaves admission/device stages gated | close playback contract for the requested emitter/cue |
| Selection observation | observation | SMS-Q05, SMS-O01 | runtime confirmation only | matching original-game capture | 2026-10-29: static stage producers located but matching eligible fixture and attributable capture not completed; no new observation performed | complete preparation and valid measured handler inputs/requests/results |
