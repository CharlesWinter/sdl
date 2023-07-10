Outputs logs in the very particular JSON that stackdriver requires for its structured logging

1. Calling logger.Error(err) will but an error through to error reporting. 

The actual JSON it spits out is like this:

```json
  {
   "eventTime":"2023-07-10T09:36:56.08643+01:00",
   "message":"oh no",
   "serviceContext":{
      "service":"test",
      "version":"1.0.0"
   },
   "severity":"error",
   "stack_trace":"oh no\ngoroutine 7 [running]:\nruntime/debug.Stack()\n\t/opt/homebrew/Cellar/go/1.20.5/libexec/src/runtime/debug/stack.go:24 +0x64\ngithub.com/CharlesWinter/sdl.Logger.Error({0x14000130070, {{0x102c08603, 0x4}, {0x102c086b9, 0x5}}}, {0x102cbafd8?, 0x1400005b030})\n\t/Users/charleswinter/workspace/sdl/logger.go:84 +0x44\ngithub.com/CharlesWinter/sdl_test.TestLoggingErrors.func1(0x14000138000)\n\t/Users/charleswinter/workspace/sdl/logger_test.go:26 +0x168\ntesting.tRunner(0x14000138000, 0x102cb9c20)\n\t/opt/homebrew/Cellar/go/1.20.5/libexec/src/testing/testing.go:1576 +0x10c\ncreated by testing.(*T).Run\n\t/opt/homebrew/Cellar/go/1.20.5/libexec/src/testing/testing.go:1629 +0x368\n",
   "timestamp":"2023-07-10T09:36:56.086434+01:00"
}
```

2. Using the request logger will make the requests format nicely in stackdriver, with well formed status codes etc.


