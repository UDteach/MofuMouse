# Prompt patterns

Use an exact reviewed image as the pose/identity authority. Replace species and
color descriptions with the current brief; do not copy degu body proportions to
other animals. These are task-local patterns, not fixed model/API contracts.

## Photoreal identity for Flow

Generate one natural photographic [SPECIES, BASE COAT], full body facing right,
almost side profile with the far-side feet readable. Natural eye and ear size,
species-correct torso, feet and tail, realistic fine fur and soft neutral studio
lighting. A relaxed four-foot stance, generous margin around whiskers, ears, paws
and the entire tail. A perfectly uniform light neutral gray background, no ground
texture, horizon, props, floor shadow, text, border or extra animals. The reference
animal is for species traits only; recreate it as a real animal photograph, not
pixel art, plush, cartoon or an enlarged old sprite.

When a transparent identity is needed, request true alpha with the built-in tool
instead. Transparent output is not itself a safe Flow input: the earlier idle
trial developed a dark halo. Prepare and inspect a gray-background derivative via
the authorized ImageGen editing route if that artifact occurs.

## Flow walk

Animate the one [SPECIES] in the start image. Locked-off camera, full body facing
right, fixed horizontal position and scale. Continuous relaxed walking in place
for the clip, enough repeated movement to select a complete cycle from the middle.
Natural planted support and lifted recovery feet, exactly two front and two hind
legs. The far-side hind paw beneath the belly visibly moves forward, passes through
the middle and moves rearward. Preserve the same animal's head, ears, eyes, coat,
proportions and tail. Small natural weight shifts only; all paws and the entire
tail stay in frame. Uniform light gray background and constant illumination.
No camera motion, zoom, turning, travelling across the frame, cuts, fades, text,
added limbs or objects.

For rabbits or another distinct gait, rewrite the locomotion description for that
species. Do not force the degu's alternating walk or cycle duration onto it.

## Flow idle

The same animal remains in a calm alert pause, a near-hold continuation of its
reference stance. Feet stay planted, body position and scale remain constant.
Subtle natural chest/belly breathing, tiny nose/ear motion and an occasional blink.
Maintain photographic identity, coat and complete tail. No walking, hopping,
large head/body bobbing, camera movement or background change.

Start/end images may help an idle hold but do not guarantee a seamless loop. Check
actual blink count, body drift and the seam. Keep the idle at a natural pace,
independently of the user's selected walking speed.

## Single-pose coat edit

Edit reference 1 into one true-alpha PNG of exactly ONE animal in ONE pose.
Reference 1 is the exact pose master. Preserve its silhouette, head, paw positions,
tail curve, scale, lighting and anchor. Change only the fur to [VERIFIED COAT
DESCRIPTION]. Reference 2, if provided, defines coat color/pattern only. Maintain
natural eyes and species-appropriate ears/feet unless the approved coat reference
requires a specific eye color. Exactly four limbs, two hind legs total; no
additional paws. Retain photoreal texture and the same individual. Complete
feet/ears/tail, clean true-alpha edges. No sheet, grid, multiple poses, text,
numbers, checkerboard pixels, shadows, scenery or extra animals.

The historical successful eight-pose prompt remains in
`assets/source/animal-cursor-coat-sheet-pilot/degu/blue/v1/generation.prompt.txt`.
The eight-pose sheet is historical, not a newly authorized route. Individual poses
remain the default. For the current explicitly bounded four-pose exception, use
workflow stage 5A and adapt the prompt to exactly the four existing cells; do not
leave contradictory ONE-pose/no-sheet instructions in a four-cell request. Keep
the exact source-to-cell mapping, fixed boundaries and shared transform. For a
patterned coat, lock each marking's location in body coordinates across calls;
pose preservation alone is insufficient.
