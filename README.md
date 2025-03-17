# Hisame (氷雨)

Hisame is a TUI tool to help you manage your AniList.

# Todo

- Fix the init flow.  Currently network calls are being made to init the AniList client before the TUI renders.  Ideally this would have some sort of loading modal while it happens.  Requires all the init stuff to be refactored to be async and use messages to report on results.
- Fix title width in login view.  Margins probably messing with it.  Title is a little shorter than the borders around the content.
- Make the help screen scrollable when it has too much text
- Use proper fuzzy finding for search query filter
- Make anime list columns more dynamic and reactive to available width.  Allow some to shrink (to a minimum), and assign a priority to them for order to cull when there's not enough space.
- Allow user to specify which columns are shown in the list view