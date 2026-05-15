# daily-tasks

CLI to log time per project with subtasks and export to JSON.

## Install

```bash
pip install -e ".[dev]"
```

## Usage

```bash
daily-tasks --help
daily-tasks project add "My project" --subtask "Standup" --subtask "Coding"
daily-tasks project list
daily-tasks log
daily-tasks export --output export.json
```

Default database: `%LOCALAPPDATA%\\daily_tasks\\data.db` on Windows. Override with `--db PATH`.

```bash
daily-tasks project add "My project" -d "What this project is about"
daily-tasks project set-description "My project" -d "What this project is about"
daily-tasks config set-author --name "Hernan Gomez" --email "you@company.com" --sn-id "SE40634"
```