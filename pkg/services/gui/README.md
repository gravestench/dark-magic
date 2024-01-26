# GUI Manager
The purpose of this [Service Mesh](https://github.com/gravestench/servicemesh) service is to provide a GUI which
integrates with the raylib renderer and exposes itself as a renderable layer.

The GUI manager maintains a "node tree", where each node can (optionally) do 
the following: (1) have a function that yields an `image.Image`, and (2) have
an update function.

The GUI manager has a configurable refresh rate. There is a root node that all 
nodes are attached to by default. The root node will update all its children. 
When the renderer picks up the GUI manager, the resulting image is a composite 
of the root node and all of its child images.

## Dependencies
This service has dependencies on all other Diablo2 file-loader services:
* [raylib renderer](../raylibRenderer) - waits for the renderer to init
* [sprite manager](../spriteManager)


## Integration with other services
This service integrates with the following services:
* [lua service](../lua)
* [raylib renderer](../raylibRenderer) - implements the RenderableLayer interface

The raylib renderer integration is mandatory considering that the renderer
init is part of the dependency resoltuion loop.

_______
This service exports an integration interface `ManagesGui` with an alias 
`Dependencncy` which are intended to be used by other services for dependency
resolution (see servicemesh.HasDependencies), and expose just the methods which 
other services should use.
```golang
type Dependency = ManagesGui

type ManagesGui interface {
    NewNode() node
    Nodes() []node
    Update()
}
```

Other services can use this GUI manager as a dependency for creating UI nodes. 
The actual implementation of the update function and the function that yields 
the `image.Image` for the node are not a concern of this service. Other services
or modules would use this abstraction to build actual UI stuff, this service
is just managing the node tree and compositing the root node into an image for 
the renderer.

The GUI `node` is an exported interface: 
```golang
type inputHandler interface {
    InputVector() InputState
    SetInputVector(InputState)
    ClearInputVectors()
    
    Handler() func()
    SetHandler(func())
}

type element interface {
    inputHandler
    
    UUID() uuid.UUID
    
    Enable()
    Disable()
    IsEnabled() bool
    
    Position() (point image.Point)
    SetPosition(point image.Point)
}

type node interface {
    element
    
    Parent() node
    SetParent(node)
    
    Children() []node
    AddChild(child node)
    RemoveChild(child node)
    
    LayerIndexOf(child node) int
    SetLayerIndexOf(child node, index int)
    
    BringToTop(child node)
    BringToBottom(child node)
    Raise(child node)
    Lower(child node)
    
    HandleInput(InputState) (terminate bool)
    
    Image() image.Image
    ImageFunc() func() image.Image
    SetImageFunc(func() image.Image)
    
    Update()
    UpdateFunc() func()
    SetUpdateFunc(func())
    
    GetRelativePosition(target node) (relativePosition image.Point, found bool)
}
```

## Lua service integration
Nothing is implemented for the lua integration, yet.