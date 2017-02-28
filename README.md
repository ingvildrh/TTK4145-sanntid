# [TTK4145-Sanntidsprogrammering](https://www.ntnu.no/studier/emner/TTK4145)

## DISCLAIMER 
This is a course I'm currently undertaking, so it will be changing from week to week and is in no means a finished repo. 

## Course Content 
Programming formalisms for real-time systems; POSIX, Ada, Java and Go. Threads/processes, synchronization and communication. Shared variable-based synchronization and resource control. Fault Tolerance, availability and consistency. Message-based synchronization, CSP and formal methods. Exercises and project. 

In this course, we have a term project. I am in a group with two other students. The project counts 25% of the final grade. Here's a brief description of the project and our solution so far: 

## Elevator project
### Description
In this project, we have to create software for controlling `n` elevators working in parallel across `m` floors. There are some main requirments, here summarised in bullet points: 

  - **No orders are lost** 
  - **Multiple elevators should be more efficient than one** 
  - **An Inidividual elevator should behave sensibly and efficiently**
  - **The lights should function as expected**
  
In the project, we start with `1 <= n <= 3` elevators, and `m == 4` floors. However, we should avoid hard-coding these values, and we aim to write the project where adding a floor or an elevator requires minimal work. The system will however _not_ be tested for `n > 3` or `m != 4`. There are also some unspecified behaviours we have to decide for ourselves: 

  - **Which orders are cleared when stopping at a floor**
  - **How the elevator behaves when it cannot connect to the network during initialization**
  - **How the hall (call up, call down) buttons work when the elevator is disconnected from the network**
  - **Stop button and obstruction switch are disabled**
 
Lastly, there are some permitted assumptions: 
  - **At least one elevator is always working normally**
  - **No multiple simultaneous errors: Only one error happens at a time, but the system must still return to a fail-safe state after this error**
  - **No network partitioning: Situations where there are multiple sets of two or more elevators with no connection between them can be ignored**

For full details on each point, the driver files, or the full specs of the project: head over to [`TTK4145`](https://github.com/TTK4145/Project#elevator-project)

## Our solution
We are writing our soloution in `Google GO`. This is a new language for us, so it requires some learning. We've decided to use a _"fleeting master"_ together with **UDP broadcasting**. This means that **all elevators knows about all orders** and that **all elevators knows about all other elevators `state`, `direction` and `floor`**. The elevator that receives an external order, will be the one to decide which elevator should execute the order. This decision, along with the order, is broadcasted to all other elevators on the network. The order is then implicitly acknowledged between the elevators before execution. Timers will ensure that the order is executed at some point.

_This `README` will be updated after completion_
