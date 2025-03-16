# Hisame (氷雨)

Hisame is a TUI tool to help you manage your AniList.

# Todo

- Fix the init flow.  Currently network calls are being made to init the AniList client before the TUI renders.  Ideally this would have some sort of loading modal while it happens.  Requires all the init stuff to be refactored to be async and use messages to report on results.
- Fix title width in login view.  Margins probably messing with it.  Title is a little shorter than the borders around the content.