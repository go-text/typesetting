# typesetting

This library provides typesetting capabilities in pure Go. It is appropriate for use in GUI applications, and is shared by multiple Go UI toolkits including [Fyne](https://fyne.io), [Gio](https://gioui.org), and [Ebitengine](https://ebitengine.org).

## Development cycle

This project, although already used in production by UI toolkits, still evolves rapidly. As such, the library uses unstable versions v0.x.y : the required breaking changes will bump the minor version number (x); the bug fixes and performance improvements the patch number (y).

## Review guidelines
 
Go-text is a collaboration between many individuals and projects, it is important to us that
designs and decisions are right for the broadest possible audience.
As a result the project will always have 3 maintainers that represent different projects
(currently Fyne.io, Gio and an independent developer).

### API and Architectural decisions

Changes to any core go-text repositories (not including utility or generator repos) will require
sign-off from at least 2 of these 3 maintainers to be approved.
"typesetting" and "render" are currently considered core.
Upon approval the second thumbs up will typically merge the change into the repository.

Decisions or API discussions are best carried out within the context of a GitHub issue or
pull request for greatest visibility in the future.

### Smaller changes and quality of life improvements

For speed of acceptance on smaller issues it is not always required to have complete consensus.
When a change is deemed to be of minor impact (for example documentation corrections, trivial
bug fixes and straight forward refactoring of content) an expedited review is supported.
In this situation the contribution only requires a single approval (not from the individual
proposing the change).

If in doubt please seek approval of two maintainers - and feel free to ask questions in the
#go-text channel of gophers Slack server.
