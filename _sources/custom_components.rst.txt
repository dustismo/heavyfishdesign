=================
Custom Components
=================


Creating Components via Markup
==============================

It's easy to create custom components via the HFD markup language.  Simply add a 
``custom_component`` attribute to any component.  The new component will be available
via the ``import`` attribute on the document

Example:

Here is a simple line component:

    `<https://github.com/dustismo/heavyfishdesign/blob/master/designs/joints/line.hfd>`_

It requires the attributes ``to`` and ``from`` when being imported

.. code-block:: JSON
   :linenos:
   :emphasize-lines: 7-13

    {
        "parts": [
            {
                "components": [
                    {
                        "type": "draw",
                        "custom_component": {
                            "type": "line" 
                        },
                        "commands": [
                            {
                                "command": "move",
                                "to": "from"
                            },
                            {
                                "command": "line",
                                "to": "to"
                            }
                        ]  
                    }
                ]
            }
        ]
    }


Then to access the new component, simply import it and use it. 

.. code-block:: JSON
   :linenos:
   :emphasize-lines: 3, 16-20

    {
        "imports": [
            {"path": "designs/common/line.hfd"}
        ],
        "params": {
            "material_width": 20,
            "material_height": 12,
            "material_thickness": 0.2,
            "width": 2,
            "height" 3
        },
        "parts": [
            {
                "id" : "line_example",
                "components": [
                    {
                        "type": "line",
                        "from" : "width,height",
                        "to" : "width / 2,height / 2"
                    }
                ]
            }
        ]
    }
