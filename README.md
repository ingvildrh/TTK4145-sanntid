# [TTK4145-Sanntidsprogrammering](https://www.ntnu.no/studier/emner/TTK4145)

## Course Content 
Programming formalisms for real-time systems; POSIX, Ada, Java and Go. Threads/processes, synchronization and communication. Shared variable-based synchronization and resource control. Fault Tolerance, availability and consistency. Message-based synchronization, CSP and formal methods. Exercises and project. 

In this course, we had a term project, where I was in a group with two other students. The project counts 25% of the final grade. Here's a brief description of the project and our solution: 

## Elevator project
### Description
In this project, we had to create software for controlling `n` elevators working in parallel across `m` floors. There were some main requirments, here summarised in bullet points: 

  - **No orders are lost** 
  - **Multiple elevators should be more efficient than one** 
  - **An Inidividual elevator should behave sensibly and efficiently**
  - **The lights should function as expected**
  
In the project, we start with `1 <= n <= 3` elevators, and `m == 4` floors. However, we should avoid hard-coding these values, and we aimed to write the project where adding a floor or an elevator required minimal work. The system will however _not_ be tested for `n > 3` or `m != 4`. There are also some unspecified behaviours we had to decide for ourselves: 

  - **Which orders are cleared when stopping at a floor**
  - **How the elevator behaves when it cannot connect to the network during initialization**
  - **How the hall (call up, call down) buttons work when the elevator is disconnected from the network**
 
Lastly, there were some permitted assumptions: 

  - **At least one elevator is always working normally**
  - **No multiple simultaneous errors: Only one error happens at a time, but the system must still return to a fail-safe state after this error**
  - **No network partitioning: Situations where there are multiple sets of two or more elevators with no connection between them can be ignored**
  - **Stop button and obstruction switch are disabled**

For full details on each point, the driver files, or the full specs of the project: head over to [`TTK4145`](https://github.com/TTK4145/Project#elevator-project) (the description/project might have changed as the course is held every spring)

## Our solution
We wrote our soloution in `Google GO`. This was a new language for us, so it required some learning. We decided to use a _"fleeting master"_ together with **UDP broadcasting**. This means that **all elevators knows about all orders** and that **all elevators knows about all other elevators `state`, `direction` and `floor`**. The elevator that receives an external order, will be the one to decide which elevator should execute the order. This decision, along with the order, is broadcasted to all other elevators on the network. The order is then implicitly acknowledged between the elevators before execution. If an elevator lost network or failed to finish the order in a certain time, the other elevators would take over the order. If an elevator is operating normally, only without network, it functioned as a locally run elevator. More details can be found in the subfolders of this project.

In terms of performance, the system performed great. Network packet loss was for example not a problem for our solution in relation to the project requirments. However, there were points of improvement that could be done. For example, one module of the project ended up becoming somewhat of a "fix-it-all"-module, where more divisions of tasks would improve the code neatness and readability. Also, even though scaling the system wouldn't require a lot of work, data types shared between modules and changing these could have been benefited of being more dynamic. 

This project took a lot of work to complete, but it gave great experience and was a lot of fun. Our end result is something we're satisified with. 

_This `README` will be updated after completion of the course_
