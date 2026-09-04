# Contributing to Play Console CLI

Thanks for your interest in contributing!

## Adding Your App to the Wall of Apps

If you use Play Console CLI to ship your Android app, add it to our Wall of Apps!

### Steps:

1. Fork this repository
2. Edit `docs/wall-of-apps.json`
3. Add your app in this format:

```json
{
  "app": "Your App Name",
  "link": "https://play.google.com/store/apps/details?id=your.package.name",
  "creator": "YourGitHubUsername",
  "icon": "https://optional-icon-url.png",
  "platform": ["Android"]
}
```

4. Submit a Pull Request

## Bug Reports

Open an issue with:
- Description of the problem
- Steps to reproduce
- Expected behavior
- Actual behavior
- `gpc version` output

## Feature Requests

Open an issue describing:
- The feature you'd like
- Why it would be useful
- Example usage

## Code Contributions

1. Fork the repo
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Run tests: `make test`
5. Commit: `git commit -m "Add my feature"`
6. Push: `git push origin feature/my-feature`
7. Open a Pull Request

## Development Setup

```bash
git clone https://github.com/AndroidPoet/playconsole-cli.git
cd playconsole-cli
make build
./playconsole-cli version
```

On Windows, use PowerShell instead of `make`:

```powershell
git clone https://github.com/AndroidPoet/playconsole-cli.git
Set-Location playconsole-cli
go build -o bin/playconsole-cli.exe ./cmd/playconsole-cli
go test ./...
.\bin\playconsole-cli.exe version
```

## Code Style

- Run `go fmt` before committing
- Follow existing patterns in the codebase
- Add tests for new features
