# Package Manager Publishing Setup Guide

This guide explains how to set up automated publishing to Homebrew, Scoop, and Winget package managers.

## Prerequisites

- Admin access to the Devolvio-B-V GitHub organization
- Ability to create repositories and secrets

## 1. Homebrew Setup (macOS/Linux)

### One-Time Setup

1. **Create the Homebrew tap repository**
   - Go to https://github.com/organizations/Devolvio-B-V/repositories/new
   - Repository name: `homebrew-tap`
   - Description: "Homebrew formulae for Devolvio B.V. tools"
   - Visibility: Public
   - Click "Create repository"
   - **Note**: Leave the repository empty - GoReleaser will populate it automatically

2. **Generate a Personal Access Token (PAT)**
   - Go to https://github.com/settings/tokens/new
   - Token name: `GoReleaser Homebrew Tap`
   - Expiration: No expiration (or choose your preference)
   - Select scopes:
     - ✅ `repo` (Full control of private repositories)
   - Click "Generate token"
   - **Important**: Copy the token immediately (you won't see it again)

3. **Add the token as a repository secret**
   - Go to https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/settings/secrets/actions
   - Click "New repository secret"
   - Name: `HOMEBREW_TAP_GITHUB_TOKEN`
   - Value: Paste the token from step 2
   - Click "Add secret"

### How It Works

When you create a release tag (e.g., `v2.0.0`):
1. GitHub Actions triggers the release workflow
2. GoReleaser builds the binaries
3. GoReleaser automatically creates/updates the formula in `homebrew-tap`
4. Users can install with: `brew install devolvio-b-v/tap/cf-dlp-decode`

## 2. Scoop Setup (Windows)

### One-Time Setup

1. **Create the Scoop bucket repository**
   - Go to https://github.com/organizations/Devolvio-B-V/repositories/new
   - Repository name: `scoop-bucket`
   - Description: "Scoop bucket for Devolvio B.V. tools"
   - Visibility: Public
   - Click "Create repository"
   - **Note**: Leave the repository empty - GoReleaser will populate it automatically

2. **Generate a Personal Access Token (PAT)**
   - Follow the same steps as Homebrew above, or reuse the same token
   - Token name: `GoReleaser Scoop Bucket` (if creating a new one)
   - Scopes: `repo`

3. **Add the token as a repository secret**
   - Go to https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/settings/secrets/actions
   - Click "New repository secret"
   - Name: `SCOOP_BUCKET_GITHUB_TOKEN`
   - Value: Paste the token
   - Click "Add secret"

### How It Works

When you create a release tag:
1. GoReleaser automatically creates/updates the manifest in `scoop-bucket`
2. Users can install with:
   ```powershell
   scoop bucket add devolvio-b-v https://github.com/Devolvio-B-V/scoop-bucket
   scoop install cf-dlp-decode
   ```

## 3. Winget Setup (Windows)

Winget requires manual or semi-automated submission to the official Microsoft repository.

### Option A: Semi-Automated with wingetcreate (Recommended)

1. **Install wingetcreate** (one-time)
   ```powershell
   winget install wingetcreate
   ```

2. **Fork the winget-pkgs repository** (one-time)
   - Go to https://github.com/microsoft/winget-pkgs
   - Click "Fork" in the top-right corner

3. **For each release**, run:
   ```powershell
   # Replace VERSION with your release version (e.g., 2.0.0)
   $VERSION = "2.0.0"
   
   # Generate manifest and submit PR
   wingetcreate update Devolvio-B-V.cf-dlp-decode `
     --version $VERSION `
     --urls "https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/releases/download/v$VERSION/cf-dlp-decode-windows-amd64.exe" `
     --submit
   ```

4. **Wait for Microsoft review**
   - Microsoft team reviews the PR (typically 1-3 business days)
   - They may request changes or automatically merge
   - Once merged, users can install with: `winget install Devolvio-B-V.cf-dlp-decode`

### Option B: Fully Manual Submission

1. **Create manifest files**
   - Clone your fork of https://github.com/microsoft/winget-pkgs
   - Create manifests in `manifests/d/Devolvio-B-V/cf-dlp-decode/<version>/`
   - Required files:
     - `Devolvio-B-V.cf-dlp-decode.installer.yaml`
     - `Devolvio-B-V.cf-dlp-decode.locale.en-US.yaml`
     - `Devolvio-B-V.cf-dlp-decode.yaml`
   - See https://github.com/microsoft/winget-pkgs/blob/master/AUTHORING_MANIFESTS.md

2. **Submit Pull Request**
   - Commit and push to your fork
   - Open a PR to microsoft/winget-pkgs
   - Wait for review and merge

### Option C: Automated Workflow (Future Enhancement)

We can add a GitHub Actions workflow to automatically create the wingetcreate PR. Let me know if you'd like this added.

## Verification

After setting up:

1. **Create a test release**
   ```bash
   git tag v2.0.0-test
   git push origin v2.0.0-test
   ```

2. **Check GitHub Actions**
   - Go to https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/actions
   - Verify the "Release" workflow runs successfully

3. **Verify package manager repositories**
   - Homebrew: Check https://github.com/Devolvio-B-V/homebrew-tap for the new formula
   - Scoop: Check https://github.com/Devolvio-B-V/scoop-bucket for the new manifest

## Troubleshooting

### Homebrew/Scoop: "Token authentication failed"
- Verify the PAT has the correct permissions (`repo` scope)
- Ensure the token hasn't expired
- Check the secret name matches exactly: `HOMEBREW_TAP_GITHUB_TOKEN` or `SCOOP_BUCKET_GITHUB_TOKEN`

### Homebrew/Scoop: Repository not found
- Ensure the repository exists and is public
- Verify the repository name matches exactly: `homebrew-tap` or `scoop-bucket`

### Winget: PR rejected
- Ensure all required manifest files are present
- Verify URLs are correct and accessible
- Check that version numbers match across all files

## Additional Resources

- [GoReleaser Documentation](https://goreleaser.com/)
- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Scoop Bucket Documentation](https://github.com/ScoopInstaller/Scoop/wiki/Buckets)
- [Winget Package Submission](https://github.com/microsoft/winget-pkgs/blob/master/AUTHORING_MANIFESTS.md)
- [wingetcreate Documentation](https://github.com/microsoft/winget-create)
