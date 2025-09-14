Issue when cleaning up YT-DLP fragments between servers - maybe create a local directory per server to store fragments etc
Could happen when a bot joins a new channel during a fragment cleanup, then the bot in the new channel fails at voice connection.
logs: 
2025/09/14 04:59:03 [DG0] voice.go:403:wsListen() voice endpoint <>.discord.media:443 websocket closed unexpectantly, websocket: close 4016: Unknown encryption mode.
2025/09/14 04:59:03 [DG0] voice.go:199:Close() error closing websocket, websocket: close sent
2025/09/14 04:59:14 Voice connection is invalid or disconnected, aborting stream
2025/09/14 04:59:14 Queue is empty.
2025/09/14 04:59:16 Bot was disconnected from voice channel in guild: 1392997664542691419
2025/09/14 04:59:16 Cleaning up player state for disconnected guild: 1392997664542691419

  --
2025/09/14 05:00:03 Audio Stats - Frames: 13120, Timeouts: 0, Errors: 0, Duration: 4m40.000645002s
2025/09/14 05:00:08 Audio Stats - Frames: 13370, Timeouts: 0, Errors: 0, Duration: 4m45.000559373s
2025/09/14 05:00:12 Audio stream completed, 52125644 bytes processed
2025/09/14 05:00:12 Audio stream cleanup completed

2025/09/14 05:00:12 Stream finished after 13574 frames
2025/09/14 05:00:12 Removed fragment file: --Frag2

