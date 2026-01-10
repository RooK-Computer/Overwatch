# Overwatch - monitor GPIO Pins for activity

Overwatch is a small tool we use in Rook to implement the on-device Keys.
It monitors the three GPIO Pins for a Low signal (including debouncing) and executes a program upon pressing them.

Which commands to execute and which Pins to monitor is configured by a simple JSON configuration file.
To run overwatch standalone, simply run `overwatch /path/to/config/file`.

## Configuration Sample

A configuration file has to look like this:

```json
{
    "pins": [
        {
            "pin_number": 17,
            "command": "/usr/bin/command1",
            "pressed_file": "/path/to/file1"
        },
        {
            "pin_number": 27,
            "command": "/usr/bin/command2",
            "pressed_file": "/path/to/file2"
        },
        {
            "pin_number": 22,
            "command": "/usr/bin/command3",
            "pressed_file": "/path/to/file3"
        }
    ]
}
```

The files given in the `pressed_file` property are created when a button is pressed and deleted once the button is released. You can use those files to poll for the buttons outside of the scripts executed by overwatch. The `pressed_file` properties are optional and can be omitted.

## Sample systemd Unit File

To run Overwatch as a systemd service, you can create a unit file like the following:

```ini
[Unit]
Description=Overwatch GPIO Monitor
After=network.target

[Service]
ExecStart=/path/to/overwatch /path/to/config/file
Restart=on-failure
User=root

[Install]
WantedBy=multi-user.target
```

Save this file as `/etc/systemd/system/overwatch.service`, then enable and start the service using:

```bash
sudo systemctl enable --now overwatch.service
```
# Development

Packages are created with [nfpm](https://nfpm.goreleaser.com/)

```
VERSION=1.0.0 make package
```
