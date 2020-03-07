This is Heavy Fish Design.  It's for designing things that scale (get it, get it?).  

HFD is a json based design language.  This package contains a compiler to SVG files, and various useful tools for working with 2d models.  

Features:
    * A fully functional JSON based language for designing parameterized 2d vector drawings.
    * Set of useful transformations including joining, scaling, simplify and more.
    * Outputs clear and concise svg
    * Layout engine will arrange parts to fit within the specified material size
    * Parts can be automatically split to fit within the cut material.

Usage:

The runnable contains a simple server for displaying in the browser, as well as commands for rendering designs on the command line. 

Local Render:

To render an hfd file to the corresponding svg files use:

    $ go run main.go render --path=designs/drawer_organizers/silverware.json --output_file=designs_rendered/silverware

Basic server operation:

go run main.go serve

then open browser to:

    http://localhost:2003/json?file=<path_to_hfd_file>

ex: 
    http://localhost:2003/json?file=designs/box/box_three_sided.json



Runbook
========



When making code changes, first run the unit tests then the comparison of all local designs (this will compare all the local designs to the last render)

    $ go test ./...
    $ go run main.go diff_test

This will output any rendering differences your change causes.  Double check that changes are expected.  if so, run:

    $ go run main.go designs_updated

and commit

The Code
========

The SVG Package:

Most of the code operates on Path objects. A Path contains a sequence of drawing operations, which map to SVG drawing operations. To simplify writing transforms, only a subset of svg path commands are supported. This subset is full featured though, and all other commands are automatically converted to this set when parsed (For instance Quadratic curves are automatically converted to Bezier).

The DOM Package:

Individual pieces are all layed out into a document or planset. The document will attempt to maximize the usuable space on the material. Elements are layed out using a simple bin packing algorithm and potentially rotated. Additionally elements can be split into parts if the element does not fit on the material (a butt joint is added so the elements can be reassembled)
