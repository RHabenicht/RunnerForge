# RunnerForge

Standalone Shadowrun Third Edition character generator for Windows 10/11. Version 0.2.0 is a development build, with remaining rules work listed in DEVELOPMENT_STATUS.txt.

## NSRCG (New Shadowrun 3 Character Generator) – Community Update

RunnerForge is a standalone community redevelopment of the classic NSRCG character creation and management utility for Shadowrun Third Edition, maintained by **NightLife (RHabenicht)**. It is an independent Go-based application, not a claim of ownership of the original Visual Basic program or its data.

### What's new in RunnerForge

- Modern native Windows application and per-user installer, with optional desktop shortcut and preservation of user-created saves on uninstall.
- Priority and adjustable Build Point creation, standard metatypes, 14 Companion metavariants, and an optional external NightLife pack containing all 16 rebuilt races (see [race-pack documentation](race-packs/README.md)).
- Custom race editor, reusable/importable/exportable race definitions, attribute modifiers and editable point costs.
- Starting ghoul transformation and seven shapeshifter species with separate physical forms, recorded GM decisions and advancement.
- Active, Knowledge and Language skill accounting; specializations, literacy, optional Knowledge points, Spell Points and Edge/Flaw accounting.
- Cyberware grade price and Essence calculations, Bio Index and starting Magic calculations.
- Editable possessions, equipment quantities and prices, contacts and prepaid lifestyles; unresolved prices are flagged.
- Career snapshots, karma accounting and stepwise skill, specialization and attribute advancement with purchase history.
- Legacy character import with retained source records, JSON saves and HTML/PDF character-sheet exports.
- Catalog filtering, 35 checked equipment additions, source documentation and automated regression, native UI and installer tests.

### Licensing & copyright notice

Original RunnerForge code, scripts, original documentation and independently authored modifications owned by NightLife (RHabenicht) are made available under the [MIT License](LICENSE). The grant covers only material for which the project maintainer holds the necessary rights. It does not relicense legacy NSRCG material, third-party code, catalog records derived from third-party materials, artwork, trademarks or Shadowrun publications. Existing third-party license and attribution requirements remain in force.

