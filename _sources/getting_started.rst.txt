Getting Started
===============

installation
------------------------------------------------------------------------------------------
TODO

HFD
----------------

Every design is based on one or more HFD files.  HFD is syntactically JSON but allows
comments.

Every HFD file has the following structure

.. code-block:: JSON

    {
        "imports": [
            {"path": "some/custom_component.json"},
            {"path": "another/custom/component.json"}
        ],
        "params": {
            "offset": ".0035",
            "material_width": 10,
            "material_height": 12,
            "measurement_units": "in",
            "material_thickness": 0.2,
            "some_param1": "0",
            "another_param": 3.25, 
        },
        "parts": [
            {
                "id": "some_identifier",
                "components": [
                    {
                        "type": "stem_with_topper",
                        "stem_height" : "material_thickness * ((num_layers * 2) + num_stem_layers + 2)",
                        "stem_topper_svg": "star_svg"
                    }
                ]
            },
            {
                "id": "spacer",
                "repeat" : {
                    "total" : "num_layers + num_stem_layers"
                },
                "components": [
                    {
                        "type": "spacer",
                        "width": "stem_width * 2"
                    }
                ]
            }
        ]
    }

Each document is divided into sections:

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

