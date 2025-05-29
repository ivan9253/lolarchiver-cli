# LoLArchiver CLI

A CLI tool for the [LoL Archiver API](https://lolarchiver.com/). Requires paid subscription [Web+API](https://lolarchiver.com/subscription).

No external dependencies, small and simple codebase.

## Installation

```bash
go install github.com/ivan9253/lolarchiver-cli/cmd/lolarchiver-cli@latest
```

## Configuration

Before using the CLI, you need to set your API key:

```bash
lolarchiver-cli config set-api-key YOUR_API_KEY
```

## Usage

### YouTube Tools

Requires paid API subscription.

#### Get User Comments

```bash
lolarchiver-cli youtube comments --user-id USER_ID --offset 0
# or
lolarchiver-cli youtube comments --handle HANDLE --offset 0
# or
lolarchiver-cli youtube comments --channel-id CHANNEL_ID --offset 0
```

#### Get Comment Replies

```bash
lolarchiver-cli youtube replies --comment-id COMMENT_ID
```

### Twitter Tools

#### Get User History

```bash
lolarchiver-cli twitter --handle HANDLE
# or
lolarchiver-cli twitter --id USER_ID
# or
lolarchiver-cli twitter --handle HANDLE --by-old
```

### Twitch Tools

#### Get User Messages

```bash
lolarchiver-cli twitch messages --username USERNAME --server superserver2 --offset 0
```

#### Get User Timeouts

```bash
lolarchiver-cli twitch timeouts --username USERNAME --offset 0
```

#### Get User History

```bash
lolarchiver-cli twitch history --username USERNAME --mode username
# or
lolarchiver-cli twitch history --username USERNAME --mode utype
# or
lolarchiver-cli twitch history --username USERNAME --mode btype
```

#### Get User Followage

```bash
lolarchiver-cli twitch followage --username USERNAME
```

#### Get User Followers

```bash
lolarchiver-cli twitch followers --username USERNAME
```

### Kick Tools

#### Get User Messages

```bash
lolarchiver-cli kick messages --username USERNAME --offset 0
```

#### Get User Timeouts

```bash
lolarchiver-cli kick timeouts --username USERNAME
```

#### Get User Mod Channels

```bash
lolarchiver-cli kick mods --username USERNAME
```

#### Get User Subscribers

```bash
lolarchiver-cli kick subscribers --username USERNAME
```

### Reverse Lookup Tools

#### Phone Lookup

```bash
lolarchiver-cli reverse phone --phone PHONE_NUMBER
# or
lolarchiver-cli reverse phone --phone PHONE_NUMBER --insecure
```

#### Email Lookup

```bash
lolarchiver-cli reverse email --email EMAIL_ADDRESS
# or
lolarchiver-cli reverse email --email EMAIL_ADDRESS --insecure
```

### Database Search

```bash
lolarchiver-cli database --query SEARCH_QUERY
# or
lolarchiver-cli database --query SEARCH_QUERY --exact
```

## License

MIT 