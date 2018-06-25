# mmpost
Like pbcopy/pbpaste for [Mattermost](https://mattermost.com/)


## Config 
A config file will be created on first run.
$HOME/.config/mmpost/config.json
```json
{
        "server": "",
        "pat": "",
        "team": "",
        "channel": "",
        "maxlines": 50,
        "lang": "",
        "filename": ""
}
```

### Get a Personal Access Token on mattermost server.

Account Setings -> Security -> Personal Access Tokens -> Create New Token.


## Usage examples

Less than maxlines with a lang
```bash
cat file.go | mmpost -lang go
```
More than maxlines (will be an attachment)
```bash
cat file.go | mmpost -lang go --filename file.go
```
See config flags
```bash
mmpost -h
```
File over 16300 bytes required to be an attachment so you have to use --filename
```bash
cat file.go | mmpost -lang go --filename file.go
```
