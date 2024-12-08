# Repowatcher
A Tool written in Golang that displays the status of your local repos in Waybar.

## Installation

### Requirements
* Golang 1.23.3
* Git tools installed
* Waybar >= 0.11

### Building

Build the binary with

```
git clone https://github.com/scqcasc/repowatcher.git
cd repowatcher
go build
cp ~/.local/bin/repowatcher
```



### Configuration
Copy the sample.config.json to ~/.local/share/repowatcher/config.json.
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
        "format": "Repos: {}"
    },
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
