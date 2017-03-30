# Elevator Governor module

This module has multiple responsibillities that is somewhat natural to keep in the same module. Those responsibillities include:
 - Setting button lights depending on acknowledged orders in the queue
 - Keeping track of every elevators state, position, direction and orders
 - Keeping track of every order, remote and local, in a queue
 - Calculating cost for a new order registered on local elevator external button
    - Deciding who should execute the order and forwarding that information

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
