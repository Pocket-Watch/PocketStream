# PocketStream

https://github.com/Pocket-Watch/PocketWatch integration for desktop live-streaming

## Dependencies
 - ffmpeg in PATH
 - pocketstream in PATH
 - go-netstat

## Setting up OBS
1. Navigate to **File** >> **Settings** >> **Stream**
2. Change Service to `Custom...`
3. Change destination Server to `rtmp://localhost:9000`
4. Adjust streaming settings: Set Keyframe interval to 2s

## Setting up Lua script for OBS [Optional]
1. Navigate to **Tools** >> **Scripts**
2. Press **+** and select `pocketstream.lua`
3. On the right you'll be able to configure input parameters such as pocketwatch token