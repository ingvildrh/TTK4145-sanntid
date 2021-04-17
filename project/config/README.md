# Configuration package

This package contains definitions of types and global constants used by the other modules. 
Our system can be customized into using different elevator hardware. By changing NumElevators we can set the maximum number of elevators operating. The number of buttons and floors is also possible to modify and the system will work as normal. 
```
const (
	NumElevators     = 3
	NumFloors    int = 4
	NumButtons   int = 3
)
```
