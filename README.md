# harbour-whisperfish-message-restore

Script to restore signal messages from Signal Desktop to [Whisperfish for SailfishOS](https://whisperfish.rubdos.be/).

## Limitations

Currently only text-only private messages are restored.
* Attachments are ignored.
* Group Messages are currently also not copied. Most of the code has been prepared, but I am having issues with
  messages from senders that are no longer group members.
* This script depends on the database format of both, signal desktop and whisperfish. Updates of either
  may break functionality here.

## Usage

Before starting, ensure that whisperfish on your phone is not running anymore.
Disable background mode and the autostart systemd service and quit the application.

1. Export signal-desktop sqlite DB using [tbvdm/sigtop](https://github.com/tbvdm/sigtop): `sigtop export-database signal-desktop.sqlite`
2. Acquire the whisperfish sqlite DB
    - If running this script directly on the phone, the DB can be found in `.local/share/be.rubdos/harbour-whisperfish/db/harbour-whisperfish.db`
    - If running on a desktop, copy the DB (e.g. via SCP)
3. Run the script: `./harbour-whisperfish-message-restore --signal-desktop-file=./signal-desktop.sqlite --whisperfish-file=./harbour-whisperfish.db`
4. Copy the whisperfish sqlite DB back to it's original location
