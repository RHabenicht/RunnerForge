RunnerForge — Windows 10/11 desktop acceptance

Use the RunnerForge-Windows-development artifact from a successful GitHub
Actions run. Extract it, then run RunnerForge-Setup-0.2.0.exe. Installation
is per user; an administrator account is not required.

Automated Windows Server 2022 checks cover native editor operations,
character round trips, PDF generation, startup/close, installation and
uninstallation. The following direct Windows 10/11 checks remain:

1. Install, launch from the Start menu, and inspect at 100%, 125% and 150%
   display scaling. Resize the main window and inspect both lower panels.
2. Create Priority and Build Point characters. Raise/lower the BP budget,
   choose a metavariant, and create/export/reload a custom race.
3. Open Edit / remove. Add and edit Active, Knowledge and Language skills;
   enter a creation specialization and check its base/specialized ratings.
4. Edit equipment prices and quantities. Add alpha cyberware and bioware,
   then compare spending, Essence, Bio Index and Magic with source values.
5. Save/reopen the character. Check notes, contacts, magic gear, sources,
   custom traits, magic choices and point allowances.
6. In the Character entry group, apply details and choose Begin career.
   Use karma to learn a skill, buy a specialization, and raise an attribute.
   Confirm the creation snapshot remains unchanged in the saved record.
7. Export HTML and PDF. Inspect all sections, Unicode names, long notes,
   page breaks and validation findings. PDF requires Microsoft Edge.
8. Import one of go/srcapp/testdata/arch*.sr3. The importer preserves original
   bytes; compare visible field mapping with the original character sheet.
9. Check unsaved-change prompts on replacement/exit and unapplied entry edits.
10. Uninstall, verify user saves remain, reinstall and reopen a saved character.

Do not equate passing these checks with complete SR3 rules coverage.
DEVELOPMENT_STATUS.txt records the remaining implementation and data audit.
TEST_REPORT.txt is a historical November 2026 baseline report, not the current
release acceptance record. Current automated evidence is in GitHub Actions.
