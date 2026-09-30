---
name: mofumouse-animal-animation
description: Produce and resume MofuMouse photoreal animal cursor assets using built-in ImageGen, Google Flow video, transparent frame export, coat sheets, retiming and Windows playback checks. Use for this project's animal motions and coat variants, not unrelated animal illustrations or global Codex configuration.
---

# MofuMouse animal animation

This is a project-owned Skill. Resolve `../../..` from this directory to the
MofuMouse repository root. Keep global/provider Skills and referenced projects read-only.

Read the latest user instruction, root AGENTS and the relevant task ledger first.
Then read only the needed stages of the detailed
[production workflow](../../../docs/workflows/animal-cursor-production.md).
The current [species/ownership plan](../../../docs/animal-cursor-production-priorities.md)
contains priorities, not permission to start every future production wave.

## Choose the smallest route

| Request | Route |
|---|---|
| New species or realistic remake | Workflow stages 1–4, then 7 if Windows output is requested |
| Coat variants of an accepted motion | Stages 0, 5 and 6; reuse accepted poses |
| Faster/slower playback | Stage 6 only; no ImageGen or Flow calls |
| Run a completed animal on Windows | Stage 7; verify sources before switching |
| Resume, inventory or prioritize | Stage 0 and the current ledger; no speculative regeneration |

## Current production contract

- Use natural photographic anatomy, fur and eye proportions. AnimalsDesktop and
  LINE assets can inform species traits and verified colors; their illustration style
  and old low-resolution pixels do not define this new artwork.
- Generate each species as its own animal. A recolored degu is not a chinchilla.
- Built-in ImageGen is the image-generation/editing route. Read its current provider
  Skill and inspect local references with `view_image` before using them. Never
  silently substitute an API, model runner or procedural animal drawing.
- The user accepted Flow-video-derived motion. Keep video-frame provenance distinct
  from ImageGen provenance. Individual-pose editing is the default. The replacement
  chat-supplied AGENTS on 2026-09-29 explicitly permits a bounded four-pose edit
  trial under workflow stage 5A: one 2x2 sheet of existing same-motion poses, one
  image at a time. Record the current parent stage/cap before using this exception;
  it does not renew consumed caps, authorize larger sheets or adopt the method.
  If a later direct chat instruction forbids multi-pose generation, follow that
  instruction and clarify before another multi-pose call.
- Existing accepted eight-cell blue/sand sources and four-pose pilot evidence remain
  reusable history. Preserve their actual generation counts and cell-to-source
  mappings; do not relabel those outputs as individually generated poses.
- Preserve accepted originals and source hashes. Make new revisions for deliberate
  appearance changes; copy built-in outputs into the repo and keep generated originals.
- Code may decode, extract alpha, crop, pack, normalize, export and retime according
  to the selected pipeline. Do not draw/overlay replacement limbs or synthesize
  a gait by copying one pose and moving the whole body up and down.
- Use one common crop/scale/anchor per motion. Do not resize or recenter each frame
  to hide body drift. Check the actual feet and anatomy, not only IoU.
- Review source size and 64/96px, light/dark backgrounds and the last-to-first seam.
  Exactly two hind legs; the small bent far-side hind leg under the belly must move
  through front/middle/rear positions without becoming an extra paw.
- Keep an asset's real frame count, timestamps, alpha and generation count explicit.
  A build or native render count is not visual approval.

## Reliable reuse

The existing scripts are documented with their actual scope in the workflow.
Several are degu-specific and have fixed paths/intervals; do not run them for a new
species by pretending they accept generic options. Extend or add a species-scoped
helper only after inspecting its inputs and callers.

For the existing degu, playback speed is **relative to the original video**:
1.0 → 1.5 → 2.25. The current 30-pose native/APNG cycle is 556ms; the separate
eight-pose blue GIF is 560ms. These are an example's accepted values, not defaults
for all species or idle animations. Prefer active manifests over stale prose.

Use [prompt patterns](references/prompts.md) when preparing a new generation and
[behavior checks](references/validation-cases.md) when changing this Skill or flow.
The read-only `scripts/check_animal_cursor_state.py` checks the selected degu and
optional coat manifest; it does not grant runtime promotion or visual acceptance.

## Ownership and completion

Use the user-authorized existing chat for a delegated coat batch. Give concrete
inputs, output directories, script ownership, retry bounds and acceptance criteria.
Other workers share the checkout: do not revert their edits or race shared manifests.
An explicit pose/revision/call cap remains in force across turns, Goal continuations
and compactions. A generic instruction to continue the larger Goal does not renew
a consumed generation cap. After a stop-on-failure result, save and audit the
existing evidence; a further ImageGen/Flow call needs a new bounded instruction
from the coordinating parent. The parent may issue that instruction within the
user-authorized production scope without requesting the same user approval again.
Read status with compact thread snapshots; distinguish sent, running and completed.
If sending is ambiguous, inspect state before retrying the same request.

A generation pass ends with saved originals/copies, prompt and hash provenance,
transparent outputs, a sequence review and an honest ready/candidate/rejected state.
For an explicitly requested Windows result, also build, perform the native smoke
check, replace only the owned prototype and verify the launched path/process.
Keep generated candidates, source acceptance and runtime promotion separate.
