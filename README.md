# Repowatcher
A Tool written in Golang that displays the status of your local repos in Waybar.  At a glance you can see if any of your repositories
need updated or pushed.

When you click the Repos lable it opens a Rofi window allowing you to select the repo you want to update.

This was developed and tested in Fedora Workstation 41 using Hyprland.

This can run in a daemon mode, but for Waybar you can run it with the --once flag to have Waybar manage the re-runs.

## Installation

### Requirements
* Golang 1.23.3
* Git tools installed
* Waybar >= 0.11
* Rofi >= 1.7.5
* Lazygit

### Building

Build the binary with these steps:

```
git clone https://github.com/scqcasc/repowatcher.git
cd repowatcher
go build
cp repowatcher ~/.local/bin/repowatcher
cp repowatcher-handler ~/.local/bin/repowatcher-handler
```



### Configuration
Copy the sample.config.json to ~/.local/share/repowatcher/config.json.
```
mkdir -p ~/.local/share/repowatcher
cp sample.config.json ~/.local/share/repowatcher/config.json
```
Update it with the repos you want to watch.

The config structure is
```
{
  "repositories": [
    {
      "name": "Some_Name",
      "location": "/the/path/to/Some_Name"
    }
  ],
  "poll_interval": 10
}
```

Update your waybar config

* Add the repowatcher module
```
"custom/git-repowatcher": {
        "exec": "~/.local/bin/repowatcher -once",
        "interval": 10,
        "tooltip": true,
        "return-type": "json",
        "format": "Repos",
        "on-click": "~/.local/bin/repowatcher-handler"
    },
```
Add the module somewhere in your Waybar:

```
"modules-right": ["custom/spotify", "custom/git-repowatcher",  "custom/weather",  "network", "cpu", "tray", "clock","pulseaudio","battery", "custom/power"],
```

* Add css styling:
```
#custom-git-repowatcher {
    color: green; /* Default to green */
}
#custom-git-repowatcher.red {
    color: red;
}
#custom-git-repowatcher.yellow {
    color: yellow;
}
```

Inform waybar about it's new configuration:
```
killall -SIGUSR2 waybar
```

## TODO
Add need to pull notification