The original NSRCG was developed by **McMackie**, building on earlier work by **Paolo Marcucci**, with contributions from numerous community members. Their rights and credits are retained. The [historical NSRCG site](https://nsrcg.neocities.org/) documents the original credits and conditions for obtaining and distributing its source code. The original executables, Visual Basic components, base DAT files and other legacy material are not granted an MIT license by this project.

Attempts to establish communication concerning the legacy project were made but did not result in contact or a grant of permission. This is a record of the attempt, not authorization to redistribute or relicense legacy files. Material without established redistribution rights must not be included in public release packages. A release-content audit is required before distribution.

Shadowrun and its associated intellectual property remain the property of their respective rights holders. RunnerForge is an unofficial community tool, not a replacement for the published rulebooks, and is not affiliated with or endorsed by the rights holders.

## Character creation

- Priority and revised Companion Build Points, with an adjustable point budget.
- Standard metatypes and 14 Companion metavariants at their published costs.
- Custom races with editable names, point costs, six attribute modifiers and descriptive traits. Race definitions can be imported, exported and stored with a character.
- Active skill costs based on linked natural attributes; Knowledge and Language allowances; creation specializations and derived read/write ratings.
- Optional extra Knowledge points, Spell Point purchases/sales, and Edge/Flaw point accounting.
- Cyberware grade price/Essence calculations and bioware's effect on starting Magic.
- Equipment quantity and price accounting. Unknown prices remain unresolved rather than counting as free.
- Editable contact purchases and prepaid lifestyles count toward resources. Set a contact's price to zero for an explicitly free contact; a blank price requires review. Lifestyles use price per paid period and quantity, or quantity 1 for a lump-sum purchase. Prices and free-contact eligibility require source/GM review; no automatic allowance is applied.

Choose **Edit / remove…** to edit a selected skill or possession. The group selector also opens character notes, karma, tradition and extra-point settings. Select **Character**, apply any edits, and choose **Begin career** to preserve a creation snapshot and enable karma advancement. Skills, specializations and attributes can then be raised one step at a time. New skills cost one karma at rating 1; additional specializations use the SR3 advancement costs. Each purchase records its cost.

Starting ghouls: choose Build Points, build the ordinary character, then open **Edit / remove… → Ghoul creation**. Enter the actual third-stage Willpower (6) test successes (two extra dice, no Karma). The editor offers the explicit GM-approved two-point attribute alternative. Purchases remain stored before transformation; final attributes, racial limits, Essence, Magic and skill costs use the applicable rules. Starting cyberware is assumed installed after transformation and costs double Essence. Choices lock when beginning a career. Career-acquired infection, geasa and combat/condition effects are not automated.

Shapeshifters: open **Edit / remove… → Shapeshifter creation**, choose a species and buy its animal physical attributes separately. Main-window attributes are human-form purchases. Select the active form and movement there; both forms are saved and exported. Skill costs use human form. Resources are fixed at ¥5,000; cyberware and free starting contacts are disallowed. After career start, **Raise attribute** advances the active form only. Seals require an explicit GM interpretation of the conflicting printed tables. See [source decisions and implementation limits](docs/SPECIAL_CHARACTERS_2026-11-13.md).

## Files and Windows installation

JSON character files retain custom definitions and legacy import bytes. HTML and PDF sheets include all stored possession groups, notes, sources, karma and validation findings. PDF export uses a separate local Microsoft Edge process. Character data stays on the computer.

Legacy contacts and lifestyles are available in **Edit / remove…** and on exported sheets. Imported lifestyles retain their saved total as one purchase, avoiding a second multiplication by prepaid months. Missing or invalid saved prices remain unresolved. Original records remain attached to each imported entry.

The Windows workflow builds the executable and a per-user Inno Setup installer. Download the **RunnerForge-Windows-development** artifact from a successful [Actions run](https://github.com/RHabenicht/RunnerForge/actions). The installer adds RunnerForge to the Start menu and offers an optional desktop shortcut. User-created files are preserved by uninstall.

CI tests use Windows Server 2022 and exercise native controls, PDF output, installer launch and uninstallation. Direct Windows 10/11 acceptance testing remains outstanding.

## Catalog and limits

The embedded catalog contains 6,484 active records, including 35 individually checked additions. The original quarantine archive contains 2,633 records, of which 24 have reviewed active counterparts. Source filtering does not establish that each record was verified against its book.

Custom racial traits and many catalog modifiers are reference text. Compound implants, augmentation stacking, advanced magic, otaku and full legacy-field fidelity still need implementation or verification. The application does not yet provide complete SR3 legality checking.

## Development

Build with Go using the version in go/go.mod. Python 3 is required for catalog tools. Inno Setup 6 builds the installer on Windows.

```sh
python -m unittest discover -s . -p 'test_*.py'
cd go
go test ./...
go vet ./...
go build -ldflags="-H windowsgui" -o RunnerForge.exe ./cmd/nsrcgii
```

From the repository root, recheck the existing catalog and quarantine while preserving IDs:

```sh
python build_catalog.py --catalog go/srcapp/catalog.json
```

Rules implemented in this checkpoint use SR3 pp. 54–59 and 244–245, revised Shadowrun Companion pp. 13–16 and 38–43, and Man & Machine pp. 44–45 and 78. See DEVELOPMENT_STATUS.txt for validation coverage and remaining work.

## Acknowledgments

Thanks to the maintainers of these projects for their work on Shadowrun Third Edition character tools.

- **finsterdexter/sr3data** provided a reference for understanding the original NSRCG data formats and catalog structure. [sr3data repository](https://github.com/finsterdexter/sr3data)
- **apampuch/PySRCG** provided an implementation reference for SR3 character creation and character data modeling. Calculations require verification against the primary sourcebooks. [PySRCG repository](https://github.com/apampuch/PySRCG)

The November 5, 2026 project audit recorded PySRCG as MIT-licensed and found no declared repository license for sr3data. The audit and integration notes document how these references informed development.
