# Overwatch - monitor GPIO Pins for activity

Overwatch is a small tool we use in Rook Mk2 to implement the on-device Keys. 
It monitors the three GPIO Pins for a Low signal (including debouncing) and executes a program upon pressing them.

Which commands to execute and which Pins to monitor is configured by a simple JSON configuration file.
To run overwatch standalone, simply run `overwatch /path/to/config/file`.

## Configuration Sample

A configuration file has to look like this:

```json
{
    "pins": [
        {
            "pin": 17,
            "action": "/usr/bin/command1"
        },
        {
            "pin": 27,
            "action": "/usr/bin/command2"
        },
        {
            "pin": 22,
            "action": "/usr/bin/command3"
        }
    ]
}
```
