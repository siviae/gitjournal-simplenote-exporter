# GitJournal Simplenote Exporter
This is an application for import notes from [Simplenote](https://simplenote.com/) to [GitJournal](https://gitjournal.io/). Deleted notes are exported as commited and then removed.
# Usage
```
git clone https://github.com/isae/gitjournal-simplenote-exporter && cd gitjournal-simplenote-exporter
go run main.go -input <path-to-zip-downloaded-from-simplenote> -output <path-to-git-repo>
```
Also, see ``go run main.go --help``
