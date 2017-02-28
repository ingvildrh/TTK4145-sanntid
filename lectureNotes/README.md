# Lecture notes TTK4145 Sanntid

Hi! this is in no way complete notes from the class, as I started typing these notes on the computer over half-way through the semester. Not sure how to structure it yet and whether if I should keep it up, so don't expect too much of this.

## Lecture 28.02.2017

**Learning goals**
- Ability to create (error free) multi thread programs with shared variable synchronization.
- Thorough understanding of pitfalls, patterns, and standard applications of shared variable synchronization.
- Understanding of synchronization mechanisms in the context of the kernel/HW.
- Ability to correctly use the synchronization mechanisms in POSIX, Java and ADA (incl. knowledge of requeue and entry families)

**Three Parts:**
  1. How it works, how it is implemented
    - RT kernel, test & set, disable interrupt, timer interrupt, spin locks, blocking suspend & resume, events & conditions
  2. Skills & Insights.
    - Semaphores, the problems & application. The little book of semaphores.
  3. Synchronization, state of the art.
    - ADA
    - JAVA
    - POSIX

**A hard-to-find bug:**
```C
void allocate(int priority) {
  Wait(M);
  if(busy){
    Signal(M);
    Wait(PS[priority]);
  }
  busy = true;
  Signal(M);
}
```

```C
void deallocate(){
  Wait(M);
  busy = false;
  waiting = GetValue(PS[1]);
  if(waiting>0) Signal(PS[1]);
} else {
    waiting=GetValue(PS[0]);
    if (waiting>0) Signal(PS[0]);
    else {
      Signal(M);
    }
  }
}
```
"What's the bug? **Deadlock!** another classic example: ```wait(A)``` A is zero. Will wait forever. And ```recv(S)```, same here"


- "We will come back to **livelocks** later. It's more difficult to identy than deadlocks"
- "**Starvation** classic example is that we have not a lot of memory and we give memory to all the small requests, but the one that needs almost _all_ the memory will never be run. This is starvation"
- "**Race conditions** is any bug of these, which happens if you have bad luck / timing"
- "**Mutual exclusion**: We fix this easily with semaphores"
- "**Bounded buffer**: If you try to get something from an empty buffer, then you're blocked. It will suspend you. Trying to put something in a full buffer, will also block you / suspend you, until there's space in the buffer" 
