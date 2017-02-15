# Elevator Governor module

This module is perhaps the largest and most diffuse module. It has multiple responsibillities that is somewhat natural to keep in the same module. Those responsibillities include:
 - Setting button lights depending on acknowledged orders in the queue
 - Keeping track of every elevators state, position, direction and orders
 - Keeping track of every order, remote and local, in a queue
 - Calculating cost for a new order registered on local elevator external button
    - Deciding who should execute the order and forwarding that information

This module has not received proper naming, and is so far a stillborn baby (read: no code), buried without a burial. Hopefully the module won't turn into a [botchling](http://witcher.wikia.com/wiki/Botchling) due to the lack of proper naming.
If so, I hope it won't be too late to lift the curse and transform it to a [lubberkin](http://witcher.wikia.com/wiki/Lubberkin). Time will tell.

Orders Queue | Elevator 1 | Elevator 2 | Elevator 3
----------- | ---------- | ---------- | ----------
Floor 4     | ---- / :arrow_down: / :four: | ---- / :arrow_down: / :four: | ---- / :arrow_down: / :four:
Floor 3     | :arrow_up: / :arrow_down: / :three: | :arrow_up: / :arrow_down: / :three: |  :arrow_up: / :arrow_down: / :three:
Floor 2     | :arrow_up: / :arrow_down: / :two: | :arrow_up: / :arrow_down: / :two: |  :arrow_up: / :arrow_down: / :two:
Floor 1     | :arrow_up: / ---- / :one: | :arrow_up: / ---- / :one: |  :arrow_up: / ---- / :one:

Elevator State | Elevator 1 | Elevator 2 | Elevator 3
--------------- | ---------- | ---------- | ----------
Floor 4 | :arrow_forward: / :zzz: / :clock10: | :arrow_forward: / :zzz: / :clock10: |  :arrow_forward: / :zzz: / :clock10:
Floor 3     | :arrow_forward: / :zzz: / :clock10: | :arrow_forward: / :zzz: / :clock10: | :arrow_forward: / :zzz: / :clock10:
Floor 2     | :arrow_forward: / :zzz: / :clock10: | :arrow_forward: / :zzz: / :clock10: | :arrow_forward: / :zzz: / :clock10:
Floor 1     | :arrow_forward: / :zzz: / :clock10: | :arrow_forward: / :zzz: / :clock10: |  :arrow_forward: / :zzz: / :clock10:
