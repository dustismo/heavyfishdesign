=================
Expressions
=================

HFD has a built in expression language that handles parameters, math and a handful of 
useful functions. 

------------------------------------------------------------------------------------------

Available functionality in expressions
======================================

* ``+-*/()`` Normal arithmetic operators.
* ``sqrt(arg)`` Square Root of arg
* ``distance(p1,p2)`` OR ``distance(x1,y1,x2,y2)`` The distance between the points (x1,y1) and (x2,y2)
* ``angle(p1,p2)`` OR ``angle(x1,y1,x2,y2)`` The angle in degrees of the line described by p1,p2
* ``mmToInch(arg)`` Converts to inches from mm
* ``inchToMM(arg)`` Converts to mm from inches

Parameters
==========

Parameters in any expression are looked up based on the current context. 

The order of lookup is as follows

* Local attributes
* Local params
* Parent attributes
* Parent params
* ... Repeat until Document
* Document params

