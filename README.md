# Tuidoist

A fast, efficient TUI todoist application for macOS, Linux, and Windows.

Inline help is available for key commands. Press `?` for more keys!

- Optimisticly updating UI (it's fast)
- Offline support with persistent storage (resyncs when you're back online)
- Most todoist features supported (Projects, filters, reschedule, etc)

![demo](assets/tuidoist.gif)

### Getting started

Download binary from [here](https://github.com/psethwick/tuidoist/releases).

Alternatively, you can use the go toolchain:

```bash
go install github.com/psethwick/tuidoist
```

On first startup it will ask for a todoist api key, which you can get from
[here](https://app.todoist.com/app/settings/integrations/developer).

By default your api key is saved in your home directory. If you prefer you can
set `TUIDOIST_TOKEN` environment variable instead.

### Thanks to the following projects:

- https://github.com/sachaos/todoist
- https://github.com/charmbracelet/bubbletea
- https://github.com/erikgeiser/promptkit
- https://github.com/charmbracelet/bubbles
- https://github.com/treilik/bubblelister
- https://github.com/mistakenelf/teacup
- https://github.com/charmbracelet/vhs

### Contributing

- Bug reports: please open an issue, will do my best to address quickly
- Feature requests.. will be considered, open an issue!
- Pull requests: open an issue for discussion first before larger or
  user-visible changes
