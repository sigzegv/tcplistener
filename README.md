Start any number of headless chromes you want

ie :
```
$YOUR_CHROME_BIN --headless --disable-gpu --remote-debugging-port=9222 https://www.chromestatus.com/
$YOUR_CHROME_BIN --headless --disable-gpu --remote-debugging-port=9223 https://www.myip.com/
```

Configure the proxies in the config variable in tcplistener.go

Then run tcpserver with :
```
go run tcplistener.go
```

Default port is 5555

Check help with -h command parameter to change port

Query headless servers with
`curl http://localhost:5555/json -H "Api-Key: MYKEY"`

This will return default api response like this
```
[ {
   "description": "",
   "devtoolsFrontendUrl": "/devtools/inspector.html?ws=localhost:5555/devtools/page/FFC8AF0B78925EB157C04C61DD6E15E5",
   "id": "FFC8AF0B78925EB157C04C61DD6E15E5",
   "title": "xxxxxxxxxxxxx",
   "type": "page",
   "url": "https://xxxxxxxxxxxxxxx",
   "webSocketDebuggerUrl": "ws://localhost:5554/devtools/page/FFC8AF0B78925EB157C04C61DD6E15E5"
} ]
```
