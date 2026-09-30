# MofuMouse Agent Rules

## Current Production Workflow

- Direct instructions supplied in the current chat take precedence over this file. If the chat-supplied AGENTS forbids multi-pose generation while this file records an earlier four-pose exception, do not infer a new override: use individual-pose generation and seek clarification before another multi-pose call. Preserve existing pilot outputs and continue independent Flow/QA work.
- For this project's animal motion, coat, export and playback work, read `.agents/skills/mofumouse-animal-animation/SKILL.md` and the relevant stages of `docs/workflows/animal-cursor-production.md`.
- The user-approved motion pipeline is photoreal identity → Google Flow motion → directly decoded transparent PNG sequence/APNG/GIF. Individual-pose editing remains the default. A later explicit user instruction on 2026-09-28 authorizes a four-pose-per-image trial and workflow; use the bounded pilot in workflow stage 5A. This does not authorize arbitrary larger sheets or imply pilot acceptance.
- Use AnimalsDesktop and LINE as species/color references. New animals may be remade as realistic photographic animals; do not inherit low-resolution pixels or turn one species into another by recoloring.
- Keep video-derived source records and ImageGen source records distinct. Record source hashes, actual frame times and the actual generation count. Technical comparison sheets assembled from existing frames are not ImageGen multi-pose generation.

## Asset Rules

- ImageGen production assets must be generated one image at a time.
- Generate new coat poses individually by default. Exception: the user-authorized four-pose edit trial in workflow stage 5A may use one 2×2 sheet of existing poses. Preserve its original output, source-to-cell mapping and actual call count. Promote the method only after geometry, anatomy, alpha and color-continuity review; do not treat four samples as a complete 30-pose walk.
- Do not use procedural drawing, manual limb overlays, or synthetic leg patches for production mascot sprites.
- For video-derived motion, save the original video and decoded source PNGs with frame indices/timestamps under `assets/source/video-cycle/`. Preserve existing `assets/source/walk-cycle/` material as historical ImageGen work.
- If an ImageGen frame has a visible problem, regenerate that single frame as a new revision. For defective video motion, revise the video input/prompt and select a new valid cycle. Do not patch limbs by hand or repeat unchanged failures indefinitely.
- Keep selected files as copies in the repo and leave the original generated files in `C:\Users\user\.codex\generated_images\...` untouched.
- Review walk frames as a sequence before accepting them. Both rear legs must visibly move, and no leg should look oversized or pasted on.
- Review accepted walk frames at the app runtime size as well as source size. The far-side rear leg must remain readable at 64px.
- The far-side rear leg must have a clearly readable stride arc. Its paw position needs obvious front, middle, and rear extremes across the accepted walk cycle.
- Judge readable gait and continuity together. Do not discard accepted consecutive video frames merely to force an eight-frame sheet; report an eight-pose coat preview separately from the full video motion.
- The far-side rear leg means the small bent hind leg visible under the body, like the user's cropped reference: it curves down from the belly and lands with a small paw. Animate that specific leg, not only the foreground hind leg or a generic dark extra limb.
- Walking frames must show exactly two rear legs total: the foreground rear leg and the far-side rear leg. Do not accept a frame where the hindquarters appear to have a third rear leg, extra paw, or duplicate partial limb.
