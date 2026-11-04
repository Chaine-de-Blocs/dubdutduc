# Hash → MIDI composition rules

Every Bitcoin block becomes one fixed 16-bar chiptune. The tonic follows the circle of fifths by block height; the hash seeds everything else. The same hash and height always produce the same Standard MIDI File. Composition lives in `internal/music` (`Compose` → `Piece` → `MIDI()`).

## How hash and height are used

The **height** sets the tonic. The **hash** is a seed for a deterministic RNG that picks mode, groove, melody, and the rest. The composer walks a fixed pipeline and, at every fork, asks that RNG for the next number.

```
block height  →  Root = (7 × height) mod 12   (circle of fifths, C at height 0)

block hash
    │
    ├─ 32 hex bytes? → use as seed
    └─ otherwise → SHA-256(hash) as seed
    │
    ▼
ChaCha8 PRNG  →  stream of ints/floats
    │
    ▼
choose(mode), choose(progression), choose(rhythm), pick(pitch), …
    │
    ▼
MIDI notes (same hash + height ⇒ same piece)
```

Weighted picks use `IntN(totalWeight)`; uniform lists use `xs[IntN(len)]`; coin-flips use a float in `[0,1)`. Theory tables stay fixed; only which row wins depends on the hash.

### Draw order (approximate)

1. Mode, meter, BPM, ensemble (root comes from height, not the PRNG)  
2. Open / middle / cadence chord-degree rows  
3. Melody rhythm patterns  
4. Melody pitch branches and tie-breaks  
5. Comping style or arp direction/step  
6. Bass pattern, drum groove/fill, velocity jitter  

## 1. Seeding (determinism)

```
hash (hex, lowercased, trimmed)
        │
        ├─ if it is 32 valid hex bytes → use those 32 bytes as the seed
        └─ otherwise → SHA-256(hash string) as the seed
        │
        ▼
  ChaCha8 PRNG (math/rand/v2)
```

All “choices” below (except root) are draws from this PRNG. Theory tables (scales, progressions, rhythms, ensembles) are fixed in code; only which option is picked depends on the hash.

## 2. Global parameters (drawn once)

| Parameter | How it is chosen |
|-----------|------------------|
| **Root** | `(7 × height) mod 12` — consecutive blocks are a perfect fifth apart, anchored at C for height 0 |
| **Mode** | All 7 modes, weighted: Ionian/Dorian/Mixolydian/Aeolian 16, Phrygian 14, Lydian 12, Locrian 10 |
| **Meter** | Weighted: 4/4 (weight 3) or 3/4 (weight 1) |
| **BPM** | Uniform pick from `{100, 112, 124, 136, 148, 160, 172, 184, 196, 208}` |
| **Ensemble** | Uniform pick of lead / chords / bass chip instruments (drums always on) |

Timing uses **PPQ = 480** ticks per quarter note. One sixteenth = 120 ticks. A bar is `beats × 480` ticks.

## 3. Form: always 16 bars

Four phrases of 4 bars: **A · A · B · cadence**.

```
bars  1–4   phrase A   (open progression)
bars  5–8   phrase A   (same chords; melody mostly restated)
bars  9–12  phrase B   (middle / contrast progression)
bars 13–16  cadence    (last bar = BridgeChord: V / next block's tonic)
```

Harmony degrees are 0-based scale degrees. For each mode, the composer picks:

1. one **open** 4-chord row (used for phrases A and A again),
2. one **middle** 4-chord row (phrase B),
3. one **cadence** 4-chord row (phrase 4); the **final bar is replaced** by `BridgeChord`.

Because tonics walk the circle of fifths by height, the next block’s tonic is a perfect fifth above the current root. `BridgeChord` hangs on that pitch (diatonic V/v or V7 when stable; otherwise a major triad on the next tonic). The last melody note lands on the same pitch class, so block *n* opens into block *n+1*.

Example (Ionian open): `{0, 4, 5, 3}` → I–V–vi–IV.

