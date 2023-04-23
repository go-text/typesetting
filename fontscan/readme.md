# Description and purpose of the package

This package provides two ways to locate and load a `font.Font`, which is the
fundamental object needed by `go-text` for shaping and text rendering.
The two methods are quite different in term of use cases, complexity and possible implementation, but they still answer the same general question : _how can I load a font file if I don't know its exact location ?_

## Use cases

Let's describe the two major use cases we want to cover.

### UI toolkits

The first (and simpler) use case is to locate on the user machine a font file to use to render text in GUI applications. In this case, the developper specifies a couple of family (maybe just "serif" if he doesn't really care), and instead of having to bundle the font file with the application, the font is loaded at the application startup.

One of the requirement is that the lookup time is tiny so that the initial rendering on screen is not delayed.

On the flip side, it is not expected that a large number of fonts will be loaded this way. The rune coverage needed is assumed to be known at build time.

### Markup language renderers

The second use case deals with the rendering of markup documents (think HTML or SVG), where the author provide _hints_ about the fonts that should be used. In this case, the two main requirements are

- use the whole system fonts to find the best match with respect to the author intention
- handle a large rune coverage : many scripts may be present in the same document, so that we can't specify at build time the set of fonts we will use

On the flip side, such renderers startup time is less crucial, so that a slow loading step (say 1 second) is acceptable.

## Overview of the API

For the first task, the package provide the

`FindFont(family string, aspect Aspect) (font.Font, error)`

function, which walk through font directories and search for `family` in filenames.
Among matched font files, the first font matching `aspect` is returned.

For the second task, the `FontMap` type is provided. It should be created for each text shaping task and be filled either with system fonts (by calling `UseSystemFonts`) or with user-provided font files (using `AddFont`), or both.
To leverage all the system fonts, the first usage of `UseSystemFonts` triggers a scan which builds a font index. Its content is saved on disk so that subsequent usage by the same app are not slowed down by this step.

Once initialized, the font map is used to select fonts matching a `Query` with `SetQuery`. A query is defined by one or several families and an `Aspect`, containining style, weight, stretchiness. Finally, the font map satisfies the `shaping.Fontmap` interface, so that is may be used with `shaping.SplitByFace`.

## Zoom on the implementation

### Font directories

Fonts are searched by walking the file system, in the folders returned by `DefaultFontDirectories`, which are platform dependent.
The current list is copied from [fontconfig](https://gitlab.freedesktop.org/fontconfig/fontconfig) and [go-findfont](github.com/flopp/go-findfont).

### Font family substitutions

A key concept of the implementation (inspired by [fontconfig](https://gitlab.freedesktop.org/fontconfig/fontconfig)) is the idea to enlarge the requested family with similar known families.
This ensure that suitable font fallbacks may be provided even if the required font is not available.
It is implemented by a list of susbtitutions, each of them having a test and a list of additions.

Simplified example : if the list of susbtitutions is

- Test: the input family is Arial, Addition: Arimo
- Test: the input family is Arimo, Addition: sans-serif
- Test: the input family is sans-serif, Addition: DejaVu Sans et Verdana

then,

- for the Arimo input family, [Arimo, sans-serif, DejaVu Sans, Verdana] would be matched
- for the Arial input family, [Arial, Arimo, sans-serif, DejaVu Sans, Verdana] would be matched

To respect the user request, the order of the list is significant (first entries have higher priority).

Both `FindFont` and `FontMap.SetQuery` apply a list of hard-coded subsitutions, extracted from
Fontconfig configurations files.

### Style matching

`FindFont` and `FontMap.SetQuery` takes an optionnal argument describing the style of
the required font (style, weight, stretchiness).

When no exact match is found, the [CSS font selection rules](https://drafts.csswg.org/css-fonts/#font-prop) are applied to return the closest match.
As an example, if the user asks for `(Italic, ExtraBold)` but only `(Normal, Bold)` and `(Oblique, Bold)`
are available, the `(Oblique, Bold)` would be returned.

### System font index

The `FontMap` type requires more information than the font paths to be able to quickly and accurately
match a font against family, aspect, and rune coverage query. This information is provided by a list of font summaries,
which are lightweight enough to be loaded and queried efficiently.
The initial scan required to build this index is significantly slow (say between 0.5 and 1 sec on a laptop), meaning this approach is not usable by defaut in GUI applications.

Once the first scan has been done, however, the subsequent launches are fast : at the first call of `UseSystemFonts`, the index is loaded from an on-disk cache, and its integrity is checked against the
current file system state to detect font installation or suppression.

## Performance overview

Performance is a key goal of the package. This section roughly describes the performance of each matching task.

### Font file name matching

The fastest way to match a font is to use its file name, which usually contains its family and sometimes its style.
Walking font directories only using the file names is fast enough to be used at app startup.
However, the family and style fetched may be a bit inaccurate, and the rune/language coverage of the font is not available.

### Family and aspect matching

Font files expose in their content their family and aspect. Reading it requires to open the file and to parse part of its content, adding some overhead.

### Rune coverage matching

Lastly, fetching the rune (or even language) coverage of the fonts is even slower, since the encoding (cmap) table must be parsed.

### Conclusion

The faster `FindFont` uses the first approach to narrow the set of candidate files, then applies the second to select the best match. On a Linux laptop, the total lookup time is around 0.1 sec

The slower but more complete `FontMap.UseSystemFonts` method uses the second and third approaches, amortizing its cost by caching on the disk the result of the scan. On a Linux laptop, the initial lookup time is around 1 sec.