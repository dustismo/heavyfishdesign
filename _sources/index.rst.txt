=================================
Heavy Fish Design (the docs)
=================================

Welcome. Heavy Fish Design is a JSON-like language for designing and layout of parts for laser cutters or CNC routers. Once you get the hang of it, we think you will have fun with all of things you can do. 

Highlights include: 
   * easy to use offsetting (i.e. adjusting for kerf).  
   * Automatic layout of parts within the size of your material. 
   * Splitting across multiple files if parts do not fit within the material. 
   * Full featured expression language for easy scaling of components. 
   * Easy to compose your designs into reusuable custom components.  


.. toctree::
   :maxdepth: 1

   getting_started
   components
   transforms
   part_transformers
   custom_components
   expressions


========
Examples
========

There are loads of examples available in the repo. See `HERE <https://github.com/dustismo/heavyfishdesign/blob/master/designs>`_ for the various HFD files, and to see `HERE <https://github.com/dustismo/heavyfishdesign/blob/master/designs_rendered>`_ for the rendered output.

Game Cabinet
------------

Designed by: Dustin Norlander, Gus Norlander

`HFD File <https://github.com/dustismo/heavyfishdesign/blob/master/designs/game_cabinet/cabinet.hfd>`_


.. code-block::

   hfd-mac render designs/game_cabinet/cabinet.hfd

Game cabinet we designed this for the picade from `adafruit <https://www.adafruit.com/product/2707>`_ but it can be easily customizable for other configurations

.. image:: _static/game_cab.jpg
  :width: 400


Christmas Tree
--------------

Designed by: Nadia Swift

`HFD File <https://github.com/dustismo/heavyfishdesign/blob/master/designs/xmas_tree/nadia.hfd>`_


.. code-block::

   hfd-mac render designs/xmas_tree/nadia.hfd

Christmas tree design, demonstrates a repeatable element that changes based on index.  Easily customizable by providing your own branch and star designs.

.. image:: _static/xmas_tree.jpg
  :width: 400
