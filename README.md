# Documize Community Fork

For the original repository, see https://github.com/documize/community

## Changes

### Attachments get dumped to object storage

See [09931004](https://github.com/tunezilla/documize-community/commit/09931004e8b3786ef37c0fbfccc114d56f9271ea)

To configure:

- Set `DOCUMIZEBUCKET=some-bucket-here`
- If using Minio, set `DOCUMIZEMINIO=http://id:secret@minio:9000`
- Otherwise, your AWS S3 config will be pulled from [default locations](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html)

### More DB options

See [e81d81](https://github.com/tunezilla/documize-community/commit/e81d81be09e24969909be4cefa023c70f422f2d2), [49e77c](49e77c60aaaa74b4bc4905b0cd72b0dea9e2e97a)

| Env Var                   | Func               | Default       |
|---------------------------|--------------------|---------------|
| DOCUMIZEDBMAXIDLECONNS    | [SetMaxIdleConns](https://pkg.go.dev/database/sql#DB.SetMaxIdleConns)    | 30            |
| DOCUMIZEDBMAXOPENCONNS    | [SetMaxOpenConns](https://pkg.go.dev/database/sql#DB.SetMaxOpenConns)    | 100           |
| DOCUMIZEDBCONNMAXLIFETIME | [SetConnMaxLifetime](https://pkg.go.dev/database/sql#DB.SetConnMaxLifetime) | 14400         |
| DOCUMIZEDBCONNMAXIDLETIME | [SetConnMaxIdleTime](https://pkg.go.dev/database/sql#DB.SetConnMaxIdleTime) | 0 (unlimited) |

### Stuff is stubbed

We [stop checking for changelogs and news](https://github.com/tunezilla/documize-community/commit/f2cdc751d4a6d65a87f0a2f5f7f621c3b447831f) and [don't care about the "What's New" dot](https://github.com/tunezilla/documize-community/commit/6c1e51ee346578dd5f200916f7c6b657699d209d)

### Page loader

[A progress bar shows between page loads](https://github.com/tunezilla/documize-community/commit/4e785287348e4684fd594eb76340b8c01a24cbd8)

### Others

[Full list](https://github.com/tunezilla/documize-community/compare/documize:community:master...master)
