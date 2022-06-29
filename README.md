# GONitroChecker
A promotional discord nitro checker written in Golang.

Supports proxies in format username:password@hostname:port

How to setup it properly:
1. Download the whole source code.
2. Download Golang.
3. Put your promotional nitro codes in /input/promos.txt
4. Put your proxies in /input/proxies.txt
5. Run a cmd in the folder where the source code is and type "go run ."
6. There will be shown a message "Press ENTER to start checking", just press ENTER and wait till it finish checking your promotional codes.
7. You can find your results (valids only) in /output/valid.txt


Format supported for promotionals code is:
https://promos.discord.gg/[code]
