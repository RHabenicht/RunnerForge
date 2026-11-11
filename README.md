# RunnerForge

Standalone Shadowrun Third Edition character generator for Windows 10/11, currently in development.

The Go application supports basic Priority and Build Point character creation, local JSON saves, legacy .sr3 imports, and HTML export. Comprehensive rules validation, a finished installer, and Windows runtime acceptance testing remain pending.

Build Point mode uses the revised Shadowrun Companion (pp. 13–14) for the standard metatypes, magic choices, resource tiers, and attribute costs. Active skill costs use linked attributes from the approved catalog. Optional point adjustments and full character eligibility checks remain pending.

The embedded catalog contains 6,449 records citing approved sources. Another 2,633 records are preserved in a separate review file. Source filtering does not establish that each record has been verified against its book.


## Metavariants and custom races

Build Point mode includes the 14 metavariants in the revised Shadowrun Companion (pp. 38–43). The published cost is 5 BP above the parent metatype, for a total of 10 BP for dwarf and ork variants or 15 BP for elf and troll variants.

To create a house-rule race, select a starting metatype and choose **Edit race…**. Change its name, Build Point cost, six attribute modifiers, and descriptive traits, then choose **Use race**. **Cancel** discards edits. **Export race…** saves a reusable JSON definition; **Load race…** imports one. Use a distinct name for a house-rule version of a built-in race. The example below prices a custom Night One at the same 10 BP as an elf.

```json
{
  "name": "Night One (house rules)",
  "build_point_cost": 10,
  "modifiers": {"Quickness": 2, "Charisma": 2},
  "traits": "Low-light vision, Mild sunlight allergy"
}
```

Enter the complete racial modifiers as adjustments to base attributes. Omitted attributes have a zero adjustment. RunnerForge includes the chosen race cost in the BP total and stores the imported definition with the character. Racial traits appear as reference text in the summary and HTML sheet.

## Development

Requires Go 1.23 or later and Python 3.

Run tests from the repository root.

```sh
python -m unittest discover -s . -p 'test_*.py'
cd go
go test ./...
go vet ./...
```

Build from the go directory on Windows.

```sh
go build -ldflags="-H windowsgui" -o RunnerForge.exe ./cmd/nsrcgii
```

Recheck the catalog from the repository root.

```sh
python build_catalog.py --catalog go/srcapp/catalog.json
```

This reads both the embedded catalog and go/data/quarantine.json, preserving existing record IDs.

The Windows CI suite runs Go tests, vet, an executable build, and native window startup/close checks. It also exercises custom-race editing and character save/load through native Windows controls. These checks use Windows Server 2022; Windows 10/11 desktop acceptance and PDF export testing remain outstanding.

DEVELOPMENT_STATUS.txt describes the latest checkpoint. Other historical reports describe earlier development stages.

## Acknowledgments

Thanks to the maintainers of these projects for their work on Shadowrun Third Edition character tools.

- **finsterdexter/sr3data** provided a reference for understanding the original NSRCG data formats and catalog structure. [sr3data repository](https://github.com/finsterdexter/sr3data)
- **apampuch/PySRCG** provided an implementation reference for SR3 character creation and character data modeling. Calculations require verification against the primary sourcebooks. [PySRCG repository](https://github.com/apampuch/PySRCG)

The November 5, 2026 project audit recorded PySRCG as MIT-licensed and found no declared repository license for sr3data. The audit and integration notes document how these references informed development.
