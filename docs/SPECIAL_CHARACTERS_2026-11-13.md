# Companion special-character implementation

Verified against the user's attached revised *Shadowrun Companion* PDF on November 13, 2026. References below are printed page numbers; PDF page numbers are two higher. The source PDF is not committed.

## Starting ghouls

Pages 13–14 and 33–34 govern starting-ghoul creation. The ordinary metahuman is purchased first, using Build Points. The ghoul surcharge is 10 BP in addition to the metatype's cost. Skill costs, Knowledge and Language pools use ordinary attributes before transformation, as the worked example on p.14 explicitly demonstrates.

The editor records actual successes on the third-stage Willpower (6) test. Starting characters receive two extra dice; Karma cannot be used. The p.33 success bands are applied in addition to the p.34 common changes. Only the common changes adjust racial limits. The p.34 Jenny example is a regression fixture: final B/Q/S/C/I/W are 6/5/4/1/2/3, Essence and Magic are 5, Reaction is 3, and racial limits are 11/6/9/4/4/6.

A result below 1 in Charisma or Intelligence remains an NPC outcome unless the GM approves the alternative. The player must explicitly name each two-point sacrifice. No purchase points are refunded or silently reassigned. Substitutions are frozen at career start so later advancement actually raises the final attribute.

Starting cyberware is assumed installed after transformation (p.33); its graded Essence charge is doubled. Starting Essence loses one point, which also reduces starting Magic. This implementation does not apply infection to an existing career character: that requires separate treatment of broken cyberware, absorbed bioware and the infection stages. Geasa and ongoing blindness/allergy/perception/diet/combat effects remain reference information.

## Shapeshifters

Pages 13–14 and 34–37 specify seven species and a 25 BP cost. Three animal physical purchases are stored separately from the six human/shared purchases. Both physical sets count toward the attribute allocation. Mental modifiers are shared. Skill costs use human-form attributes, including when animal form is selected. Knowledge allowance is Intelligence ×3 and Language allowance is Intelligence.

Resources are fixed at ¥5,000. BP resource purchases are disallowed; Priority creation requires Race C and Resources A (B for full magicians). Starting contacts must be purchased. Cyberware is rejected. Essence is 8, while awakened starting Magic is 6. Bioware/regeneration interactions remain a review item pending the applicable source rules.

The p.36 species modifiers supply form-specific attributes and the animal Reaction bonus. Eagle initiative applies only while flying; seal initiative only while swimming. Human form receives neither animal Reaction nor initiative bonuses. Animal form marks adept powers inactive. Powers' individual mechanical effects remain part of the outstanding general adept-power implementation.

Advancement charges the applicable form's attribute cost and raises only that form's physical purchase. The history identifies Human or Animal. Species, limit interpretation and direct animal purchase edits lock at career start; form and movement selection remain available. Both forms appear in exported sheets.

Regeneration, silver allergy/vulnerability, natural attacks/armor/reach, movement multipliers, social limitations and astral restrictions are carried as source-cited reference information. They are not a combat/condition simulation.

## Seal table conflict

The page images were inspected, not just OCR. On p.36, seal animal modifiers are Body +2, Quickness +1, Strength +1. On p.37, the animal racial limits are Body 7, Quickness 6, Strength 7. Thus the Body/Quickness limits do not equal 6 plus the printed modifiers.

There is no silent erratum. Seal configuration requires an explicit recorded GM choice:

- **Modifiers:** use p.36 statistics, deriving Body/Quickness limits 8/7.
- **Limits:** retain p.36 statistics and enforce p.37 limits 7/6; creation purchases must fit those limits.

Strength remains limit 7 in either case. The choice appears on the sheet and cannot change during a career. This is a campaign interpretation of conflicting tables, not a claim that either publication has been corrected by errata.

## Verification and remaining access

Regression coverage includes the ghoul worked example, infection bands, ordinary skill budgets, explicit substitutions, Essence depletion, implant grade interaction, career snapshots, all seven shapeshifter species, independent forms, Priority/BP resources, paid contacts, conditional initiative, separate advancement and JSON/HTML round trips. Windows acceptance exercises both native creation editors and animal advancement; PDF and installer checks run in the existing Windows workflow.

The Companion attachment was downloaded successfully. Both supplied Matrix attachments exceeded the executor's 32 MiB transfer limit before content could be read. Otaku rules must be verified from a smaller PDF extract of the applicable Matrix section before implementation; older Companion restrictions must not substitute for Matrix.

Verified implementation commits: ghoul `0295ff39027beb0646107a5fe8f7eb468be4ce55`, shapeshifter `95460acaa89bc5ed8df46cdb724c66c83a904494`. All local checks passed, including four Python tests. [Windows acceptance run 43444144321](https://github.com/RHabenicht/RunnerForge/actions/runs/43444144321) passed every step, including native editors, PDF export and installer install/launch/uninstall with save retention. This is Windows Server 2022 evidence; direct Windows 10/11 visual/DPI acceptance remains outstanding.
