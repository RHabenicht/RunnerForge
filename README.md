# RunnerForge

Standalone Shadowrun Third Edition character generator for Windows 10/11. Version 0.2.0 is a development build, with remaining rules work listed in DEVELOPMENT_STATUS.txt.

## Character creation

- Priority and revised Companion Build Points, with an adjustable point budget.
- Standard metatypes and 14 Companion metavariants at their published costs.
- Custom races with editable names, point costs, six attribute modifiers and descriptive traits. Race definitions can be imported, exported and stored with a character.
- Active skill costs based on linked natural attributes; Knowledge and Language allowances; creation specializations and derived read/write ratings.
- Optional extra Knowledge points, Spell Point purchases/sales, and Edge/Flaw point accounting.
- Cyberware grade price/Essence calculations and bioware's effect on starting Magic.
- Equipment quantity and price accounting. Unknown prices remain unresolved rather than counting as free.

Choose **Edit / remove…** to edit a selected skill or possession. The group selector also opens character notes, karma, tradition and extra-point settings. Select **Character**, apply any edits, and choose **Begin career** to preserve a creation snapshot and enable karma advancement. Skills, specializations and attributes can then be raised one step at a time. New skills cost one karma at rating 1; additional specializations use the SR3 advancement costs. Each purchase records its cost.

## Files and Windows installation

JSON character files retain custom definitions and legacy import bytes. HTML and PDF sheets include all stored possession groups, notes, sources, karma and validation findings. PDF export uses a separate local Microsoft Edge process. Character data stays on the computer.

The Windows workflow builds the executable and a per-user Inno Setup installer. Download the **RunnerForge-Windows-development** artifact from a successful [Actions run](https://github.com/RHabenicht/RunnerForge/actions). The installer adds RunnerForge to the Start menu and offers an optional desktop shortcut. User-created files are preserved by uninstall.

CI tests use Windows Server 2022 and exercise native controls, PDF output, installer launch and uninstallation. Direct Windows 10/11 acceptance testing remains outstanding.

## Catalog and limits

The embedded catalog contains 6,449 records citing approved sources. Another 2,633 records remain quarantined. Source filtering does not establish that each record was verified against its book.

Custom racial traits and many catalog modifiers are reference text. Compound implants, augmentation stacking, advanced magic, special character types and full legacy-field fidelity still need implementation or verification. The application does not yet provide complete SR3 legality checking.

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
