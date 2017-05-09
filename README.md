Vickrey–Clarke–Groves auction
===

This program generates random bids for n agents and m items and runs the auction algorithm.
It then outputs optimal allocations and price to pay.

Allocation to agent 0 is in fact allocation to "nobody", since it is possible for the optimal allocation to
not give an item to anyone. Items are traditionally enumerated from 0 to m-1.


How to run?
======

* Install Go (tested on Go v1.8)
* Execute: `go run main.go n m` (eg. `go run main.go n m`)
