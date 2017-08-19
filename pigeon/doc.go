/*
Package pigeon contains the interfaces and definitions of Pigeon service.

   ┌───────┐      ┌───────────┐      ┌────────┐      ┌────────────┐
   │       │ ───► │ Scheduler │ ───► │        │ ───► │ Dispatcher │
   │       │      └───────────┘      │        │      └────────────┘
   │       │      ┌───────────┐      │        │      ┌────────────┐
   │ Store │ ───► │ Scheduler │ ───► │ Merger │ ───► │ Dispatcher │
   │       │      └───────────┘      │        │      └────────────┘
   │       │      ┌───────────┐      │        │      ┌────────────┐
   │       │ ───► │ Scheduler │ ───► │        │ ───► │ Dispatcher │
   └───────┘      └───────────┘      └────────┘      └────────────┘

*/
package pigeon
