# RunnerForge — Codex handoff

Prepared November 13, 2026. Repository: https://github.com/RHabenicht/RunnerForge (private).

## Start here

Continue this existing Go/Win32 project. Read this file, DEVELOPMENT_STATUS.txt, README.md, README_TESTING.txt and docs/EQUIPMENT_REVIEW_2026-11-12.md, then inspect the code and run the existing tests. Work until the requested features are complete; do not stop after each small increment to ask whether to continue. Report concrete blockers and actual test results. Do not claim complete SR3 coverage from a passing build.

The latest tested implementation commit is `2d38a8ed5bb859a2dceaf2516dcb7273f93ae63c`, now on main. This handoff is a later documentation-only commit. Clone main rather than rebuilding from a chat transcript. The former name was NSRCG-II; the product is RunnerForge. The separate VB project, NSRCG maintenance project, and The Shop are not this project and must not be edited.

## User requirements

- Standalone installable Windows 10/11 SR3 character generator.
- Adjustable creation point budgets and editable equipment/character values.
- Companion metahuman variants included at their printed point costs.
- Custom races with editable costs, modifiers and traits; import/export reusable definitions.
- The user has a recalculated metahuman-variant spreadsheet to import when located. It has NOT yet been imported. Do not assert that it does not exist.
- Use supplied sourcebooks. Preserve source/page references. Replace obsolete rules when a newer applicable source supersedes them; Rigger 3 Revised supersedes Rigger 2.
- Test the work yourself. Keep GitHub current. Do not request routine confirmations already covered by these instructions.

## Current implementation

The Go module is under `go/` (module name `nsrcgii`; retained for compatibility). Native Windows controls use Win32 syscalls, not a web wrapper.

`go/srcapp/` contains character calculations and persistence. Relevant files: core.go, build_points.go, races.go, skills.go, inventory.go, augmentation.go, magic.go, advancement.go, specializations.go, files.go, legacy.go and sheet.html.

`go/cmd/nsrcgii/` contains the GUI: main_windows.go, race_editor_windows.go, contents_windows.go, layout_windows.go and Windows-only acceptance tests. Tests on Linux do not execute these Windows files.

Implemented: Priority/Build Point methods; skill costs and free Knowledge/Language pools; creation specializations/literacy; 14 Companion variants; custom races; inventory prices/quantities/editing; spell-point purchases/sales; edge/flaw point accounting; cyberware grades/used prices; Bio Index; career snapshots; attribute/skill/specialization advancement with karma history; HTML/PDF export; atomic JSON saving; per-user Inno installer.

Save format remains `nsrcgii.v2`. Legacy bytes/text and warnings are preserved. Avoid destructive format migrations. Custom-race definitions are embedded in saves. Creation snapshots preserve pre-career state. Additional specializations have independent ratings.

Equipment update: 35 book-checked entries from Sprawl Survival Guide and Target: Awakened Lands, including 11 household drones and 10 Counterfeiting utility ratings. Total active entries: 6,484. See the equipment review for scope, source references, corrected drone statistics and the documented paint-table column transposition.

`equipment_supplement.json` is embedded alongside `catalog.json`; LoadCatalogue merges them and rejects ID collisions. Do not blindly enable all source codes in build_catalog.py. The original quarantine archive has 2,633 records; 24 now have reviewed active counterparts. Its raw size is not the unresolved count. Preserve old IDs. Missing prices remain validation issues rather than zero-cost gear.

## Verified build and installer

GitHub Actions run: https://github.com/RHabenicht/RunnerForge/actions/runs/43324415437

All steps succeeded on Windows Server 2022: native tests, Go vet, GUI build, window launch/close, installer build, install/launch/uninstall with user-save retention.

Artifact: https://github.com/RHabenicht/RunnerForge/actions/runs/43324415437/artifacts/13992191909

Artifact name: RunnerForge-Windows-development. Contains RunnerForge.exe and RunnerForge-Setup-0.2.0.exe. It expires February 10, 2027. SHA256 of artifact: c418fa32abb8d704fb168c273f682b9d1bb7894580530f73476e7629febbabf0.

Workflow: `.github/workflows/windows-save-smoke.yml`. Installer sources: installer/RunnerForge.iss and installer/test-install.ps1. Version 0.2.0 is a development build, not a completed rules release. Direct Windows 10/11 visual and high-DPI acceptance remains outstanding.

Standard checks from `go/`, using the Go version declared in go.mod:

```sh
go test -buildvcs=false ./...
go vet -buildvcs=false ./...
GOOS=windows GOARCH=amd64 go vet -buildvcs=false ./...
GOOS=windows GOARCH=amd64 go build -buildvcs=false -ldflags='-H windowsgui' -o RunnerForge.exe ./cmd/nsrcgii
```

On Windows, run `go test -timeout 3m -v ./...` to include native controls and Edge PDF export. From repository root, run `python -m unittest discover -s . -p 'test_*.py'`. Four importer tests passed. Linux Go tests, vet, Windows cross-vet/build and the Windows workflow passed for the tested commit. No Windows 10/11 manual visual test has been performed.

## Remaining work

