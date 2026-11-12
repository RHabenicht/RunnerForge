# Equipment review — November 12, 2026

RunnerForge now loads 6,484 active catalog entries: the 6,449 legacy-filtered entries plus 35 individually verified equipment entries. Source: the user's newly supplied Sprawl Survival Guide and Target: Awakened Lands PDFs. The newly supplied Critters contents and searchable text contain creature rules and statistics, with no equipment purchase tables identified.

## Added entries

| Source | Entries | Reference |
| --- | ---: | --- |
| Sprawl Survival Guide household drones | 11 | pp.121–123; statistics and prices p.122 |
| Counterfeiting utility, ratings 1–10 | 10 | pp.124–125; SR3 pp.223 and 304 for size/pricing |
| Electronic paper: standard 20 Mp and tabloid 50 Mp | 2 | p.123 |
| ColorChange paint, one liter; paint controller | 2 | p.123 |
| Telecom drone-control package; credstick beacon signaler | 2 | pp.120, 123 |
| Altyerre, Kuman-Nhepa, Wudu'aku | 3 | Target: Awakened Lands p.109 |
| Unenchanted churinga | 1 | Target: Awakened Lands p.108 |
| Raw/refined Australian plant and animal enchanting materials | 4 | Target: Awakened Lands p.108 |

## Verified corrections and decisions

- Preserve all 24 recovered legacy identifiers (11 drones, 10 utility variants and 3 compounds), so saved catalog references remain stable. New records use IDs 100001–100011.
- Correct V.R. House Helper's display name to VOR Robotics House Helper, as named in the descriptive text; retain the old table name and similar models as aliases in its fields.
- Preserve the printed `2 (4)` Pilot values for the Butlerbot and Barback. Include fuel capacity, economy, chassis and the other table columns missing from the old import.
- Paper costs 0.1 nuyen per Mp: 2 nuyen for a standard 20 Mp sheet, 5 for a 50 Mp tabloid sheet.
- Paint is sold by the liter. At approximately 10 square meters per liter and 75 nuyen per square meter, the one-liter entry costs 750 nuyen. The controller costs 2,000 nuyen including setup.
- The paint/controller rows on p.123 place numeric street-index values under Availability and time-based values under Street Index. The catalog records 4/3 days and 4/7 days as availability, with street indices 1.5 and 1 respectively. This is a documented interpretation of the transposed table cells, not a claim of published errata.
- Counterfeiting has multiplier 10. Base size is rating squared times 10 Mp; standard SR3 program price/availability bands apply. J is an available option, not preinstalled. Saved possessions receive the actual selected utility rating.
- The drone-control package, beacon signaler, compounds, churinga and enchanting materials have no fixed price in their cited passages. They remain explicitly unpriced and trigger the existing price-required validation; they are not silently free.
- Existing core-book telecom entries remain available. The new consumer drone-control package is a separate accessory, not a second telecom purchase.
- Magical compounds retain their vectors, benefits and drawbacks as reference notes. Adding one to inventory does not permanently modify character attributes or automatically consume it. Unenchanted churinga has no automatic focus benefit.

## Catalog maintenance

`go/srcapp/equipment_supplement.json` is the checked source layer loaded alongside the generated legacy catalog. It survives a legacy catalog regeneration. `LoadCatalogue` rejects colliding IDs rather than silently duplicating equipment.

`go/data/quarantine.json` remains the original legacy archive; 24 of its records now have checked counterparts in the active supplement. Its raw row count is therefore not the number of still-unresolved records. No blanket approval was added to the legacy importer for either book: future unreviewed rows remain quarantined.

Only equipment is covered by this update. New spells, totems, metamagic and detailed lifestyle rules in these books are outside this equipment pass.

## Validation

Regression tests cover catalog search and source filtering, all recovered IDs, utility rating/price bands, purchase totals, paint quantities, explicit unknown-price validation, save/load preservation and duplicate-ID rejection. Run the standard Go and Python suites plus Windows build/vet and the Windows acceptance workflow before release.
