# Elevator State Machine module

This is our finite state machine of the system. It has `3 states` and `3 events`.
What does the finite state machine do? It takes local orders from the queue,
which it then proceeds to execute. It handles motor direction, as well
as door timeout and door light. It also has a timer to prevent starvation, for example
if the motor power is lost. 

States | Events
------------ | -------------
IDLE| New Order
MOVING | Floor Reached
DOOROPEN | Door Timeout

**`IDLE`** - _Elevator is stationary, at a floor with closed doors, awaiting orders_  
**`MOVING`** - _Elevator is moving and is either between floors or at a floor going past it_  
**`DOOROPEN`** - _Elevator is at a floor with the door open_  
**`New Order`** - _There's a new order added to the queue_  
**`Floor Reached`** - _The elevator has reached a new floor_  
**`Door Timeout`** - _The door has been open for set period and is timed out; the doors should close_
