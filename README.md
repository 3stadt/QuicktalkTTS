# QuickTalkTTS

QuickTalkTTS is a text to speech program with a very narrow use case, using AWS Polly.

It was created to be used with TeamSpeak in case of having a sore throat and thus being unable to speak.

**This is currently only working for german text.**

## Prerequisites

- Amazon AWS Account
- [AWS Polly](https://aws.amazon.com/polly/) set up, you need an access key and a secret
- go 1.19
- cgo must be set up, see the [fyne.io docs](https://developer.fyne.io/started/#prerequisites)

## Installation

Clone or download this repository, the install all packages:

```bash
go mod tidy
```

Copy `.aws_credentials.dist` to `.aws_credentials` and insert your AWS polly key.

Then either use `go run .` or compile it into an executable [using fyne](https://developer.fyne.io/started/packaging).

## Usage

The interface should be self explanatory. To pipe the sound into TeamSpeak or discord, use a "virtual audio cable".

## License

[MIT](https://choosealicense.com/licenses/mit/)