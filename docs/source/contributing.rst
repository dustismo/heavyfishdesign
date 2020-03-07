Contributing
==========================

Contributions are welcome.  See github.com/dustismo/hfd

Components
--------------------------

Components can be written by implementing the Component interface, creating a factory and 
registering the factory.  


Transforms
--------------------------

Transforms simply do some operations on a Path and return a new path or an error.  Transforms
implement the PathTransforms interface, then register a TransformFactory to make it available
