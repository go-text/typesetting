# Description and purpose of the package

This package provides two ways to locate and load a `font.Face`, which is the
fundamental object needed by `go-text` for shaping and rendering text.
The two methods are quite different in term of use cases, complexity and possible implementation, but they still answer the same general question : _how can I load a font file if I don't know its exact location ?_

## Use cases

Let's describe the two major use cases we want to cover.

### UI toolkits

The first (and simpler) use case is to locate on the user machine a font file to use to render text in graphical toolkits. In this case, the developper specifies a couple of family (maybe just "serif" if if doesn't really care), and instead of having to bundle the font file with the application, the font is loaded at the application startup.

One of the requirement is that the lookup time is tiny so that the initial rendering on screen is not delayed.

On the flip side, it is not expected that a large number of fonts will be loaded this way. The rune coverage needed is assumed to be known at build time.

### Markup language renderers

The second use case deals with the rendering of markup documents (think HTML or SVG), where the author provide _hints_ about the font that should be used. In this case, the two main requirements are

- use the whole system fonts to find the best match with respect to the author intention
- handle a large rune coverage : many scripts may be present in the same document, so that we can't specify at build time the set of fonts we will use

On the flip side, such renderers startup time is less crucial, so that a slow loading step (say 1 second) is acceptable.

## Overview of the API

For the first task, the package provide the

`FindFont(family string) (font.Face, error)`

function, which walk through font directories and search for `family` in filenames.
Among matched font files, the first with a regular style is returned.

For the second task, the `FontMap` type is provided. It should be created for each text shaping task and be filled either with system fonts (by calling `UseSystemFonts`) or with user-provided font files (using `AddFont`).
To leverage all the system fonts, the first usage of `UseSystemFonts` triggers a scan, building a font index, whose content is saved on disk so that subsequent usage of the same app are not slowed down by this step.

Once initialized, the font map is used to select fonts matching a `Query` with `SetQuery`. A query is defined by one or several families and an `Aspect`, containining style, weight, stretchiness. Finally, the font map satisfy the `shaping.Fontmap` interface, so that is may be used with `shaping.SplitByFace`.

## Zoom on the implementation

### Font directories

// TODO:

### Font family substitutions

// TODO:

### System font index

// TODO:

## Possible integration in go-text

I think the two use proposed solutions have enough elements in common so that it make sense to include both
in go-text.
However, since it is actually a big addition, we could also only add some parts and keep the remaining parts of fontscan in a third party repository. Some alternatives :

- only provide FindFont, with no substitutions. It would be very close to what existing package provides (see the credits part); the main additional feature being the possiblity to match font by aspect. TODO: not implemented yet

- provide FindFont and support for substitutions. That would make the FindFont quite powerful on its own.
