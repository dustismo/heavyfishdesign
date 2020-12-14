This is Heavy Fish Design  
==========================

It's for designing things that scale (get it, get it?).  

At its heart, HFD is a json based design language.  This package contains a compiler to SVG files, and various useful tools for working with 2d models.  

Visit [heavyfishdesign.com](http://heavyfishdesign.com) for full documentation

## Features:

* A fully functional JSON based language for designing parameterized 2d vector drawings
* Set of useful transformations including joining, scaling, simplify and more
* Outputs clear and concise svg
* Layout engine will arrange parts to fit within the specified material size
* Parts can be automatically split to fit within the cut material.

## Usage:


The runnable contains a simple server for displaying in the browser, as well as commands for rendering designs on the command line. 

### Local Render:

To render an hfd file to the corresponding svg files use:

    $ go run main.go render --path=designs/drawer_organizers/silverware.hfd --output_file=designs_rendered/silverware

### Basic server operation:

To run a local server to see svg's rendered in the browser, do this.  This is useful to use during design, but note that by default, the server only displays the first rendered svg document (i.e. if your document spans multiple pages only the first is desplayed)

    go run main.go serve

then open browser to:

    http://localhost:2003/json?file=<path_to_hfd_file>

ex: 
    http://localhost:2003/json?file=designs/box/box_three_sided.hfd

*****

Runbook
========


When making code changes, first run the unit tests then the comparison of all local designs (this will compare all the local designs to the last render)

    $ go test ./...
    $ go run main.go diff_test

This will output any rendering differences your change causes.  Double check that changes are expected.  if so, run:

    $ go run main.go designs_updated

and commit

Build the docs:

    $ cd docs && make html && cd ..


VSCODE

To set the .hfd file associations edit settings.hfd and add:

    "files.associations": {
        "*.hfd": "json"
    }

RELEASE

To create a release

    $ git tag v1.0.X
    $ git push origin master --tags

