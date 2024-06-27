# logger

The `logger` package provides "event manager", "basic event definitions", and "a simple log transmitter".

- **Event Manager**: Used to receive events and dispatch them to all subscribers.
- **Event Definition**: Defines the most basic events, allowing for extension.
- **Log Transmitter**: The event manager can derive a log transmitter, 
allowing log events to be passed to the log transmitter when events occur.