# daily-tasks (Go)

Go port of the daily-tasks CLI using **Cobra** and **Bubble Tea**. Uses the **same SQLite database** as the Python app.

## Build

```powershell
cd go_version
go build -o daily-tasks.exe ./cmd/daily-tasks
```

Or run `.\build.ps1`.

## Default database

Same as Python:

- Windows: `%LOCALAPPDATA%\daily_tasks\data.db`
- Other: `~/.daily_tasks/data.db`

Override with `--db PATH` on any command.

## Commands

```text
daily-tasks --help
daily-tasks config set-author -n "Your Name" -e "you@company.com" --sn-id "SE40634"
daily-tasks config show-author
daily-tasks project add "My project" -d "Description" -s "Subtask A"
daily-tasks project list
daily-tasks project set-description -p 1 "New description"
daily-tasks project set-description -p 1 -d "New description"
daily-tasks project set-description "My project" -d "New description"
daily-tasks project set-default-subtasks -p 1 -s "A" -s "B"
daily-tasks log
daily-tasks log --date 2026-05-18
daily-tasks export
daily-tasks export -o export.json
daily-tasks export --all
daily-tasks export --date 2026-05-18
daily-tasks export --from 2026-05-01 --to 2026-05-31
daily-tasks ticket set-ritm 2026-05-18 RITM123456
```

## Interactive log wizard

`daily-tasks log` opens a Bubble Tea UI:

1. Multi-select projects (space / enter)
2. Ticket title and description (description defaults from author profile)
3. Per project: hours, subtask checklist (defaults checked, past days unchecked), optional extras
4. Task description is taken from the project record automatically

Use **Windows Terminal** or another modern terminal for best results. The wizard uses a **Tokyo Night Neon** lipgloss theme.

## Export behavior

- **Default** (`daily-tasks export`): exports **today only** (local date).
- **`--all`**: export every ticket in the database.
- **`--date YYYY-MM-DD`**: export a single day.
- **`--from` / `--to`**: inclusive date range. If you pass only `--from`, that **single day** is exported (not “from that day onward”).

## Tests

```powershell
go test ./...
```
