# Changelog

All notable changes to Hisame will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2025-05-11

### Added
- Added a menu system to make actions more discoverable.  Press 'enter' on the anime list view screen to see it
- Anime list view menu supports playing next episode
- Anime list view menu supports selecting an episode
- Anime list view menu supports showing anime details
- Anime list view menu includes an option to refresh anime data
- Anime list view menu includes an option to close the menu
- Anime list view menu includes an option to quit the application

### Changed
- MPV window title now shows the episode number & anime title by default instead of part of the episode URL
- Default status filters are now current + repeating, instead of just current
- Pressing 'enter' on the anime list menu no longer plays the next episode.  It now opens a context menu

## [0.2.2] - 2025-04-13

### Fixed
- Now more selectively passes key presses to the viewport of the anime details screen.  No more surprising undocumented behaviour

## [0.2.1] - 2025-04-13

### Added
- Important keybinds now shown at the bottom of each view
- Add a 'details' view for each anime to see more information about it

## [0.2.0] - 2025-04-09

### Added
- Help screen is now scrollable, supporting up/down, pgup/pgdown and home/end, as well as mouse wheel scrolling

### Changed
- Made keybind handling more consistent across the application with a central place to define them all.
- Some keybindings have been updated:
  - On the anime list view, 'enter' will now play the next episode for the highlighted show.  'p' still works.
  - On the anime list view & episode search view, '/' can be used to enter search mode.  ctrl+f still works.
  - When in search mode, 'esc' now consistently exists search mode and cancels the search you had.  ctrl+f also does this
  - When in search mode 'enter' now consistently applies the current search and moves focus back to the list.
- Help screen content updated

### Fixed
- Initialisation error which never removed the initial loading model from the display stack.  This meant you could press 'esc' on the auth or list view and fall back to a loading screen where you couldn't take any action.

## [0.1.0] - 2025-04-06
### Initial Release
- Authentication with AniList
- View and filter anime lists
- Play episodes with MPV integration
- Track episode progress
