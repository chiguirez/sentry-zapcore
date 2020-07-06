# sentry-zapcore
a simpler and cleaner integration for [zap logger](https://github.com/uber-go/zap) and [sentry-go](https://github.com/getsentry/sentry-go)

```golang
	logger, err := zap.NewDevelopment(
		core.WithSentry(sentry.ClientOptions{
			Dsn: "http://whatever@really.com/1337",
		}),
	)
```
