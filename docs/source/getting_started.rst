===============
Getting Started
===============


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

parts
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Each document is comprised of a set of Parts.  Each part should be considered a single unit that can be created independently of the rest.  The document will automagically move parts around to fit within the material size.

Each Part has the following properties

* ``id`` Unique id, not required, but can help with debugging sometimes.  If an id is not supplied a unique id will be randomly assigned 
* ``components`` List of the individual components
* ``repeat`` This is useful to repeat the part multiple times.  Each rendered part will have an additional variable called ``part_index`` which could be used to change each part based on index.
    repeat is structured like:

 .. code-block:: JSON

    "repeat" : {
        "total" : "num_layers"
    }

* ``part_transforms`` part specific transforms, which may or may not split the part into multiple parts.  See part_transformers.rst