1. Special-character creation: otaku and shapeshifters with separate forms and their actual rules. Starting-ghoul creation is implemented; career-acquired infection remains.
2. Compound implants and capacities; augmentation stacking and effective attribute modifiers; cultured bioware; equipment dependencies.
3. Magical eligibility, tradition/aspect restrictions, foci bonding, spirits, initiation and remaining edge/flaw effects.
4. Advancement for spells/foci, training, earned-karma tracking and karma-pool progression.
5. Vehicle/deck customization and availability/legality checks.
6. Book-by-book catalog validation and full legacy import fidelity. Do not treat quarantined data as verified just because it parses.
7. Locate/import the user's rebuilt-race sheet.
8. Windows 10/11 manual visual and DPI acceptance, then final release packaging.

Contacts/lifestyles now support explicit editable purchase prices and quantities in resource accounting, save/load, native editing and exports. Legacy contacts and lifestyle purchase totals are imported with original records retained; imported lifestyles use quantity 1 to avoid multiplying their saved total by prepaid months again. Missing prices require review. Automatic contact allowances, source-based prices, and detailed lifestyle rules remain unimplemented. Only RCT/INI equipment modifier strings currently affect derived calculations; other augmentation effects still need implementation. Custom race traits are descriptive. Advanced magic, otaku and shapeshifters remain unimplemented. Starting ghouls now have BP surcharge, recorded third-stage successes, explicit GM substitutions, pre-transformation skill costs/pools, common-only racial-limit changes, Essence/Magic loss, doubled post-transformation implant Essence, persistence, export and locked career choices. Companion printed pp.14 and 33–34 are verified from the attached PDF. Career infection, geasa and combat/condition effects remain unimplemented.

## Primary sources and access

The books were supplied through ChatGPT Library. They are NOT committed to GitHub. A separate Codex environment may not have Library access; use the same connected files when available or request the specific source files needed. Never fabricate table values. The equipment review document preserves the evidence used for this update without requiring the PDFs to build the app.

Known Library file IDs:

| Book | Library file ID |
| --- | --- |
| SR3 core | libfile_0bae032ee5e8819185f03ff84c2794e0 |
| Revised Companion | libfile_2a22506a5d448191bab0528aac00cde6 |
| Matrix | libfile_4943700068f881919a47338eb534969e |
| Man & Machine | libfile_f998b8d2e3588191802a3969b6289546 |
| Magic in the Shadows | libfile_8891205b37788191a87b7c8295177b7f |
| Cannon Companion | libfile_984823b5b1e88191afd570b1e979f028 |
| Rigger 3 Revised | libfile_5ee35bf21c648191b863c245fa66170f |
| State of the Art 2063 | libfile_79d41f07cda0819195b52f9829eab29a |
| State of the Art 2064 | libfile_77bbb47f30c88191b4d337a29d2d87d2 |
| Target: Wastelands | libfile_100611515ad88191a4aeca289290b152 |
| Year of the Comet | libfile_89223e0e577c8191a0c426c68cc464e0 |
| Sprawl Survival Guide | libfile_b175fbe4cc088191b7b149828212cf52 |
| Target: Awakened Lands | libfile_2246be676f0c8191b6d82c0e4ca2a224 |
| Critters | libfile_1284e4ede12081919831fea6b91f6d25 |

## Special-character research to resume

These are research leads, not completed behavior. Re-read the cited rules before implementation.

- Companion printed pp.33–34: starting ghouls use Build Points, +10 BP above base metatype; ordinary character built before transformation. Willpower TN6 with two extra dice determines additional changes; no Karma. Common adjustments +2 Body, +1 Strength, -1 Charisma, -1 Intelligence, -1 Essence, -1 Magic; running multiplier +1. Infection-test adjustments do not alter racial maxima. Cyberware installed after transformation costs double Essence. Handle attributes reduced to zero and explicit GM alternatives, rather than silently reallocating points.
- Companion printed pp.34–37: seven shapeshifters; separate purchased physical attributes for animal/human forms and shared mental attributes. BP cost25; resources5,000. Knowledge Int*3; Language Int. Essence8 but awakened Magic6. No free contacts/cyberware. Skill costs use human form. Printed seal animal-form modifiers and racial-maxima table appear inconsistent; resolve rather than guessing. Form-specific adept effects and initiative must be modeled.
- Matrix printed pp.136–138 supersedes older Companion otaku restrictions: any metatype, Otaku Priority A replacing Magic or +30 BP. No magical ability. Personal resources5,000; extra resource allocation improves tribal lifestyle and complex-form allowance. Mental maxima +1, physical maxima -1; optional weak-body tradeoff has extra constraints. Computer6–8, Etiquette<=4 with Matrix/Street specializations only. Channels Access/Control/Index/Files/Slave linked to Willpower; free points ceil((Int+Will+Cha)/3), sorted starting maxima6/5/4/3/3. Persona formula table needs image inspection because OCR was incomplete. Free basic datajack/ASIST still have Essence costs. Path and complex-form rules need complete implementation.

## Previous environment details

The prior working copy was `/workspace/scratch/b58601a50dbd/nsrcg` and was not a git checkout. GitHub commits were created through the connector. A downloaded Go compiler was at `/workspace/scratch/b58601a50dbd/toolchain/go/bin/go`. These paths are session-specific; do not assume they exist in a new Codex environment. The authoritative durable code is on GitHub.