Special case: if a bar is degree **IV** (4) and the next bar is **I** (0), and the IV seventh is a dominant seventh in that key, the IV chord is written as a **dominant 7th** (secondary-dominant color into I).

## 4. Chord voicing

Each bar gets a close-position voicing of the chord tones inside MIDI range **47–64**, preferring motion near center **55** and minimal movement from the previous bar’s voicing (sum of nearest-note distances + drift from center).

Those MIDI pitches become the harmony track’s note material (block chords or arpeggio).

## 5. Melody notes

### Rhythm

Per meter, the composer picks weighted **body** and **phrase-end** rhythms (lists of sixteenth lengths; negative = rest). Layout:

| Bars | Rhythm sources |
|------|----------------|
| 1–4 | `a, b, a, end` |
| 5–8 | same `a, b, a` with a new `end` — bars 5–7 **copy melody notes from bars 1–3** shifted by 4 bars |
| 9–12 | new `x, y, x, end` |
| 13–16 | `a, b, y, final` (weighted cadence rhythms; whole-bar hold is rare) |

### Pitch (per non-rest slot)

Melody stays on scale pitches in a window around middle C (roughly one octave+, keyed off the tonic). Contour state tracks previous pitch, last direction, unresolved leaps, and repeats.

For each attack in the bar:

| Condition | Pitch rule |
|-----------|------------|
| Last note of the piece | Next block’s tonic (perfect fifth up) |
| Downbeat of the bar | Chord tone (40% chance: avoid repeating previous pitch) |
| Long note (≥ 6 sixteenths) | Chord tone, avoid repeat |
| Strong beat (beat 3 in 4/4) and 70% roll | Chord tone |
| Last note of a non-final bar that fills the bar | **Approach tone**: scale step next to the next bar’s target chord tone |
| Otherwise | **Walk**: mostly step ±1 in the scale, sometimes repeat, skip, or leap to a chord tone; after a leap, resolve opposite |

Melody note lengths are shortened to **85%** of the rhythmic slot (articulation gap). Velocities swell through each 4-bar phrase (sine envelope), accent downbeats, and add a tiny PRNG jitter.

## 6. Harmony track (comping / arp)

- If the ensemble’s chord instrument is **chip-arp**: each bar (including the last) is filled with an up (70%) or down arpeggio of the bar voicing, step 2 sixteenths (35% chance step 1).
- Otherwise (**chip-chord**): a pulsed comping rhythm for the meter runs through every bar; every hit plays **all voicing pitches** together. Note length ≈ 70% of the slot.

## 7. Bass notes

A meter-specific pattern of root / fifth / rest slots (sixteenth lengths) for every bar, including the last. Root and fifth are placed in MIDI range **34–48**, nearest the previous root. Notes shortened to **92%** of the slot.

## 8. Drum notes (GM channel 10)

Pitches: kick **36**, snare **38**, closed hat **42**, open hat **46**.

- Body bars: weighted groove patterns for the meter.
- Every 4th bar (phrase end, not final): a fill pattern.
- Final bar: ending figure.

Velocity comes from the pattern plus a small PRNG offset (±2).

## 9. MIDI file layout

`Piece.MIDI()` writes a **format-1 SMF**:

1. Conductor track — title `Block <hash>`, tempo from BPM, time signature from meter.
2. Melody, harmony, bass — program changes from the ensemble; note on/off from composed notes.
3. Drums — channel 10 (index 9), no program change.

Ticks are absolute from the piece start; duration is in ticks as composed above.

## 10. What the hash does *not* change

These are fixed for every hash (only selections among them vary):

- Tonic (from height via the circle of fifths)
- Piece length (16 bars) and A–A–B–cadence form
- Available modes, meters, tempo list, ensembles
- Progression tables, rhythm catalogs, bass/drum patterns
- Registers (melody / chord / bass ranges)
- Melodic grammar (chord tones on strong places, leap resolution, approach into the next bar, final bridge to next tonic)

So the hash is a **seed for constrained musical choices**, not a direct bit→pitch map. Two different hashes diverge because the PRNG stream picks different rows from the same rulebook; they never invent pitches outside the scale/chord rules above.
