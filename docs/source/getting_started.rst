===============
Getting Started
===============

Installation
============

HFD is a command line tool with available binaries for mac and windows (and simple to build for anyone with golang experience).  Simply download the package, and open a terminal to render your first design.  The zip file contains many example designs, which are easily customizable by changing any of the parameters near the top of the .hfd file.

`DOWNLOAD NOW <https://dustismo.github.io/heavyfishdesign/_static/heavyfishdesign.zip>`_

Open a terminal (these instructions for mac, but should be similar on windows):

.. code-block::

    $ cd ~/Downloads
    $ unzip heavyfishdesign.zip
    $ cd heavyfishdesign
    $ ./hfd-mac render designs/xmas_tree/nadia.hfd


The easiest way to get started to modify one of the existing designs. Start by changing some of the top level params and seeing how that impacts the rendered SVG. 

HFD
===

Every design is based on one or more HFD files.  HFD is syntactically JSON but allows
comments.

Every HFD file has the following structure

.. code-block:: JSON

    {
        "imports": [
            {"path": "some/custom_component.hfd"},
            {"path": "another/custom/component.hfd"}
        ],
        "params": {
            // These fields are recommended for all designs
            "offset": ".0035",          // the cutting kerf / 2
            "material_width": 18,       // size of the material we are cutting from
            "material_height": 11,      // size of material
            "material_thickness": 0.2,  // thickness of material
            "measurement_units": "in",  // either "in" or "mm"

            // more design specific params
            "some_param1": "0",
            "another_param": 3.25, 
        },
        "parts": [
            // One or more parts
            {
                "id": "some_identifier",
                "components": [
                    {
                        "type": "stem_with_topper",
                        "stem_height" : "material_thickness * ((num_layers * 2) + num_stem_layers + 2)",
                        "stem_topper_svg": "star_svg"
                    }
                ]
            }
        ]
    }

Each document is divided into the following sections:

* ``imports`` Imports allows importing custom components from other documents as well as importing siple SVG files. 
* ``params`` List of parameters which impact the design.
* ``parts`` The individual parts of the design. Each part is expected to be a descrete item that can be cut out of the material.

imports
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Imports allows importing custom components from other documents
as well as importing siple SVG files. See docs for custom components

Import options:

.. code-block:: JSON

    {
        "path": //<required> <string> the path to the component or svg being imported. 
                // currently this must be a file path, but should 
                // eventually support urls
        "type": //<optional> <string> default is `component` the other option is `svg`
        "alias": // <optional> <map>  
            // this is a map used to rename custom component imports
            // for instance:
            // "box_side": "house_front"
            // would rename the imported component "box_side" so it could 
            // be referenced in this document as "house_front" 
            // this can be useful for clarity, or to resolve naming conflicts
    }

