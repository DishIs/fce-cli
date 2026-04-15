# fce — FreeCustom.Email CLI

```
______             _____           _                    _____                _ _ 
|  ___|           /  __ \         | |                  |  ___|              (_) |
| |_ _ __ ___  ___| /  \/_   _ ___| |_ ___  _ __ ___   | |__ _ __ ___   __ _ _| |
|  _| '__/ _ \/ _ \ |   | | | / __| __/ _ \| '_ ` _ \  |  __| '_ ` _ \ / _` | | |
| | | | |  __/  __/ \__/\ |_| \__ \ || (_) | | | | | |_| |__| | | | | | (_| | | |
\_| |_|  \___|\___|\____/\__,_|___/\__\___/|_| |_| |_(_)____/_| |_| |_|\__,_|_|_|
                                                                                 
                                                                                 
  FreeCustom.Email
  disposable inbox API
```

Manage disposable inboxes, extract OTPs, and stream real-time email events from your terminal — in under 30 seconds.

---

## Install

```bash
curl -fsSL freecustom.email/install.sh | sh
```

*(Or use your preferred package manager below)*

**macOS/Linux (Homebrew)**
```bash
brew tap DishIs/homebrew-tap
brew install fce
```

**Windows (Scoop)**
```powershell
scoop bucket add fce https://github.com/DishIs/scoop-bucket
scoop install fce
```

**Windows (Chocolatey)**
```powershell
choco install fce
```

**Shell Script (macOS/Linux)**
```bash
curl -sSfL https://raw.githubusercontent.com/DishIs/fce-cli/main/scripts/install.sh | sh
```

**Go install**
```bash
go install github.com/DishIs/fce-cli@latest
```

---

## Update

When a new version is released, you can update the CLI using your package manager:

**Homebrew**
```bash
brew update
brew upgrade fce
```

**Scoop**
```powershell
scoop update fce
```

**Chocolatey**
```powershell
choco upgrade fce
```

**NPM**
```bash
npm install -g fcemail@latest
```

**Shell Script**
Simply re-run the installation command or use the built-in update:
```bash
fce update
```

---

## Uninstall

To remove the CLI and all local configuration:

1. **Clear Config & Credentials**
   ```bash
   fce uninstall
   ```
   *(This clears your API key and local cache)*

2. **Remove the Binary**
   - **Homebrew**: `brew uninstall fce`
   - **Scoop**: `scoop uninstall fce`
   - **Choco**: `choco uninstall fce`
    - **NPM**: `npm uninstall -g fcemail`
    - **Manual**: `sudo rm /usr/local/bin/fce`

Or download a binary from [Releases](https://github.com/DishIs/fce-cli/releases).

---

## Quick start

```bash
# 1. Login — opens your browser
fce login

# 2. Watch a random inbox for emails in real time
fce watch random

# 3. Or watch a specific one
fce watch mytest@ditmail.info
```

---

## Commands

| Command | Description | Plan required |
|---------|-------------|---------------|
| `fce login` | Authenticate via browser | Any |
| `fce logout` | Remove stored credentials | Any |
| `fce status` | Account info, plan, inbox counts | Any |
| `fce usage` | Request usage for current period | Any |
| `fce inbox list` | List registered inboxes | Any |
| `fce inbox add <addr>` | Register a new inbox | Any |
| `fce inbox add random` | Register a random inbox | Any |
| `fce inbox remove <addr>` | Unregister an inbox | Any |
| `fce messages <inbox> [id]` | List messages or view a specific message | Any |
| `fce domains` | List available domains | Any |
| `fce watch [inbox\|random]` | Stream emails via WebSocket | **Startup+** |
| `fce otp <inbox>` | Get latest OTP from an inbox | **Growth+** |
| `fce dev` | Instantly register a dev inbox and start watching | Any |
| `fce update` | Update the CLI to the latest version | Any |
| `fce uninstall` | Remove all local config and credentials | Any |
| `fce version` | Show version info | Any |


---

## Automation, CI/CD & AI Agents

`fce-cli` provides native support for scripting, automation, CI/CD pipelines, and agentic AI workflows. You can strictly format the output to `json` or `csv` and suppress all terminal UI components using global flags.

**Global Flags:**
* `--format`, `-f` : Set output format to `text` (default), `json`, or `csv`.
* `--limit`, `-l` : Limit the number of results returned (0 for all).
* `--silent`, `-s` : Suppress non-essential output (automatically enabled for json/csv).

```bash
# Get your account status in JSON format
fce status --format json

# List the last 5 emails in CSV format
fce messages test@ditube.info --format csv --limit 5

# Extract OTP silently in a CI/CD pipeline
OTP_JSON=$(fce otp test@ditube.info --format json)
```

### Examples

```bash
# Register + watch a random inbox
fce inbox add random
fce watch random

# Watch a specific inbox (Startup plan+)
fce watch alerts@ditmail.info

# Get the latest OTP (Growth plan+)
fce otp mytest@ditmail.info

# Check quota
fce usage

# List all your inboxes
fce inbox ls
```

---

## Authentication

`fce login` opens your browser to `www.freecustom.email`. Sign in with GitHub, Google, or a magic link — a new API key is created and stored securely in your OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service / libsecret).

You can also set the `FCE_API_KEY` environment variable to skip the keychain entirely — useful in CI:

```bash
export FCE_API_KEY=fce_your_key_here
fce status
```

---

## Plan limits

| Feature | Free | Developer | Startup | Growth | Enterprise |
|---------|------|-----------|---------|--------|------------|
| All basic commands | ✓ | ✓ | ✓ | ✓ | ✓ |
| `fce watch` (WebSocket) | ✗ | ✗ | ✓ | ✓ | ✓ |
| `fce otp` | ✗ | ✗ | ✗ | ✓ | ✓ |

Upgrade at: https://www.freecustom.email/api/pricing

---

## Build from source

```bash
git clone https://github.com/DishIs/fce-cli
cd fce
go build -o fce .
./fce --help
```

**Cross-platform release build** (requires [goreleaser](https://goreleaser.com)):
```bash
goreleaser build --clean --snapshot
# Binaries in dist/
```

---

## CI usage

```yaml
# GitHub Actions example
- name: Get OTP
  env:
    FCE_API_KEY: ${{ secrets.FCE_API_KEY }}
  run: |
    fce inbox add random > /tmp/inbox.txt
    INBOX=$(cat /tmp/inbox.txt | grep -o '[a-z0-9@.]*')
    # trigger your app to send email to $INBOX
    OTP=$(fce otp $INBOX)
    echo "OTP: $OTP"
```

---

## License

MIT © FreeCustom.Email

### Observability Commands

You can view the event timeline and debug insights for any inbox directly from the CLI.

```bash
# View the event timeline and latencies
fce timeline <email@domain.com>

# View delivery insights and failure flags (Requires Growth+)
fce insights <email@domain.com>
```
